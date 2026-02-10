package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tiaanduplessis/envy/internal/config"
)

func NewInitCmd(store *config.Store) *cobra.Command {
	var envs []string
	var paths []string
	var defaultEnv string

	cmd := &cobra.Command{
		Use:   "init <project>",
		Short: "Create a new project configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if store.Exists(name) {
				return fmt.Errorf("project %q already exists", name)
			}

			p, err := config.NewProject(name, envs, defaultEnv)
			if err != nil {
				return err
			}

			for _, path := range paths {
				if p.Paths == nil {
					p.Paths = make(map[string]map[string]map[string]string)
				}
				p.Paths[path] = make(map[string]map[string]string)
			}

			if err := store.Save(p); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created project %q\n", name)
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&envs, "env", nil, "environments to create (default: dev)")
	cmd.Flags().StringSliceVar(&paths, "path", nil, "monorepo subpath stubs")
	cmd.Flags().StringVar(&defaultEnv, "default-env", "", "default environment")

	return cmd
}
