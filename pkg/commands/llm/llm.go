package llm

import (
	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/spf13/cobra"
)

func NewLLMCmd(iostream *iostreams.IOStreams) *cobra.Command {
	options := &commandOptions{}

	cmd := &cobra.Command{
		Use:   "llm",
		Short: "Manage local LLM runtimes",
		Annotations: map[string]string{
			"group": "llm",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().StringVarP(&options.configFile, "config", "c", "", "Path to the llm-enabled configuration file")
	cmd.AddCommand(NewStartCmd(iostream, options))
	cmd.AddCommand(NewStopCmd(iostream, options))
	cmd.AddCommand(NewStatusCmd(iostream, options))
	cmd.AddCommand(NewLogsCmd(iostream, options))
	cmd.AddCommand(NewDoctorCmd(iostream, options))

	return cmd
}
