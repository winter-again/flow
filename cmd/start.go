package cmd

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/winter-again/flow/internal/tmux"
)

func init() {
	rootCmd.AddCommand(startCmd)
	// TODO: make these default to ""?
	startCmd.Flags().StringP("name", "n", "default", "tmux server socket name")
	startCmd.Flags().StringP("path", "p", "/tmp/tmux-1000/default", "tmux server socket path")
	startCmd.MarkFlagsMutuallyExclusive("name", "path")
}

// NOTE: only accept one of socket name or socket path
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Create or attach to a tmux server",
	Long:  `Either create a tmux server and attach to a default session or attach to a session in an existing server. If no flags passed, will use the default server.`,
	Run:   startServer,
}

func startServer(cmd *cobra.Command, args []string) {
	socketName, _ := cmd.Flags().GetString("name")
	socketPath, _ := cmd.Flags().GetString("path")

	server := tmux.NewServer(socketName, socketPath)
	log.Printf("new server struct: %q", server)

	// NOTE: treat a server without sessions as equivalent to no server at all
	// rely on stderr to tell us
	_, stderr, err := server.Attach()
	if err != nil {
		if strings.TrimSpace(stderr) == "no sessions" {
			_, _, err := server.Create()
			if err != nil {
				log.Printf("Problem creating server & session -> %s : %s", server.SocketName, server.SocketPath)
				log.Fatal(err)
			}

			_, _, err = server.Attach()
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
