package utils

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/XiaoConstantine/mycli/pkg/iostreams"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"gopkg.in/yaml.v2"
)

var execCommandContext = exec.CommandContext

type UserUtils interface {
	GetCurrentUser() (*user.User, error)
	IsAdmin(ctx context.Context, u *user.User) bool
}

type RealUserUtils struct{}

func (RealUserUtils) GetCurrentUser() (*user.User, error) {
	return user.Current()
}

func (RealUserUtils) IsAdmin(ctx context.Context, u *user.User) bool {
	cmd := execCommandContext(ctx, "groups", u.Username)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error checking groups: %v\n", err)
		return false
	}
	return strings.Contains(string(output), "admin")
}

// func GetCurrentUser() (*user.User, error) {
// 	return user.Current()
// }

// isAdmin checks if the given user is a member of the "admin" group.
// It uses the "groups" command to list the groups the user belongs to,
// and returns true if the output contains the "admin" group.
// func IsAdmin(ctx context.Context, u *user.User) bool {
// 	cmd := execCommandContext(ctx, "groups", u.Username)
// 	output, err := cmd.Output()
// 	if err != nil {
// 		fmt.Printf("Error checking groups: %v\n", err)
// 		return false
// 	}
// 	return strings.Contains(string(output), "admin")
// }

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
	Name             string `yaml:"name"`
	ConfigURL        string `yaml:"config_url,omitempty"`
	InstallPath      string `yaml:"install_path"`
	ConfigureCommand string `yaml:"configure_command,omitempty"`
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
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return "", fmt.Errorf("invalid URL: %v", err)
	}
	if parsedURL.Host != "github.com" {
		return inputURL, nil // Not a GitHub URL, return as-is
	}

	// Preserve trailing slash
	hasTrailingSlash := strings.HasSuffix(parsedURL.Path, "/")

	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 3 {
		return "", fmt.Errorf("invalid GitHub URL format")
	}

	username := pathParts[0]
	repo := pathParts[1]
	branch := "main"
	var filePath []string

	if len(pathParts) > 3 && (pathParts[2] == "blob" || pathParts[2] == "tree") {
		branch = pathParts[3]
		filePath = pathParts[4:]
	} else {
		filePath = pathParts[2:]
	}

	// Reconstruct the URL for raw content
	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s",
		username,
		repo,
		branch,
		strings.Join(filePath, "/"))

	// Add trailing slash if original URL had one
	if hasTrailingSlash && len(filePath) > 0 {
		rawURL += "/"
	}

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

func PrintWelcomeMessage(iostream *iostreams.IOStreams, version string) {
	cs := iostream.ColorScheme()
	out := iostream.Out

	asciiLogo := getRandomASCIILogo()

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, cs.Blue("Welcome to MyCLI!"))
	fmt.Fprintln(out, cs.Blue(asciiLogo))
	fmt.Fprintln(out, cs.Green("Your personal machine bootstrapping tool"))
	fmt.Fprintln(out, "")
	fmt.Fprintf(out, "Version: %s\n", cs.Yellow(version))
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

// CompareVersions compares two version strings.
// It returns:
//
//	-1 if v1 < v2
//	 0 if v1 == v2
//	 1 if v1 > v2
func CompareVersions(v1, v2 string) int {
	v1Parts := strings.Split(strings.TrimPrefix(v1, "v"), ".")
	v2Parts := strings.Split(strings.TrimPrefix(v2, "v"), ".")

	for i := 0; i < len(v1Parts) && i < len(v2Parts); i++ {
		v1Part := strings.Split(v1Parts[i], "-")
		v2Part := strings.Split(v2Parts[i], "-")

		n1, err1 := strconv.Atoi(v1Part[0])
		n2, err2 := strconv.Atoi(v2Part[0])

		if err1 != nil || err2 != nil {
			// If we can't convert to integers, compare as strings
			if v1Part[0] < v2Part[0] {
				return -1
			} else if v1Part[0] > v2Part[0] {
				return 1
			}
		} else {
			if n1 < n2 {
				return -1
			} else if n1 > n2 {
				return 1
			}
		}

		// If the numeric parts are equal, but one has a pre-release version, it's considered lower
		if len(v1Part) == 1 && len(v2Part) > 1 {
			return 1
		} else if len(v1Part) > 1 && len(v2Part) == 1 {
			return -1
		} else if len(v1Part) > 1 && len(v2Part) > 1 {
			// If both have pre-release versions, compare them
			if v1Part[1] < v2Part[1] {
				return -1
			} else if v1Part[1] > v2Part[1] {
				return 1
			}
		}
	}

	// If we've gotten this far and haven't returned, check for different lengths
	if len(v1Parts) < len(v2Parts) {
		return -1
	} else if len(v1Parts) > len(v2Parts) {
		return 1
	}

	// Versions are equal
	return 0
}
