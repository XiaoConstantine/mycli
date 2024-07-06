package utils

import (
	"context"
	"fmt"
	"math/rand"
	"mycli/pkg/iostreams"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"gopkg.in/yaml.v2"
)

var execCommandContext = exec.CommandContext

func GetCurrentUser() (*user.User, error) {
	return user.Current()
}

// isAdmin checks if the given user is a member of the "admin" group.
// It uses the "groups" command to list the groups the user belongs to,
// and returns true if the output contains the "admin" group.
func IsAdmin(ctx context.Context, u *user.User) bool {
	cmd := execCommandContext(ctx, "groups", u.Username)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error checking groups: %v\n", err)
		return false
	}
	return strings.Contains(string(output), "admin")
}

type ToolConfig struct {
	Tools     []Tool          `yaml:"tools"`
	Configure []ConfigureItem `yaml:"configure"`
}

type Tool struct {
	Name           string `yaml:"name"`
	Method         string `yaml:"method,omitempty"` // Optional, for specifying 'cask' or other Homebrew methods
	InstallCommand string `yaml:"install_command,omitempty"`
}

type ConfigureItem struct {
	Name        string `yaml:"name"`
	ConfigURL   string `yaml:"config_url"`
	InstallPath string `yaml:"install_path"`
}

// LoadToolsConfig loads tool configuration from a YAML file.
func LoadToolsConfig(filename string) (*ToolConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config ToolConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// GetConfigureItem retrieves a specific configuration item by name.
func (tc *ToolConfig) GetConfigureItem(name string) (*ConfigureItem, error) {
	for _, item := range tc.Configure {
		if item.Name == name {
			return &item, nil
		}
	}
	return nil, fmt.Errorf("configuration for %s not found", name)
}

func GetSubcommandNames(cmd *cobra.Command) []string {
	var names []string
	for _, subcmd := range cmd.Commands() {
		names = append(names, subcmd.Use)
	}
	return names
}

func GetOsInfo() map[string]string {
	var uts unix.Utsname
	if err := unix.Uname(&uts); err != nil {
		panic(err)
	}

	sysname := unix.ByteSliceToString(uts.Sysname[:])
	release := unix.ByteSliceToString(uts.Release[:])
	user, _ := user.Current()

	return map[string]string{"sysname": sysname, "release": release, "user": user.Username}
}

func ConvertToRawGitHubURL(inputURL string) (string, error) {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}

	if parsedURL.Host != "github.com" {
		return inputURL, nil // Not a GitHub URL, return as-is
	}

	pathParts := strings.Split(parsedURL.Path, "/")
	if len(pathParts) < 5 {
		return "", fmt.Errorf("invalid GitHub URL format")
	}

	// Reconstruct the URL for raw content
	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s",
		pathParts[1],                     // username
		pathParts[2],                     // repository
		pathParts[4],                     // branch (usually "master" or "main")
		strings.Join(pathParts[5:], "/")) // file path

	return rawURL, nil
}

func getRandomASCIILogo() string {
	logos := []string{
		`
    __  ___      ________    ____
   /  |/  /_  __/ ____/ /   /  _/
  / /|_/ / / / / /   / /    / /
 / /  / / /_/ / /___/ /____/ /
/_/  /_/\__, /\____/_____/___/
       /____/
`,
		`
 ________            _________    ___
|\\   __  \\         |\\___   ___\\ |\\  \\
\\ \\  \\|\\  \\  ____  \\|___ \\  \\_| \\ \\  \\
 \\ \\   __  \\|\\  __\\     \\ \\  \\   \\ \\  \\
  \\ \\  \\ \\  \\ \\ \\__/__    \\ \\  \\   \\ \\  \\____
   \\ \\__\\ \\__\\ \\______\\    \\ \\__\\   \\ \\_______\\
    \\|__|\\|__|\\|______|     \\|__|    \\|_______|
`,
	}
	return logos[rand.Intn(len(logos))]
}

func PrintWelcomeMessage(iostream *iostreams.IOStreams) {
	cs := iostream.ColorScheme()
	out := iostream.Out

	asciiLogo := getRandomASCIILogo()

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, cs.Blue("Welcome to MyCLI!"))
	fmt.Fprintln(out, cs.Blue(asciiLogo))
	fmt.Fprintln(out, cs.Green("Your personal machine bootstrapping tool"))
	fmt.Fprintln(out, "")
	fmt.Fprintf(out, "Version: %s\n", cs.Yellow("1.0.0"))
	fmt.Fprintf(out, "Current time: %s\n", cs.Yellow(time.Now().Format("2006-01-02 15:04:05")))
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Available commands:")
	fmt.Fprintln(out, "  - install: Install software tools")
	fmt.Fprintln(out, "  - config: Configure your system")
	fmt.Fprintln(out, "  - update: Update MyCLI")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, cs.Yellow("Tip: Use 'mycli --help' to see all available commands and options."))
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Let's get started with setting up your machine...")
	fmt.Fprintln(out, "")
}
