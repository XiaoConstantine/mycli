/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"mycli/iostreams"
	"os"

	"github.com/spf13/cobra"
)

func NewRootCmd(iostream *iostreams.IOStreams) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   "mycli",
		Short: "A brief description of your application",
		Long: `A longer description that spans multiple lines and likely contains
	examples and usage of using your application. For example:

	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(cmd, args)
		},
	}
	rootCmd.PersistentFlags().Bool("help", false, "Show help for command")
	if os.Getenv("GH_COBRA") == "" {
		rootCmd.SilenceErrors = true
		rootCmd.SilenceUsage = true

		// this --version flag is checked in rootHelpFunc
		// rootCmd.Flags().Bool("version", false, "Show gh version")

		// rootCmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		// 	rootHelpFunc(f, c, args)
		// })
		// rootCmd.SetUsageFunc(func(c *cobra.Command) error {
		// 	return rootUsageFunc(f.IOStreams.ErrOut, c)
		// })
		// rootCmd.SetFlagErrorFunc(rootFlagErrorFunc)
	}

	return rootCmd, nil
}
