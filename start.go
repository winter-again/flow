package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
	"github.com/winter-again/flow/internal/tmux"
)

func Start() *cli.Command {
	socketName, socketPath := tmux.GetDefaultSocket()

	return &cli.Command{
		Name:    "start",
		Aliases: []string{"s"},
		Usage:   "Start and attach to new tmux server for flow to manage",
		// TODO: this seems to check exclusivity, but is the behavior correct? It prints the
		// help and a warning under
		// It could be a bug closed by this PR: https://github.com/urfave/cli/issues/2146
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
		Action: func(ctx context.Context, cmd *cli.Command) error {
			server := tmux.NewServer(socketName, socketPath)

			_, _, err := server.Start()
			if err != nil {
				return cli.Exit(fmt.Errorf("error while starting server with socket name '%s' and socket path '%s': %w", server.SocketName, server.SocketPath, err), 1)
			}
			_, _, err = server.Attach("")
			if err != nil {
				return cli.Exit(fmt.Errorf("error while attaching to server with socket name '%s' and socket path '%s': %w", server.SocketName, server.SocketPath, err), 1)
			}
			return nil
		},
	}
}
