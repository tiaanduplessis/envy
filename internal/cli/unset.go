package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tiaanduplessis/envy/internal/config"
)

func NewUnsetCmd(store *config.Store) *cobra.Command {
	var env string
	var path string

	cmd := &cobra.Command{
		Use:   "unset <project> KEY [KEY ...]",
		Short: "Remove variables from a project",
		Example: `  envy unset my-app DB_HOST
  envy unset my-app DB_HOST DB_PORT --env staging
  envy unset my-app PORT --path services/api`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			keys := args[1:]

			p, err := store.Load(name)
			if err != nil {
				return err
			}

			targetEnv := config.ResolveEnv(env, p.DefaultEnv)

			removed := 0
			for _, key := range keys {
				var ok bool
				if path != "" {
					ok = p.DeletePathVar(path, targetEnv, key)
				} else {
					ok = p.DeleteVar(targetEnv, key)
				}
				if ok {
					removed++
				}
			}

			if err := store.Save(p); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Removed %d variable(s) from %q [%s]\n",
				removed, name, targetEnv)
			return nil
		},
	}

	cmd.Flags().StringVar(&env, "env", "", "target environment")
	cmd.Flags().StringVar(&path, "path", "", "target monorepo subpath")

	return cmd
}
