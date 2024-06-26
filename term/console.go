package term

import (
	"errors"
	"os"
)

func enableVirtualTerminalProcessing(f *os.File) error {
	return errors.New("not implemented")
}

func openTTY() (*os.File, error) {
	return os.Open("/dev/tty")
}

func hasAlternateScreenBuffer(hasTrueColor bool) bool {
	// on non-Windows, we just assume that alternate screen buffer is supported in most cases
	return os.Getenv("TERM") != "dumb"
}
