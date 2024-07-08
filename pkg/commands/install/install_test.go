package install

import (
	"testing"

	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestNewInstallCmd(t *testing.T) {
	t.Run("Basic command creation", func(t *testing.T) {
		ios, _, _, _ := iostreams.Test()
		cmd := NewInstallCmd(ios)
		assert.NotNil(t, cmd)
		assert.Equal(t, "install", cmd.Use)
		assert.Equal(t, "Install software", cmd.Short)
	})

	// 	t.Run("Non-interactive mode with valid config", func(t *testing.T) {
	// 		ios, _, stdout, stderr := iostreams.Test()
	// 		cmd := NewInstallCmd(ios)

	// 		// Create a temporary directory
	// 		tempDir, err := os.MkdirTemp("", "test-config")
	// 		assert.NoError(t, err)
	// 		defer os.RemoveAll(tempDir) // Clean up

	// 		// Create a temporary config file
	// 		configContent := []byte(`
	// tools:
	//   - name: example-tool
	// `)
	// 		configPath := filepath.Join(tempDir, "config.yaml")
	// 		err = os.WriteFile(configPath, configContent, 0644)
	// 		assert.NoError(t, err)

	// 		// Set flags
	// 		cmd.Flags().Set("non-interactive", "true")
	// 		cmd.Flags().Set("config", configPath)
	// 		cmd.Flags().Set("force", "true")

	// 		// Execute the command
	// 		err = cmd.Execute()

	// 		assert.NoError(t, err)
	// 		assert.Contains(t, stdout.String(), "Running all installation subcommands...")
	// 		assert.NotContains(t, stderr.String(), "Error:")
	// 	})

	// t.Run("Non-interactive mode with invalid config", func(t *testing.T) {
	// 	ios, _, _, stderr := iostreams.Test()
	// 	cmd := NewInstallCmd(ios)
	// 	cmd.SetArgs([]string{"--non-interactive", "--config", "testdata/nonexistent_config.yaml"})
	// 	err := cmd.Execute()
	// 	assert.Error(t, err)
	// 	assert.Contains(t, stderr.String(), "Error: Config file does not exist at path:")
	// })

	// t.Run("Interactive mode with Everything choice", func(t *testing.T) {
	// 	ios, stdin, stdout, _ := iostreams.Test()
	// 	cmd := NewInstallCmd(ios)
	// 	stdin.WriteString("Everything\n")
	// 	stdin.WriteString("testdata/valid_config.yaml\n")
	// 	stdin.WriteString("n\n") // No force reinstall
	// 	err := cmd.Execute()
	// 	assert.NoError(t, err)
	// 	assert.Contains(t, stdout.String(), "Running all installation subcommands...")
	// })

	// t.Run("Interactive mode with specific subcommand choice", func(t *testing.T) {
	// 	ios, stdin, stdout, _ := iostreams.Test()
	// 	cmd := NewInstallCmd(ios)
	// 	stdin.WriteString("tools\n")
	// 	stdin.WriteString("testdata/valid_config.yaml\n")
	// 	stdin.WriteString("n\n") // No force reinstall
	// 	err := cmd.Execute()
	// 	assert.NoError(t, err)
	// 	assert.Contains(t, stdout.String(), "Running installation for: tools...")
	// })
}
