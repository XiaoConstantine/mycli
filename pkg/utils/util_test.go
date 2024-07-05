package utils

import (
	"context"
	"os"
	"os/exec"
	"os/user"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfigureItem(t *testing.T) {
	config := ToolConfig{
		Configure: []ConfigureItem{
			{Name: "git", ConfigURL: "https://example.com/git", InstallPath: "/usr/local/bin"},
			{Name: "vim", ConfigURL: "https://example.com/vim", InstallPath: "/usr/local/bin"},
		},
	}

	item, err := config.GetConfigureItem("git")
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, "https://example.com/git", item.ConfigURL)

	_, err = config.GetConfigureItem("nonexistent")
	assert.Error(t, err)
}

func TestLoadToolsConfig(t *testing.T) {
	// Temporary file to mimic a YAML config file
	tmpfile, err := os.CreateTemp("", "config*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name()) // clean up

	content := `
tools:
  - name: "zsh"
  - name: "kubectl"
`
	_, err = tmpfile.Write([]byte(content))
	assert.NoError(t, err)
	err = tmpfile.Close()
	assert.NoError(t, err)

	// Test loading the config
	config, err := LoadToolsConfig(tmpfile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Len(t, config.Tools, 2)
	assert.Equal(t, "zsh", config.Tools[0].Name)
	assert.Equal(t, "kubectl", config.Tools[1].Name)

	// Test file not found error
	_, err = LoadToolsConfig("nonexistent.yaml")
	assert.Error(t, err)
}

// Mock for the exec.Command
type MockCommand struct {
	mock.Mock
}

func (m *MockCommand) CommandContext(ctx context.Context, command string, args ...string) *exec.Cmd {
	args = m.Called(ctx, command, args).Get(0).([]string)
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestIsAdmin(t *testing.T) {
	// Save the current execCommandContext and restore it after the tests
	oldExecCommandContext := execCommandContext
	defer func() { execCommandContext = oldExecCommandContext }()

	tests := []struct {
		name        string
		mockOutput  string
		shouldError bool
		want        bool
	}{
		{
			name:       "User is admin",
			mockOutput: "username admin wheel",
			want:       true,
		},
		{
			name:       "User is not admin",
			mockOutput: "username wheel",
			want:       false,
		},
		{
			name:        "Command fails",
			mockOutput:  "",
			shouldError: true,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the execCommandContext function
			execCommandContext = func(ctx context.Context, command string, args ...string) *exec.Cmd {
				cmd := exec.Command("echo", tt.mockOutput)
				cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"} // Ensure it uses the test environment
				return cmd
			}

			// Create a user and pass it to IsAdmin
			u := &user.User{Username: "username"}
			ctx := context.Background()
			got := IsAdmin(ctx, u)
			assert.Equal(t, tt.want, got, "IsAdmin did not return expected value")
		})
	}
}
