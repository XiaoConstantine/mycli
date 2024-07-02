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

func NewInstallCmd(iostream *iostreams.IOStreams) *cobra.Command {
	cs := iostream.ColorScheme()
	var configFile string

	cmd := &cobra.Command{
		Use:   "tools",
		Short: "Install software from a YAML configuration file",
		Long:  `Reads a list of software tools and casks to install from a YAML file.`,
		Annotations: map[string]string{
			"group": "install",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			span, ctx := tracer.StartSpanFromContext(cmd.Context(), "installhomebrew")
			defer span.Finish()

			config, err := utils.LoadToolsConfig(configFile)
			if err != nil {
				fmt.Fprintf(iostream.ErrOut, cs.Red("Error loading configuration: %v\n"), err)
				return err
			}
			return InstallToolsFromConfig(iostream, config, ctx)
		},
	}
	cmd.Flags().StringVarP(&configFile, "config", "c", "config.yaml", "Path to the configuration file")
	return cmd
}

func InstallToolsFromConfig(iostream *iostreams.IOStreams, config *utils.ToolsConfig, ctx context.Context) error {
	for _, tool := range config.Tools {
		fmt.Printf("Installing tool %s...\n", tool)
		if err := runBrewInstall(iostream, tool, false, ctx); err != nil {
			return err
		}
	}
	for _, cask := range config.Casks {
		fmt.Printf("Installing cask %s...\n", cask)
		if err := runBrewInstall(iostream, cask, true, ctx); err != nil {
			return err
		}
	}
	fmt.Println("All requested tools and casks have been installed successfully.")
	return nil
}

// runBrewInstall runs the brew install command for a given tool or cask.
func runBrewInstall(iostream *iostreams.IOStreams, name string, isCask bool, ctx context.Context) error {
	var cmd *exec.Cmd
	if isCask {
		cmd = exec.CommandContext(ctx, "brew", "install", "--cask", name)
	} else {
		cmd = exec.CommandContext(ctx, "brew", "install", name)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
