//go:build integration
// +build integration

package integration

import (
	"mycli/pkg/commands/root"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigureCommand(t *testing.T) {
	// Set up a temporary directory to act as the home directory
	fakeHomeDir, err := os.MkdirTemp("", "fake-home")
	require.NoError(t, err)
	defer os.RemoveAll(fakeHomeDir)

	// Save the original home directory and set it back after the test
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set the fake ome directory
	os.Setenv("HOME", fakeHomeDir)
	// // Create a 'test-config' directory inside the fake home directory
	// testConfigDir := fakeHomeDir + "/test-config"
	// err = os.Mkdir(testConfigDir, 0755)
	// require.NoError(t, err, "Failed to create test-config directory")

	// Set up a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "mycli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configPath := filepath.Join(tempDir, "test_config.yml")
	testConfig := `
configure:
  - name: test-tool
    config_url: https://raw.githubusercontent.com/username/repo/main/test-config/init.yaml
    install_path: ~/test-config/init.yaml
`
	err = os.WriteFile(configPath, []byte(testConfig), 0644)
	require.NoError(t, err)

	// Set up mock HTTP server for the config file
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock config content"))
	}))
	defer mockServer.Close()

	// Replace the GitHub URL in the config with our mock server URL
	testConfig = strings.Replace(testConfig, "https://raw.githubusercontent.com/username/repo/main/test-config/init.yaml", mockServer.URL, 1)
	err = os.WriteFile(configPath, []byte(testConfig), 0644)
	require.NoError(t, err)

	// Run the CLI in non-interactive mode
	args := []string{"--non-interactive", "configure", "--config", configPath, "--force"}
	exitCode := root.Run(args)

	// Assert that the exit code is as expected
	assert.Equal(t, root.ExitCode(0), exitCode)

	// Verify that the config file was processed
	require.NoError(t, err)
	expectedConfigPath := filepath.Join(fakeHomeDir, "test-config/init.yaml")

	// Check if the file exists
	_, err = os.Stat(expectedConfigPath)
	assert.NoError(t, err)

	// Check the content of the created file
	content, err := os.ReadFile(expectedConfigPath)
	require.NoError(t, err)
	assert.Equal(t, "mock config content", string(content))

	// Clean up the created file
	os.Remove(expectedConfigPath)
}

func TestInstallCommand(t *testing.T) {
	// Set up a temporary directory to act as the home directory
	fakeHomeDir, err := os.MkdirTemp("", "fake-home")
	require.NoError(t, err)
	defer os.RemoveAll(fakeHomeDir)

	// Save the original home directory and set it back after the test
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set the fake home directory
	os.Setenv("HOME", fakeHomeDir)

	// Set up a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "mycli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test config file for tools
	configPath := filepath.Join(tempDir, "test_config.yaml")
	testConfig := `
tools:
  - name: test-tool
    cask: test-cask
`
	err = os.WriteFile(configPath, []byte(testConfig), 0644)
	require.NoError(t, err)

	// Test install homebrew
	t.Run("Install Homebrew", func(t *testing.T) {
		args := []string{"--non-interactive", "install", "homebrew"}
		exitCode := root.Run(args)
		assert.Equal(t, root.ExitCode(0), exitCode)

		// Check if Homebrew is installed
		// This is a simple check and might need to be adjusted based on your actual installation process
		// _, err := os.Stat("/usr/local/bin/brew")
		_, err := os.Stat("/opt/homebrew/bin/brew")

		assert.NoError(t, err, "Homebrew was not installed")
	})

	// Test install Xcode
	t.Run("Install Xcode", func(t *testing.T) {
		args := []string{"--non-interactive", "install", "xcode"}
		exitCode := root.Run(args)
		assert.Equal(t, root.ExitCode(0), exitCode)

		// Check if Xcode is installed
		// This is a simple check and might need to be adjusted based on your actual installation process
		_, err := os.Stat("/Applications/Xcode.app")
		assert.NoError(t, err, "Xcode was not installed")
	})

	// Test install tools
	// t.Run("Install Tools", func(t *testing.T) {
	// 	args := []string{"--non-interactive", "install", "tools", "--config", configPath, "--force"}
	// 	exitCode := root.Run(args)
	// 	assert.Equal(t, root.ExitCode(0), exitCode)

	// 	// Check if the test-tool was installed
	// 	// This is a simple check and might need to be adjusted based on your actual installation process
	// 	expectedToolDir := filepath.Join(fakeHomeDir, "test-tool")
	// 	_, err := os.Stat(expectedToolDir)
	// 	assert.NoError(t, err, "Expected tool directory was not created")
	// })

	// // Test install everything
	// t.Run("Install Everything", func(t *testing.T) {
	// 	args := []string{"--non-interactive", "install", "--config", configPath, "--force"}
	// 	exitCode := root.Run(args)
	// 	assert.Equal(t, root.ExitCode(0), exitCode)

	// 	// Check if Homebrew is installed
	// 	_, err := os.Stat("/opt/homebrew/bin/brew")
	// 	assert.NoError(t, err, "Homebrew was not installed")

	// 	// Check if Xcode is installed
	// 	_, err = os.Stat("/Applications/Xcode.app")
	// 	assert.NoError(t, err, "Xcode was not installed")

	// 	// Check if the test-tool was installed
	// 	expectedToolDir := filepath.Join(fakeHomeDir, "test-tool")
	// 	_, err = os.Stat(expectedToolDir)
	// 	assert.NoError(t, err, "Expected tool directory was not created")
	// })

	// Clean up any created files or directories
	os.RemoveAll(filepath.Join(fakeHomeDir, "test-tool"))
}
