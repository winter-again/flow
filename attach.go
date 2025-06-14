package main

import (
	"context"
	"fmt"
	"log"

	"github.com/urfave/cli/v3"
	"github.com/winter-again/flow/internal/tmux"
)

func Attach() *cli.Command {
	var socketName string
	var socketPath string
	var target string

	// TODO: should this be run here? what about in init()?
	defaultSocketName, defaultSocketPath := tmux.GetDefaultSocket()

	return &cli.Command{
		Name:    "attach",
		Aliases: []string{"a"},
		Usage:   "Attach to existing tmux server",
		MutuallyExclusiveFlags: []cli.MutuallyExclusiveFlags{
			{
				Flags: [][]cli.Flag{
					{
						&cli.StringFlag{
							Name:        "name",
							Aliases:     []string{"n"},
							Value:       defaultSocketName,
							Usage:       "tmux server socket name",
							Destination: &socketName,
						},
					},
					{
						&cli.StringFlag{
							Name:        "path",
							Aliases:     []string{"p"},
							Value:       defaultSocketPath,
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
			// log.Printf("CLI socketName: %s\n", socketName)
			// log.Printf("CLI socketPath: %s\n", socketPath)

			server := tmux.NewServer(socketName, socketPath)
			// log.Printf("targeting server: %q\n", server)

			_, _, err := server.Attach(target)
			if err != nil {
				log.Fatal(fmt.Errorf("error while attaching to server with socket name '%s' and socket path '%s': %w", server.SocketName, server.SocketPath, err))
			}
			return nil
		},
	}
}
