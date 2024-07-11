package xcode

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/XiaoConstantine/mycli/pkg/utils"

	"github.com/spf13/cobra"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var execCommandContext = exec.CommandContext

// NewCmdXcode creates a new cobra.Command that installs Xcode on the user's system.
// The command runs the "xcode-select --install" command, which prompts the user to install Xcode.
// The command output and errors are forwarded to the user's terminal.
func NewCmdXcode(iostream *iostreams.IOStreams, statsCollector *utils.StatsCollector) *cobra.Command {
	cs := iostream.ColorScheme()
	var stats *utils.Stats

	cmd := &cobra.Command{
		Use:   "xcode",
		Short: cs.GreenBold("Install xcode"),
		// Long:   actionsExplainer(cs),
		Hidden: true,
		Annotations: map[string]string{
			"group": "install",
		},
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			startTime := time.Now()
			stats = &utils.Stats{
				Name:      "Xcode",
				Operation: "install",
			}
			span, ctx := tracer.StartSpanFromContext(cmd.Context(), "install_xcode")
			defer span.Finish()
			if isXcodeAlreadyInstalled(ctx) {
				duration := time.Since(startTime)
				stats.Duration = duration
				stats.Status = "success"
				statsCollector.AddStat(stats)
				span.SetTag("status", "success")
				span.Finish()
				fmt.Println("Xcode is already installed.")
				return nil // Early exit if Xcode is already installed
			}
			installCmd := execCommandContext(ctx, "xcode-select", "--install")

			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr
			installCmd.Stdin = os.Stdin

			err := installCmd.Run()
			duration := time.Since(startTime)
			stats.Duration = duration
			if err != nil {
				fmt.Fprintf(iostream.ErrOut, "Failed to install xcode: %v\n", err)
				stats.Status = "error"
				statsCollector.AddStat(stats)

				span.SetTag("error", true)
				span.Finish(tracer.WithError(err))
				return err
			}
			stats.Status = "success"
			statsCollector.AddStat(stats)

			span.SetTag("status", "success")
			return nil
		},
	}

	return cmd
}

// isXcodeAlreadyInstalled checks if Xcode is already installed by looking for its directory.
func isXcodeAlreadyInstalled(ctx context.Context) bool {
	cmd := execCommandContext(ctx, "xcode-select", "-p")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false // xcode-select command failed, likely Xcode not installed
	}

	outputStr := strings.TrimSpace(string(output))
	fmt.Printf("xcode-select -p output: %s\n", outputStr)
	// Check output for a known path component, like "/Applications/Xcode.app"
	// Check for various possible paths
	knownPaths := []string{
		"/Applications/Xcode.app",
		"/Library/Developer/CommandLineTools",
		"/Applications/Xcode.app/Contents/Developer",
		"/Users/runner/Library/Developer/Xcode", // Common path in GitHub Actions
	}

	for _, path := range knownPaths {
		if strings.Contains(outputStr, path) {
			return true
		}
	}

	return false
}
