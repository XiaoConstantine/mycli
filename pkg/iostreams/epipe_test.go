package iostreams

import (
	"errors"
	"fmt"
	"syscall"
	"testing"
)

func TestIsEpipeError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "EPIPE error",
			err:      syscall.EPIPE,
			expected: true,
		},
		{
			name:     "Wrapped EPIPE error",
			err:      fmt.Errorf("wrapped error: %w", syscall.EPIPE),
			expected: true,
		},
		{
			name:     "Non-EPIPE error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEpipeError(tt.err)
			if result != tt.expected {
				t.Errorf("isEpipeError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}
