package extensions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExtensionsDir(t *testing.T) {
	home, err := os.UserHomeDir()
	assert.NoError(t, err)

	expected := filepath.Join(home, ".mycli", "extensions")
	result := GetExtensionsDir()

	assert.Equal(t, expected, result)
}

func TestExtensionExecute(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mycli-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a mock executable file
	mockExecutable := filepath.Join(tempDir, ExtensionPrefix+"mock")
	mockContent := []byte("#!/bin/sh\necho 'Mock executed'")
	err = os.WriteFile(mockExecutable, mockContent, 0755)
	assert.NoError(t, err)

	ext := &Extension{
		Name: "mock",
		Path: mockExecutable,
	}

	// Test execution
	err = ext.Execute([]string{"arg1", "arg2"})
	assert.NoError(t, err)

	// Test execution with non-existent file
	ext.Path = filepath.Join(tempDir, "non-existent")
	err = ext.Execute([]string{})
	assert.Error(t, err)
}

func TestIsExecutable(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mycli-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test cases
	testCases := []struct {
		name     string
		fileName string
		perms    os.FileMode
		expected bool
	}{
		{"Executable file", "exec", 0755, true},
		{"Non-executable file", "non-exec", 0644, false},
		{"Directory", "dir", 0755, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(tempDir, tc.fileName)
			if tc.name == "Directory" {
				err = os.Mkdir(path, tc.perms)
			} else {
				err = os.WriteFile(path, []byte("test content"), tc.perms)
			}
			assert.NoError(t, err)

			result := isExecutable(path)
			assert.Equal(t, tc.expected, result)
		})
	}

	// Test non-existent file
	t.Run("Non-existent file", func(t *testing.T) {
		result := isExecutable(filepath.Join(tempDir, "non-existent"))
		assert.False(t, result)
	})
}

func TestIsExtension(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mycli-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test cases
	testCases := []struct {
		name     string
		fileName string
		setup    func(string) error
		expected bool
	}{
		{
			name:     "Valid extension",
			fileName: ExtensionPrefix + "test",
			setup: func(path string) error {
				return os.WriteFile(path, []byte("test content"), 0755)
			},
			expected: true,
		},
		{
			name:     "Non-executable extension",
			fileName: ExtensionPrefix + "test",
			setup: func(path string) error {
				if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
					return err
				}
				// Explicitly remove execute permissions
				return os.Chmod(path, 0644)
			},
			expected: false,
		},
		{
			name:     "Non-prefix file",
			fileName: "test",
			setup: func(path string) error {
				return os.WriteFile(path, []byte("test content"), 0755)
			},
			expected: false,
		},
		{
			name:     "Directory with prefix",
			fileName: ExtensionPrefix + "dir",
			setup: func(path string) error {
				return os.Mkdir(path, 0755)
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(tempDir, tc.fileName)
			err := tc.setup(path)
			assert.NoError(t, err)

			result := IsExtension(path)
			assert.Equal(t, tc.expected, result)
		})
	}

	// Test non-existent file
	t.Run("Non-existent file", func(t *testing.T) {
		result := IsExtension(filepath.Join(tempDir, "non-existent"))
		assert.False(t, result)
	})
}
