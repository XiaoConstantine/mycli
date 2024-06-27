package install

import (
	"fmt"
	"mycli/pkg/iostreams"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

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
