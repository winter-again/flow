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
	// todo: add socket flags?
}

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch tmux sessions using a popup",
	Long:  `Pick an existing tmux session to switch to or create a new one from common directories`,
	RunE:  switchSession,
}

// TODO: figure out the error handling here
// we want it to return nil (and thus exit code 0) if user cancels via esc or ctrl-c

func switchSession(cmd *cobra.Command, args []string) error {
	if !tmux.InsideTmux() {
		log.Fatal("Not running inside tmux")
	}

	// todo: this command should only work from within tmux and should use the current server only
	// thus I think safe to assume that there exists at least one session (the currently attached session)

	// todo: fix all err handling
	server, err := tmux.GetCurrentServer()
	if err != nil {
		log.Fatal(err)
	}

	sessions, err := server.GetSessions()
	if err != nil {
		// log.Fatal(err)
		return nil
	}

	session, err := selSession(sessions)
	if err != nil {
		// log.Fatal(err)
		return nil
	}

	if server.SessionExists(session.Name) {
		err := switchSess(session)
		if err != nil {
			// log.Fatal(err)
			return nil
		}
	} else {
		// NOTE: here session is from fd output and the parsed session name from the path
		// was used to check if session exists; guaranteed to have both Name and Path
		// TODO: is it confusing that use a *Session to create/return a new *Session?
		newSession, err := server.CreateSession(session.Name, session.Path)
		if err != nil {
			// TODO: need to handle?
			// log.Fatal(err)
			return nil
		}
		err = switchSess(newSession)
		if err != nil {
			// log.Println(err)
			return nil
		}
	}
	return nil
}

// use fzf-tmux to pick a session that either exists for switching
// or should be created
func selSession(sessions []*tmux.Session) (*tmux.Session, error) {
	// TODO: make some of these consts or vars outside of func?
	// TODO: should we do all the external program checks at once?
	// TODO: should we even rely on fd? note that fd is used in the reload fzf call, so may
	// not be able to call something from Go
	// could at least offer ability to pick find or fd to reduce deps
	_, err := exec.LookPath("fd")
	if err != nil {
		return &tmux.Session{}, errors.New("Couldn't find fd in the PATH")
	}

	fdDirs := viper.GetStringSlice("fd.dirs")
	fdArgs := viper.GetStringSlice("fd.args")
	fdDirsStr := strings.Join(fdDirs, " ")
	fdArgsStr := strings.Join(fdArgs, " ")
	fdCmd := fmt.Sprintf("fd . %s %s --type d", fdDirsStr, fdArgsStr)

	fzfTmuxWidth := viper.GetString("fzf-tmux.width")
	fzfTmuxLength := viper.GetString("fzf-tmux.length")
	fzfTmuxBorder := viper.GetString("fzf-tmux.border")

	fzfTmuxPrevCmd := viper.GetStringSlice("fzf-tmux.preview_dir_cmd")
	fzfTmuxPrevCmdStr := strings.Join(fzfTmuxPrevCmd, " ")
	fzfTmuxPrevPos := viper.GetString("fzf-tmux.preview_pos")
	fzfTmuxPrevSize := viper.GetString("fzf-tmux.preview_size")
	fzfTmuxPrevBorder := viper.GetString("fzf-tmux.preview_border")

	// TODO: make at least appearance configurable
	// can use append for some of that; otherwise user's fzf config is used?
	// I think it also deps on where called from; if called from zsh then
	// inherits that env, but if we call via run-shell, then it's sh env

	// TODO: do we need to be more careful with the commands being set to the keybinds?
	// i.e., keeping stuff in line with expected behavior of Go side of things
	args := []string{
		"--layout",
		"reverse",    // display from top; overrides user fzf config
		"--no-multi", // disable multi-select
		"-p",         // popup window size, req. tmux 3.2+
		fmt.Sprintf("%s,%s", fzfTmuxWidth, fzfTmuxLength),
		"--prompt",
		" Sessions: ",
		"--header",
		"\033[1;34m<tab>\033[m: common dirs / \033[1;34m<shift-tab>\033[m: sessions / \033[1;34m<ctrl-k>\033[m: kill session",
		"--preview",
		"active_pane_id=$(tmux display-message -t {} -p '#{pane_id}'); tmux capture-pane -ep -t $active_pane_id",
		"--bind",
		fmt.Sprintf("tab:reload(%s)+change-prompt( Common dirs: )+change-preview(%s {})+change-preview-label(Files)", fdCmd, fzfTmuxPrevCmdStr),
		"--bind",
		"shift-tab:reload(tmux list-sessions -F '#{session_name}')+change-prompt( Sessions)+change-preview(active_pane_id=$(tmux display-message -t {} -p '#{pane_id}'); tmux capture-pane -ep -t $active_pane_id)+change-preview-label(Currently active pane)",
		"--bind",
		"ctrl-k:execute(tmux kill-session -t {})+reload(tmux list-sessions -F '#{session_name}')",
		"--preview-label",
		"Currently active pane",
		"--preview-window",
		fmt.Sprintf("%s,%s,border-%s", fzfTmuxPrevPos, fzfTmuxPrevSize, fzfTmuxPrevBorder),
		"--border",
		fmt.Sprintf("%s", fzfTmuxBorder),
		"--no-separator",
		// TODO: if configurable, then have to figure out how to inherit user's fzf defaults
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

	// NOTE: fzf-tmux is wrapper script from fzf
	fzfTmux, err := exec.LookPath("fzf-tmux")
	if err != nil {
		return &tmux.Session{}, errors.New("Couldn't find fzf-tmux in the PATH")
	}
	fzfTmuxCmd := exec.Command(fzfTmux, args...)

	stdin, err := fzfTmuxCmd.StdinPipe()
	if err != nil {
		return &tmux.Session{}, errors.New("Problem with stdin pipe")
	}

	var sessionList []string
	for _, session := range sessions {
		sessionList = append(sessionList, session.Name)
	}
	sessionStr := strings.Join(sessionList, "\n")

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, sessionStr)
	}()

	// TODO: is there where we have to handle a ctrl-c or escape?
	out, err := fzfTmuxCmd.CombinedOutput()
	if err != nil {
		return &tmux.Session{}, errors.New("Problem running fzf-tmux command")
	}

	selection := strings.TrimSpace(string(out))
	if tmux.IsPath(selection) {
		return &tmux.Session{
			Name: filepath.Base(selection),
			Path: selection,
		}, nil
	}
	return &tmux.Session{
		Name: selection,
	}, nil
}

// TODO: this technically checks session ID before name
// prefix name with "=" to force an exact match

// call to tmux to switch session based on name
func switchSess(session *tmux.Session) error {
	args := []string{
		"switch-client",
		"-t",
		session.Name,
	}
	_, _, err := tmux.Cmd(args)
	if err != nil {
		return err
	}
	return nil
}
