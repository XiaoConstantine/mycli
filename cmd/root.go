/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"mycli/pkg/commands/install"
	"mycli/pkg/iostreams"

	"os"

	"github.com/spf13/cobra"
)

func NewRootCmd(iostream *iostreams.IOStreams) (*cobra.Command, error) {
	cs := iostream.ColorScheme()
	ctx := context.Background()

	rootCmd := &cobra.Command{
		Use:           cs.GreenBold("mycli"),
		Short:         "Bootstrap my machine",
		Long:          `Internal CLI help bootstrap my machine.`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	rootCmd.AddGroup(
		&cobra.Group{
			ID:    "install",
			Title: "Install commands",
		})
	rootCmd.AddCommand(install.NewCmdXcode(iostream))
	rootCmd.AddCommand(install.NewCmdHomeBrew(iostream))

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
	if _, err := rootCmd.ExecuteContextC(ctx); err != nil {
		fmt.Fprintf(iostream.ErrOut, "Failed to execute root command: %v\n", err)
	}
	return rootCmd, nil
}
