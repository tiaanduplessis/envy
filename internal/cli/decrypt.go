package cli

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tiaanduplessis/envy/internal/config"
	"github.com/tiaanduplessis/envy/internal/crypto"
)

func NewDecryptCmd(store *config.Store) *cobra.Command {
	return newDecryptCmd(store, nil)
}

func newDecryptCmd(store *config.Store, getPass func(string) (string, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "decrypt <project>",
		Short:   "Disable encryption on a project",
		Example: "  envy decrypt my-app",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			p, err := store.LoadRaw(name)
			if err != nil {
				return err
			}

			if !p.IsEncrypted() {
				return fmt.Errorf("project %q is not encrypted", name)
			}

			passphrase, err := resolvePassphrase(getPass, "Enter passphrase for "+name+": ")
			if err != nil {
				return err
			}

			salt, err := base64.StdEncoding.DecodeString(p.Encryption.Salt)
			if err != nil {
				return fmt.Errorf("decoding salt: %w", err)
			}
			key := crypto.DeriveKey(passphrase, salt, p.Encryption.Params)

			if err := decryptAllValues(p, key); err != nil {
				return err
			}

			p.Encryption = nil
			p.UpdatedAt = time.Now().UTC()

			if err := store.SaveRaw(p); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Decrypted project %q\n", name)
			return nil
		},
	}
	return cmd
}
