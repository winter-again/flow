package tmux

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// check if $TMUX environment var is set, meaning running inside tmux
func InsideTmux() bool {
	if os.Getenv("TMUX") == "" {
		return false
	}
	return true
}

type Session struct {
	Id      string // unique session ID
	Name    string // name of session
	Path    string // working directory of session
	Windows int    // number of windows in session
}

// check if session exists based on its name
func (session *Session) Exists() bool {
	// TODO: should we do this edge-case here?
	if session.Name == "" {
		return false
	}

	// NOTE: `has-session` will either report error and exit with 1 or exit with 0
	args := []string{
		"has-session",
		"-t",
		session.Name,
	}
	_, _, err := Cmd(args)
	if err != nil {
		return false
	}
	return true
}

// check if a session with given name exists
func SessionExists(sessionName string) bool {
	// TODO: should we do this edge-case here?
	if sessionName == "" {
		return false
	}

	// NOTE: `has-session` will either report error and exit with 1 or exit with 0
	args := []string{
		"has-session",
		"-t",
		sessionName,
	}
	_, _, err := Cmd(args)
	if err != nil {
		return false
	}
	return true
}

// TODO: is there any redundancy with Exists()?
// this also seems to be addressing when session doesn't exist
// maybe this should actually use the Exists() to check? Seems better than doing
// this iter?
// Get tmux session by name
func GetSession(sessionName string) (*Session, error) {
	// TODO: can we guarantee that names are unique?
	sessions, err := GetSessions()
	if err != nil {
		return &Session{}, errors.New("Couldn't get session")
	}
	for _, session := range sessions {
		if session.Name == sessionName {
			return session, nil
		}
	}
	return &Session{}, fmt.Errorf("Session %q doesn't exist", sessionName)
}

// Get all tmux sessions
func GetSessions() ([]*Session, error) {
	// TODO: have to cover non-default sockets?
	// TODO: return err too?
	args := []string{
		"list-sessions",
		"-F",
		"#{session_id};#{session_name};#{session_path};#{session_windows}",
	}
	sessions, _, err := Cmd(args)
	if err != nil {
		return []*Session{}, errors.New("Couldn't get sessions")
	}
	return parseSessions(sessions), nil
}

// Run tmux command with given args; return stdout and stderr
func Cmd(args []string) (string, string, error) {
	tmux, err := exec.LookPath("tmux")
	if err != nil {
		return "", "", errors.New("Couldn't find tmux in the PATH")
	}

	cmd := exec.Command(tmux, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	return outStr, errStr, err
}

// parse string of session info
func parseSessions(sessionStr string) []*Session {
	// TODO: do we need to consider any errors when splitting? shouldn't it always
	// return at least one session?
	sessionsSplit := strings.Split(strings.TrimSpace(sessionStr), "\n")

	sessionList := make([]*Session, 0, len(sessionsSplit))
	for _, s := range sessionsSplit {
		// TODO: should ";" be global const?
		// TODO: remove "$" from session ID?
		fields := strings.Split(s, ";")
		session := &Session{
			Id:      fields[0],
			Name:    fields[1],
			Path:    fields[2],
			Windows: stringToInt(fields[3]),
		}
		sessionList = append(sessionList, session)
	}
	return sessionList
}

// TODO: could this ever fail or receive some bad value?
func stringToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err)
	}
	return i
}

func IsPath(session string) bool {
	match, err := regexp.MatchString("^.*/(.*/)*$", session)
	if err != nil {
		log.Fatal(err)
	}
	return match
}

// create and return new tmux session with given name and path
func CreateSession(sessionName string, sessionPath string) (*Session, error) {
	if sessionName == "" || strings.Contains(sessionName, ".") || strings.Contains(sessionName, ":") {
		return &Session{}, fmt.Errorf("Session names can't be empty and can't contain colons or periods: %s", sessionName)
	}

	args := []string{
		"new-session",
		"-d",
		"-s",
		sessionName,
		"-c",
		sessionPath,
	}
	_, _, err := Cmd(args)
	if err != nil {
		return &Session{}, err
	}

	// TODO: can we guarantee session names are unique so that we can guarantee
	// retrieval by name?
	session, err := GetSession(sessionName)
	if err != nil {
		return &Session{}, err
	}
	return session, nil
}
