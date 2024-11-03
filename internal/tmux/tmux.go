package tmux

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// Checks if $TMUX environment var is set, meaning running inside tmux
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

// Creates Server spec based on socket name or path or just the default.
// This guarantees that both socket name and path are set
func NewServer(socketName string, socketPath string) *Server {
	defaultSocketName := "default"
	sockDir := getSocketDir()
	UID := getUID()
	defaultSocketPath := fmt.Sprintf("%s/tmux-%s/%s", sockDir, UID, defaultSocketName)

	// note: if socket path is not the default, then socket name should be ignored
	// b/c socket path already specifies the name; this matches tmux behavior
	// otherwise, if socket name is given, then the socket path must use that name
	// but other parts of the path remain default
	if socketPath != defaultSocketPath {
		return &Server{
			SocketName: filepath.Base(socketPath), // set name to whatever was given for socketPath
			SocketPath: socketPath,
		}
	} else if socketName != defaultSocketName {
		return &Server{
			SocketName: socketName,
			SocketPath: fmt.Sprintf("%s/tmux-%s/%s", sockDir, UID, socketName),
		}
	} else {
		// todo: think about this case; how does it get reached? caller would have
		// to explicitly specify default values for the args
		return &Server{
			SocketName: defaultSocketName,
			SocketPath: defaultSocketPath,
		}
	}
}

// Starts a new tmux server with a single session
// using either socket name or socket path
func (server *Server) Create() (string, string, error) {
	if InsideTmux() {
		// todo: case of creating a diff server?
		log.Fatal("Shouldn't nest tmux sessions")
	}

	_, defaultSocketPath := getDefaultSocket()

	log.Printf("default socket path: %s\n", defaultSocketPath)
	log.Printf("given socket path: %s\n", server.SocketPath)

	// todo: should this check if this specific server already exists and not create session
	// since we're really relying on `new-session` to do the server creation?
	// use `tmux info`? abuse `tmux run-shell`?

	// todo: consider need for -d (detached)
	// -d: new session is attached to the current terminal *unless* -d is given
	// I think this means immediate attachment of client to session instead of just
	// creating the server+session in the background
	var args []string
	// note: this creates server with single default session
	if server.SocketPath != defaultSocketPath {
		args = []string{
			"-S", // socket path
			server.SocketPath,
			"new-session",
			"-d", // detached
			"-s", // session name
			viper.GetString("flow.init_session_name"),
		}
	} else {
		args = []string{
			"-L", // socket name
			server.SocketName,
			"new-session",
			"-d", // detached
			"-s", // session name
			viper.GetString("flow.init_session_name"),
		}
	}

	stdout, stderr, err := Cmd(args)
	if err != nil {
		return stdout, stderr, err
	}
	return stdout, stderr, nil
}

// Attaches to session in the given server.
// Should prefer the most recently used unattached session
func (server *Server) Attach() (string, string, error) {
	if InsideTmux() {
		log.Fatal("Shouldn't nest tmux sessions")
	}

	// todo: should allow specific target session?

	// note: attach-session will try to create server, but this will fail
	// if no sessions specified in the tmux config file
	// note: not specifying target session will pref the most recently used
	// unattached session
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

func getDefaultSocket() (string, string) {
	defaultSocketName := "default"
	sockDir := getSocketDir()
	UID := getUID()
	return defaultSocketName, fmt.Sprintf("%s/tmux-%s/%s", sockDir, UID, defaultSocketName)
}

// Gets the tmux server socket directory, first checking
// if TMUX_TMPDIR environment var set. Otherwise, returns the
// default socket directory
func getSocketDir() string {
	if sockDir := os.Getenv("TMUX_TMPDIR"); sockDir != "" {
		return sockDir
	}
	return "/tmp"
}

// Gets the current UID
func getUID() string {
	currUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return currUser.Uid
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
// TODO: I think this always returns same order
// should we have ability to manipulate order or customize?

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

// Runs tmux command with given args; returns stdout and stderr
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
