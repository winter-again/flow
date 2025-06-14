package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"slices"
	"strings"

	"github.com/urfave/cli/v3"
)

func Find() *cli.Command {
	return &cli.Command{
		Name:  "find",
		Usage: "List candidate directories for roots of new tmux sessions",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			findDirs := K.Strings("find.dirs")

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
			return nil
		},
	}
}
