package utils

import (
	"fmt"
	"os/exec"
	"os/user"
	"strings"
)

func GetCurrentUser() (*user.User, error) {
	return user.Current()
}

// isAdmin checks if the given user is a member of the "admin" group.
// It uses the "groups" command to list the groups the user belongs to,
// and returns true if the output contains the "admin" group.
func IsAdmin(u *user.User) bool {
	cmd := exec.Command("groups", u.Username)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error checking groups: %v\n", err)
		return false
	}
	return strings.Contains(string(output), "admin")
}
