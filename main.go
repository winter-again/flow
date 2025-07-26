package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/urfave/cli/v3"

	"github.com/winter-again/flow/internal/tmux"
)

var k = koanf.New(".")

func main() {
	// TODO: create global --debug flag for logging?
	if err := loadConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// NOTE: is this any better than rereading the config file in that package?
	tmux.InitSessionName = k.String("flow.init_session_name")

	cmd := &cli.Command{
		Name:    "flow",
		Version: "v0.1.1",
		Usage:   "CLI for managing tmux sessions",
		Commands: []*cli.Command{
			Start(),
			Attach(),
			Switch(),
			Find(),
			// TODO: Kill() should also handle clean up of server/socket files
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1) // NOTE: this might be redundant?
	}
}

func loadConfig() error {
	// TODO: should allow user to config this from fzf-tmux instead?
	k.Load(confmap.Provider(map[string]any{
		"flow.init_session_name":   "0",
		"fzf-tmux.length":          "60%",
		"fzf-tmux.width":           "80%",
		"fzf-tmux.border":          "rounded",
		"fzf-tmux.preview_size":    "60%",
		"fzf-tmux.preview_border":  "rounded",
		"fzf-tmux.preview_dir_cmd": []string{"ls"},
		"fzf-tmux.preview_pos":     "right",
		// TODO: check if $HOME can be used
		"find.dirs": []string{"$HOME"},
	}, "."), nil)

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// TODO: what about --config flag for custom loc?
	config := filepath.Join(home, ".config/flow/config.toml")
	if err := k.Load(file.Provider(config), toml.Parser()); err != nil {
		return fmt.Errorf("error loading config file: %w", err)
	}
	return nil
}
