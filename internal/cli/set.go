package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tiaanduplessis/envy/internal/config"
)

func NewSetCmd(store *config.Store) *cobra.Command {
	var env string
	var path string

	cmd := &cobra.Command{
		Use:   "set <project> KEY=VALUE [KEY=VALUE ...]",
		Short: "Set variables in a project",
		Example: `  envy set my-app DB_HOST=localhost DB_PORT=5432
  envy set my-app DB_HOST=staging-db --env staging
  envy set my-app PORT=3000 --path services/api`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			pairs := args[1:]

			p, err := store.Load(name)
			if err != nil {
				return err
			}

			targetEnv := config.ResolveEnv(env, p.DefaultEnv)

			for _, pair := range pairs {
				key, value, ok := strings.Cut(pair, "=")
				if !ok {
					return fmt.Errorf("invalid format %q: expected KEY=VALUE", pair)
				}
				if key == "" {
					return fmt.Errorf("empty key in %q", pair)
				}

				if path != "" {
					p.SetPathVar(path, targetEnv, key, value)
				} else {
					p.SetVar(targetEnv, key, value)
				}
			}

			if err := store.Save(p); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Set %d variable(s) in %q [%s]\n",
				len(pairs), name, targetEnv)
			return nil
		},
	}

	cmd.Flags().StringVar(&env, "env", "", "target environment")
	cmd.Flags().StringVar(&path, "path", "", "target monorepo subpath")

	return cmd
}
