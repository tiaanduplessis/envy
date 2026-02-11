package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func NewDocCmd(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:    "doc <output-dir>",
		Short:  "Generate man pages",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			header := &doc.GenManHeader{
				Title:   "ENVY",
				Section: "1",
			}
			return doc.GenManTree(root, header, args[0])
		},
	}
}
