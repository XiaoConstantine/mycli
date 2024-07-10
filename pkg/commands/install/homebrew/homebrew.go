package homebrew

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/XiaoConstantine/mycli/pkg/utils"

	"github.com/XiaoConstantine/mycli/pkg/iostreams"

	"github.com/spf13/cobra"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var execCommandContext = exec.CommandContext

// NewCmdHomeBrew creates a new cobra.Command that installs Homebrew on the system.
// It checks if the current user is an administrator, and if so, runs the Homebrew
// installation script using the current user's credentials. If the current user
// is not an administrator, it prints an error message and exits.
func NewCmdHomeBrew(iostream *iostreams.IOStreams, userUtils utils.UserUtils) *cobra.Command {
	cs := iostream.ColorScheme()

	cmd := &cobra.Command{
		Use:   "homebrew",
		Short: cs.GreenBold("Install homebrew, require admin privileges, make sure enable this via privileges app"),
		// Long:   actionsExplainer(cs),
		Hidden:        true,
		SilenceErrors: true,
		Annotations: map[string]string{
			"group": "install",
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			currentUser, _ := userUtils.GetCurrentUser()
			span, ctx := tracer.StartSpanFromContext(cmd.Context(), "install_homebrew")
			defer span.Finish()

			isAdmin := userUtils.IsAdmin(ctx, currentUser)
			if IsHomebrewInstalled(ctx) {
				fmt.Fprintln(iostream.Out, "Homebrew is installed.")
				span.SetTag("status", "success")
				span.Finish()
				return nil
			}

			if !isAdmin {
				fmt.Fprintln(iostream.ErrOut, cs.Red("You need to be an administrator to install Homebrew. Please run this command from an admin account."))
				span.SetTag("error", true)
				span.Finish(tracer.WithError(os.ErrPermission))
				return os.ErrPermission
			}

			fmt.Fprint(iostream.Out, cs.Green("Installing homebrew with su current user, enter your password when prompt\n"))
			installCmd := execCommandContext(ctx, "su", currentUser.Username, "-c", `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`)

			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr
			installCmd.Stdin = os.Stdin

			err := installCmd.Run()
			if err != nil {
				fmt.Fprintf(iostream.ErrOut, "Failed to install Homebrew: %v\n", err)
				span.SetTag("error", true)
				span.Finish(tracer.WithError(err))
				return err
			}

			// Update PATH after successful installation
			if err := updatePath(ctx, iostream); err != nil {
				fmt.Fprintf(iostream.ErrOut, "Warning: %v\n", err)
				// We don't return here because Homebrew is still installed successfully
			} else {
				fmt.Fprintln(iostream.Out, cs.Green("Homebrew installed successfully and PATH updated."))
				fmt.Fprintln(iostream.Out, "If you don't see the changes immediately, you may need to restart your terminal or manually source your shell configuration file.")
			}
			span.SetTag("status", "success")
			return nil
		},
	}

	return cmd
}

// IsHomebrewInstalled checks if Homebrew is installed on the system.
func IsHomebrewInstalled(ctx context.Context) bool {
	// The 'which' command searches for the Homebrew executable in the system path.
	cmd := execCommandContext(ctx, "which", "brew")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false // 'which' did not find the Homebrew binary, or another error occurred
	}

	// Check the output. If it contains the path to the brew executable, Homebrew is installed.
	return strings.Contains(string(output), "/brew")
}

func updatePath(ctx context.Context, iostream *iostreams.IOStreams) error {
	homebrewPath := "/opt/homebrew/bin" // Default path for Apple Silicon Macs
	if _, err := os.Stat(homebrewPath); os.IsNotExist(err) {
		homebrewPath = "/usr/local/bin" // Default path for Intel Macs
	}

	currentPath := os.Getenv("PATH")
	if !strings.Contains(currentPath, homebrewPath) {
		newPath := fmt.Sprintf("%s:%s", homebrewPath, currentPath)
		if err := os.Setenv("PATH", newPath); err != nil {
			return fmt.Errorf("failed to update PATH: %v", err)
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	shellConfigFiles := []string{".zshrc"}
	updatedFile := ""
	for _, file := range shellConfigFiles {
		configPath := filepath.Join(homeDir, file)
		if _, err := os.Stat(configPath); err == nil {
			content, err := os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("failed to read %s: %v", file, err)
			}

			if !strings.Contains(string(content), fmt.Sprintf("export PATH=\"%s:$PATH\"", homebrewPath)) {
				newContent := fmt.Sprintf("%s\nexport PATH=\"%s:$PATH\"\n", string(content), homebrewPath)
				if err := os.WriteFile(configPath, []byte(newContent), 0644); err != nil {
					return fmt.Errorf("failed to update %s: %v", file, err)
				}
				fmt.Fprintf(iostream.Out, "Updated %s with Homebrew path\n", file)
				return nil // Exit after updating the first file that needs it
			}
		}
	}

	if updatedFile != "" {
		// Attempt to source the updated file
		cmd := execCommandContext(ctx, "zsh", "-c", fmt.Sprintf("source %s", updatedFile))
		cmd.Stdout = iostream.Out
		cmd.Stderr = iostream.ErrOut
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(iostream.ErrOut, "Warning: Failed to source updated configuration: %v\n", err)
			fmt.Fprintln(iostream.Out, "You may need to restart your terminal or manually source your shell configuration file.")
		} else {
			fmt.Fprintln(iostream.Out, "Shell configuration has been sourced.")
		}
	}

	return nil
}
