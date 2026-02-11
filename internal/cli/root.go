package cli

import (
	"github.com/spf13/cobra"
	"github.com/tiaanduplessis/envy/internal/config"
)

func NewRootCmd(store *config.Store) *cobra.Command {
	var llm bool

	cmd := &cobra.Command{
		Use:           "envy",
		Short:         "Manage .env files from a centralised config store",
		Long:          "Envy is a local-first CLI for managing .env files across projects from a centralised YAML config store.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if llm {
				generateLLMOutput(cmd.Root(), cmd.OutOrStdout())
				return nil
			}
			return cmd.Help()
		},
	}

	cmd.Flags().BoolVar(&llm, "llm", false, "print comprehensive CLI reference for LLMs")

	cmd.AddCommand(
		NewInitCmd(store),
		NewScanCmd(store),
		NewSetCmd(store),
		NewUnsetCmd(store),
		NewGetCmd(store),
		NewListCmd(store),
		NewShowCmd(store),
		NewLoadCmd(store),
		NewUpdateCmd(store),
		NewDeleteCmd(store),
		NewEnvCmd(store),
		NewDiffCmd(store),
		NewEncryptCmd(store),
		NewDecryptCmd(store),
		NewRekeyCmd(store),
	)

	return cmd
}
