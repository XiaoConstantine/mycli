package homebrew

import (
	"context"
	"errors"
	"os/exec"
	"regexp"
	"testing"

	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/XiaoConstantine/mycli/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for execCommandContext.
type mockCommandContext struct {
	mock.Mock
}

func (m *mockCommandContext) CommandContext(ctx context.Context, name string, arg ...string) *exec.Cmd {
	args := m.Called(ctx, name, arg)
	return args.Get(0).(*exec.Cmd)
}

func TestNewInstallToolsCmd(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	statsCollector := utils.NewStatsCollector()
	cmd := NewInstallToolsCmd(ios, statsCollector)

	assert.NotNil(t, cmd)
	assert.Equal(t, "tools", cmd.Use)
	assert.Equal(t, "Install software from a YAML configuration file", cmd.Short)
}

func TestInstallToolsFromConfig(t *testing.T) {
	tests := []struct {
		name             string
		config           *utils.ToolConfig
		force            bool
		executorError    error
		expectedPatterns []*regexp.Regexp
		expectedError    error
		mockSetup        func(*mockCommandContext)
	}{
		{
			name: "Successful installation",
			config: &utils.ToolConfig{
				Tools: []utils.Tool{
					{Name: "tool1"},
					{Name: "tool2", Method: "cask"},
					{Name: "tool3", InstallCommand: "custom install command"},
				},
			},
			force:         false,
			executorError: nil,
			expectedPatterns: []*regexp.Regexp{
				regexp.MustCompile(`Installing tool1 using Homebrew with brew install...`),
				regexp.MustCompile(`Installing tool2 using Homebrew with brew install --cask...`),
				regexp.MustCompile(`Installing tool3 using custom command custom install command...`),
			},
			expectedError: nil,
			mockSetup: func(mockCmd *mockCommandContext) {
				mockCmd.On("CommandContext", mock.Anything, "sh", []string{"-c", "brew install tool1"}).
					Return(exec.Command("echo", "Installation of tool1 succeeded"))
				mockCmd.On("CommandContext", mock.Anything, "sh", []string{"-c", "brew install --cask tool2"}).
					Return(exec.Command("echo", "Installation of tool2 succeeded"))
				mockCmd.On("CommandContext", mock.Anything, "sh", []string{"-c", "custom install command"}).
					Return(exec.Command("echo", "Installation of tool3 succeeded"))
			},
		},
		{
			name: "Failed installation",
			config: &utils.ToolConfig{
				Tools: []utils.Tool{
					{Name: "tool1"},
				},
			},
			force:         false,
			executorError: errors.New("installation failed"),
			expectedPatterns: []*regexp.Regexp{
				regexp.MustCompile(`Installing tool1 using Homebrew with brew install...`),
				regexp.MustCompile(`Failed to install`),
			},
			expectedError: errors.New("exit status 1"),
			mockSetup: func(mockCmd *mockCommandContext) {
				failCmd := exec.Command("false")
				mockCmd.On("CommandContext", mock.Anything, "sh", []string{"-c", "brew install tool1"}).
					Return(failCmd)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ios, _, out, errOut := iostreams.Test()
			mockCmd := &mockCommandContext{}
			oldExecCommandContext := execCommandContext
			execCommandContext = mockCmd.CommandContext
			defer func() { execCommandContext = oldExecCommandContext }()
			tt.mockSetup(mockCmd)

			// Execute
			_, err := InstallToolsFromConfig(ios, tt.config, context.Background(), tt.force)
			// Assert
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			// Assert
			output := out.String() + errOut.String()
			for _, pattern := range tt.expectedPatterns {
				assert.True(t, pattern.MatchString(output), "Expected pattern not found: %s", pattern.String())
			}
			mockCmd.AssertExpectations(t)
		})
	}
}

func TestExecuteCommand(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		expectedErr bool
		mockSetup   func(*mockCommandContext)
	}{
		{
			name:        "Valid command",
			command:     "echo 'Hello, World!'",
			expectedErr: false,
			mockSetup: func(mockCmd *mockCommandContext) {
				successCmd := exec.Command("echo", "Hello, World!")
				mockCmd.On("CommandContext", mock.Anything, "sh", []string{"-c", "echo 'Hello, World!'"}).
					Return(successCmd)
			},
		},
		{
			name:        "Invalid command",
			command:     "invalid_command",
			expectedErr: true,
			mockSetup: func(mockCmd *mockCommandContext) {
				failCmd := exec.Command("false")
				mockCmd.On("CommandContext", mock.Anything, "sh", []string{"-c", "invalid_command"}).
					Return(failCmd)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCmd := &mockCommandContext{}
			oldExecCommandContext := execCommandContext
			execCommandContext = mockCmd.CommandContext
			defer func() { execCommandContext = oldExecCommandContext }()

			tt.mockSetup(mockCmd)

			err := executeCommand(tt.command, context.Background())

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCmd.AssertExpectations(t)
		})
	}
}
