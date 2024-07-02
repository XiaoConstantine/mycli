package install

import (
	"errors"
	"fmt"
	"mycli/pkg/commands/install/homebrew"
	"mycli/pkg/commands/install/xcode"
	"mycli/pkg/iostreams"

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
			if len(args) == 0 { // No specific subcommand was provided
				fmt.Fprintln(iostream.Out, cs.GreenBold("Running all installation subcommands..."))
				for _, subcmd := range cmd.Commands() {
					subSpan, subCtx := tracer.StartSpanFromContext(ctx, "install_"+subcmd.Use)

					subcmd.SetContext(subCtx)
					fmt.Fprintf(iostream.Out, cs.Gray("Installing %s...\n"), subcmd.Use)
					if err := subcmd.RunE(subcmd, nil); err != nil {
						subSpan.SetTag("status", "failed")
						subSpan.SetTag("error", err)
						subSpan.Finish()
						fmt.Fprintf(iostream.ErrOut, cs.Red("Failed to install %s: %v\n"), subcmd.Use, err)
						return err // or continue based on your policy
					}
					subSpan.SetTag("status", "success")
					subSpan.Finish()
				}
				fmt.Fprintln(iostream.Out, cs.GreenBold("All installations completed successfully."))
				return nil
			}

			// If arguments are provided, let Cobra handle the command execution.
			return errors.New("subcommand required")
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
