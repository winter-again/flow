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
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

const tmuxFormatSep string = ";"

// InsideTmux checks if $TMUX environment var is set, meaning running inside tmux
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

// NewServer creates Server spec based on socket name or path or just the default.
func NewServer(socketName string, socketPath string) *Server {
	defaultSocketName := "default"
	socketDir := getSocketDir()
	UID := getUID()
	defaultSocketPath := fmt.Sprintf("%s/tmux-%s/%s", socketDir, UID, defaultSocketName)

	// NOTE: if socket path is not the default, then socket name should be ignored
	// b/c socket path already specifies the name; this matches tmux behavior
	// Otherwise, if socket name is given, then the socket path must use that name
	// but other parts of the path remain default
	if socketPath != defaultSocketPath {
		return &Server{
			SocketName: filepath.Base(socketPath), // set name to whatever was given for socketPath
			SocketPath: socketPath,
		}
	} else if socketName != defaultSocketName {
		return &Server{
			SocketName: socketName,
			SocketPath: fmt.Sprintf("%s/tmux-%s/%s", socketDir, UID, socketName),
		}
	} else {
		return &Server{
			SocketName: defaultSocketName,
			SocketPath: defaultSocketPath,
		}
	}
}

// GetCurrentServer retrieves the server for the current session
func GetCurrentServer() (*Server, error) {
	args := []string{
		"list-clients",
		"-F",
		"#{socket_path}",
	}
	serverInfo, _, err := Cmd(args)
	if err != nil {
		return &Server{}, errors.New("Problem fetching server")
	}

	socketPath := strings.TrimSpace(serverInfo)
	return &Server{
		SocketName: filepath.Base(socketPath),
		SocketPath: socketPath,
	}, nil
}

// Start starts a new tmux server with a single session using either the
// socket name or the socket path
func (server *Server) Start() (string, string, error) {
	if InsideTmux() {
		return "", "", errors.New("Shouldn't nest tmux sessions")
	}

	// NOTE: assumes that server is running if the socket exists,
	// though it's possible to just delete the socket while server runs
	if _, err := os.Stat(server.SocketPath); err == nil {
		return "", "", errors.New("Server already exists")
	}

	_, defaultSocketPath := GetDefaultSocket()

	var args []string
	// NOTE: creates server with single default session
	// refrain from using `start-server` command
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

// Attach attaches to session in the given server.
// If no target session is given, tmux will pref most recently used unattached session
func (server *Server) Attach(sessionName string) (string, string, error) {
	if InsideTmux() {
		return "", "", errors.New("Shouldn't nest tmux sessions")
	}

	// NOTE: attach-session will try to create server, but this will fail
	// if no sessions specified in the tmux config file
	args := []string{
		"-S",
		server.SocketPath,
		"attach-session",
	}

	if sessionName != "" {
		if !server.SessionExists(sessionName) {
			return "", "", errors.New("Session doesn't exist")
		}
		args = append(args, "-t", sessionName)
	}

	stdout, stderr, err := Cmd(args)
	if err != nil {
		return stdout, stderr, err
	}
	return stdout, stderr, nil
}

// GetDefaultSocket returns default tmux socket name and path
func GetDefaultSocket() (string, string) {
	defaultSocketName := "default"
	sockDir := getSocketDir()
	UID := getUID()
	return defaultSocketName, fmt.Sprintf("%s/tmux-%s/%s", sockDir, UID, defaultSocketName)
}

// getSocketDir retrieves the tmux server socket directory, first checking
// if TMUX_TMPDIR environment var set. Otherwise, returns the
// default socket directory
func getSocketDir() string {
	if sockDir := os.Getenv("TMUX_TMPDIR"); sockDir != "" {
		return sockDir
	}
	return "/tmp"
}

// getUID retrieves the current UID
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

// GetSession retrieves a tmux session by name
func (server *Server) GetSession(sessionName string) (*Session, error) {
	if !server.SessionExists(sessionName) {
		return &Session{}, fmt.Errorf("Session %s doesn't exist", sessionName)
	}

	sessions, err := server.GetSessions()
	if err != nil {
		return &Session{}, err
	}

	for _, session := range sessions {
		if session.Name == sessionName {
			return session, nil
		}
	}
	return &Session{}, fmt.Errorf("Session %s doesn't exist", sessionName)
}

// GetSessions retrieves all tmux sessions
func (server *Server) GetSessions() ([]*Session, error) {
	format := []string{
		"#{session_id}",
		"#{session_name}",
		"#{session_path}",
		"#{session_windows}",
	}
	args := []string{
		"-S",
		server.SocketPath,
		"list-sessions",
		"-F",
		strings.Join(format, tmuxFormatSep),
	}
	sessions, _, err := Cmd(args)
	if err != nil {
		return []*Session{}, fmt.Errorf("Couldn't retrieve sessions: %w", err)
	}

	parsedSessions, err := parseSessions(sessions)
	if err != nil {
		return []*Session{}, fmt.Errorf("Couldn't retrieve sessions: %w", err)
	}
	return parsedSessions, nil
}

// parseSessions parses returned tmux session data into Session struct
func parseSessions(sessionsOutput string) ([]*Session, error) {
	sessions := strings.Split(strings.TrimSpace(sessionsOutput), "\n")

	sessionsParsed := make([]*Session, len(sessions))
	for i, s := range sessions {
		fields := strings.Split(s, tmuxFormatSep)
		nWins, err := strconv.Atoi(fields[3])
		if err != nil {
			return []*Session{}, errors.New("Error parsing number of windows per session")
		}
		session := &Session{
			Id:      fields[0],
			Name:    fields[1],
			Path:    fields[2],
			Windows: nWins,
		}
		sessionsParsed[i] = session
	}
	return sessionsParsed, nil
}

// SessionExists checks if session exists based on its name
func (server *Server) SessionExists(sessionName string) bool {
	if sessionName == "" {
		return false
	}

	// NOTE: `has-session` will either report error and exit with 1 or exit with 0
	args := []string{
		"-S",
		server.SocketPath,
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

// CreateSession creates a tmux session based on name and working directory
func (server *Server) CreateSession(sessionName string, sessionPath string) (*Session, error) {
	if sessionName == "" || strings.Contains(sessionName, ":") {
		return &Session{}, fmt.Errorf("Session names can't be empty and can't contain colons: %s", sessionName)
	}

	if strings.Contains(sessionName, ".") {
		sessionName = strings.ReplaceAll(sessionName, ".", "_")
	}

	args := []string{
		"-S",
		server.SocketPath,
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

	session, err := server.GetSession(sessionName)
	if err != nil {
		return &Session{}, err
	}
	return session, nil
}

// IsValidPath checks if a given session name is actually a valid path
func IsValidPath(session string) bool {
	_, err := os.Stat(session)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

// Cmd runs a tmux command with given args; returns stdout and stderr
func Cmd(args []string) (string, string, error) {
	tmux, err := exec.LookPath("tmux")
	if err != nil {
		return "", "", errors.New("Couldn't find tmux in the PATH")
	}

	cmd := exec.Command(tmux, args...)

	var stdout, stderr bytes.Buffer
	// NOTE: setting stdin makes it so that creating and attach to server works
	// but does it make sense?
	cmd.Stdin = os.Stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	return outStr, errStr, err
}
