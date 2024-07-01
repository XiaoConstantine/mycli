/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"mycli/pkg/commands/install"
	"mycli/pkg/iostreams"
	"mycli/pkg/utils"

	"os"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/spf13/cobra"
)

type exitCode int

const (
	exitOK      exitCode = 0
	exitError   exitCode = 1
	exitCancel  exitCode = 2
	exitAuth    exitCode = 4
	exitPending exitCode = 8
)

func NewRootCmd(iostream *iostreams.IOStreams) (*cobra.Command, error) {
	cs := iostream.ColorScheme()

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

	installCmd := install.NewInstallCmd(iostream)

	rootCmd.AddCommand(installCmd)
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

func Run() exitCode {
	iostream := iostreams.System()
	stderr := iostream.ErrOut
	ctx := context.Background()
	rootCmd, err := NewRootCmd(iostream)

	if err != nil {
		fmt.Fprintf(stderr, "failed to create root command: %s\n", err)
		return exitError
	}
	if command, err := rootCmd.ExecuteContextC(ctx); err != nil {
		var pagerPipeError *iostreams.ErrClosedPagerPipe
		var noResultsError utils.NoResultsError

		if err == utils.SilentError {
			return exitError
		} else if err == utils.PendingError {
			return exitPending
		} else if utils.IsUserCancellation(err) {
			if errors.Is(err, terminal.InterruptErr) {
				// ensure the next shell prompt will start on its own line
				fmt.Fprint(stderr, "\n")
			}
			return exitCancel
		} else if errors.As(err, &pagerPipeError) {
			// ignore the error raised when piping to a closed pager
			return exitOK
		} else if errors.As(err, &noResultsError) {
			if iostream.IsStdoutTTY() {
				fmt.Fprintln(stderr, noResultsError.Error())
			}
			// no results is not a command failure
			return exitOK
		}
		fmt.Fprintf(stderr, "Error %v", command)

	}
	return exitOK
}
