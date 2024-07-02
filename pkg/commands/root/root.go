/*
Copyright © 2024 Xiao Cui <constantine124@gmail.com>
*/
package root

import (
	"context"
	"errors"
	"fmt"
	"mycli/pkg/commands/install"
	"mycli/pkg/iostreams"
	"mycli/pkg/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/spf13/cobra"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
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

	utils.PrintWelcomeMessage(iostream)
	rootCmd, err := NewRootCmd(iostream)

	tracer.Start(
		tracer.WithService("mycli"),
		tracer.WithEnv("development"),
		tracer.WithServiceVersion("1.0.0"),
		tracer.WithLogStartup(false),
		tracer.WithDebugMode(false),
		tracer.WithAgentAddr("localhost:8126"),
	)
	defer tracer.Stop()

	// Get all available commands
	var options []string
	for _, cmd := range rootCmd.Commands() {
		options = append(options, cmd.Use)
	}

	// Prompt user to select a command
	var selectedOption string
	prompt := &survey.Select{
		Message: "Choose a command to run:",
		Options: options,
	}
	survey.AskOne(prompt, &selectedOption)

	// Confirm if user wants to run the install command
	var confirm bool
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("Do you want to run the '%s' command?", selectedOption),
	}
	survey.AskOne(confirmPrompt, &confirm)

	if !confirm {
		fmt.Println("Operation cancelled by user.")
		return exitOK
	}

	var configPath string
	if selectedOption == "install" {
		// Prompt for config file path
		configPrompt := &survey.Input{
			Message: "Enter the path to the config file:",
			Default: "config.yaml",
		}
		survey.AskOne(configPrompt, &configPath)
		configPath = os.ExpandEnv(configPath)
		// Replace ~ with home directory
		if strings.HasPrefix(configPath, "~") {
			home, err := os.UserHomeDir()
			if err == nil {
				configPath = filepath.Join(home, configPath[1:])
			}
		}

		// Get the absolute path
		absPath, err := filepath.Abs(configPath)
		if err == nil {
			configPath = absPath
		}
		fmt.Println(configPath)
		// Validate the file path
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			fmt.Fprintf(stderr, "Error: Config file does not exist at path: %s\n", configPath)
			return exitError
		}
	}

	// Set the args for the root command
	if selectedOption == "install" {
		rootCmd.SetArgs([]string{selectedOption, "--config", configPath})
		var selectedCmd *cobra.Command
		for _, cmd := range rootCmd.Commands() {
			if cmd.Use == selectedOption {
				selectedCmd = cmd
				break
			}
			selectedCmd.SetArgs([]string{selectedOption, "--config", configPath})
		}
	} else {
		rootCmd.SetArgs([]string{selectedOption})
	}

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
