package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
	"github.com/winter-again/flow/internal/tmux"
)

func Attach() *cli.Command {
	var target string

	socketName, socketPath := tmux.GetDefaultSocket()

	return &cli.Command{
		Name:    "attach",
		Aliases: []string{"a"},
		Usage:   "Attach to existing tmux server and session",
		MutuallyExclusiveFlags: []cli.MutuallyExclusiveFlags{
			{
				Flags: [][]cli.Flag{
					{
						&cli.StringFlag{
							Name:        "name",
							Aliases:     []string{"n"},
							Value:       socketName,
							Usage:       "tmux server socket name",
							Destination: &socketName,
						},
					},
					{
						&cli.StringFlag{
							Name:        "path",
							Aliases:     []string{"p"},
							Value:       socketPath,
							Usage:       "tmux server socket path",
							Destination: &socketPath,
						},
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "target",
				Aliases:     []string{"t"},
				Usage:       "Target tmux session. Defaults to most recently used unattached session.",
				Destination: &target,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			server := tmux.NewServer(socketName, socketPath)

			_, _, err := server.Attach(target)
			if err != nil {
				cli.Exit(fmt.Errorf("error while attaching to server with socket name '%s' and socket path '%s': %w", server.SocketName, server.SocketPath, err), 1)
			}
			return nil
		},
	}
}
