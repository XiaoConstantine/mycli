package utils

import (
	"context"
	"os"
	"os/exec"
	"os/user"
	"testing"

	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/sys/unix"
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

// Mock for the exec.Command.
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

func TestGetCurrentUser(t *testing.T) {
	u, err := GetCurrentUser()
	assert.NoError(t, err)
	assert.NotNil(t, u)
}

func TestGetSubcommandNames(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.AddCommand(&cobra.Command{Use: "test1"})
	cmd.AddCommand(&cobra.Command{Use: "test2"})

	names := GetSubcommandNames(cmd)
	assert.Equal(t, []string{"test1", "test2"}, names)
}

func TestGetOsInfo(t *testing.T) {
	info := GetOsInfo()
	assert.NotEmpty(t, info["sysname"])
	assert.NotEmpty(t, info["release"])
	assert.NotEmpty(t, info["user"])
}

func TestConvertToRawGitHubURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "Valid GitHub URL",
			input:    "https://github.com/user/repo/blob/master/file.txt",
			expected: "https://raw.githubusercontent.com/user/repo/master/file.txt",
			hasError: false,
		},
		{
			name:     "Non-GitHub URL",
			input:    "https://example.com/file.txt",
			expected: "https://example.com/file.txt",
			hasError: false,
		},
		{
			name:     "Invalid URL",
			input:    "not a url",
			expected: "",
			hasError: true,
		},
		{
			name:     "Invalid GitHub URL format",
			input:    "https://github.com/user/repo",
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertToRawGitHubURL(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetRandomASCIILogo(t *testing.T) {
	logo := getRandomASCIILogo()
	assert.NotEmpty(t, logo)
}

func TestPrintWelcomeMessage(t *testing.T) {
	// Create a mock IOStreams
	ios, _, out, _ := iostreams.Test()

	// Call the function
	PrintWelcomeMessage(ios)
	// Assert that something was written to the output
	assert.NotEmpty(t, out.String())

	// You can also check for specific content if needed
	assert.Contains(t, out.String(), "Welcome to MyCLI!")
	assert.Contains(t, out.String(), "Your personal machine bootstrapping tool")
	assert.Contains(t, out.String(), "Available commands:")
}

// Mock for unix.Uname
func mockUname(uts *unix.Utsname) error {
	copy(uts.Sysname[:], []byte("MockOS"))
	copy(uts.Release[:], []byte("1.0"))
	return nil
}

func TestGetOsInfoWithMock(t *testing.T) {

	info := GetOsInfo()
	// Check that the returned map contains the expected keys
	assert.Contains(t, info, "sysname")
	assert.Contains(t, info, "release")
	assert.Contains(t, info, "user")

	// Check that the values are not empty
	assert.NotEmpty(t, info["sysname"])
	assert.NotEmpty(t, info["release"])
	assert.NotEmpty(t, info["user"])
}
