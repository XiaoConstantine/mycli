package llm

import (
	"fmt"
	"os"

	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/spf13/cobra"
)

func NewLogsCmd(iostream *iostreams.IOStreams, options *commandOptions) *cobra.Command {
	var follow bool
	var lines int

	cmd := &cobra.Command{
		Use:   "logs [profile]",
		Short: "Show logs for a local LLM runtime",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := loadLLMConfig(options.configFile)
			if err != nil {
				return err
			}

			profileName, _ := resolveProfile(cfg, "")
			if len(args) == 1 {
				profileName = args[0]
			}

			paths := resolveInstancePaths(cfg, profileName)
			state, err := loadState(paths.StatePath)
			if err != nil {
				return err
			}

			if _, err := os.Stat(paths.LogPath); err == nil {
				if follow {
					cmd := execCommandContext(ctx, "tail", "-f", paths.LogPath)
					cmd.Stdout = iostream.Out
					cmd.Stderr = iostream.ErrOut
					return cmd.Run()
				}

				tail, err := readLastLines(paths.LogPath, lines)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintln(iostream.Out, tail)
				return nil
			}

			if state != nil && state.TMuxWindow != "" {
				captured, err := captureTmuxPane(ctx, profileName)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintln(iostream.Out, captured)
				return nil
			}

			return fmt.Errorf("no log output found for profile %q", profileName)
		},
	}

	cmd.Flags().BoolVar(&follow, "follow", false, "Follow log output")
	cmd.Flags().IntVar(&lines, "lines", 40, "Number of log lines to show")

	return cmd
}
