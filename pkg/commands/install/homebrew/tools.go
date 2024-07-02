package homebrew

import (
	"context"
	"fmt"
	"mycli/pkg/iostreams"
	"mycli/pkg/utils"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func NewInstallToolsCmd(iostream *iostreams.IOStreams) *cobra.Command {
	cs := iostream.ColorScheme()
	var configFile string
	var force bool

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
				return err
			}
			return InstallToolsFromConfig(iostream, config, ctx, force)
		},
	}
	cmd.Flags().StringVarP(&configFile, "config", "c", "config.yaml", "Path to the configuration file")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force reinstall of casks")
	return cmd
}

func InstallToolsFromConfig(iostream *iostreams.IOStreams, config *utils.ToolConfig, ctx context.Context, force bool) error {
	cs := iostream.ColorScheme()
	parentSpan, ctx := tracer.StartSpanFromContext(ctx, "install_tools")
	defer parentSpan.Finish()
	for _, tool := range config.Tools {
		toolSpan, toolCtx := tracer.StartSpanFromContext(ctx, fmt.Sprintf("install_%s", tool.Name))

		fmt.Fprintf(iostream.Out, cs.Green("Installing tool %s...\n"), tool)
		if tool.InstallCommand != "" {
			fmt.Fprintf(iostream.Out, "Installing %s using custom command %s...\n", tool.Name, tool.InstallCommand)
			if err := executeCommand(tool.InstallCommand, toolCtx); err != nil {
				fmt.Fprintf(iostream.ErrOut, cs.Red("Failed to install %s: %v\n"), tool.Name, err)
				toolSpan.SetTag("status", "failed")
				toolSpan.SetTag("error", err)
				return err
			}
		} else {
			// Default to Homebrew installation
			command := "brew install"
			if tool.Method == "cask" {
				command += " --cask"
			}
			if force {
				command += " --force"
			}
			fmt.Printf("Installing %s using Homebrew with %s...\n", tool.Name, command)
			if err := executeCommand(fmt.Sprintf("%s %s", command, tool.Name), toolCtx); err != nil {
				toolSpan.SetTag("status", "failed")
				toolSpan.SetTag("error", err)
				fmt.Fprintf(iostream.ErrOut, cs.Red("Failed to install %s: %v\n"), tool.Name, err)
				return err
			}
		}
		toolSpan.SetTag("status", "success")
		toolSpan.Finish()
	}
	fmt.Fprintln(iostream.Out, cs.GreenBold("All requested tools and casks have been installed successfully."))
	return nil
}

func executeCommand(command string, ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
