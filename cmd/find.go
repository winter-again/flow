package cmd

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(findCmd)
}

var findCmd = &cobra.Command{
	Use:   "find",
	Short: "List candidate directories for creating new sessions",
	Long:  `List all direct child directories for the list of parent directories specified in config file. For use in creating new tmux sessions from a pre-specified set of common working directories.`,
	Run:   find,
}

// find collects all top-level subdirectories of the provided target directories
// and prints the sorted list to stdout; it's solely for use in fzf-tmux popup window
func find(cmd *cobra.Command, args []string) {
	findDirs := viper.GetStringSlice("find.dirs")

	// TODO: should there be more validation of find.dirs data?
	// e.g., ignore duplicates, handle empty slice?

	var childDirs []string
	for _, parent := range findDirs {
		if strings.HasPrefix(parent, "~/") {
			user, err := user.Current()
			if err != nil {
				log.Fatal(err)
			}
			parent = filepath.Join(user.HomeDir, parent[2:])
		}

		file, err := os.Open(parent)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		dirs, err := file.Readdirnames(0)
		if err != nil {
			log.Fatal(err)
		}

		path, err := filepath.Abs(parent)
		if err != nil {
			log.Fatal(err)
		}

		for _, dir := range dirs {
			childDirs = append(childDirs, filepath.Join(path, dir))
		}
	}

	slices.Sort(childDirs)
	out := strings.Join(childDirs, "\n")
	fmt.Println(out)
}
