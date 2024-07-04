package configure

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"mycli/pkg/iostreams"
	"mycli/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigureToolsCmd(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cmd := NewConfigureCmd(ios)

	assert.NotNil(t, cmd)
	assert.Equal(t, "configure", cmd.Use)
	assert.Equal(t, "configure", cmd.Annotations["group"])
}

func TestConfigureToolsFromConfig(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "test-configure")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test server to serve configuration files
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("test configuration content"))
		if err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer testServer.Close()

	// Create a test configuration
	config := &utils.ToolConfig{
		Configure: []utils.ConfigureItem{
			{
				Name:        "test-tool",
				ConfigURL:   testServer.URL,
				InstallPath: filepath.Join(tempDir, "test-tool-config"),
			},
		},
	}

	ios, _, stdout, stderr := iostreams.Test()
	err = ConfigureToolsFromConfig(ios, config, context.Background(), false)

	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "Configuring test-tool...")
	assert.Contains(t, stdout.String(), "All requested tools have been configured successfully.")
	assert.Empty(t, stderr.String())

	// Check if the file was created and has the correct content
	content, err := os.ReadFile(filepath.Join(tempDir, "test-tool-config"))
	require.NoError(t, err)
	assert.Equal(t, "test configuration content", string(content))
}

func TestConfigureTool(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "test-configure-tool")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test server to serve configuration files
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("test configuration content"))
		if err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer testServer.Close()

	testCases := []struct {
		name        string
		item        utils.ConfigureItem
		force       bool
		expectError bool
	}{
		{
			name: "Successful configuration",
			item: utils.ConfigureItem{
				Name:        "test-tool",
				ConfigURL:   testServer.URL,
				InstallPath: filepath.Join(tempDir, "test-tool-config"),
			},
			force:       false,
			expectError: false,
		},
		{
			name: "File already exists - no force",
			item: utils.ConfigureItem{
				Name:        "existing-tool",
				ConfigURL:   testServer.URL,
				InstallPath: filepath.Join(tempDir, "existing-tool-config"),
			},
			force:       false,
			expectError: true,
		},
		{
			name: "File already exists - with force",
			item: utils.ConfigureItem{
				Name:        "existing-tool",
				ConfigURL:   testServer.URL,
				InstallPath: filepath.Join(tempDir, "existing-tool-config"),
			},
			force:       true,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// For the "existing file" tests, create the file first
			if tc.item.Name == "existing-tool" {
				err := os.WriteFile(tc.item.InstallPath, []byte("existing content"), 0644)
				require.NoError(t, err)
			}

			err := configureTool(tc.item, context.Background(), tc.force)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Check if the file was created and has the correct content
				content, err := os.ReadFile(tc.item.InstallPath)
				require.NoError(t, err)
				assert.Equal(t, "test configuration content", string(content))
			}
		})
	}
}

func TestExpandTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	testCases := []struct {
		input    string
		expected string
	}{
		{"~/test", filepath.Join(home, "test")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"~user/test", "~user/test"}, // should not expand for other users
		{"~", home},                  // just ~ should expand to home
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := expandTilde(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
