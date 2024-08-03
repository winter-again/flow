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
	"github.com/winter-again/flow/internal/tmux"
)

func init() {
	// register command and its flags, opts, etc.
	rootCmd.AddCommand(switchCmd)
}

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch tmux sessions",
	Long:  `Pick an existing tmux session to switch to or create a new one from common directories`,
	Run:   switchSession,
}

// switch tmux session based on selection
func switchSession(cmd *cobra.Command, args []string) {
	// TODO: should we be able to run commands from outside tmux?
	// TODO: should this go inside every command?
	if !tmux.InsideTmux() {
		log.Fatal("Not running inside tmux")
	}

	sessions, err := tmux.GetSessions()
	if err != nil {
		log.Fatal(err)
	}

	session, err := selSession(sessions)
	if err != nil {
		log.Fatal(err)
	}

	if session.Exists() {
		err := switchSess(session)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// NOTE: here session is from fd output and the parsed session name from the path
		// was used to check if session exists; guaranteed to have both Name and Path
		// TODO: is it confusing that use a *Session to create/return a new *Session?
		newSession, err := tmux.CreateSession(session.Name, session.Path)
		// log.Printf("created session with name '%s' and path '%s'", session.Name, session.Path)
		if err != nil {
			log.Fatal(err)
		}
		err = switchSess(newSession)
		if err != nil {
			log.Fatal(err)
		}
	}
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
	// TODO: should make the paths that get fed into fd configurable
	// what about the fd command in general?
	fdDirs := []string{
		"~/Documents/Bansal-lab",
		"~/Documents/code",
	}
	fdDirsStr := strings.Join(fdDirs, " ")
	fdCmd := fmt.Sprintf("fd . %s --min-depth 1 --max-depth 1 --type d", fdDirsStr)
	// TODO: make at least appearance configurable
	// can use append for some of that; otherwise user's fzf config is used?
	// TODO: do we need to be more careful with the commands being set to the keybinds?
	// i.e., keeping stuff in line with expected behavior of Go side of things
	args := []string{
		"--layout",
		"reverse",
		"--no-multi",
		"-p",
		"80%,60%",
		"--prompt",
		" Sessions: ",
		"--header",
		"\033[1;34m<tab>\033[m: common dirs / \033[1;34m<shift-tab>\033[m: sessions / \033[1;34m<ctrl-k>\033[m: kill session",
		"--bind",
		fmt.Sprintf("tab:change-preview-window(hidden)+change-prompt( Common dirs: )+reload(%s)", fdCmd),
		"--bind",
		"shift-tab:preview(~/.local/bin/tmux-switcher-preview.sh {})+change-prompt( Sessions)+reload(tmux list-sessions -F '#{session_name}')",
		"--bind",
		"ctrl-k:execute(tmux kill-session -t {})+reload(tmux list-sessions -F '#{session_name}')",
		// TODO: should --preview call an actual command of this proj instead?
		"--preview",
		"active_pane_id=$(tmux display-message -t {} -p '#{pane_id}') && tmux capture-pane -ep -t $active_pane_id",
		"--preview-window",
		"right:65%",
		"--preview-window",
		"border-left",
		"--preview-label",
		"Currently active pane",
		"--border",
		"rounded",
		"--no-separator",
		// --color=fg:#cacaca,bg:-1,hl:underline:#8a98ac \
		// --color=fg+:#f0f0f0,bg+:#262626,hl+:underline:#8f8aac \
		// --color=info:#c6a679,prompt:#8f8aac,pointer:#f0f0f0 \
		// --color=marker:#8aac8b,spinner:#8aac8b \
		// --color=gutter:-1,border:#8a98ac,header:-1 \
		// --color=preview-fg:#f0f0f0,preview-bg:-1
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
