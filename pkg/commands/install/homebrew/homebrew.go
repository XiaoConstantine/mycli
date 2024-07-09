package homebrew

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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
