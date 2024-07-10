package install

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewInstallCmd(t *testing.T) {
	t.Run("Basic command creation", func(t *testing.T) {
		ios, _, _, _ := iostreams.Test()
		cmd := NewInstallCmd(ios)
		assert.NotNil(t, cmd)
		assert.Equal(t, "install", cmd.Use)
		assert.Equal(t, "Install software", cmd.Short)
	})
}

// Helper function to replace subcommands with mocks.
func mockSubcommands(cmd *cobra.Command, outcomes map[string]error) {
	for _, subcmd := range cmd.Commands() {
		// Capture the subcommand name and the intended mock outcome
		if outcome, ok := outcomes[subcmd.Use]; ok {
			subcmd.RunE = func(cmd *cobra.Command, args []string) error {
				return outcome // Return the mock outcome when the subcommand is run
			}
		}
	}
}

func TestNewInstallCmd_NonInteractiveMode_Force(t *testing.T) {
	ios, _, outBuf, errBuf := iostreams.Test()
	cmd := NewInstallCmd(ios)

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "test-config-install")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir) // Clean up

	// Create a temporary config file
	configContent := []byte(`
tools:
  - name: example-tool
`)
	configPath := filepath.Join(tempDir, "test-config.yaml")
	err = os.WriteFile(configPath, configContent, 0644)
	assert.NoError(t, err)

	// Set flags to simulate non-interactive mode
	cmd.SetArgs([]string{"--non-interactive", "--config", configPath, "--force"})

	// Disable base commands
	cmd.Root().CompletionOptions.DisableDefaultCmd = true
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	// Map of subcommand uses to their desired mock outcomes
	outcomes := map[string]error{
		"homebrew": nil,                                     // Mock homebrew command to succeed
		"xcode":    errors.New("xcode installation failed"), // Mock xcode command to fail
		"tools":    nil,
	}

	mockSubcommands(cmd, outcomes)

	err = cmd.Execute()
	assert.Error(t, err) // Expect an error because xcode installation is mocked to fail
	assert.Contains(t, errBuf.String(), "xcode installation failed")
	assert.Contains(t, outBuf.String(), "Running all installation subcommands...\n") // Check output for successful mock
}

func TestNewInstallCmd_FileNotExist(t *testing.T) {
	ios, _, _, errBuf := iostreams.Test()
	cmd := NewInstallCmd(ios)

	// Set flags to simulate file not found
	cmd.SetArgs([]string{"--non-interactive", "--config", "/non/existent/path.yaml"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, errBuf.String(), "Error: Config file does not exist at path")
}

func TestNewInstallCmdFlagsErrorsEdgeCases(t *testing.T) {
	ios, _, outBuf, errBuf := iostreams.Test()
	cmd := NewInstallCmd(ios)

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "test-config-install")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir) // Clean up

	// Create a temporary config file
	configContent := []byte(`
tools:
  - name: example-tool
`)
	configPath := filepath.Join(tempDir, "test-config.yaml")
	err = os.WriteFile(configPath, configContent, 0644)
	assert.NoError(t, err)

	// Setup mock outcomes with no specific subcommand behavior
	// Map of subcommand uses to their desired mock outcomes
	outcomes := map[string]error{
		"homebrew": nil,                                     // Mock homebrew command to succeed
		"xcode":    errors.New("xcode installation failed"), // Mock xcode command to fail
		"tools":    nil,
	}

	mockSubcommands(cmd, outcomes)

	tests := []struct {
		name        string
		args        []string
		setupFunc   func()
		cleanupFunc func()
		expectErr   bool
		errorMsg    string
		successMsg  string
	}{
		{
			name:      "Non-interactive with missing config file",
			args:      []string{"--non-interactive", "--config", "/fake/path.yaml"},
			expectErr: true,
			errorMsg:  "Config file does not exist at path",
		},
		{
			name: "Valid non-interactive force install",
			args: []string{"--non-interactive", "--force", "--config", configPath},
			setupFunc: func() {
				// Mock environment variable and file path resolution
				os.Setenv("HOME", "/tmp")
				outcomes := map[string]error{
					"homebrew": nil, // Mock homebrew command to succeed
					"xcode":    nil, // Mock xcode command to fail
					"tools":    nil,
				}
				mockSubcommands(cmd, outcomes)

			},
			cleanupFunc: func() {
				os.Unsetenv("HOME")
				mockSubcommands(cmd, map[string]error{})
			},
			expectErr:  false,
			successMsg: "All installations completed successfully.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupFunc != nil {
				tc.setupFunc()
			}
			outBuf.Reset()
			errBuf.Reset()

			// Disable base commands
			cmd.Root().CompletionOptions.DisableDefaultCmd = true
			cmd.SetHelpCommand(&cobra.Command{Hidden: true})

			cmd.SetArgs(tc.args)
			err := cmd.Execute()

			if tc.expectErr {
				assert.Error(t, err)
				assert.Contains(t, errBuf.String(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, outBuf.String(), tc.successMsg)
			}

			if tc.cleanupFunc != nil {
				tc.cleanupFunc()
			}
		})
	}
}
