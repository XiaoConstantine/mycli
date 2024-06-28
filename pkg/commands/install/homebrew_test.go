package install

import (
	"mycli/pkg/iostreams"
	"os/exec"
	"os/user"
	"testing"
)

func TestNewCmdHomeBrew(t *testing.T) {
	tests := []struct {
		name           string
		isAdmin        bool
		installSuccess bool
		wantError      bool
	}{
		{
			name:           "non-admin user",
			isAdmin:        false,
			installSuccess: true,
			wantError:      true,
		},
		{
			name:           "admin user, install success",
			isAdmin:        true,
			installSuccess: true,
			wantError:      false,
		},
		{
			name:           "admin user, install failure",
			isAdmin:        true,
			installSuccess: false,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			io, _, outBuf, errBuf := iostreams.Test()
			cs := io.ColorScheme()

			cmd := NewCmdHomeBrew(io)
			cmd.SetArgs([]string{})

			currentUser := &user.User{
				Username: "testuser",
				Gid:      "80",
			}
			if !tt.isAdmin {
				currentUser.Gid = "1000"
			}
			getCurrentUserMock = func() (*user.User, error) {
				return currentUser, nil
			}

			execCommandMock = func(command string, args ...string) *exec.Cmd {
				if tt.installSuccess {
					return exec.Command("true")
				}
				return exec.Command("false")
			}

			err := cmd.Execute()
			if (err != nil) != tt.wantError {
				t.Errorf("NewCmdHomeBrew() error = %v, wantError %v", err, tt.wantError)
			}

			if tt.isAdmin && tt.installSuccess {
				expectedOutput := cs.Green("Installing homebrew with su current user, enter your password when prompt\n")
				if outBuf.String() != expectedOutput {
					t.Errorf("unexpected output: %s", outBuf.String())
				}
			} else if !tt.isAdmin {
				expectedOutput := cs.Red("You need to be an administrator to install Homebrew. Please run this command from an admin account.\n")
				if errBuf.String() != expectedOutput {
					t.Errorf("unexpected output: %s", errBuf.String())
				}
			}
		})
	}
}

var (
	getCurrentUserMock func() (*user.User, error)
	execCommandMock    func(string, ...string) *exec.Cmd
)
