package main

import (
	"context"
	"fmt"
	"log"

	"github.com/urfave/cli/v3"
	"github.com/winter-again/flow/internal/tmux"
)

func Start() *cli.Command {
	var socketName string
	var socketPath string

	// TODO: should this be run here? what about in init()?
	defaultSocketName, defaultSocketPath := tmux.GetDefaultSocket()

	return &cli.Command{
		Name:    "start",
		Aliases: []string{"s"},
		Usage:   "Start tmux server for flow",
		// TODO: this seems to check exclusivity, but is the behavior correct? It prints the
		// help and a warning under
		// it could be a bug closed by this PR: https://github.com/urfave/cli/issues/2146
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
		Action: func(ctx context.Context, cmd *cli.Command) error {
			server := tmux.NewServer(socketName, socketPath)
			log.Printf("targeting server: %+v\n", server)

			_, _, err := server.Start()
			if err != nil {
				log.Fatal(fmt.Errorf("error while starting server with socket name '%s' and socket path '%s': %w", server.SocketName, server.SocketPath, err))
			}
			_, _, err = server.Attach("")
			if err != nil {
				log.Fatal(fmt.Errorf("error while attaching to server with socket name '%s' and socket path '%s': %w", server.SocketName, server.SocketPath, err))
			}
			return nil
		},
	}
}
