package homebrew

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/XiaoConstantine/mycli/pkg/utils"

	"github.com/spf13/cobra"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// NewInstallToolsCmd creates and returns a cobra.Command for the 'tools' subcommand of the install command.
//
// The tools subcommand allows users to install specific software tools defined in a configuration file.
// It supports both interactive and non-interactive modes, and can install multiple tools at once.
//
// Usage:
//
//	mycli install tools [flags]
//	mycli install tools [tool1] [tool2] ... [flags]
//
// Flags:
//
//	-c, --config string   Path to the configuration file (default "~/.mycli/config.yaml")
//	-f, --force           Force reinstall of tools even if they are already installed
//	--non-interactive     Run in non-interactive mode
//
// The function sets up the command's flags and its Run function. It uses the provided IOStreams
// for input/output operations and a StatsCollector for gathering installation statistics.
//
// Parameters:
//   - iostream: An iostreams.IOStreams instance for handling input/output operations.
//   - statsCollector: A pointer to a StatsCollector for gathering installation statistics.
//
// Returns:
//   - *cobra.Command: A pointer to the created cobra.Command for the tools subcommand.
//
// Example:
//
//	// Creating the tools subcommand
//	toolsCmd := install.toolsCmd(iostreams.System(), &StatsCollector{})
//	installCmd.AddCommand(toolsCmd)
func NewInstallToolsCmd(iostream *iostreams.IOStreams, statsCollector *utils.StatsCollector) *cobra.Command {
	cs := iostream.ColorScheme()
	var configFile string
	var force bool
	var nonInteractive bool
	var toolStats []*utils.Stats

	cmd := &cobra.Command{
		Use:   "tools",
		Short: "Install software from a YAML configuration file",
		Long:  `Reads a list of software tools and casks to install from a YAML file.`,
		Annotations: map[string]string{
			"group": "install",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			span, ctx := tracer.StartSpanFromContext(cmd.Context(), "install_tools")
			defer span.Finish()

			config, err := utils.LoadToolsConfig(configFile)
			if err != nil {
				fmt.Fprintf(iostream.ErrOut, cs.Red("Error loading configuration: %v\n"), err)
				return utils.ConfigNotFoundError
			}
			toolStats, err = InstallToolsFromConfig(iostream, config, ctx, force)
			for _, item := range toolStats {
				statsCollector.AddStat(item)
			}
			return err
		},
	}
	cmd.Flags().StringVarP(&configFile, "config", "c", "config.yaml", "Path to the configuration file")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force reinstall of casks")
	cmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Run in non-interactive mode")
	return cmd
}

// InstallToolsFromConfig installs tools based on the provided configuration.
//
// This function is responsible for the actual installation process of the tools.
// It reads the tool definitions from the config, checks if they need to be installed,
// and executes the installation commands.
//
// Parameters:
//   - iostream: An iostreams.IOStreams instance for I/O operations.
//   - config: A pointer to the ToolConfig containing tool definitions.
//   - ctx: A context.Context for handling cancellation and timeouts.
//   - force: A boolean indicating whether to force reinstallation of tools.
//   - statsCollector: A pointer to a StatsCollector for gathering installation statistics.
//   - requestedTools: A slice of strings containing names of specific tools to install.
//     If empty, all tools in the config will be considered.
//
// Returns:
//   - error: An error if the installation process fails, nil otherwise.
func InstallToolsFromConfig(iostream *iostreams.IOStreams, config *utils.ToolConfig, ctx context.Context, force bool) ([]*utils.Stats, error) {
	cs := iostream.ColorScheme()
	var stats []*utils.Stats

	parentSpan, ctx := tracer.StartSpanFromContext(ctx, "install_tools")
	defer parentSpan.Finish()
	// Log tool installation details
	for _, tool := range config.Tools {
		toolSpan, toolCtx := tracer.StartSpanFromContext(ctx, fmt.Sprintf("install_%s", tool.Name))
		toolStartTime := time.Now()
		toolStat := utils.Stats{
			Name:      tool.Name,
			Operation: "Install",
		}

		fmt.Fprintf(iostream.Out, cs.Green("Installing tool %s...\n"), tool)
		if tool.InstallCommand != "" {
			fmt.Fprintf(iostream.Out, "Installing %s using custom command %s...\n", tool.Name, tool.InstallCommand)
			if err := executeCommand(tool.InstallCommand, toolCtx); err != nil {
				fmt.Fprintf(iostream.ErrOut, cs.Red("Failed to install %s: %v\n"), tool.Name, err)
				toolStat.Status = "error"
				toolStat.Duration = time.Since(toolStartTime)
				stats = append(stats, &toolStat)
				toolSpan.SetTag("status", "failed")
				toolSpan.SetTag("error", err)
				return stats, err
			}
			// Run post-install commands if they exist
			if tool.PostInstall != nil && len(tool.PostInstall) > 0 {
				for _, cmd := range tool.PostInstall {
					expandedCmd := os.ExpandEnv(cmd) // Expand environment variables in the command
					if err := executeCommand(expandedCmd, ctx); err != nil {
						fmt.Fprintf(iostream.ErrOut, "Failed to run post-install command for %s: %v\n", tool.Name, err)
						// Decide whether to continue or return based on the error
					}
				}
			}
			toolDuration := time.Since(toolStartTime)
			toolStat.Status = "success"
			toolStat.Duration = toolDuration
			stats = append(stats, &toolStat)
		} else {
			// Default to Homebrew installation
			command := "brew install"
			if tool.Method == "cask" {
				command += " --cask"
			}
			if force {
				command += " --force"
			}
			fmt.Fprintf(iostream.Out, "Installing %s using Homebrew with %s...\n", tool.Name, command)
			if err := executeCommand(fmt.Sprintf("%s %s", command, tool.Name), toolCtx); err != nil {
				toolStat.Status = "error"
				toolStat.Duration = time.Since(toolStartTime)
				stats = append(stats, &toolStat)

				toolSpan.SetTag("status", "failed")
				toolSpan.SetTag("error", err)
				fmt.Fprintf(iostream.ErrOut, cs.Red("Failed to install %s: %v\n"), tool.Name, err)
				return stats, err
			}

			// Run post-install commands
			// Run post-install commands if they exist
			if tool.PostInstall != nil && len(tool.PostInstall) > 0 {
				for _, cmd := range tool.PostInstall {
					expandedCmd := os.ExpandEnv(cmd) // Expand environment variables in the command
					if err := executeCommand(expandedCmd, ctx); err != nil {
						fmt.Fprintf(iostream.ErrOut, "Failed to run post-install command for %s: %v\n", tool.Name, err)
						// Decide whether to continue or return based on the error
					}
				}
			}
			toolDuration := time.Since(toolStartTime)
			toolStat.Status = "success"
			toolStat.Duration = toolDuration
			stats = append(stats, &toolStat)
		}
		toolSpan.SetTag("status", "success")
		toolSpan.Finish()
	}

	fmt.Fprintln(iostream.Out, cs.GreenBold("All requested tools and casks have been installed successfully."))
	return stats, nil
}

func executeCommand(command string, ctx context.Context) error {
	cmd := execCommandContext(ctx, "sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
