package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/winter-again/flow/internal/tmux"
)

func init() {
	rootCmd.AddCommand(attachCmd)

	defaultSocketName, defaultSocketPath := tmux.GetDefaultSocket()
	attachCmd.Flags().StringP("name", "n", defaultSocketName, "tmux server socket name")
	attachCmd.Flags().StringP("path", "p", defaultSocketPath, "tmux server socket path")
	attachCmd.MarkFlagsMutuallyExclusive("name", "path")
}

var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach to tmux server",
	Long:  `Attach to an existing tmux server. Will prefer the most recently used session`,
	Run:   attachServer,
}

func attachServer(cmd *cobra.Command, args []string) {
	// todo: if mutually exclusive, what does empty one return? ""?
	socketName, _ := cmd.Flags().GetString("name")
	socketPath, _ := cmd.Flags().GetString("path")

	log.Printf("CLI socketName: %s\n", socketName)
	log.Printf("CLI socketPath: %s\n", socketPath)

	server := tmux.NewServer(socketName, socketPath)
	log.Printf("targeting server: %q\n", server)

	_, _, err := server.Attach()
	if err != nil {
		log.Printf("Problem attaching to server %s : %s\n", server.SocketName, server.SocketPath)
		log.Fatal(err)
	}
}
