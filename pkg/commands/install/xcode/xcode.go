package xcode

import (
	"context"
	"fmt"
	"mycli/pkg/iostreams"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// NewCmdXcode creates a new cobra.Command that installs Xcode on the user's system.
// The command runs the "xcode-select --install" command, which prompts the user to install Xcode.
// The command output and errors are forwarded to the user's terminal.
func NewCmdXcode(iostream *iostreams.IOStreams) *cobra.Command {
	cs := iostream.ColorScheme()

	cmd := &cobra.Command{
		Use:   "install xcode",
		Short: cs.GreenBold("Install xcode"),
		// Long:   actionsExplainer(cs),
		Hidden: true,
		Annotations: map[string]string{
			"group": "install",
		},
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			span, ctx := tracer.StartSpanFromContext(cmd.Context(), "installxcode")
			defer span.Finish()
			if isXcodeAlreadyInstalled(ctx) {
				fmt.Println("Xcode is already installed.")
				return nil // Early exit if Xcode is already installed
			}
			installCmd := exec.CommandContext(ctx, "xcode-select", "--install")

			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr
			installCmd.Stdin = os.Stdin

			err := installCmd.Run()
			if err != nil {
				// fmt.Printf("Failed to install xcode: %v\n", err)
				span.SetTag("error", true)
				span.Finish(tracer.WithError(err))
				return err
			}
			return nil
		},
	}

	return cmd
}

// isXcodeAlreadyInstalled checks if Xcode is already installed by looking for its directory.
func isXcodeAlreadyInstalled(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "xcode-select", "-p")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false // xcode-select command failed, likely Xcode not installed
	}
	// Check output for a known path component, like "/Applications/Xcode.app"
	return strings.Contains(string(output), "/Applications/Xcode.app") || strings.Contains(string(output), "CommandLineTools")
}
