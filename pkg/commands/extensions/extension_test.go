package extensions

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/spf13/cobra"
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

func TestExtensionInstallCommand(t *testing.T) {
	// Create a temporary directory to simulate the extensions directory
	tempDir, err := os.MkdirTemp("", "mycli-test-extensions")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Mock the GetExtensionsDir function to return our temp directory
	oldGetExtensionsDir := getExtensionDir
	getExtensionDir = func() string { return tempDir }
	defer func() { getExtensionDir = oldGetExtensionsDir }()

	// Create a mock iostream
	iostreams, _, _, _ := iostreams.Test()
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("Mock repository content")); err != nil {
			return
		}
	}))
	defer server.Close()

	// Mock the exec.Command function to avoid actual git clone
	oldExecCommand := execCommand
	execCommand = func(name string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", name}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}
	defer func() { execCommand = oldExecCommand }()

	// Create the extension install command
	cmd := newExtensionInstallCmd(iostreams)

	testCases := []struct {
		name           string
		repoURL        string
		expectedExtDir string
	}{
		{
			name:           "Normal repository URL",
			repoURL:        server.URL + "/user/test-extension.git",
			expectedExtDir: "mycli-test-extension",
		},
		{
			name:           "Repository URL without .git suffix",
			repoURL:        server.URL + "/user/another-extension",
			expectedExtDir: "mycli-another-extension",
		},
		{
			name:           "Repository URL with mycli- prefix",
			repoURL:        server.URL + "/user/mycli-prefixed-extension.git",
			expectedExtDir: "mycli-prefixed-extension",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// Run the command
			cmd.SetArgs([]string{tc.repoURL})
			err := cmd.Execute()
			assert.NoError(t, err)

			// Check if the extension directory was created correctly
			extDir := filepath.Join(tempDir, tc.expectedExtDir)
			_, err = os.Stat(extDir)
			assert.NoError(t, err, "Extension directory should exist")

			// Check that a directory with .git suffix was not created
			_, err = os.Stat(extDir + ".git")
			assert.True(t, os.IsNotExist(err), "Directory with .git suffix should not exist")
		})
	}
}

func TestRunExtension(t *testing.T) {
	// Create a temporary directory to simulate the extensions directory
	tempDir, err := os.MkdirTemp("", "mycli-test-extensions")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Mock the getExtensionDir function
	oldGetExtensionDir := getExtensionDir
	getExtensionDir = func() string { return tempDir }
	defer func() { getExtensionDir = oldGetExtensionDir }()

	// Create a mock extension
	mockExtName := "mock-extension"
	mockExtDir := filepath.Join(tempDir, "mycli-"+mockExtName)
	mockExtPath := filepath.Join(mockExtDir, "mycli-"+mockExtName)
	err = os.MkdirAll(mockExtDir, 0755)
	assert.NoError(t, err)

	// Create a mock executable
	mockExtContent := []byte("#!/bin/sh\necho \"Mock extension executed with args: $@\"")
	err = os.WriteFile(mockExtPath, mockExtContent, 0755)
	assert.NoError(t, err)

	// Test cases
	testCases := []struct {
		name           string
		extName        string
		args           []string
		expectedError  bool
		expectedOutput string
	}{
		{
			name:           "Existing extension",
			extName:        mockExtName,
			args:           []string{"arg1", "arg2"},
			expectedError:  false,
			expectedOutput: "Mock extension executed with args: arg1 arg2\n",
		},
		{
			name:           "Non-existent extension",
			extName:        "non-existent",
			args:           []string{},
			expectedError:  true,
			expectedOutput: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runExtension(tc.extName, tc.args)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedOutput, output)
			}
		})
	}
}

func TestNewExtensionRunCmd(t *testing.T) {
	// Create a temporary directory to simulate the extensions directory
	tempDir, err := os.MkdirTemp("", "mycli-test-extensions")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Mock the getExtensionDir function
	oldGetExtensionDir := getExtensionDir
	getExtensionDir = func() string { return tempDir }
	defer func() { getExtensionDir = oldGetExtensionDir }()

	// Create a mock extension
	mockExtName := "mock-extension"
	mockExtDir := filepath.Join(tempDir, "mycli-"+mockExtName)
	mockExtPath := filepath.Join(mockExtDir, "mycli-"+mockExtName)
	err = os.MkdirAll(mockExtDir, 0755)
	assert.NoError(t, err)

	// Create a mock executable
	mockExtContent := []byte("#!/bin/sh\necho \"Mock extension executed with args: $@\"")
	err = os.WriteFile(mockExtPath, mockExtContent, 0755)
	assert.NoError(t, err)

	// Create the run command
	runCmd := newExtensionRunCmd()

	testCases := []struct {
		name           string
		args           []string
		expectedError  bool
		expectedOutput string
		expectedErrMsg string
	}{
		{
			name:           "Run existing extension",
			args:           []string{mockExtName, "arg1", "arg2"},
			expectedError:  false,
			expectedOutput: "Mock extension executed with args: arg1 arg2\n",
		},
		{
			name:           "Run non-existent extension",
			args:           []string{"non-existent"},
			expectedError:  true,
			expectedErrMsg: "extension 'non-existent' not found",
		},
		{
			name:           "No arguments",
			args:           []string{},
			expectedError:  true,
			expectedErrMsg: "extension name is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute the command
			cmd := &cobra.Command{}
			cmd.SetArgs(tc.args)
			err := runCmd.RunE(cmd, tc.args)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			if tc.expectedError {
				assert.Error(t, err)
				if tc.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedOutput, output)
			}
		})
	}
}

// TestHelperProcess isn't a real test. It's used to mock exec.Command in the main test.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// Simulate the behavior of git clone by creating a directory
	if os.Args[3] == "git" && os.Args[4] == "clone" {
		err := os.MkdirAll(os.Args[6], 0755)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(1)
}
