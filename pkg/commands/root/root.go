/*
Copyright © 2024 Xiao Cui <constantine124@gmail.com>
*/
package root

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/XiaoConstantine/mycli/pkg/build"
	"github.com/XiaoConstantine/mycli/pkg/commands/extensions"
	"github.com/XiaoConstantine/mycli/pkg/commands/install"
	"github.com/XiaoConstantine/mycli/pkg/commands/update"
	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/XiaoConstantine/mycli/pkg/utils"

	"github.com/XiaoConstantine/mycli/pkg/commands/configure"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/spf13/cobra"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type NoOpLogger struct{}

func (l NoOpLogger) Log(msg string) {}

type ExitCode int

const (
	exitOK      ExitCode = 0
	exitError   ExitCode = 1
	exitCancel  ExitCode = 2
	exitAuth    ExitCode = 4
	exitPending ExitCode = 8
)

var nonInteractive bool

// NewRootCmd creates and returns the root command for mycli.
// It sets up all subcommands and flags, and defines the main run logic.
func NewRootCmd(iostream *iostreams.IOStreams) (*cobra.Command, error) {
	cs := iostream.ColorScheme()

	rootCmd := &cobra.Command{
		Use:           cs.GreenBold("mycli"),
		Short:         "Bootstrap my machine",
		Long:          `Internal CLI help bootstrap my machine.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       fmt.Sprintf("%s (%s) - built on %s", build.Version, build.Commit, build.Date),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if the first argument is an extension
			if len(args) == 0 {
				return nil
			}
			extName := args[0]
			extDir := extensions.GetExtensionsDir()
			extPath := filepath.Join(extDir, extensions.ExtensionPrefix+extName)

			if _, err := os.Stat(extPath); err == nil {
				ext := &extensions.Extension{Name: extName, Path: extPath}
				return ext.Execute(args[1:])
			}

			return nil
		},
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

	rootCmd.AddGroup(&cobra.Group{
		ID:    "update",
		Title: "Update command",
	})

	rootCmd.AddGroup(&cobra.Group{
		ID:    "extension",
		Title: "Extension commands",
	})

	installCmd := install.NewInstallCmd(iostream)
	configureCmd := configure.NewConfigureCmd(iostream)
	updateCmd := update.NewUpdateCmd(iostream)

	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(updateCmd)

	// Add extension management command
	extCmd := extensions.NewCmdExtension(iostream)
	if extCmd != nil {
		rootCmd.AddCommand(extCmd)
	} else {
		return nil, fmt.Errorf("failed to create extension command")
	}
	rootCmd.PersistentFlags().Bool("help", false, "Show help for command")
	rootCmd.PersistentFlags().BoolVar(&nonInteractive, "non-interactive", false, "Run in non-interactive mode")

	return rootCmd, nil
}

// Run executes the main logic of mycli. It handles argument parsing,
// command execution, and returns an appropriate exit code.
//
// args: Command-line arguments passed to mycli.
// Returns: An ExitCode indicating the result of the command execution.
func Run(args []string) ExitCode {
	iostream := iostreams.System()
	stderr := iostream.ErrOut
	ctx := context.Background()

	rootCmd, err := NewRootCmd(iostream)
	utils.PrintWelcomeMessage(iostream, rootCmd.Version)

	if err != nil {
		fmt.Fprintf(iostream.ErrOut, "Failed to create root command")
		return exitError
	}

	// Check for updates
	hasUpdate, latestVersion, err := update.CheckForUpdatesFunc(iostream)
	if err != nil {
		fmt.Fprintf(stderr, "Failed to check for updates: %s\n", err)
	} else if hasUpdate {
		fmt.Fprintf(iostream.Out, "A new version of mycli is available: %s (current: %s)\n", latestVersion, build.Version)
		fmt.Fprintf(iostream.Out, "Run 'mycli update' to update\n\n")
	}

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

	rootCmd.SetArgs(args)
	// Hack: Check for non-interactive mode
	if len(args) > 0 && args[0] == "--non-interactive" {
		nonInteractive = true
	}
	// If args are provided or --help flag is set, execute the command directly
	if len(args) > 0 {
		if _, err := rootCmd.ExecuteContextC(ctx); err != nil {
			handleExecutionError(err, iostream)
			return exitError
		}
		return exitOK
	}
	return runInteractiveMode(rootCmd, ctx, iostream, options)
}

func runInteractiveMode(rootCmd *cobra.Command, ctx context.Context, iostream *iostreams.IOStreams, options []string) ExitCode {
	var selectedOption string
	prompt := &survey.Select{
		Message: "Choose a command to run:",
		Options: options,
	}
	if err := survey.AskOne(prompt, &selectedOption); err != nil {
		fmt.Fprintf(iostream.ErrOut, "failed to select a command: %s\n", err)
		return exitError
	}

	// Confirm if user wants to run the install command
	var confirm bool
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("Do you want to run the '%s' command?", selectedOption),
	}
	if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
		fmt.Fprintf(iostream.ErrOut, "failed to confirm the command: %s\n", err)
		return exitError
	}

	if !confirm {
		fmt.Println("Operation cancelled by user.")
		return exitOK
	}
	rootCmd.SetArgs([]string{selectedOption})

	if _, err := rootCmd.ExecuteContextC(ctx); err != nil {
		handleExecutionError(err, iostream)
		return exitError
	}
	return exitOK
}

func handleExecutionError(err error, iostream *iostreams.IOStreams) {
	var pagerPipeError *iostreams.ErrClosedPagerPipe
	var noResultsError utils.NoResultsError

	if err == utils.SilentError {
		return
	} else if err == utils.PendingError {
		return
	} else if utils.IsUserCancellation(err) {
		if errors.Is(err, terminal.InterruptErr) {
			// ensure the next shell prompt will start on its own line
			fmt.Fprint(iostream.ErrOut, "\n")
		}
		return
	} else if errors.As(err, &pagerPipeError) {
		// ignore the error raised when piping to a closed pager
		return
	} else if errors.As(err, &noResultsError) {
		if iostream.IsStdoutTTY() {
			fmt.Fprintln(iostream.ErrOut, noResultsError.Error())
		}
		// no results is not a command failure
		return
	} else if err == utils.ConfigNotFoundError {
		fmt.Fprintln(iostream.ErrOut, iostream.ColorScheme().Red("Config file is invalid for tools install, skipping..."))
		return
	}

	fmt.Fprintf(iostream.ErrOut, "Error: %v\n", err)
}
