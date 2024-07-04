/*
Copyright © 2024 Xiao Cui <constantine124@gmail.com>
*/
package root

import (
	"context"
	"errors"
	"fmt"
	"mycli/pkg/commands/configure"
	"mycli/pkg/commands/install"
	"mycli/pkg/iostreams"
	"mycli/pkg/utils"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/spf13/cobra"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type NoOpLogger struct{}

func (l NoOpLogger) Log(msg string) {}

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

	rootCmd.AddGroup(
		&cobra.Group{
			ID:    "configure",
			Title: "Configure commands",
		})

	installCmd := install.NewInstallCmd(iostream)
	configureCmd := configure.NewConfigureCmd(iostream)

	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(configureCmd)
	rootCmd.PersistentFlags().Bool("help", false, "Show help for command")
	return rootCmd, nil
}

func Run() exitCode {
	iostream := iostreams.System()
	stderr := iostream.ErrOut
	ctx := context.Background()

	utils.PrintWelcomeMessage(iostream)
	rootCmd, err := NewRootCmd(iostream)

	os_info := utils.GetOsInfo()
	// todo: make optional for tracing
	tracer.Start(
		tracer.WithService("mycli"),
		tracer.WithEnv("development"),
		tracer.WithServiceVersion("1.0.0"),
		tracer.WithLogStartup(false),
		tracer.WithDebugMode(false),
		tracer.WithLogger(NoOpLogger{}),
		tracer.WithAgentAddr("localhost:8126"),
		tracer.WithGlobalTag("system", os_info),
	)
	defer tracer.Stop()

	span, ctx := tracer.StartSpanFromContext(ctx, "root")
	defer span.Finish()

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
	if err := survey.AskOne(prompt, &selectedOption); err != nil {
		fmt.Fprintf(stderr, "failed to select a command: %s\n", err)
		return exitError
	}

	// Confirm if user wants to run the install command
	var confirm bool
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("Do you want to run the '%s' command?", selectedOption),
	}
	if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
		fmt.Fprintf(stderr, "failed to confirm the command: %s\n", err)
		return exitError
	}

	if !confirm {
		fmt.Println("Operation cancelled by user.")
		return exitOK
	}
	rootCmd.SetArgs([]string{selectedOption})
	if err != nil {
		fmt.Fprintf(stderr, "failed to create root command: %s\n", err)
		return exitError
	}
	if _, err := rootCmd.ExecuteContextC(ctx); err != nil {
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
		} else if err == utils.ConfigNotFoundError {
			fmt.Fprintln(stderr, iostream.ColorScheme().Red("Config file is invalid for tools install, skipping..."))
			return exitOK
		}
	}
	return exitOK
}
