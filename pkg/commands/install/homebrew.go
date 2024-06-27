package install

import (
	"fmt"
	"mycli/pkg/iostreams"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func NewCmdHomeBrew(iostream *iostreams.IOStreams) *cobra.Command {
	cs := iostream.ColorScheme()

	cmd := &cobra.Command{
		Use:   "install",
		Short: cs.GreenBold("Install homebrew, require admin privileges, make sure enable this via privileges app"),
		// Long:   actionsExplainer(cs),
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			currentUser, _ := getCurrentUser()

			if !isAdmin(currentUser) {
				fmt.Println(cs.Red("You need to be an administrator to install Homebrew. Please run this command from an admin account."))
				os.Exit(1)
			}

			fmt.Fprintf(iostream.Out, cs.Green("Installing homebrew with su current user, enter your password when prompt\n"))
			installCmd := exec.Command("su", currentUser.Username, "-c", `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`)

			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr
			installCmd.Stdin = os.Stdin

			err := installCmd.Run()
			if err != nil {
				fmt.Printf("Failed to install Homebrew: %v\n", err)
				return err
			}
			return nil
		},
	}

	return cmd
}