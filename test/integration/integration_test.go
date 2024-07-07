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
