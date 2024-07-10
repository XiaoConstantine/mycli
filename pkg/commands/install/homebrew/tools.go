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
			toolDuration := time.Since(toolStartTime)
			toolStat.Status = "success"
			toolStat.Duration = toolDuration
			stats = append(stats, &toolStat)
		}
		toolSpan.SetTag("status", "success")
		toolSpan.Finish()
	}

	// totalDuration := time.Since(startTime)
	// installedTools = append(installedTools, []string{"Total", totalDuration.String()})
	// // Print summary of installed tools in a table
	// table := tablewriter.NewWriter(iostream.Out)
	// table.SetHeader([]string{"Tool", "Duration", "Status"})
	// for _, v := range installedTools {
	// 	table.Append(v)
	// }
	// table.Render()
	fmt.Fprintln(iostream.Out, cs.GreenBold("All requested tools and casks have been installed successfully."))
	return stats, nil
}

func executeCommand(command string, ctx context.Context) error {
	cmd := execCommandContext(ctx, "sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
