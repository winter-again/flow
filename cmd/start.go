package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/winter-again/flow/internal/tmux"
)

func init() {
	rootCmd.AddCommand(startCmd)

	defaultSocketName, defaultSocketPath := tmux.GetDefaultSocket()
	startCmd.Flags().StringP("name", "n", defaultSocketName, "tmux server socket name")
	startCmd.Flags().StringP("path", "p", defaultSocketPath, "tmux server socket path")
	startCmd.MarkFlagsMutuallyExclusive("name", "path") // NOTE: only accept one of socket name or socket path
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start tmux server",
	Long:  `Start a new tmux server with either a given socket name or path. Attach to its default session.`,
	Run:   startAttachServer,
}

// startAttachServer starts a tmux server and attaches to a default session
func startAttachServer(cmd *cobra.Command, args []string) {
	// TODO: if mutually exclusive, what does empty one return? ""?
	socketName, _ := cmd.Flags().GetString("name")
	socketPath, _ := cmd.Flags().GetString("path")

	log.Printf("CLI socketName: %s\n", socketName)
	log.Printf("CLI socketPath: %s\n", socketPath)

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
}
