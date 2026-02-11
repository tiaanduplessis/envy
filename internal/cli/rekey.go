package cli

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tiaanduplessis/envy/internal/config"
	"github.com/tiaanduplessis/envy/internal/crypto"
)

func NewRekeyCmd(store *config.Store) *cobra.Command {
	return newRekeyCmd(store, nil, nil)
}

func newRekeyCmd(store *config.Store, getOldPass func(string) (string, error), getNewPass func() (string, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rekey <project>",
		Short: "Change the encryption passphrase for a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			p, err := store.LoadRaw(name)
			if err != nil {
				return err
			}

			if !p.IsEncrypted() {
				return fmt.Errorf("project %q is not encrypted", name)
			}

			oldPass, err := resolvePassphrase(getOldPass, "Enter current passphrase: ")
			if err != nil {
				return err
			}

			oldSalt, err := base64.StdEncoding.DecodeString(p.Encryption.Salt)
			if err != nil {
				return fmt.Errorf("decoding salt: %w", err)
			}
			oldKey := crypto.DeriveKey(oldPass, oldSalt, p.Encryption.Params)

			if err := decryptAllValues(p, oldKey); err != nil {
				return err
			}

			newPass, err := resolveNewPassphrase(getNewPass)
			if err != nil {
				return err
			}

			newSalt, err := crypto.GenerateSalt()
			if err != nil {
				return err
			}

			newParams := crypto.DefaultParams()
			newKey := crypto.DeriveKey(newPass, newSalt, newParams)

			if err := encryptAllValues(p, newKey); err != nil {
				return err
			}

			p.Encryption = &config.EncryptionConfig{
				Enabled: true,
				Salt:    base64.StdEncoding.EncodeToString(newSalt),
				Params:  newParams,
			}
			p.UpdatedAt = time.Now().UTC()

			if err := store.SaveRaw(p); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Re-encrypted project %q with new passphrase\n", name)
			return nil
		},
	}
	return cmd
}
