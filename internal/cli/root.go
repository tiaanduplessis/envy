package cli

import (
	"github.com/spf13/cobra"
	"github.com/tiaanduplessis/envy/internal/config"
)

// NewRootCmd creates the top-level envy command with all subcommands wired in.
func NewRootCmd(store *config.Store) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "envy",
		Short: "Manage .env files from a centralised config store",
		Long:  "Envy is a local-first CLI for managing .env files across projects from a centralised YAML config store.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		NewInitCmd(store),
		NewScanCmd(store),
		NewSetCmd(store),
		NewGetCmd(store),
		NewListCmd(store),
		NewShowCmd(store),
		NewLoadCmd(store),
		NewUpdateCmd(store),
		NewDeleteCmd(store),
		NewEnvCmd(store),
		NewDiffCmd(store),
	)

	return cmd
}
