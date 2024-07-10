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

// Mock UserUtils for testing.
type MockUserUtils struct {
	mock.Mock
}

func (m *MockUserUtils) GetCurrentUser() (*user.User, error) {
	args := m.Called()
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserUtils) IsAdmin(ctx context.Context, u *user.User) bool {
	args := m.Called(ctx, u)
	return args.Bool(0)
}

func TestMockUserUtils(t *testing.T) {
	m := &MockUserUtils{}

	// Test GetCurrentUser
	expectedUser := &user.User{Username: "testuser"}
	m.On("GetCurrentUser").Return(expectedUser, nil)
	u, err := m.GetCurrentUser()
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, u)

	// Test IsAdmin
	ctx := context.Background()
	m.On("IsAdmin", ctx, expectedUser).Return(true)
	isAdmin := m.IsAdmin(ctx, expectedUser)
	assert.True(t, isAdmin)

	m.AssertExpectations(t)
}

func TestRealUserUtils_IsAdmin(t *testing.T) {
	oldExecCommandContext := execCommandContext
	defer func() { execCommandContext = oldExecCommandContext }()

	tests := []struct {
		name       string
		mockOutput string
		want       bool
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execCommandContext = func(ctx context.Context, command string, args ...string) *exec.Cmd {
				return exec.Command("echo", tt.mockOutput)
			}

			utils := RealUserUtils{}
			u := &user.User{Username: "username"}
			ctx := context.Background()
			got := utils.IsAdmin(ctx, u)
			assert.Equal(t, tt.want, got, "IsAdmin did not return expected value")
		})
	}
}

func TestRealUserUtils_GetCurrentUser(t *testing.T) {
	utils := RealUserUtils{}
	u, err := utils.GetCurrentUser()
	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.NotEmpty(t, u.Username)
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
		wantErr  bool
	}{
		{
			name:     "Basic GitHub URL",
			input:    "https://github.com/username/repo/blob/main/file.txt",
			expected: "https://raw.githubusercontent.com/username/repo/main/file.txt",
			wantErr:  false,
		},
		{
			name:     "GitHub URL without explicit branch",
			input:    "https://github.com/username/repo/file.txt",
			expected: "https://raw.githubusercontent.com/username/repo/main/file.txt",
			wantErr:  false,
		},
		{
			name:     "GitHub URL with subdirectory",
			input:    "https://github.com/username/repo/blob/main/dir/file.txt",
			expected: "https://raw.githubusercontent.com/username/repo/main/dir/file.txt",
			wantErr:  false,
		},
		{
			name:     "GitHub URL with multiple subdirectories",
			input:    "https://github.com/username/repo/blob/main/dir1/dir2/file.txt",
			expected: "https://raw.githubusercontent.com/username/repo/main/dir1/dir2/file.txt",
			wantErr:  false,
		},
		{
			name:     "GitHub URL with 'tree' instead of 'blob'",
			input:    "https://github.com/username/repo/tree/main/dir",
			expected: "https://raw.githubusercontent.com/username/repo/main/dir",
			wantErr:  false,
		},
		{
			name:     "GitHub URL with different branch",
			input:    "https://github.com/username/repo/blob/develop/file.txt",
			expected: "https://raw.githubusercontent.com/username/repo/develop/file.txt",
			wantErr:  false,
		},
		{
			name:     "GitHub URL with trailing slash",
			input:    "https://github.com/username/repo/blob/main/dir/",
			expected: "https://raw.githubusercontent.com/username/repo/main/dir/",
			wantErr:  false,
		},
		{
			name:     "Non-GitHub URL",
			input:    "https://gitlab.com/username/repo/blob/main/file.txt",
			expected: "https://gitlab.com/username/repo/blob/main/file.txt",
			wantErr:  false,
		},
		{
			name:    "Invalid URL",
			input:   "not-a-url",
			wantErr: true,
		},
		{
			name:    "GitHub URL with too few parts",
			input:   "https://github.com/username",
			wantErr: true,
		},
		{
			name:     "GitHub URL with query parameters",
			input:    "https://github.com/username/repo/blob/main/file.txt?raw=true",
			expected: "https://raw.githubusercontent.com/username/repo/main/file.txt",
			wantErr:  false,
		},
		{
			name:     "GitHub URL with hash",
			input:    "https://github.com/username/repo/blob/main/file.txt#L10",
			expected: "https://raw.githubusercontent.com/username/repo/main/file.txt",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertToRawGitHubURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToRawGitHubURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ConvertToRawGitHubURL() = %v, want %v", got, tt.expected)
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

func TestCompareVersions(t *testing.T) {
	testCases := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"v1.0.0", "v1.0.0", 0},
		{"v1.0.0", "v1.0.1", -1},
		{"v1.1.0", "v1.0.0", 1},
		{"v2.0.0", "v1.9.9", 1},
		{"v0.9.9", "v1.0.0", -1},
		{"v1.0.0", "v1.0.0-alpha", 1},
		{"v1.0.0-beta", "v1.0.0-alpha", 1},
		{"v1.0.0-rc1", "v1.0.0-rc2", -1},
	}

	for _, tc := range testCases {
		t.Run(tc.v1+" vs "+tc.v2, func(t *testing.T) {
			result := CompareVersions(tc.v1, tc.v2)
			assert.Equal(t, tc.expected, result)
		})
	}
}
