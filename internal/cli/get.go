package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tiaanduplessis/envy/internal/config"
)

func NewGetCmd(store *config.Store) *cobra.Command {
	var env string
	var path string

	cmd := &cobra.Command{
		Use:   "get <project> <KEY>",
		Short: "Get a variable's resolved value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			key := args[1]

			p, err := store.Load(name)
			if err != nil {
				return err
			}

			targetEnv := config.ResolveEnv(env, p.DefaultEnv)
			vars, err := config.ResolveVars(p, targetEnv, path)
			if err != nil {
				return err
			}

			value, ok := vars[key]
			if !ok {
				return fmt.Errorf("variable %q not found in environment %q", key, targetEnv)
			}

			fmt.Fprintln(cmd.OutOrStdout(), value)
			return nil
		},
	}

	cmd.Flags().StringVar(&env, "env", "", "environment")
	cmd.Flags().StringVar(&path, "path", "", "monorepo subpath (resolves with inheritance)")

	return cmd
}
