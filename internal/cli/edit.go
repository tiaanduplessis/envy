package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/tiaanduplessis/envy/internal/config"
)

func NewEditCmd(store *config.Store) *cobra.Command {
	return newEditCmd(store, nil)
}

func newEditCmd(store *config.Store, editorFn func(path string) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "edit <project>",
		Short:   "Open a project config file in your editor",
		Example: `  envy edit my-app`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if !store.Exists(name) {
				return fmt.Errorf("project %q not found", name)
			}

			path := store.ProjectPath(name)

			if editorFn != nil {
				return editorFn(path)
			}

			editor := os.Getenv("VISUAL")
			if editor == "" {
				editor = os.Getenv("EDITOR")
			}
			if editor == "" {
				editor = "vi"
			}

			c := exec.Command(editor, path)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}

	return cmd
}
