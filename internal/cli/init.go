package cli

import (
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tiaanduplessis/envy/internal/config"
	"github.com/tiaanduplessis/envy/internal/crypto"
)

func NewInitCmd(store *config.Store) *cobra.Command {
	return newInitCmd(store, nil)
}

func newInitCmd(store *config.Store, getPass func() (string, error)) *cobra.Command {
	var envs []string
	var paths []string
	var defaultEnv string
	var encrypt bool

	cmd := &cobra.Command{
		Use:   "init <project>",
		Short: "Create a new project configuration",
		Example: `  envy init my-app
  envy init my-app --env dev --env staging --env prod
  envy init my-app --encrypt`,
		Args: cobra.ExactArgs(1),
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

			if encrypt {
				passphrase, err := resolveNewPassphrase(getPass)
				if err != nil {
					return err
				}

				salt, err := crypto.GenerateSalt()
				if err != nil {
					return err
				}

				params := crypto.DefaultParams()
				p.Encryption = &config.EncryptionConfig{
					Enabled: true,
					Salt:    base64.StdEncoding.EncodeToString(salt),
					Params:  params,
				}

				key := crypto.DeriveKey(passphrase, salt, params)
				if err := encryptAllValues(p, key); err != nil {
					return err
				}

				if err := store.SaveRaw(p); err != nil {
					return err
				}
			} else {
				if err := store.Save(p); err != nil {
					return err
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created project %q\n", name)
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&envs, "env", nil, "environments to create (default: dev)")
	cmd.Flags().StringSliceVar(&paths, "path", nil, "monorepo subpath stubs")
	cmd.Flags().StringVar(&defaultEnv, "default-env", "", "default environment")
	cmd.Flags().BoolVar(&encrypt, "encrypt", false, "enable encryption on the new project")

	return cmd
}
