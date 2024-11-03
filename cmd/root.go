package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// base command when called without any subcommands
var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "flow",
		Short: "CLI for managing tmux sessions",
		Long:  `A simple CLI wrapper for managing tmux sessions. Switch between sessions or create them on the fly from specified directories via a popup window`,
		// use if bare app has action associated
		// Run: func(cmd *cobra.Command, args []string) {},
	}
)

// TODO: how do these affect the control flow?
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	// TODO: check how this is triggered and what cfgFile represents
	// runtime-loaded config?
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// home, err := os.UserHomeDir()
		// cobra.CheckErr(err)

		viper.SetConfigName("flow")
		viper.SetConfigType("toml")

		// TODO: is there a better way of specifying this?
		// can we use $XDG_CONFIG_HOME?
		// viper.AddConfigPath(fmt.Sprintf("%s/.config/flow/", home))
		viper.AddConfigPath("$HOME/.config/flow")

		viper.SetDefault("fd.args", []string{"--min-depth", "1", "--max-depth", "1"})

		viper.SetDefault("flow.init_session_name", "0")

		viper.SetDefault("fzf-tmux.width", "80%")
		viper.SetDefault("fzf-tmux.length", "60%")
		viper.SetDefault("fzf-tmux.border", "rounded")
		viper.SetDefault("fzf-tmux.preview_dir_cmd", []string{"ls"})
		viper.SetDefault("fzf-tmux.preview_pos", "right")
		viper.SetDefault("fzf-tmux.preview_size", "60%")
		viper.SetDefault("fzf-tmux.preview_border", "rounded")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			home, _ := os.UserHomeDir()
			log.Fatalf("Config file not found at %s", fmt.Sprintf("%s/.config/flow/flow.toml", home))
		} else {
			log.Fatal("Error reading config file %w", err)
		}
	}
}
