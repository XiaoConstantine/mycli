package utils

import (
	"errors"
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
)

func TestFlagErrorf(t *testing.T) {
	err := FlagErrorf("test error: %d", 42)

	if err == nil {
		t.Fatal("Expected non-nil error")
	}

	flagErr, ok := err.(*FlagError)
	if !ok {
		t.Fatal("Error is not a *FlagError")
	}

	expected := "test error: 42"
	if flagErr.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, flagErr.Error())
	}
}

func TestFlagErrorWrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := FlagErrorWrap(originalErr)

	if err == nil {
		t.Fatal("Expected non-nil error")
	}

	flagErr, ok := err.(*FlagError)
	if !ok {
		t.Fatal("Error is not a *FlagError")
	}

	if flagErr.Error() != originalErr.Error() {
		t.Errorf("Expected error message '%s', got '%s'", originalErr.Error(), flagErr.Error())
	}

	if errors.Unwrap(flagErr) != originalErr {
		t.Error("Unwrapped error does not match original error")
	}
}

func TestIsUserCancellation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"CancelError", CancelError, true},
		{"InterruptErr", terminal.InterruptErr, true},
		{"Other error", errors.New("other error"), false},
		{"Nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsUserCancellation(tt.err)
			if result != tt.expected {
				t.Errorf("IsUserCancellation(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestMutuallyExclusive(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		conditions []bool
		wantErr    bool
	}{
		{"No true conditions", "Error message", []bool{false, false, false}, false},
		{"One true condition", "Error message", []bool{true, false, false}, false},
		{"Multiple true conditions", "Error message", []bool{true, true, false}, true},
		{"All true conditions", "Error message", []bool{true, true, true}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MutuallyExclusive(tt.message, tt.conditions...)
			if (err != nil) != tt.wantErr {
				t.Errorf("MutuallyExclusive() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err.Error() != tt.message {
				t.Errorf("MutuallyExclusive() error message = %v, want %v", err.Error(), tt.message)
			}
		})
	}
}

func TestNoResultsError(t *testing.T) {
	message := "No results found"
	err := NewNoResultsError(message)

	if err.Error() != message {
		t.Errorf("Expected error message '%s', got '%s'", message, err.Error())
	}
}

func TestPredefinedErrors(t *testing.T) {
	errors := []error{SilentError, CancelError, PendingError, ConfigNotFoundError}
	messages := []string{"SilentError", "CancelError", "PendingError", "Config file not found"}

	for i, err := range errors {
		if err.Error() != messages[i] {
			t.Errorf("Expected error message '%s', got '%s'", messages[i], err.Error())
		}
	}
}
