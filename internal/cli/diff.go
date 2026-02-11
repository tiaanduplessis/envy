package cli

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/tiaanduplessis/envy/internal/config"
	"github.com/tiaanduplessis/envy/internal/dotenv"
	"github.com/tiaanduplessis/envy/internal/util"
)

type DiffResult struct {
	Added   map[string]string    // keys only in right
	Removed map[string]string    // keys only in left
	Changed map[string][2]string // keys in both with different values: [left, right]
}

func ComputeDiff(left, right map[string]string) DiffResult {
	result := DiffResult{
		Added:   make(map[string]string),
		Removed: make(map[string]string),
		Changed: make(map[string][2]string),
	}

	for k, lv := range left {
		rv, ok := right[k]
		if !ok {
			result.Removed[k] = lv
		} else if lv != rv {
			result.Changed[k] = [2]string{lv, rv}
		}
	}

	for k, rv := range right {
		if _, ok := left[k]; !ok {
			result.Added[k] = rv
		}
	}

	return result
}

func (d DiffResult) IsEmpty() bool {
	return len(d.Added) == 0 && len(d.Removed) == 0 && len(d.Changed) == 0
}

func NewDiffCmd(store *config.Store) *cobra.Command {
	var envs []string
	var reveal bool

	cmd := &cobra.Command{
		Use:   "diff <project>",
		Short: "Compare environments or local .env against stored config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			out := cmd.OutOrStdout()

			p, err := store.Load(name)
			if err != nil {
				return err
			}

			var left, right map[string]string
			var leftLabel, rightLabel string

			if len(envs) == 2 {
				left, err = config.ResolveVars(p, envs[0], "")
				if err != nil {
					return err
				}
				right, err = config.ResolveVars(p, envs[1], "")
				if err != nil {
					return err
				}
				leftLabel = envs[0]
				rightLabel = envs[1]
			} else if len(envs) == 1 {
				f, err := os.Open(".env")
				if err != nil {
					return fmt.Errorf("opening .env: %w", err)
				}
				defer f.Close()

				left, err = dotenv.Parse(f)
				if err != nil {
					return fmt.Errorf("parsing .env: %w", err)
				}

				right, err = config.ResolveVars(p, envs[0], "")
				if err != nil {
					return err
				}
				leftLabel = ".env"
				rightLabel = envs[0]
			} else {
				return fmt.Errorf("provide one --env to diff against local .env, or two --env flags to compare environments")
			}

			diff := ComputeDiff(left, right)

			if diff.IsEmpty() {
				fmt.Fprintln(out, "No differences.")
				return nil
			}

			formatValue := func(v string) string {
				if reveal {
					return v
				}
				return util.MaskValue(v)
			}

			keys := sortedKeys(diff.Removed)
			for _, k := range keys {
				fmt.Fprintf(out, "- %s=%s (%s)\n", k, formatValue(diff.Removed[k]), leftLabel)
			}

			keys = sortedKeys(diff.Added)
			for _, k := range keys {
				fmt.Fprintf(out, "+ %s=%s (%s)\n", k, formatValue(diff.Added[k]), rightLabel)
			}

			changedKeys := make([]string, 0, len(diff.Changed))
			for k := range diff.Changed {
				changedKeys = append(changedKeys, k)
			}
			sort.Strings(changedKeys)
			for _, k := range changedKeys {
				pair := diff.Changed[k]
				fmt.Fprintf(out, "~ %s: %s -> %s\n", k, formatValue(pair[0]), formatValue(pair[1]))
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVar(&envs, "env", nil, "environment(s) to compare")
	cmd.Flags().BoolVar(&reveal, "reveal", false, "show actual values instead of masked")

	return cmd
}
