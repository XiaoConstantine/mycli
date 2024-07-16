package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/XiaoConstantine/mycli/pkg/build"
	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/XiaoConstantine/mycli/pkg/utils"
	"github.com/spf13/cobra"
)

var CheckForUpdatesFunc = CheckForUpdates

// UpdateCmd creates and returns a cobra.Command for the 'update' command of mycli.
//
// The update command checks for and installs the latest version of mycli.
// It compares the current version with the latest available version and
// performs the update if a newer version is found.
//
// Usage:
//
//	mycli update
//
// The function sets up the command's flags and its Run function. It uses the provided
// IOStreams for input/output operations and interacts with the user to confirm the update.
//
// Parameters:
//   - iostream: An iostreams.IOStreams instance for handling input/output operations.
//
// Returns:
//   - *cobra.Command: A pointer to the created cobra.Command for the update command.
//
// Example:
//
//	// Creating the update command
//	updateCmd := update.UpdateCmd(iostreams.System())
//	rootCmd.AddCommand(updateCmd)
func NewUpdateCmd(iostream *iostreams.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update mycli to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			installDir, err := ensureInstallDirectory()
			if err != nil {
				return err
			}

			if err := ensurePathInZshrc(installDir); err != nil {
				return err
			}

			return updateCLI(iostream)
		},
	}
	return cmd
}

func updateCLI(iostream *iostreams.IOStreams) error {
	currentVersion := build.Version

	// Ensure .mycli/bin directory exists
	installDir, err := ensureInstallDirectory()
	if err != nil {
		return fmt.Errorf("failed to ensure install directory: %w", err)
	}

	// Get the latest release info
	release, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to get latest release: %w", err)
	}

	// Check if update is needed
	if utils.CompareVersions(currentVersion, release.TagName) >= 0 {
		fmt.Fprintln(iostream.Out, "You're already using the latest version of mycli.")
		return nil
	}

	// Determine the asset to download based on the current OS and architecture
	assetURL := ""
	for _, asset := range release.Assets {
		if strings.Contains(strings.ToLower(asset.Name), fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)) {
			assetURL = asset.BrowserDownloadURL
			break
		}
	}

	if assetURL == "" {
		return fmt.Errorf("no suitable release found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download the new binary
	resp, err := http.Get(assetURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	// Determine the install location
	installPath := filepath.Join(installDir, "mycli")

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "mycli-update")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Copy the downloaded content to the temporary file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write update to temporary file: %w", err)
	}
	tmpFile.Close()

	// Make the temporary file executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to make new binary executable: %w", err)
	}

	// Replace the old binary with the new one
	if err := os.Rename(tmpFile.Name(), installPath); err != nil {
		return fmt.Errorf("failed to replace old binary: %w", err)
	}

	fmt.Fprintf(iostream.Out, "mycli has been updated successfully to version %s!\n", release.TagName)
	return nil
}

var ensureInstallDirectory = func() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	installDir := filepath.Join(home, ".mycli", "bin")
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create install directory: %w", err)
	}

	return installDir, nil
}

func ensurePathInZshrc(installDir string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	zshrcPath := filepath.Join(home, ".zshrc")
	zshrcContent, err := os.ReadFile(zshrcPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read .zshrc: %w", err)
	}

	pathLine := fmt.Sprintf("export PATH=\"$PATH:%s\"", installDir)
	if !strings.Contains(string(zshrcContent), pathLine) {
		f, err := os.OpenFile(zshrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open .zshrc: %w", err)
		}
		defer f.Close()

		if _, err := f.WriteString("\n" + pathLine + "\n"); err != nil {
			return fmt.Errorf("failed to write to .zshrc: %w", err)
		}

		fmt.Println("Added .mycli/bin to your PATH in .zshrc. Please restart your terminal or run 'source ~/.zshrc' to apply the changes.")
	}

	return nil
}

const repoURL = "https://api.github.com/repos/XiaoConstantine/mycli/releases/latest"

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

var getLatestRelease = func() (*githubRelease, error) {
	resp, err := http.Get(repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release JSON: %w", err)
	}

	return &release, nil
}

func CheckForUpdates(iostream *iostreams.IOStreams) (bool, string, error) {
	currentVersion := build.Version

	// Get the latest release info
	release, err := getLatestRelease()
	if err != nil {
		return false, "", fmt.Errorf("failed to get latest release: %w", err)
	}

	hasUpdate := utils.CompareVersions(currentVersion, release.TagName) < 0
	return hasUpdate, release.TagName, nil
}
