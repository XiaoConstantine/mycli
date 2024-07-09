package homebrew

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"os/user"
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

type mockUtils struct {
	mock.Mock
}

func (m *mockUtils) GetCurrentUser() (*user.User, error) {
	args := m.Called()
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockUtils) IsAdmin(ctx context.Context, user *user.User) bool {
	args := m.Called(ctx, user)
	return args.Bool(0)
}

func TestNewCmdHomeBrew(t *testing.T) {
	tests := []struct {
		name           string
		isAdmin        bool
		isInstalled    bool
		installSuccess bool
		expectedOutput string
		expectedError  error
	}{
		{
			name:           "Already installed",
			isAdmin:        true,
			isInstalled:    true,
			installSuccess: true,
			expectedOutput: "Homebrew is installed.\n",
			expectedError:  nil,
		},
		{
			name:           "Not admin",
			isAdmin:        false,
			isInstalled:    false,
			installSuccess: false,
			expectedOutput: "You need to be an administrator to install Homebrew. Please run this command from an admin account.\n",
			expectedError:  os.ErrPermission,
		},
		{
			name:           "Install success",
			isAdmin:        true,
			isInstalled:    false,
			installSuccess: true,
			expectedOutput: "Installing homebrew with su current user, enter your password when prompt\n",
			expectedError:  nil,
		},
		{
			name:           "Install failure",
			isAdmin:        true,
			isInstalled:    false,
			installSuccess: false,
			expectedOutput: "Installing homebrew with su current user, enter your password when prompt\nFailed to install Homebrew: exit status 1\n",
			expectedError:  errors.New("exit status 1"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockCmd := &mockCommandContext{}
			oldExecCommandContext := execCommandContext
			execCommandContext = mockCmd.CommandContext
			defer func() { execCommandContext = oldExecCommandContext }()

			mockUtil := &mockUtils{}
			ios, _, out, errOut := iostreams.Test()
			cmd := NewCmdHomeBrew(ios, mockUtil)

			// Mock GetCurrentUser
			mockUtil.On("GetCurrentUser").Return(&user.User{Username: "testuser"}, nil)

			// Mock IsAdmin
			mockUtil.On("IsAdmin", mock.Anything, mock.Anything).Return(tt.isAdmin)

			// Mock IsHomebrewInstalled
			if tt.isInstalled {
				mockCmd.On("CommandContext", mock.Anything, "which", []string{"brew"}).
					Return(exec.Command("echo", "/opt/homebrew/bin/brew"))
			} else {
				mockCmd.On("CommandContext", mock.Anything, "which", []string{"brew"}).
					Return(exec.Command("false"))
			}

			// Mock install command
			if !tt.isInstalled && tt.isAdmin {
				if tt.installSuccess {
					mockCmd.On("CommandContext", mock.Anything, "su", mock.Anything).
						Return(exec.Command("echo", "Homebrew installed successfully"))
				} else {
					failCmd := exec.Command("false")
					mockCmd.On("CommandContext", mock.Anything, "su", mock.Anything).
						Return(failCmd)
				}
			}

			// Execute
			err := cmd.RunE(cmd, []string{})
			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			output := out.String() + errOut.String()
			assert.Contains(t, output, tt.expectedOutput)
			mockCmd.AssertExpectations(t)
			mockUtil.AssertExpectations(t)
		})
	}
}
