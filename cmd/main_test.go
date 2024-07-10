package main

import (
	"testing"

	"github.com/XiaoConstantine/mycli/pkg/commands/root"
	"github.com/XiaoConstantine/mycli/pkg/commands/update"
	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	// Save the original function and defer its restoration
	originalCheckForUpdates := update.CheckForUpdatesFunc
	defer func() { update.CheckForUpdatesFunc = originalCheckForUpdates }()

	// Mock the CheckForUpdates function
	update.CheckForUpdatesFunc = func(iostream *iostreams.IOStreams) (bool, string, error) {
		return false, "v1.0.0", nil // No update available
	}
	// Define test cases
	tests := []struct {
		name     string
		args     []string
		wantCode root.ExitCode
	}{
		{
			name:     "No arguments",
			args:     []string{"--non-interactive"}, // essentially --non-interactive is equivalent to no args
			wantCode: root.ExitCode(0),              // Successful execution
		},
		{
			name:     "Invalid arguments",
			args:     []string{"--unknown"},
			wantCode: root.ExitCode(1), // Error due to unknown argument
		},
		// Add more tests for other exit codes like exitCancel, exitAuth, etc.
		{
			name:     "Cancel operation",
			args:     []string{"--cancel"},
			wantCode: root.ExitCode(1), // Simulation of a cancellation scenario
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := root.Run(tt.args)
			assert.Equal(t, tt.wantCode, code)
		})
	}
}
