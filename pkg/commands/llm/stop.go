package llm

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/spf13/cobra"
)

func NewStopCmd(iostream *iostreams.IOStreams, options *commandOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop [profile]",
		Short: "Stop a running local LLM runtime",
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
				return fmt.Errorf("no llm instance found for profile %q", profileName)
			}

			return stopInstance(ctx, iostream, *state, paths)
		},
	}

	return cmd
}

func stopInstance(ctx context.Context, iostream *iostreams.IOStreams, state InstanceState, paths InstancePaths) error {
	if state.External {
		if err := removeState(paths.StatePath); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(iostream.Out, "Removed mycli state for external %s instance. The daemon is still running.\n", state.Runtime)
		return nil
	}

	if state.Foreground && state.PID > 0 {
		process, err := os.FindProcess(state.PID)
		if err != nil {
			return err
		}
		_ = process.Signal(os.Interrupt)
		time.Sleep(3 * time.Second)
		_ = process.Signal(syscall.SIGKILL)
		if err := removeState(paths.StatePath); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(iostream.Out, "Stopped foreground llm instance for profile %q.\n", state.Profile)
		return nil
	}

	if state.TMuxWindow != "" {
		_ = sendTmuxInterrupt(ctx, state.Profile)
		time.Sleep(500 * time.Millisecond)

		healthy, _ := instanceHealthy(ctx, state)
		if healthy {
			_ = killTmuxWindow(ctx, state.Profile)
		}

		if err := removeState(paths.StatePath); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(iostream.Out, "Stopped llm instance for profile %q.\n", state.Profile)
		return nil
	}

	if err := removeState(paths.StatePath); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(iostream.Out, "Removed stale llm state for profile %q.\n", state.Profile)
	return nil
}
