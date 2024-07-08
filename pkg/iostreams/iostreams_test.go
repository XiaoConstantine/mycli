package iostreams

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStopAlternateScreenBuffer(t *testing.T) {
	ios, _, stdout, _ := Test()
	ios.SetAlternateScreenBufferEnabled(true)

	ios.StartAlternateScreenBuffer()
	fmt.Fprint(ios.Out, "test")
	ios.StopAlternateScreenBuffer()

	// Stopping a subsequent time should no-op.
	ios.StopAlternateScreenBuffer()

	const want = "\x1b[?1049htest\x1b[?1049l"
	if got := stdout.String(); got != want {
		t.Errorf("after IOStreams.StopAlternateScreenBuffer() got %q, want %q", got, want)
	}
}

func TestIOStreams_pager(t *testing.T) {
	t.Skip("TODO: fix this test in race detection mode")
	ios, _, stdout, _ := Test()
	ios.SetStdoutTTY(true)
	ios.SetPager(fmt.Sprintf("%s -test.run=TestHelperProcess --", os.Args[0]))
	t.Setenv("GH_WANT_HELPER_PROCESS", "1")
	if err := ios.StartPager(); err != nil {
		t.Fatal(err)
	}
	if _, err := fmt.Fprintln(ios.Out, "line1"); err != nil {
		t.Errorf("error writing line 1: %v", err)
	}
	if _, err := fmt.Fprintln(ios.Out, "line2"); err != nil {
		t.Errorf("error writing line 2: %v", err)
	}
	ios.StopPager()
	wants := "pager: line1\npager: line2\n"
	if got := stdout.String(); got != wants {
		t.Errorf("expected %q, got %q", wants, got)
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GH_WANT_HELPER_PROCESS") != "1" {
		return
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Printf("pager: %s\n", scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading stdin: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func TestColorEnabled(t *testing.T) {
	ios, _, _, _ := Test()
	assert.False(t, ios.ColorEnabled())

	ios.SetColorEnabled(true)
	assert.True(t, ios.ColorEnabled())
}

func TestColorSupport256(t *testing.T) {
	ios, _, _, _ := Test()
	assert.False(t, ios.ColorSupport256())

	ios.SetColorEnabled(true)
	assert.True(t, ios.ColorSupport256())
}

func TestHasTrueColor(t *testing.T) {
	ios, _, _, _ := Test()
	assert.False(t, ios.HasTrueColor())

	ios.SetColorEnabled(true)
	assert.True(t, ios.HasTrueColor())
}

func TestDetectTerminalTheme(t *testing.T) {
	ios, _, _, _ := Test()
	ios.DetectTerminalTheme()
	assert.Equal(t, "none", ios.TerminalTheme())
}

func TestIsStdinTTY(t *testing.T) {
	ios, _, _, _ := Test()
	assert.False(t, ios.IsStdinTTY())

	ios.SetStdinTTY(true)
	assert.True(t, ios.IsStdinTTY())
}

func TestIsStdoutTTY(t *testing.T) {
	ios, _, _, _ := Test()
	assert.False(t, ios.IsStdoutTTY())

	ios.SetStdoutTTY(true)
	assert.True(t, ios.IsStdoutTTY())
}

func TestIsStderrTTY(t *testing.T) {
	ios, _, _, _ := Test()
	assert.False(t, ios.IsStderrTTY())

	ios.SetStderrTTY(true)
	assert.True(t, ios.IsStderrTTY())
}

func TestSetPager(t *testing.T) {
	ios, _, _, _ := Test()
	pager := "less"
	ios.SetPager(pager)
	assert.Equal(t, pager, ios.GetPager())
}

func TestCanPrompt(t *testing.T) {
	ios, _, _, _ := Test()
	assert.False(t, ios.CanPrompt())

	ios.SetStdinTTY(true)
	ios.SetStdoutTTY(true)
	assert.True(t, ios.CanPrompt())

	ios.SetNeverPrompt(true)
	assert.False(t, ios.CanPrompt())
}

func TestStartStopProgressIndicator(t *testing.T) {
	ios, _, _, _ := Test()
	ios.progressIndicatorEnabled = true

	ios.StartProgressIndicator()
	assert.NotNil(t, ios.progressIndicator)

	ios.StopProgressIndicator()
	assert.Nil(t, ios.progressIndicator)
}

func TestRunWithProgress(t *testing.T) {
	ios, _, _, _ := Test()
	ios.progressIndicatorEnabled = true

	err := ios.RunWithProgress("Test", func() error {
		return nil
	})
	assert.NoError(t, err)
}

func TestStartStopAlternateScreenBuffer(t *testing.T) {
	ios, _, stdout, _ := Test()
	ios.SetAlternateScreenBufferEnabled(true)

	ios.StartAlternateScreenBuffer()
	ios.StopAlternateScreenBuffer()

	expected := "\x1b[?1049h\x1b[?1049l"
	assert.Equal(t, expected, stdout.String())
}

func TestRefreshScreen(t *testing.T) {
	ios, _, stdout, _ := Test()
	ios.SetStdoutTTY(true)

	ios.RefreshScreen()

	expected := "\x1b[0;0H\x1b[J"
	assert.Equal(t, expected, stdout.String())
}

func TestTerminalWidth(t *testing.T) {
	ios, _, _, _ := Test()
	assert.Equal(t, DefaultWidth, ios.TerminalWidth())
}

func TestColorScheme(t *testing.T) {
	ios, _, _, _ := Test()
	scheme := ios.ColorScheme()
	assert.NotNil(t, scheme)
}

func TestReadUserFile(t *testing.T) {
	ios, in, _, _ := Test()

	// Test reading from stdin
	in.WriteString("test input")
	content, err := ios.ReadUserFile("-")
	assert.NoError(t, err)
	assert.Equal(t, []byte("test input"), content)

	// Test reading from file
	tmpfile, err := os.CreateTemp("", "test")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte("file content"))
	assert.NoError(t, err)
	tmpfile.Close()

	content, err = ios.ReadUserFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, []byte("file content"), content)
}

func TestTempFile(t *testing.T) {
	ios, _, _, _ := Test()

	file, err := ios.TempFile("", "test")
	assert.NoError(t, err)
	assert.NotNil(t, file)
	defer os.Remove(file.Name())

	// Test with TempFileOverride
	override, err := os.CreateTemp("", "override")
	assert.NoError(t, err)
	defer os.Remove(override.Name())

	ios.TempFileOverride = override
	file, err = ios.TempFile("", "test")
	assert.NoError(t, err)
	assert.Equal(t, override, file)
}

func TestSystem(t *testing.T) {
	ios := System()
	assert.NotNil(t, ios)
	assert.NotNil(t, ios.In)
	assert.NotNil(t, ios.Out)
	assert.NotNil(t, ios.ErrOut)
}
