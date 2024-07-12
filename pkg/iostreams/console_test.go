package iostreams

import (
	"os"
	"testing"
)

func TestHasAlternateScreenBuffer(t *testing.T) {
	tests := []struct {
		name     string
		termEnv  string
		expected bool
	}{
		{
			name:     "TERM is dumb",
			termEnv:  "dumb",
			expected: false,
		},
		{
			name:     "TERM is xterm",
			termEnv:  "xterm",
			expected: true,
		},
		{
			name:     "TERM is not set",
			termEnv:  "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save the original TERM environment variable
			origTerm := os.Getenv("TERM")
			defer os.Setenv("TERM", origTerm)

			// Set the TERM environment variable for the test
			os.Setenv("TERM", tt.termEnv)

			// Call the function and check the result
			result := hasAlternateScreenBuffer()
			if result != tt.expected {
				t.Errorf("hasAlternateScreenBuffer() = %v, want %v", result, tt.expected)
			}
		})
	}
}
