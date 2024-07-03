package install

import (
	"fmt"
	"mycli/pkg/commands/install/homebrew"
	"mycli/pkg/commands/install/xcode"
	"mycli/pkg/iostreams"
	"mycli/pkg/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func NewInstallCmd(iostream *iostreams.IOStreams) *cobra.Command {
	cs := iostream.ColorScheme()
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
			prompt := &survey.Select{
				Message: "What would you like to install?",
				Options: append([]string{"Everything"}, utils.GetSubcommandNames(cmd)...),
			}
			survey.AskOne(prompt, &installChoice)

			if installChoice == "Everything" || installChoice == "tools" {
				// Prompt for config file path
				configPrompt := &survey.Input{
					Message: "Enter the path to the config file:",
					Default: "config.yaml",
				}
				survey.AskOne(configPrompt, &configPath)
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
				survey.AskOne(forcePrompt, &force)
			}

			if installChoice == "Everything" {
				// Run all install subcommands
				fmt.Fprintln(iostream.Out, cs.GreenBold("Running all installation subcommands..."))
				for _, subcmd := range cmd.Commands() {
					fmt.Printf("Running installation for %s...\n", subcmd.Use)
					subSpan, subCtx := tracer.StartSpanFromContext(ctx, "install_"+subcmd.Use)
					subcmd.SetContext(subCtx)
					if subcmd.Use == "tools" {
						subcmd.Flags().Set("config", configPath)
						if force {
							subcmd.Flags().Set("force", "true")
						}
					}

					if err := subcmd.RunE(subcmd, args); err != nil {
						fmt.Fprintf(iostream.ErrOut, "Error installing %s: %v\n", subcmd.Use, err)

						subSpan.SetTag("status", "failed")
						subSpan.SetTag("error", err)
						subSpan.Finish()
						return err
					}

					subSpan.SetTag("status", "success")
					subSpan.Finish()
				}
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
							subcmd.Flags().Set("config", configPath)
							if force {
								subcmd.Flags().Set("force", "true")
							}
						}
						if err := subcmd.RunE(subcmd, args); err != nil {
							fmt.Fprintf(iostream.ErrOut, "Error installing %s: %v\n", installChoice, err)
							subSpan.SetTag("status", "failed")
							subSpan.SetTag("error", err)
							subSpan.Finish()
							return err
						}
						subSpan.SetTag("status", "success")
						subSpan.Finish()
						break
					}
				}
			}
			return nil
		},
	}

	installCmd.AddCommand(xcode.NewCmdXcode(iostream))
	installCmd.AddCommand(homebrew.NewCmdHomeBrew(iostream))
	installCmd.AddCommand(homebrew.NewInstallToolsCmd(iostream))

	for _, subcmd := range installCmd.Commands() {
		subcmd.Flags().VisitAll(func(f *pflag.Flag) {
			if installCmd.Flags().Lookup(f.Name) == nil {
				installCmd.Flags().AddFlag(f)
			}
		})
	}
	return installCmd
}
