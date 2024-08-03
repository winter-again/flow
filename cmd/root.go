package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "flow",
	Short: "Short desc here",
	Long:  `Long desc here`,
	// use if bare app has action associated
	// Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately
// This is called by main.main(). It only needs to happen once to the rootCmd
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// this is for defining flags and config settings
// can use viper here to apply settings
// can also ADD COMMANDS here
func init() {
	// cobra.OnInitialize(initConfig)
}

// can use this to read in config via viper
func initConfig() {
	// do stuff here too
}
