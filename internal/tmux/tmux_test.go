package tmux

import (
	"os/user"
	"testing"
)

// For trying stuff
func TestMain(t *testing.T) {
	t.Logf("User ID: %s", getUID())
	t.Log(user.Current())
}
