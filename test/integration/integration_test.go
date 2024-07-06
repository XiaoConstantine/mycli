package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	expect "github.com/Netflix/go-expect"
	"github.com/stretchr/testify/require"
)

var cliPath string

func TestMain(m *testing.M) {
	cliPath = filepath.Join("..", "..", "mycli") // Adjust if necessary
	exitCode := m.Run()
	os.Exit(exitCode)
}

func runCLITest(t *testing.T, interactions func(*expect.Console)) {

	c, err := expect.NewConsole(expect.WithStdout(os.Stdout))
	c.Env["TERM"] = "dumb"
	require.NoError(t, err, "Failed to create console")
	defer c.Close()

	cmd := exec.Command(cliPath)
	cmd.Stdin = c.Tty()
	cmd.Stdout = c.Tty()
	cmd.Stderr = c.Tty()

	err = cmd.Start()
	require.NoError(t, err, "Failed to start command")

	interactions(c)

	err = cmd.Wait()
	require.NoError(t, err, "Command failed")
}

func expectPrompt(c *expect.Console, prompt string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var lastOutput string
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for prompt: %s\nLast output: %s", prompt, lastOutput)
		default:
			output, err := c.ExpectString(prompt)
			if err == nil {
				return nil
			}
			lastOutput = output
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func selectOption(c *expect.Console, option string) error {
	_, err := c.Send(option)
	if err != nil {
		return err
	}
	_, err = c.Send("\n")
	return err
}

func confirmYes(c *expect.Console) error {
	_, err := c.Send("y\n")
	if err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)
	return err
}

func confirmNo(c *expect.Console) error {
	_, err := c.Send("n\n")
	return err
}

func TestCLIConfigure(t *testing.T) {
	runCLITest(t, func(c *expect.Console) {
		require.NoError(t, expectPrompt(c, "Choose a command to run"))
		require.NoError(t, selectOption(c, "configure"))
		require.NoError(t, expectPrompt(c, "Do you want to run the 'configure' command? (y/N)"))
		// Confirm yes
		require.NoError(t, confirmYes(c))

		require.NoError(t, expectPrompt(c, "Enter config file path:"))
		_, err := c.Send("test_config.yml\n")
		require.NoError(t, err)

		require.NoError(t, expectPrompt(c, "Configuration completed successfully"))
	})
}

func TestCLIInstallEverything(t *testing.T) {
	runCLITest(t, func(c *expect.Console) {
		require.NoError(t, expectPrompt(c, "Select a command"))
		require.NoError(t, selectOption(c, "install"))

		require.NoError(t, expectPrompt(c, "Select install option"))
		require.NoError(t, selectOption(c, "everything"))

		require.NoError(t, expectPrompt(c, "Enter config file path:"))
		_, err := c.Send("test_config.yml\n")
		require.NoError(t, err)

		require.NoError(t, expectPrompt(c, "Are you sure you want to install everything? (y/N)"))
		require.NoError(t, confirmYes(c))

		require.NoError(t, expectPrompt(c, "Installation completed successfully"))
	})
}

func TestCLIInstallTools(t *testing.T) {
	runCLITest(t, func(c *expect.Console) {
		require.NoError(t, expectPrompt(c, "Select a command"))
		require.NoError(t, selectOption(c, "install"))

		require.NoError(t, expectPrompt(c, "Select install option"))
		require.NoError(t, selectOption(c, "tools"))

		require.NoError(t, expectPrompt(c, "Enter config file path:"))
		_, err := c.Send("test_config.yml\n")
		require.NoError(t, err)

		require.NoError(t, expectPrompt(c, "Installation completed successfully"))
	})
}

func TestCLIInstallXcode(t *testing.T) {
	runCLITest(t, func(c *expect.Console) {
		require.NoError(t, expectPrompt(c, "Select a command"))
		require.NoError(t, selectOption(c, "install"))

		require.NoError(t, expectPrompt(c, "Select install option"))
		require.NoError(t, selectOption(c, "xcode"))

		require.NoError(t, expectPrompt(c, "Xcode installation completed successfully"))
	})
}

func TestCLIInstallBrew(t *testing.T) {
	runCLITest(t, func(c *expect.Console) {
		require.NoError(t, expectPrompt(c, "Select a command"))
		require.NoError(t, selectOption(c, "install"))

		require.NoError(t, expectPrompt(c, "Select install option"))
		require.NoError(t, selectOption(c, "brew"))

		require.NoError(t, expectPrompt(c, "Homebrew installation completed successfully"))
	})
}
