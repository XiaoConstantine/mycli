package homebrew

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/XiaoConstantine/mycli/pkg/iostreams"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUser is a mock implementation of the User interface.
type MockUser struct {
	mock.Mock
}

func (m *MockUser) Username() string {
	args := m.Called()
	return args.String(0)
}

// MockExecCommand is a helper function to mock exec.Command.
func MockExecCommand(ctx context.Context, command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestNewCmdHomeBrew(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cmd := NewCmdHomeBrew(ios)

	assert.NotNil(t, cmd)
	assert.Equal(t, "homebrew", cmd.Use)
	assert.True(t, cmd.Hidden)
	assert.True(t, cmd.SilenceErrors)
	assert.Equal(t, "install", cmd.Annotations["group"])
}

func TestIsHomebrewInstalled(t *testing.T) {
	// Save current execCommandContext and defer its restoration
	oldExecCommandContext := execCommandContext
	defer func() { execCommandContext = oldExecCommandContext }()

	tests := []struct {
		name       string
		mockOutput string
		want       bool
	}{
		{
			name:       "Homebrew is installed",
			mockOutput: "/usr/local/bin/brew",
			want:       true,
		},
		{
			name:       "Homebrew is not installed",
			mockOutput: "",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execCommandContext = func(ctx context.Context, command string, args ...string) *exec.Cmd {
				// return MockExecCommand(ctx, "echo", tt.mockOutput)
				// Mock setup: Define what the command should return
				cmd := exec.Command("echo", tt.mockOutput)
				cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"} // Ensure it uses the test environment
				return cmd
			}
			got := IsHomebrewInstalled(context.Background())
			assert.Equal(t, tt.want, got)
		})
	}
}
