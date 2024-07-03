package utils

import (
	"fmt"
	"math/rand"
	"mycli/pkg/iostreams"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
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

type ToolConfig struct {
	Tools []Tool `yaml:"tools"`
}

type Tool struct {
	Name           string `yaml:"name"`
	Method         string `yaml:"method,omitempty"` // Optional, for specifying 'cask' or other Homebrew methods
	InstallCommand string `yaml:"install_command,omitempty"`
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

	return map[string]string{"sysname": sysname, "release": release}
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
