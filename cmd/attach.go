package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/winter-again/flow/internal/tmux"
)

func init() {
	rootCmd.AddCommand(attachCmd)

	defaultSocketName, defaultSocketPath := tmux.GetDefaultSocket()
	attachCmd.Flags().StringP("name", "n", defaultSocketName, "tmux server socket name")
	attachCmd.Flags().StringP("path", "p", defaultSocketPath, "tmux server socket path")
	attachCmd.MarkFlagsMutuallyExclusive("name", "path") // NOTE: only accept one of socket name or socket path
	attachCmd.Flags().StringP("target", "t", "", "Target tmux session. Defaults to most recently used unattached session.")
}

var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach to tmux server",
	Long:  `Attach to an existing tmux server. Will prefer the most recently used session.`,
	Run:   attachServer,
}

// attachServer attaches the current client to a given tmux session
func attachServer(cmd *cobra.Command, args []string) {
	socketName, _ := cmd.Flags().GetString("name")
	socketPath, _ := cmd.Flags().GetString("path")
	target, _ := cmd.Flags().GetString("target")

	log.Printf("CLI socketName: %s\n", socketName)
	log.Printf("CLI socketPath: %s\n", socketPath)

	server := tmux.NewServer(socketName, socketPath)
	log.Printf("targeting server: %q\n", server)

	_, _, err := server.Attach(target)
	if err != nil {
		log.Fatal(fmt.Errorf("Error while attaching to server with socket name '%s' and socket path '%s': %w", server.SocketName, server.SocketPath, err))
	}
}
