package term

import (
	"os"
	"testing"
)

func TestEnableVirtualTerminalProcessing(t *testing.T) {
	f, err := os.CreateTemp("", "test_file")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	err = enableVirtualTerminalProcessing()
	if err == nil || err.Error() != "not implemented" {
		t.Errorf("enableVirtualTerminalProcessing() error = %v, want 'not implemented'", err)
	}
}

// func TestOpenTTY(t *testing.T) {
// 	f, err := openTTY()
// 	if err != nil {
// 		// Check for specific error conditions
// 		if os.IsNotExist(err) {
// 			t.Skip("Skipping test: /dev/tty does not exist on this system")
// 		} else if err.Error() == "device not configured" {
// 			t.Skip("Skipping test: /dev/tty is not configured on this system")
// 		} else {
// 			t.Errorf("openTTY() unexpected error: %v", err)
// 		}
// 	} else {
// 		defer f.Close()
// 		if f == nil {
// 			t.Error("openTTY() returned nil file without error")
// 		} else {
// 			// Optionally, perform additional checks on the file
// 			info, err := f.Stat()
// 			if err != nil {
// 				t.Errorf("Failed to stat opened TTY: %v", err)
// 			} else if (info.Mode() & os.ModeDevice) == 0 {
// 				t.Error("Opened file is not a device")
// 			}
// 		}
// 	}
// }
