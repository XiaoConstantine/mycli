package xcode

import (
	"context"
	"errors"
	"os/exec"
	"testing"

	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/XiaoConstantine/mycli/pkg/utils"
)

func TestNewCmdXcode(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	statsCollector := utils.NewStatsCollector()

	cmd := NewCmdXcode(ios, statsCollector)

	if cmd.Use != "xcode" {
		t.Errorf("Expected Use to be 'xcode', got %s", cmd.Use)
	}

	if !cmd.Hidden {
		t.Error("Expected command to be hidden")
	}

	if cmd.Annotations["group"] != "install" {
		t.Errorf("Expected group annotation to be 'install', got %s", cmd.Annotations["group"])
	}

	if !cmd.SilenceErrors {
		t.Error("Expected SilenceErrors to be true")
	}
}

func TestIsXcodeAlreadyInstalled(t *testing.T) {
	// Save current execCommandContext and defer its restoration
	oldExecCommandContext := execCommandContext
	defer func() { execCommandContext = oldExecCommandContext }()

	tests := []struct {
		name           string
		mockOutput     string
		mockError      error
		expectedResult bool
	}{
		{
			name:           "Xcode installed",
			mockOutput:     "/Applications/Xcode.app/Contents/Developer",
			mockError:      nil,
			expectedResult: true,
		},
		{
			name:           "CommandLineTools installed",
			mockOutput:     "/Library/Developer/CommandLineTools",
			mockError:      nil,
			expectedResult: true,
		},
		{
			name:           "Xcode not installed",
			mockOutput:     "",
			mockError:      &exec.ExitError{},
			expectedResult: false,
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
			result := isXcodeAlreadyInstalled(context.Background())

			if result != tt.expectedResult {
				t.Errorf("Expected %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestRunE(t *testing.T) {
	oldExecCommandContext := execCommandContext
	defer func() { execCommandContext = oldExecCommandContext }()

	tests := []struct {
		name           string
		mockInstalled  bool
		mockInstallErr error
		expectError    bool
	}{
		{
			name:           "Xcode already installed",
			mockInstalled:  true,
			mockInstallErr: nil,
			expectError:    false,
		},
		{
			name:           "Xcode not installed, installation succeeds",
			mockInstalled:  false,
			mockInstallErr: nil,
			expectError:    false,
		},
		{
			name:           "Xcode not installed, installation fails",
			mockInstalled:  false,
			mockInstallErr: errors.New("installation failed"),
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execCommandContext = func(ctx context.Context, command string, args ...string) *exec.Cmd {
				if command == "xcode-select" && args[0] == "-p" {
					if tt.mockInstalled {
						return exec.Command("echo", "/Applications/Xcode.app/Contents/Developer")
					}
					return exec.Command("false")
				}
				if command == "xcode-select" && args[0] == "--install" {
					if tt.mockInstallErr != nil {
						return exec.Command("false")
					}
					return exec.Command("true")
				}
				return exec.Command("echo", "unexpected command")
			}

			ios, _, _, _ := iostreams.Test()
			statsCollector := utils.NewStatsCollector()

			cmd := NewCmdXcode(ios, statsCollector)
			err := cmd.RunE(cmd, []string{})

			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}
		})
	}
}
