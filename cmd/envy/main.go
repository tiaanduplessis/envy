package main

import (
	"fmt"
	"os"

	"github.com/tiaanduplessis/envy/internal/cli"
	"github.com/tiaanduplessis/envy/internal/config"
	"github.com/tiaanduplessis/envy/internal/crypto"
	"github.com/tiaanduplessis/envy/internal/util"
)

var version = "dev"

func main() {
	dir, err := util.ProjectsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	store := config.NewStore(dir)
	store.SetPassphraseFunc(crypto.GetPassphrase)
	cmd := cli.NewRootCmd(store)
	cmd.AddCommand(cli.NewVersionCmd(version))

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
