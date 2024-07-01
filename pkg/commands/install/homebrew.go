package install

import (
	"fmt"
	"mycli/pkg/iostreams"
	"mycli/pkg/utils"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// NewCmdHomeBrew creates a new cobra.Command that installs Homebrew on the system.
// It checks if the current user is an administrator, and if so, runs the Homebrew
// installation script using the current user's credentials. If the current user
// is not an administrator, it prints an error message and exits.
func NewCmdHomeBrew(iostream *iostreams.IOStreams) *cobra.Command {
	cs := iostream.ColorScheme()

	cmd := &cobra.Command{
		Use:   "install homebrew",
		Short: cs.GreenBold("Install homebrew, require admin privileges, make sure enable this via privileges app"),
		// Long:   actionsExplainer(cs),
		GroupID:       "install",
		Hidden:        true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if IsHomebrewInstalled() {
				fmt.Println("Homebrew is installed.")
				return nil
			}
			currentUser, _ := utils.GetCurrentUser()

			if !utils.IsAdmin(currentUser) {
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

// IsHomebrewInstalled checks if Homebrew is installed on the system.
func IsHomebrewInstalled() bool {
	// The 'which' command searches for the Homebrew executable in the system path.
	cmd := exec.Command("which", "brew")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false // 'which' did not find the Homebrew binary, or another error occurred
	}

	// Check the output. If it contains the path to the brew executable, Homebrew is installed.
	return strings.Contains(string(output), "/brew")
}
