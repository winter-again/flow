package tmux

import (
	"fmt"
	"os/user"
	"testing"
)

func TestNewServer(t *testing.T) {
	// TODO: prob use tmux.GetDefaultSocket() instead?
	defaultSocketName := "default"
	sockDir := getSocketDir()
	UID := getUID()
	defaultSocketPath := fmt.Sprintf("%s/tmux-%s/%s", sockDir, UID, defaultSocketName)

	expDefaultServer := Server{
		SocketName: defaultSocketName,
		SocketPath: defaultSocketPath,
	}
	gotDefaultServer := *NewServer(defaultSocketName, defaultSocketPath)
	if gotDefaultServer.SocketName != expDefaultServer.SocketName || gotDefaultServer.SocketPath != expDefaultServer.SocketPath {
		t.Errorf("Expected Server %q but got %q", expDefaultServer, gotDefaultServer)
	}

	expCustomNameServer := Server{
		SocketName: "custom_server_name",
		SocketPath: fmt.Sprintf("%s/tmux-%s/custom_server_name", sockDir, UID),
	}
	gotCustomNameServer := *NewServer("custom_server_name", defaultSocketPath)
	if gotCustomNameServer.SocketName != expCustomNameServer.SocketName || gotCustomNameServer.SocketPath != expCustomNameServer.SocketPath {
		t.Errorf("Expected Server %q but got %q", expCustomNameServer, gotCustomNameServer)
	}

	customSocketPath := "/tmp/tmux-1000/custom_dir/custom_socket"
	expCustomPathServer := Server{
		SocketName: "custom_socket",
		SocketPath: customSocketPath,
	}
	gotCustomPathServer := *NewServer(defaultSocketName, customSocketPath) // defaultSocketName should be ignored
	if gotCustomPathServer.SocketName != expCustomPathServer.SocketName || gotCustomPathServer.SocketPath != expCustomPathServer.SocketPath {
		t.Errorf("Expected Server %q but got %q", expCustomPathServer, gotCustomPathServer)
	}
}

func TestGetCurrentServer(t *testing.T) {}

func TestServerStart(t *testing.T) {}

// For trying stuff
func TestMain(t *testing.T) {
	t.Logf("User ID: %s", getUID())
	t.Log(user.Current())
}
