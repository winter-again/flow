package tmux

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

type Server struct {
	SocketName string // socket name
	SocketPath string // socket path
}

// create Server struct based on socket name or path or just the default
// guarantees that both socket name and path are set
func NewServer(socketName string, socketPath string) *Server {
	// NOTE: if socket path is given, then socket name should be ignored
	// b/c socket path already specifies the name
	tmpDir := getTmuxTmpDir()
	defaultSocketName := "default"
	defaultSocketPath := fmt.Sprintf("%s/tmux-1000/default", tmpDir)

	if socketPath != defaultSocketPath {
		return &Server{
			SocketName: filepath.Base(socketPath),
			SocketPath: socketPath,
		}
	} else if socketName != defaultSocketName {
		return &Server{
			SocketName: socketName,
			SocketPath: fmt.Sprintf("/tmp/tmux-1000/%s", socketName),
		}
	} else {
		return &Server{
			SocketName: defaultSocketName,
			SocketPath: defaultSocketPath,
		}
	}
}

// create server + default session by either socket name or socket path
func (server *Server) Create() (string, string, error) {
	if InsideTmux() {
		log.Fatal("Shouldn't nest tmux sessions")
	}

	tmpDir := getTmuxTmpDir()
	defaultSocketPath := fmt.Sprintf("%s/tmux-1000/default", tmpDir)

	log.Printf("default socket path: %s", defaultSocketPath)
	log.Printf("given socket path: %s", server.SocketPath)

	var args []string
	// NOTE: this creates server with single default session;
	// can use "server" to create bare server
	if server.SocketPath != defaultSocketPath {
		args = []string{
			"-S",
			server.SocketPath,
			"new-session",
			"-d",
		}
	} else {
		args = []string{
			"-L",
			server.SocketName,
			"new-session",
			"-d",
		}
	}

	stdout, stderr, err := Cmd(args)
	if err != nil {
		return stdout, stderr, err
	}
	return stdout, stderr, nil
}

// attach to server designated by its socket name
// allows tmux to figure out which session
func (server *Server) Attach() (string, string, error) {
	if InsideTmux() {
		log.Fatal("Shouldn't nest tmux sessions")
	}

	// NOTE: attach-session will try to create server, but this will fail
	// if no sessions specified in the config file
	args := []string{
		"-L",
		server.SocketName,
		"attach-session",
	}
	stdout, stderr, err := Cmd(args)
	if err != nil {
		return stdout, stderr, err
	}
	return stdout, stderr, nil
}

// retrieve the TMUXTMPDIR environment var
func getTmuxTmpDir() string {
	if tmpDir := os.Getenv("TMUXTMPDIR"); tmpDir != "" {
		return tmpDir
	}
	return "/tmp"
}

type Session struct {
	Id      string // unique session ID
	Name    string // name of session
	Path    string // working directory of session
	Windows int    // number of windows in session
}

// TODO: some of these should really be methods on server?

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

// TODO: should this be a Server method?
// I think session IDs and names are only unique to server

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

// TODO: should this be a Server method?
// I think session IDs and names are only unique to server

// Get all tmux sessions
func GetSessions() ([]*Session, error) {
	// TODO: have to cover non-default sockets?
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
	// NOTE: setting stdin makes it so that creating and attach to server works
	// but does it make sense
	cmd.Stdin = os.Stdin
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
