package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/tiaanduplessis/envy/internal/config"
)

func NewDeleteCmd(store *config.Store) *cobra.Command {
	return newDeleteCmd(store, nil)
}

func newDeleteCmd(store *config.Store, stdin io.Reader) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <project>",
		Short: "Delete a project",
		Example: `  envy delete my-app
  envy delete my-app --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if !store.Exists(name) {
				return fmt.Errorf("project %q not found", name)
			}

			if !force {
				reader := stdin
				if reader == nil {
					reader = os.Stdin
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Delete project %q? This cannot be undone. [y/N] ", name)
				scanner := bufio.NewScanner(reader)
				scanner.Scan()
				answer := scanner.Text()
				if answer != "y" && answer != "Y" {
					fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return nil
				}
			}

			if err := store.Delete(name); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted project %q\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "skip confirmation")
	return cmd
}
