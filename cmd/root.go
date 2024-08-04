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
		Short: "Short desc here",
		Long:  `Long desc here`,
		// use if bare app has action associated
		// Run: func(cmd *cobra.Command, args []string) {},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.SetConfigName("flow")
		viper.SetConfigType("toml")
		viper.AddConfigPath(fmt.Sprintf("%s/.config/flow/", home))
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	fmt.Printf("Got fd dirs: %s\n", viper.GetStringSlice("fd.dirs"))
}
