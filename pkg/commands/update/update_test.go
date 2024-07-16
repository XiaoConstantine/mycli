package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/XiaoConstantine/mycli/pkg/build"
	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestCheckForUpdates(t *testing.T) {
	// Save the original function and defer its restoration
	originalGetLatestRelease := getLatestRelease
	defer func() { getLatestRelease = originalGetLatestRelease }()

	testCases := []struct {
		name           string
		currentVersion string
		latestVersion  string
		expectUpdate   bool
		expectError    bool
	}{
		{"Update available", "v1.0.0", "v1.1.0", true, false},
		{"No update needed", "v1.1.0", "v1.1.0", false, false},
		{"Current version newer", "v1.2.0", "v1.1.0", false, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up the mock
			getLatestRelease = func() (*githubRelease, error) {
				return &githubRelease{TagName: tc.latestVersion}, nil
			}

			// Set the current version
			build.Version = tc.currentVersion

			ios, _, _, _ := iostreams.Test()

			hasUpdate, latestVersion, err := CheckForUpdates(ios)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectUpdate, hasUpdate)
				assert.Equal(t, tc.latestVersion, latestVersion)
			}
		})
	}
}

func TestUpdateCLI(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mycli-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Mock the install directory
	originalEnsureInstallDirectory := ensureInstallDirectory
	defer func() { ensureInstallDirectory = originalEnsureInstallDirectory }()
	ensureInstallDirectory = func() (string, error) {
		return tempDir, nil
	}

	// Create a mock binary file
	mockBinaryPath := filepath.Join(tempDir, "mycli")
	err = os.WriteFile(mockBinaryPath, []byte("old binary"), 0755)
	assert.NoError(t, err)

	// Create a mock binary content
	mockBinaryContent := []byte("new binary")

	// Create a buffer to hold the tar.gz content
	var buf bytes.Buffer

	// Create a gzip writer
	gw := gzip.NewWriter(&buf)

	// Create a tar writer
	tw := tar.NewWriter(gw)

	// Add the mock binary to the tar archive
	hdr := &tar.Header{
		Name: "mycli",
		Mode: 0755,
		Size: int64(len(mockBinaryContent)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(mockBinaryContent); err != nil {
		t.Fatal(err)
	}
	// Close the tar writer
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}

	// Close the gzip writer
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}

	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/gzip")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(buf.Bytes()); err != nil {
			return
		}
	}))
	defer server.Close()

	// Mock getLatestRelease function
	originalGetLatestRelease := getLatestRelease
	defer func() { getLatestRelease = originalGetLatestRelease }()
	getLatestRelease = func() (*githubRelease, error) {
		return &githubRelease{
			TagName: "v1.2.3",
			Assets: []githubAsset{
				{
					Name:               fmt.Sprintf("mycli_%s_%s", runtime.GOOS, runtime.GOARCH),
					BrowserDownloadURL: server.URL, // Use the mock server URL

				},
			},
		}, nil
	}

	// Set current version
	build.Version = "v1.0.0"

	// Mock iostreams
	ios, _, out, _ := iostreams.Test()

	// Run updateCLI
	err = updateCLI(ios)
	assert.NoError(t, err)

	// Check if the binary was updated
	updatedBinary, err := os.ReadFile(mockBinaryPath)
	assert.NoError(t, err)
	assert.Equal(t, "new binary", string(updatedBinary))

	// Check the output message
	assert.Contains(t, out.String(), "mycli has been updated successfully to version v1.2.3!")
}

func TestUpdateCLINoUpdateNeeded(t *testing.T) {
	// Mock getLatestRelease function
	originalGetLatestRelease := getLatestRelease
	defer func() { getLatestRelease = originalGetLatestRelease }()
	getLatestRelease = func() (*githubRelease, error) {
		return &githubRelease{
			TagName: "v1.0.0",
		}, nil
	}

	// Set current version
	build.Version = "v1.0.0"

	// Mock iostreams
	ios, _, out, _ := iostreams.Test()

	// Run updateCLI
	err := updateCLI(ios)
	assert.NoError(t, err)

	// Check the output message
	assert.Contains(t, out.String(), "You're already using the latest version of mycli.")
}
