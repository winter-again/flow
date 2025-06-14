package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/urfave/cli/v3"
)

var k = koanf.New(".")

func main() {
	k.Load(confmap.Provider(map[string]any{
		"flow.init_session_name":   "0",
		"fzf-tmux.length":          "60%",
		"fzf-tmux.width":           "80%",
		"fzf-tmux.border":          "rounded",
		"fzf-tmux.preview_size":    "60%",
		"fzf-tmux.preview_border":  "rounded",
		"fzf-tmux.preview_dir_cmd": []string{"ls"},
		"fzf-tmux.preview_pos":     "right",
		"find.dirs":                []string{"$HOME"},
	}, "."), nil)

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	config := fmt.Sprintf("%s/.config/flow/config.toml", home)

	if _, err := os.Stat(config); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Config file not found at %s", fmt.Sprintf("%s/.config/flow/config.toml", home))
		os.Exit(1)
	}

	if err := k.Load(file.Provider(config), toml.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	cmd := &cli.Command{
		Name:  "flow",
		Usage: "CLI for managing tmux sessions",
		Commands: []*cli.Command{
			Start(),
			Attach(),
			Switch(),
			Find(),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
