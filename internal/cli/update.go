package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tiaanduplessis/envy/internal/config"
	"github.com/tiaanduplessis/envy/internal/dotenv"
)

func NewUpdateCmd(store *config.Store) *cobra.Command {
	var project string
	var env string
	var path string
	var merge bool
	var file string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update stored config from a .env file",
		Example: `  envy update --project my-app --file .env --env dev
  envy update --project my-app --merge`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := store.Load(project)
			if err != nil {
				return err
			}

			targetEnv := config.ResolveEnv(env, p.DefaultEnv)

			if p.Environments[targetEnv] == nil {
				p.Environments[targetEnv] = make(map[string]string)
			}

			f, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("opening %q: %w", file, err)
			}
			defer f.Close()

			vars, err := dotenv.Parse(f)
			if err != nil {
				return fmt.Errorf("parsing %q: %w", file, err)
			}

			count := 0
			for key, value := range vars {
				if path != "" {
					if merge {
						existing := p.GetPathVars(path, targetEnv)
						if existing != nil {
							if _, ok := existing[key]; ok {
								continue
							}
						}
					}
					p.SetPathVar(path, targetEnv, key, value)
				} else {
					if merge {
						if _, ok := p.Environments[targetEnv][key]; ok {
							continue
						}
					}
					p.SetVar(targetEnv, key, value)
				}
				count++
			}

			if err := store.Save(p); err != nil {
				return err
			}

			action := "Updated"
			if merge {
				action = "Merged"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s %d variable(s) in %q [%s]\n",
				action, count, project, targetEnv)
			return nil
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "project name (required)")
	cmd.MarkFlagRequired("project")
	cmd.Flags().StringVar(&env, "env", "", "target environment")
	cmd.Flags().StringVar(&path, "path", "", "target monorepo subpath")
	cmd.Flags().BoolVar(&merge, "merge", false, "only add new keys, don't overwrite")
	cmd.Flags().StringVar(&file, "file", ".env", "file to read")

	return cmd
}
