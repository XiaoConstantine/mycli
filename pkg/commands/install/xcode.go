package install

import (
	"fmt"
	"mycli/pkg/iostreams"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
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
		Hidden:        true,
		GroupID:       "install",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if isXcodeAlreadyInstalled() {
				fmt.Println("Xcode is already installed.")
				return nil // Early exit if Xcode is already installed
			}
			fmt.Fprintf(iostream.Out, cs.Green("Installing xcode with su current user, enter your password when prompt\n"))
			installCmd := exec.Command("xcode-select", "--install")

			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr
			installCmd.Stdin = os.Stdin

			err := installCmd.Run()
			if err != nil {
				// fmt.Printf("Failed to install xcode: %v\n", err)
				return err
			}
			return nil
		},
	}

	return cmd
}

// isXcodeAlreadyInstalled checks if Xcode is already installed by looking for its directory.
func isXcodeAlreadyInstalled() bool {
	cmd := exec.Command("xcode-select", "-p")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false // xcode-select command failed, likely Xcode not installed
	}
	// Check output for a known path component, like "/Applications/Xcode.app"
	return strings.Contains(string(output), "/Applications/Xcode.app")
}
