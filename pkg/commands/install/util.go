package install

import (
	"fmt"
	"os/exec"
	"os/user"
	"strings"
)

func getCurrentUser() (*user.User, error) {
	return user.Current()
}

func isAdmin(u *user.User) bool {
	cmd := exec.Command("groups", u.Username)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error checking groups: %v\n", err)
		return false
	}
	return strings.Contains(string(output), "admin")
}
