package configure

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/XiaoConstantinue/mycli/pkg/utils"

	"github.com/XiaoConstantinue/mycli/pkg/iostreams"

	"github.com/AlecAivazis/survey/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func NewConfigureCmd(iostream *iostreams.IOStreams) *cobra.Command {
	cs := iostream.ColorScheme()
	var configFile string
	var force bool

	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure tools from a YAML configuration file",
		Long:  `Reads a list of tools to configure from a YAML file and applies their configurations.`,
		Annotations: map[string]string{
			"group": "configure",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			span, ctx := tracer.StartSpanFromContext(cmd.Context(), "configure_tools")
			defer span.Finish()
			nonInteractive, _ := cmd.Flags().GetBool("non-interactive")

			var configPath string
			var force bool

			if nonInteractive {
				configPath, _ = cmd.Flags().GetString("config")
				force, _ = cmd.Flags().GetBool("force")

				configPath = os.ExpandEnv(configPath)
				// Replace ~ with home directory
				if strings.HasPrefix(configPath, "~") {
					home, err := os.UserHomeDir()
					if err == nil {
						configPath = filepath.Join(home, configPath[1:])
					}
				}

				// Get the absolute path
				absPath, err := filepath.Abs(configPath)
				if err == nil {
					configPath = absPath
				}
				fmt.Println(configPath)
				// Validate the file path
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					fmt.Fprintf(iostream.ErrOut, "Error: Config file does not exist at path: %s\n", configPath)
					return err
				}

				config, err := utils.LoadToolsConfig(configFile)
				if err != nil {
					return err
				}
				if force {
					if err := cmd.Flags().Set("force", "true"); err != nil {
						fmt.Fprintf(iostream.ErrOut, "failed to set force flag: %s\n", err)
						return err
					}

				}

				return ConfigureToolsFromConfig(iostream, config, ctx, force)
			} else {
				// Prompt for config file path
				configPrompt := &survey.Input{
					Message: "Enter the path to the config file:",
					Default: "config.yaml",
				}
				if err := survey.AskOne(configPrompt, &configPath); err != nil {
					return os.ErrExist
				}
				configPath = os.ExpandEnv(configPath)
				// Replace ~ with home directory
				if strings.HasPrefix(configPath, "~") {
					home, err := os.UserHomeDir()
					if err == nil {
						configPath = filepath.Join(home, configPath[1:])
					}
				}

				// Get the absolute path
				absPath, err := filepath.Abs(configPath)
				if err == nil {
					configPath = absPath
				}
				fmt.Println(configPath)
				// Validate the file path
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					fmt.Fprintf(iostream.ErrOut, "Error: Config file does not exist at path: %s\n", configPath)
					return err
				}
				if err := cmd.Flags().Set("config", configPath); err != nil {
					fmt.Fprintf(iostream.ErrOut, "failed to set config flag: %s\n", err)
					return err
				}
				// Prompt for force flag
				forcePrompt := &survey.Confirm{
					Message: "Do you want to force overwrite existing configs?",
					Default: false,
				}
				if err := survey.AskOne(forcePrompt, &force); err != nil {
					return os.ErrExist
				}
				config, err := utils.LoadToolsConfig(configFile)
				if err != nil {
					fmt.Fprintf(iostream.ErrOut, cs.Red("Error loading configuration: %v\n"), err)
					return utils.ConfigNotFoundError
				}

				if force {
					if err := cmd.Flags().Set("force", "true"); err != nil {
						fmt.Fprintf(iostream.ErrOut, "failed to set force flag: %s\n", err)
						return err
					}
				}
				return ConfigureToolsFromConfig(iostream, config, ctx, force)
			}
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "config.yaml", "Path to the configuration file")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force reconfiguration of tools")

	return cmd
}

func ConfigureToolsFromConfig(iostream *iostreams.IOStreams, config *utils.ToolConfig, ctx context.Context, force bool) error {
	cs := iostream.ColorScheme()
	parentSpan, ctx := tracer.StartSpanFromContext(ctx, "configure_tools")
	defer parentSpan.Finish()

	var configuredTools [][]string
	startTime := time.Now()

	for _, item := range config.Configure {
		toolSpan, toolCtx := tracer.StartSpanFromContext(ctx, fmt.Sprintf("configure_%s", item.Name))
		toolStartTime := time.Now()

		fmt.Fprintf(iostream.Out, cs.Green("Configuring %s...\n"), item.Name)

		if err := configureTool(item, toolCtx, force); err != nil {
			fmt.Fprintf(iostream.ErrOut, cs.Red("Failed to configure %s: %v\n"), item.Name, err)
			toolSpan.SetTag("status", "failed")
			toolSpan.SetTag("error", err)
			return err
		}

		toolDuration := time.Since(toolStartTime)
		configuredTools = append(configuredTools, []string{item.Name, toolDuration.String(), "succeed"})

		toolSpan.SetTag("status", "success")
		toolSpan.Finish()
	}

	totalDuration := time.Since(startTime)
	configuredTools = append(configuredTools, []string{"Total", totalDuration.String()})

	// Print summary of configured tools in a table
	table := tablewriter.NewWriter(iostream.Out)
	table.SetHeader([]string{"Tool", "Duration", "Status"})
	for _, v := range configuredTools {
		table.Append(v)
	}
	table.Render()

	fmt.Fprintln(iostream.Out, cs.GreenBold("All requested tools have been configured successfully."))
	return nil
}

func configureTool(item utils.ConfigureItem, ctx context.Context, force bool) error {
	span, _ := tracer.StartSpanFromContext(ctx, "configure_tool")
	defer span.Finish()

	installPath := expandTilde(item.InstallPath)

	// Check if file already exists and force flag is not set
	if _, err := os.Stat(installPath); err == nil && !force {
		return fmt.Errorf("configuration file already exists at %s. Use --force to overwrite", installPath)
	}

	// Create the directory if it doesn't exist
	err := os.MkdirAll(filepath.Dir(installPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	// Convert URL if it's a GitHub URL
	convertedURL, err := utils.ConvertToRawGitHubURL(item.ConfigURL)
	if err != nil {
		return fmt.Errorf("error converting URL: %v", err)
	}
	item.ConfigURL = convertedURL
	// Validate the URL
	parsedURL, err := url.Parse(item.ConfigURL)
	fmt.Println(parsedURL)
	if err != nil {
		return fmt.Errorf("invalid configuration URL: %v", err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL scheme is missing. Please provide a complete URL including http:// or https://")
	}

	// Download the configuration file
	resp, err := http.Get(item.ConfigURL)
	if err != nil {
		return fmt.Errorf("failed to download configuration: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download configuration: HTTP status %d", resp.StatusCode)
	}

	// Create the configuration file
	out, err := os.Create(installPath)
	if err != nil {
		return fmt.Errorf("failed to create configuration file: %v", err)
	}
	defer out.Close()

	// Write the content to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write configuration: %v", err)
	}

	return nil
}

func expandTilde(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}
	if len(path) > 1 && path[1] != '/' {
		// If it's something like ~user/test, don't expand it
		return path
	}
	home := os.Getenv("HOME") // Use environment variable which is controlled in tests
	fmt.Println(home)

	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return path
		}
	}
	return filepath.Join(home, path[1:])
}
