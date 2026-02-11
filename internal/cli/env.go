package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/tiaanduplessis/envy/internal/config"
)

func NewEnvCmd(store *config.Store) *cobra.Command {
	return newEnvCmd(store, nil)
}

func newEnvCmd(store *config.Store, stdin io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Manage environments within a project",
	}

	cmd.AddCommand(
		newEnvAddCmd(store),
		newEnvRemoveCmd(store, stdin),
		newEnvListCmd(store),
		newEnvCopyCmd(store),
		newEnvFileCmd(store),
	)

	return cmd
}

func newEnvAddCmd(store *config.Store) *cobra.Command {
	return &cobra.Command{
		Use:     "add <project> <environment>",
		Short:   "Add an environment to a project",
		Example: "  envy env add my-app staging",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, envName := args[0], args[1]

			p, err := store.Load(name)
			if err != nil {
				return err
			}

			if _, ok := p.Environments[envName]; ok {
				return fmt.Errorf("environment %q already exists in project %q", envName, name)
			}

			if p.Environments == nil {
				p.Environments = make(map[string]map[string]string)
			}
			p.Environments[envName] = make(map[string]string)

			if err := store.Save(p); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Added environment %q to %q\n", envName, name)
			return nil
		},
	}
}

func newEnvRemoveCmd(store *config.Store, stdin io.Reader) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "remove <project> <environment>",
		Short: "Remove an environment from a project",
		Example: `  envy env remove my-app staging
  envy env remove my-app staging --force`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, envName := args[0], args[1]

			p, err := store.Load(name)
			if err != nil {
				return err
			}

			if _, ok := p.Environments[envName]; !ok {
				return fmt.Errorf("environment %q not found in project %q", envName, name)
			}

			if !force {
				reader := stdin
				if reader == nil {
					reader = os.Stdin
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Remove environment %q from %q? [y/N] ", envName, name)
				scanner := bufio.NewScanner(reader)
				scanner.Scan()
				answer := scanner.Text()
				if answer != "y" && answer != "Y" {
					fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return nil
				}
			}

			delete(p.Environments, envName)

			for path := range p.Paths {
				delete(p.Paths[path], envName)
			}

			if p.EnvFiles != nil {
				delete(p.EnvFiles, envName)
			}

			if err := store.Save(p); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Removed environment %q from %q\n", envName, name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "skip confirmation")
	return cmd
}

func newEnvListCmd(store *config.Store) *cobra.Command {
	return &cobra.Command{
		Use:     "list <project>",
		Short:   "List environments in a project",
		Example: "  envy env list my-app",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := store.Load(args[0])
			if err != nil {
				return err
			}

			var names []string
			for e := range p.Environments {
				names = append(names, e)
			}
			sort.Strings(names)

			for _, name := range names {
				if p.EnvFiles != nil {
					if file, ok := p.EnvFiles[name]; ok {
						fmt.Fprintf(cmd.OutOrStdout(), "%-12s (%s)\n", name, file)
						continue
					}
				}
				fmt.Fprintln(cmd.OutOrStdout(), name)
			}
			return nil
		},
	}
}

func newEnvCopyCmd(store *config.Store) *cobra.Command {
	var from, to string

	cmd := &cobra.Command{
		Use:     "copy <project>",
		Short:   "Copy variables from one environment to another",
		Example: "  envy env copy my-app --from dev --to staging",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			p, err := store.Load(name)
			if err != nil {
				return err
			}

			srcVars, ok := p.Environments[from]
			if !ok {
				return fmt.Errorf("source environment %q not found in project %q", from, name)
			}

			if p.Environments[to] == nil {
				p.Environments[to] = make(map[string]string)
			}

			for k, v := range srcVars {
				p.Environments[to][k] = v
			}

			if err := store.Save(p); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Copied %d variable(s) from %q to %q in %q\n",
				len(srcVars), from, to, name)
			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "source environment (required)")
	cmd.Flags().StringVar(&to, "to", "", "destination environment (required)")
	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("to")

	return cmd
}

func newEnvFileCmd(store *config.Store) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "Manage output filenames for environments",
	}

	cmd.AddCommand(
		newEnvFileSetCmd(store),
		newEnvFileClearCmd(store),
	)

	return cmd
}

func newEnvFileSetCmd(store *config.Store) *cobra.Command {
	return &cobra.Command{
		Use:     "set <project> <environment> <filename>",
		Short:   "Set the output filename for an environment",
		Example: "  envy env file set my-app production .env.production",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, envName, filename := args[0], args[1], args[2]

			p, err := store.Load(name)
			if err != nil {
				return err
			}

			if _, ok := p.Environments[envName]; !ok {
				return fmt.Errorf("environment %q not found in project %q", envName, name)
			}

			p.SetEnvFile(envName, filename)

			if err := store.Save(p); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Set output file for %q to %q in %q\n", envName, filename, name)
			return nil
		},
	}
}

func newEnvFileClearCmd(store *config.Store) *cobra.Command {
	return &cobra.Command{
		Use:     "clear <project> <environment>",
		Short:   "Clear the output filename for an environment",
		Example: "  envy env file clear my-app production",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, envName := args[0], args[1]

			p, err := store.Load(name)
			if err != nil {
				return err
			}

			p.ClearEnvFile(envName)

			if err := store.Save(p); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Cleared output file for %q in %q\n", envName, name)
			return nil
		},
	}
}
