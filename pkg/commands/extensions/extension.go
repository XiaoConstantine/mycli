package extensions

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/spf13/cobra"
)

const ExtensionPrefix = "mycli-"

type Extension struct {
	Name string
	Path string
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular() && (info.Mode().Perm()&0111 != 0)
}

func IsExtension(path string) bool {
	base := filepath.Base(path)
	return strings.HasPrefix(base, ExtensionPrefix) && isExecutable(path)
}

func GetExtensionsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mycli", "extensions")
}

func (e *Extension) Execute(args []string) error {
	cmd := exec.Command(e.Path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func NewCmdExtension(iostream *iostreams.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extension",
		Short: "Manage mycli extensions",
	}

	cmd.AddCommand(newExtensionInstallCmd(iostream))
	cmd.AddCommand(newExtensionListCmd(iostream))
	cmd.AddCommand(newExtensionRemoveCmd(iostream))
	cmd.AddCommand(newExtensionUpdateCmd(iostream))

	return cmd
}

func newExtensionInstallCmd(iostream *iostreams.IOStreams) *cobra.Command {
	return &cobra.Command{
		Use:   "install <repository>",
		Short: "Install a mycli extension",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo := args[0]
			extDir := GetExtensionsDir()
			extName := filepath.Base(repo)
			extPath := filepath.Join(extDir, ExtensionPrefix+extName)

			if err := os.MkdirAll(extDir, 0755); err != nil {
				return fmt.Errorf("failed to create extensions directory: %w", err)
			}

			gitCmd := exec.Command("git", "clone", repo, extPath)
			gitCmd.Stdout = iostream.Out
			gitCmd.Stderr = iostream.ErrOut

			if err := gitCmd.Run(); err != nil {
				return fmt.Errorf("failed to clone extension repository: %w", err)
			}

			fmt.Fprintf(iostream.Out, "Successfully installed extension '%s'\n", extName)
			return nil
		},
	}
}

func newExtensionListCmd(iostream *iostreams.IOStreams) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed mycli extensions",
		RunE: func(cmd *cobra.Command, args []string) error {
			extDir := GetExtensionsDir()
			entries, err := os.ReadDir(extDir)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Fprintln(iostream.Out, "No extensions installed")
					return nil
				}
				return fmt.Errorf("failed to read extensions directory: %w", err)
			}

			for _, entry := range entries {
				if entry.IsDir() && strings.HasPrefix(entry.Name(), ExtensionPrefix) {
					fmt.Fprintln(iostream.Out, entry.Name()[len(ExtensionPrefix):])
				}
			}
			return nil
		},
	}
}

func newExtensionRemoveCmd(iostream *iostreams.IOStreams) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <extension-name>",
		Short: "Remove a mycli extension",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			extName := args[0]
			extDir := GetExtensionsDir()
			extPath := filepath.Join(extDir, ExtensionPrefix+extName)

			if err := os.RemoveAll(extPath); err != nil {
				return fmt.Errorf("failed to remove extension: %w", err)
			}

			fmt.Fprintf(iostream.Out, "Successfully removed extension '%s'\n", extName)
			return nil
		},
	}
}

func newExtensionUpdateCmd(iostream *iostreams.IOStreams) *cobra.Command {
	return &cobra.Command{
		Use:   "update <extension-name>",
		Short: "Update a mycli extension",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			extName := args[0]
			extDir := GetExtensionsDir()
			extPath := filepath.Join(extDir, ExtensionPrefix+extName)

			gitCmd := exec.Command("git", "-C", extPath, "pull")
			gitCmd.Stdout = iostream.Out
			gitCmd.Stderr = iostream.ErrOut

			if err := gitCmd.Run(); err != nil {
				return fmt.Errorf("failed to update extension: %w", err)
			}

			fmt.Fprintf(iostream.Out, "Successfully updated extension '%s'\n", extName)
			return nil
		},
	}
}
