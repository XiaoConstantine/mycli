package llm

import (
	"context"
	"fmt"

	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/spf13/cobra"
)

func NewStatusCmd(iostream *iostreams.IOStreams, options *commandOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [profile]",
		Short: "Show the status of a local LLM runtime",
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
			if state == nil {
				_, _ = fmt.Fprintf(iostream.Out, "No llm instance is tracked for profile %q.\n", profileName)
				return nil
			}

			healthy, err := instanceHealthy(ctx, *state)
			if err != nil {
				return err
			}
			if !healthy {
				_ = cleanupStaleState(ctx, paths, *state)
				_, _ = fmt.Fprintf(iostream.Out, "The llm instance for profile %q is stale and has been cleaned up.\n", profileName)
				return nil
			}

			_, _ = fmt.Fprintf(iostream.Out, "Profile: %s\n", state.Profile)
			_, _ = fmt.Fprintf(iostream.Out, "Runtime: %s\n", state.Runtime)
			_, _ = fmt.Fprintf(iostream.Out, "Endpoint: %s\n", formatBaseURL(state.Host, state.Port))
			_, _ = fmt.Fprintf(iostream.Out, "Health: %s\n", state.HealthURL)
			_, _ = fmt.Fprintf(iostream.Out, "Status: %s\n", state.Status)
			if state.External {
				_, _ = fmt.Fprintln(iostream.Out, "Ownership: external")
			} else if state.TMuxWindow != "" {
				_, _ = fmt.Fprintf(iostream.Out, "tmux: %s:%s\n", state.TMuxSession, state.TMuxWindow)
			}
			_, _ = fmt.Fprintf(iostream.Out, "Logs: %s\n", state.LogPath)

			return nil
		},
	}

	return cmd
}

func instanceHealthy(ctx context.Context, state InstanceState) (bool, error) {
	backend, ok := backendByName(state.Runtime)
	if !ok {
		return false, fmt.Errorf("unknown runtime %q", state.Runtime)
	}

	if state.TMuxWindow != "" && !state.External {
		windowExists, err := tmuxWindowExists(ctx, state.Profile)
		if err != nil {
			return false, err
		}
		if !windowExists {
			return false, nil
		}
	}

	if err := backend.HealthCheck(ctx, state); err != nil {
		return false, nil
	}

	return true, nil
}
