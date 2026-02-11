package cli

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tiaanduplessis/envy/internal/config"
	"github.com/tiaanduplessis/envy/internal/crypto"
)

func NewEncryptCmd(store *config.Store) *cobra.Command {
	return newEncryptCmd(store, nil)
}

func newEncryptCmd(store *config.Store, getPass func() (string, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encrypt <project>",
		Short: "Enable encryption on a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			p, err := store.LoadRaw(name)
			if err != nil {
				return err
			}

			if p.IsEncrypted() {
				return fmt.Errorf("project %q is already encrypted", name)
			}

			passphrase, err := resolveNewPassphrase(getPass)
			if err != nil {
				return err
			}

			salt, err := crypto.GenerateSalt()
			if err != nil {
				return err
			}

			params := crypto.DefaultParams()
			key := crypto.DeriveKey(passphrase, salt, params)

			if err := encryptAllValues(p, key); err != nil {
				return err
			}

			p.Encryption = &config.EncryptionConfig{
				Enabled: true,
				Salt:    base64.StdEncoding.EncodeToString(salt),
				Params:  params,
			}
			p.UpdatedAt = time.Now().UTC()

			if err := store.SaveRaw(p); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Encrypted project %q\n", name)
			return nil
		},
	}
	return cmd
}

func encryptAllValues(p *config.Project, key []byte) error {
	for env, vars := range p.Environments {
		encrypted, err := crypto.EncryptMap(key, vars)
		if err != nil {
			return fmt.Errorf("encrypting env %q: %w", env, err)
		}
		p.Environments[env] = encrypted
	}
	for path, envMap := range p.Paths {
		for env, vars := range envMap {
			encrypted, err := crypto.EncryptMap(key, vars)
			if err != nil {
				return fmt.Errorf("encrypting path %q env %q: %w", path, env, err)
			}
			p.Paths[path][env] = encrypted
		}
	}
	return nil
}

func decryptAllValues(p *config.Project, key []byte) error {
	for env, vars := range p.Environments {
		decrypted, err := crypto.DecryptMap(key, vars)
		if err != nil {
			return fmt.Errorf("decrypting env %q: %w", env, err)
		}
		p.Environments[env] = decrypted
	}
	for path, envMap := range p.Paths {
		for env, vars := range envMap {
			decrypted, err := crypto.DecryptMap(key, vars)
			if err != nil {
				return fmt.Errorf("decrypting path %q env %q: %w", path, env, err)
			}
			p.Paths[path][env] = decrypted
		}
	}
	return nil
}

func resolveNewPassphrase(getPass func() (string, error)) (string, error) {
	if getPass != nil {
		return getPass()
	}
	return crypto.GetPassphraseWithConfirm()
}

func resolvePassphrase(getPass func(string) (string, error), prompt string) (string, error) {
	if getPass != nil {
		return getPass(prompt)
	}
	return crypto.GetPassphrase(prompt)
}
