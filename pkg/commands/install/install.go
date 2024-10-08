/*
Package install provides functionality for installing software tools and packages using mycli.

This package contains commands and utilities for managing the installation of various
development tools and software packages required for setting up a development environment.
*/
package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/XiaoConstantine/mycli/pkg/commands/install/homebrew"
	"github.com/XiaoConstantine/mycli/pkg/commands/install/xcode"
	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/XiaoConstantine/mycli/pkg/utils"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// InstallCmd creates and returns a cobra.Command for the 'install' subcommand of mycli.
//
// The install command allows users to install various software tools and packages
// either interactively or via command-line arguments. It supports installation from
// different sources including package managers (e.g., Homebrew) and custom scripts.
//
// Usage:
//
//	mycli install [flags]
//	mycli install [tool-name] [flags]
//
// Flags:
//
//	-c, --config string   Path to the configuration file (default "~/.mycli/config.yaml")
//	-f, --force           Force reinstallation of already installed tools
//	--non-interactive     Run in non-interactive mode
//
// The function sets up the command's flags, its Run function, and any subcommands.
// It uses the provided IOStreams for input/output operations.
//
// Parameters:
//   - iostream: An iostreams.IOStreams instance for handling input/output operations.
//
// Returns:
//   - *cobra.Command: A pointer to the created cobra.Command for the install subcommand.
//
// Example:
//
//	// Creating the install command
//	installCmd := install.InstallCmd(iostreams.System())
//	rootCmd.AddCommand(installCmd)
func NewInstallCmd(iostream *iostreams.IOStreams) *cobra.Command {
	cs := iostream.ColorScheme()
	statsCollector := utils.NewStatsCollector()

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install software",
		Long:  `All software installation commands.`,
		Annotations: map[string]string{
			"group": "install",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			span, ctx := tracer.StartSpanFromContext(cmd.Context(), "install")
			defer span.Finish()

			// Ask whether to install everything or specific components
			var installChoice string
			var configPath string
			var force bool
			nonInteractive, _ := cmd.Flags().GetBool("non-interactive")

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
				if force {
					if err := cmd.Flags().Set("force", "true"); err != nil {
						fmt.Fprintf(iostream.ErrOut, "failed to set force flag: %s\n", err)
						return err
					}
				}

				fmt.Fprintln(iostream.Out, cs.GreenBold("Running all installation subcommands..."))
				for _, subcmd := range cmd.Commands() {
					fmt.Printf("Running installation for %s...\n", subcmd.Use)
					if len(subcmd.Use) == 0 {
						continue
					}
					subSpan, subCtx := tracer.StartSpanFromContext(ctx, "install_"+subcmd.Use)
					subcmd.SetContext(subCtx)
					if subcmd.Use == "tools" {
						if err := subcmd.Flags().Set("config", configPath); err != nil {
							fmt.Fprintf(iostream.ErrOut, "failed to set config flag: %s\n", err)
							return err
						}
						if force {
							if err := subcmd.Flags().Set("force", "true"); err != nil {
								fmt.Fprintf(iostream.ErrOut, "failed to set force flag: %s\n", err)
								return err
							}
						}
					}

					if err := subcmd.RunE(subcmd, args); err != nil {
						fmt.Fprintf(iostream.ErrOut, "Error installing %s: %v\n", subcmd.Use, err)

						subSpan.SetTag("status", "failed")
						subSpan.SetTag("error", err)
						subSpan.Finish()
						utils.PrintCombinedStats(iostream, statsCollector.GetStats())

						return err
					}

					subSpan.SetTag("status", "success")
					subSpan.Finish()
				}
				utils.PrintCombinedStats(iostream, statsCollector.GetStats())

				fmt.Fprintln(iostream.Out, cs.GreenBold("All installations completed successfully."))
				return nil

			} else {

				prompt := &survey.Select{
					Message: "What would you like to install?",
					Options: append([]string{"Everything"}, utils.GetSubcommandNames(cmd)...),
				}
				if err := survey.AskOne(prompt, &installChoice); err != nil {
					return os.ErrExist
				}

				if installChoice == "Everything" || installChoice == "tools" {
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

					// Prompt for force flag
					forcePrompt := &survey.Confirm{
						Message: "Do you want to force reinstall of casks?",
						Default: false,
					}
					if err := survey.AskOne(forcePrompt, &force); err != nil {
						return os.ErrExist
					}
				}

				if installChoice == "Everything" {
					// Run all install subcommands
					fmt.Fprintln(iostream.Out, cs.GreenBold("Running all installation subcommands..."))
					for _, subcmd := range cmd.Commands() {
						fmt.Printf("Running installation for %s...\n", subcmd.Use)
						subSpan, subCtx := tracer.StartSpanFromContext(ctx, "install_"+subcmd.Use)
						subcmd.SetContext(subCtx)
						if subcmd.Use == "tools" {
							if err := subcmd.Flags().Set("config", configPath); err != nil {
								fmt.Fprintf(iostream.ErrOut, "failed to set config flag: %s\n", err)
								return err
							}
							if force {
								if err := subcmd.Flags().Set("force", "true"); err != nil {
									fmt.Fprintf(iostream.ErrOut, "failed to set force flag: %s\n", err)
									return err
								}
							}
						}

						if err := subcmd.RunE(subcmd, args); err != nil {
							fmt.Fprintf(iostream.ErrOut, "Error installing %s: %v\n", subcmd.Use, err)

							subSpan.SetTag("status", "failed")
							subSpan.SetTag("error", err)
							subSpan.Finish()
							utils.PrintCombinedStats(iostream, statsCollector.GetStats())

							return err
						}

						subSpan.SetTag("status", "success")
						subSpan.Finish()
					}
					utils.PrintCombinedStats(iostream, statsCollector.GetStats())

					fmt.Fprintln(iostream.Out, cs.GreenBold("All installations completed successfully."))
				} else {
					// Run the specific chosen subcommand
					fmt.Fprintln(iostream.Out, cs.GreenBold("Running installation for: %s..."), installChoice)
					for _, subcmd := range cmd.Commands() {
						if subcmd.Use == installChoice {
							fmt.Printf("Running installation for %s...\n", installChoice)

							subSpan, subCtx := tracer.StartSpanFromContext(ctx, "install_"+subcmd.Use)
							subcmd.SetContext(subCtx)
							args := []string{}
							if installChoice == "tools" {
								args = append(args, "--config", configPath)
								if err := subcmd.Flags().Set("config", configPath); err != nil {
									fmt.Fprintf(iostream.ErrOut, "failed to set config flag: %s\n", err)
									return err
								}
								if force {
									if err := subcmd.Flags().Set("force", "true"); err != nil {
										fmt.Fprintf(iostream.ErrOut, "failed to set force flag: %s\n", err)
										return err
									}
								}
							}
							if err := subcmd.RunE(subcmd, args); err != nil {
								fmt.Fprintf(iostream.ErrOut, "Error installing %s: %v\n", installChoice, err)
								subSpan.SetTag("status", "failed")
								subSpan.SetTag("error", err)
								subSpan.Finish()
								utils.PrintCombinedStats(iostream, statsCollector.GetStats())

								return err
							}
							subSpan.SetTag("status", "success")
							subSpan.Finish()
							utils.PrintCombinedStats(iostream, statsCollector.GetStats())

							break
						}
					}
				}
				return nil
			}
		},
	}
	utils := utils.RealUserUtils{}

	xcodeCmd := xcode.NewCmdXcode(iostream, statsCollector)
	homebrewCmd := homebrew.NewCmdHomeBrew(iostream, utils, statsCollector)
	toolsCmd := homebrew.NewInstallToolsCmd(iostream, statsCollector)

	installCmd.AddCommand(xcodeCmd)
	installCmd.AddCommand(homebrewCmd)
	installCmd.AddCommand(toolsCmd)

	for _, subcmd := range installCmd.Commands() {
		subcmd.Flags().VisitAll(func(f *pflag.Flag) {
			if installCmd.Flags().Lookup(f.Name) == nil {
				installCmd.Flags().AddFlag(f)
			}
		})
	}
	return installCmd
}
