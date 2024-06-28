package install

import (
	"fmt"
	"mycli/pkg/iostreams"
	"os"
	"os/exec"

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
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {

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
