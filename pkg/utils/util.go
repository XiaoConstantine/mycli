package utils

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"gopkg.in/yaml.v2"
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

type ToolsConfig struct {
	Tools []string `yaml:"tools"`
	Casks []string `yaml:"casks"`
}

// LoadToolsConfig loads tool configuration from a YAML file.
func LoadToolsConfig(filename string) (*ToolsConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config ToolsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
