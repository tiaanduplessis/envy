package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

type VersionInfo struct {
	Version   string
	Commit    string
	Date      string
	GoVersion string
}

func NewVersionCmd(info VersionInfo) *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Print version information",
		Example: "  envy version",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "envy %s\n", info.Version)
			if info.Commit != "" {
				fmt.Fprintf(out, "commit: %s\n", info.Commit)
			}
			if info.Date != "" {
				fmt.Fprintf(out, "built:  %s\n", info.Date)
			}
			if info.GoVersion != "" {
				fmt.Fprintf(out, "go:     %s\n", info.GoVersion)
			}
		},
	}
}
