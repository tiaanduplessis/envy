package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tiaanduplessis/envy/internal/config"
)

type listEntry struct {
	Name         string   `json:"name"`
	Environments []string `json:"environments"`
	PathCount    int      `json:"path_count"`
}

func NewListCmd(store *config.Store) *cobra.Command {
	var jsonOutput bool
	var quiet bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			names, err := store.List()
			if err != nil {
				return err
			}

			if len(names) == 0 {
				if !quiet && !jsonOutput {
					fmt.Fprintln(cmd.OutOrStdout(), "No projects found.")
				}
				if jsonOutput {
					fmt.Fprintln(cmd.OutOrStdout(), "[]")
				}
				return nil
			}

			if quiet {
				for _, name := range names {
					fmt.Fprintln(cmd.OutOrStdout(), name)
				}
				return nil
			}

			var entries []listEntry
			for _, name := range names {
				p, err := store.Load(name)
				if err != nil {
					continue
				}

				var envNames []string
				for e := range p.Environments {
					envNames = append(envNames, e)
				}
				sort.Strings(envNames)

				entries = append(entries, listEntry{
					Name:         name,
					Environments: envNames,
					PathCount:    len(p.Paths),
				})
			}

			if jsonOutput {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(entries)
			}

			for _, e := range entries {
				line := fmt.Sprintf("%-20s (envs: %s)", e.Name, strings.Join(e.Environments, ", "))
				if e.PathCount > 0 {
					line += fmt.Sprintf(" [%d path(s)]", e.PathCount)
				}
				fmt.Fprintln(cmd.OutOrStdout(), line)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "project names only")

	return cmd
}
