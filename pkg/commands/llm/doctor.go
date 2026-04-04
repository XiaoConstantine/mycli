package llm

import (
	"fmt"
	"os"

	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/spf13/cobra"
)

func NewDoctorCmd(iostream *iostreams.IOStreams, options *commandOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check local LLM prerequisites and environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := loadLLMConfig(options.configFile)
			if err != nil {
				return err
			}

			profileName, _ := resolveProfile(cfg, "")
			paths := resolveInstancePaths(cfg, profileName)

			fmt.Fprintln(iostream.Out, "Runtime checks:")
			if _, err := lookPath("tmux"); err == nil {
				fmt.Fprintln(iostream.Out, "- tmux: installed")
			} else {
				fmt.Fprintln(iostream.Out, "- tmux: missing")
			}

			for _, runtimeName := range []string{runtimeOllama, runtimeLlamaCpp} {
				backend, _ := backendByName(runtimeName)
				info, err := backend.Detect(ctx)
				if err != nil {
					fmt.Fprintf(iostream.Out, "- %s: error: %v\n", runtimeName, err)
					continue
				}
				if info.Available {
					fmt.Fprintf(iostream.Out, "- %s: installed (%s)\n", runtimeName, info.Version)
				} else {
					fmt.Fprintf(iostream.Out, "- %s: missing (%s)\n", runtimeName, info.InstallHint)
				}
			}

			hw := hardwareSummary(ctx)
			fmt.Fprintln(iostream.Out, "Hardware:")
			fmt.Fprintf(iostream.Out, "- os: %s\n", hw["os"])
			fmt.Fprintf(iostream.Out, "- arch: %s\n", hw["arch"])
			fmt.Fprintf(iostream.Out, "- cpu: %s\n", hw["cpu"])
			fmt.Fprintf(iostream.Out, "- ram: %s\n", hw["ram"])

			fmt.Fprintln(iostream.Out, "State directory:")
			if err := os.MkdirAll(paths.RootDir, 0o755); err != nil {
				fmt.Fprintf(iostream.Out, "- %s: not writable (%v)\n", paths.RootDir, err)
			} else {
				tempFile, err := os.CreateTemp(paths.RootDir, "doctor-*")
				if err != nil {
					fmt.Fprintf(iostream.Out, "- %s: not writable (%v)\n", paths.RootDir, err)
				} else {
					tempFile.Close()
					_ = os.Remove(tempFile.Name())
					fmt.Fprintf(iostream.Out, "- %s: writable\n", paths.RootDir)
				}
			}

			fmt.Fprintln(iostream.Out, "Ollama daemon:")
			state := InstanceState{
				Runtime:   runtimeOllama,
				Host:      defaultHost,
				Port:      11434,
				HealthURL: healthURL(runtimeOllama, defaultHost, 11434),
			}
			if err := (ollamaBackend{}).HealthCheck(ctx, state); err == nil {
				fmt.Fprintln(iostream.Out, "- healthy at http://127.0.0.1:11434")
			} else {
				fmt.Fprintln(iostream.Out, "- not reachable at http://127.0.0.1:11434")
			}

			return nil
		},
	}

	return cmd
}
