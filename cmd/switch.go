package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/winter-again/flow/internal/tmux"
)

func init() {
	rootCmd.AddCommand(switchCmd)
}

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch tmux sessions using a popup",
	Long:  `Pick an existing tmux session to switch to or create a new one from a list of common directories`,
	Run:   switchSession,
}

// switchSession switches the client to a chosen tmux session via an fzf-tmux popup
func switchSession(cmd *cobra.Command, args []string) {
	if !tmux.InsideTmux() {
		log.Fatal("Not running inside tmux")
	}

	server, err := tmux.GetCurrentServer()
	if err != nil {
		log.Fatal(err)
	}

	sessions, err := server.GetSessions()
	if err != nil {
		log.Fatal(err)
	}

	session, err := selectSession(sessions)
	if err != nil {
		if err == errFzfTmux {
			return
		}
		log.Fatal(err)
	}

	if server.SessionExists(session.Name) {
		if err := switchSess(session); err != nil {
			log.Fatal(err)
		}
	} else {
		newSession, err := server.CreateSession(session.Name, session.Path)
		if err != nil {
			log.Fatal(err)
		}

		err = switchSess(newSession)
		if err != nil {
			log.Fatal(err)
		}
	}
}

var errFzfTmux = errors.New("exited fzf-tmux")

// selectSession handles the fzf-tmux window and session selection (and potentially creation)
func selectSession(sessions []*tmux.Session) (*tmux.Session, error) {
	fdDirs := viper.GetStringSlice("find.dirs")
	fdDirsStr := strings.Join(fdDirs, " ")
	fdArgs := viper.GetStringSlice("find.args")
	fdArgsStr := strings.Join(fdArgs, " ")
	_ = fmt.Sprintf("fd . %s %s --type d", fdDirsStr, fdArgsStr)

	// TODO: how do these interact with user's settings? inherit?
	fzfTmuxWidth := viper.GetString("fzf-tmux.width")
	fzfTmuxLength := viper.GetString("fzf-tmux.length")
	fzfTmuxBorder := viper.GetString("fzf-tmux.border")
	fzfTmuxPrevCmd := viper.GetStringSlice("fzf-tmux.preview_dir_cmd")
	fzfTmuxPrevCmdStr := strings.Join(fzfTmuxPrevCmd, " ")
	fzfTmuxPrevPos := viper.GetString("fzf-tmux.preview_pos")
	fzfTmuxPrevSize := viper.GetString("fzf-tmux.preview_size")
	fzfTmuxPrevBorder := viper.GetString("fzf-tmux.preview_border")

	fzfTmuxPrompt := "Sessions: "
	useIcons := viper.GetBool("fzf-tmux.use_icons")
	if useIcons {
		fzfTmuxPrompt = " Sessions: "
	}

	// HACK: instead of relying on fd, cmd `flow find` does equivalent
	// then call fzf-tmux with ref to ths cmd to populate
	// the secondary window
	findCmd := "./bin/flow find"

	args := []string{
		"--layout",
		"reverse",    // display from top; overrides user fzf config
		"--no-multi", // disable multi-select
		"-p",         // popup window size, req. tmux 3.2+
		fmt.Sprintf("%s,%s", fzfTmuxWidth, fzfTmuxLength),
		"--prompt",
		fzfTmuxPrompt,
		"--header",
		// NOTE: hard-coded options
		"\033[1;34m<tab>\033[m: common dirs / \033[1;34m<shift-tab>\033[m: sessions / \033[1;34m<ctrl-k>\033[m: kill session",
		"--preview",
		"active_pane_id=$(tmux display-message -t {} -p '#{pane_id}'); tmux capture-pane -ep -t $active_pane_id",
		"--bind",
		// fmt.Sprintf("tab:reload(%s)+change-prompt( Common dirs: )+change-preview(%s {})+change-preview-label(Files)", fdCmd, fzfTmuxPrevCmdStr),
		fmt.Sprintf("tab:reload(%s)+change-prompt(Common dirs: )+change-preview(%s {})+change-preview-label(Files)", findCmd, fzfTmuxPrevCmdStr),
		"--bind",
		fmt.Sprintf("shift-tab:reload(tmux list-sessions -F '#{session_name}')+change-prompt(%s)+change-preview(active_pane_id=$(tmux display-message -t {} -p '#{pane_id}'); tmux capture-pane -ep -t $active_pane_id)+change-preview-label(Currently active pane)", fzfTmuxPrompt),
		"--bind",
		"ctrl-k:execute(tmux kill-session -t {})+reload(tmux list-sessions -F '#{session_name}')",
		"--preview-label",
		"Currently active pane",
		"--preview-window",
		fmt.Sprintf("%s,%s,border-%s", fzfTmuxPrevPos, fzfTmuxPrevSize, fzfTmuxPrevBorder),
		"--border",
		fzfTmuxBorder,
		"--no-separator",
		// if not specified
		// "--color",
		// "fg:#cacaca,bg:-1,hl:underline:#8a98ac",
		// "--color",
		// "fg+:#f0f0f0,bg+:#262626,hl+:underline:#8f8aac",
		// "--color",
		// "info:#c6a679,prompt:#8f8aac,pointer:#f0f0f0",
		// "--color",
		// "marker:#8aac8b,spinner:#8aac8b",
		// "--color",
		// "gutter:-1,border:#8a98ac,header:-1",
		// "--color",
		// "preview-fg:#f0f0f0,preview-bg:-1",
	}

	sessionList := make([]string, len(sessions))
	for i, session := range sessions {
		sessionList[i] = session.Name
	}
	sessionStr := strings.Join(sessionList, "\n")

	fzfTmux, err := exec.LookPath("fzf-tmux") // NOTE: fzf-tmux is wrapper script from fzf
	if err != nil {
		return &tmux.Session{}, errors.New("couldn't find fzf-tmux in the PATH")
	}
	fzfTmuxCmd := exec.Command(fzfTmux, args...)

	stdin, err := fzfTmuxCmd.StdinPipe()
	if err != nil {
		return &tmux.Session{}, fmt.Errorf("error creating stdin pipe for fzf-tmux: %w", err)
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, sessionStr)
	}()

	out, err := fzfTmuxCmd.CombinedOutput()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// NOTE: 130 = ctrl-c or esc
			if exitError.ExitCode() == 130 {
				return &tmux.Session{}, errFzfTmux
			}
			log.Println("error isn't 130")
			return &tmux.Session{}, fmt.Errorf("error running fzf-tmux command: %w", err)
		}
		log.Println("not ExitError")
		return &tmux.Session{}, fmt.Errorf("error running fzf-tmux command: %w", err)
	}

	selection := strings.TrimSpace(string(out))
	if tmux.IsValidPath(selection) {
		return &tmux.Session{
			Name: filepath.Base(selection),
			Path: selection,
		}, nil
	}
	return &tmux.Session{
		Name: selection,
	}, nil
}

// switchSess switches client to the specified tmux session
func switchSess(session *tmux.Session) error {
	// NOTE: prepending "=" to session name enforces only exact matches
	sessionName := "=" + session.Name
	args := []string{
		"switch-client",
		"-t",
		sessionName,
	}
	_, _, err := tmux.Cmd(args)
	if err != nil {
		return fmt.Errorf("Error switching tmux sessions: %w", err)
	}
	return nil
}
