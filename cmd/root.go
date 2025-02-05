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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/flow/config.toml)")

	viper.SetDefault("flow.init_session_name", "0")

	viper.SetDefault("fzf-tmux.width", "80%")
	viper.SetDefault("fzf-tmux.length", "60%")
	viper.SetDefault("fzf-tmux.border", "rounded")
	viper.SetDefault("fzf-tmux.preview_size", "60%")
	viper.SetDefault("fzf-tmux.preview_border", "rounded")
	viper.SetDefault("fzf-tmux.preview_dir_cmd", []string{"ls"})
	viper.SetDefault("fzf-tmux.preview_pos", "right")

	viper.SetDefault("find.dirs", []string{"$HOME"})
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		viper.AddConfigPath(fmt.Sprintf("%s/.config/flow", home))
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			home, err := os.UserHomeDir()
			if err != nil {
				log.Println(err)
				os.Exit(1)
			}
			fmt.Printf("Config file not found at %s", fmt.Sprintf("%s/.config/flow/config.toml", home))
			os.Exit(1)
		} else {
			fmt.Printf("error reading config file %s", err)
			os.Exit(1)
		}
	}
}
