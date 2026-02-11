package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/tiaanduplessis/envy/internal/cli"
	"github.com/tiaanduplessis/envy/internal/config"
	"github.com/tiaanduplessis/envy/internal/crypto"
	"github.com/tiaanduplessis/envy/internal/util"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

func main() {
	dir, err := util.ProjectsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	store := config.NewStore(dir)
	store.SetPassphraseFunc(crypto.GetPassphrase)
	cmd := cli.NewRootCmd(store)
	cmd.AddCommand(cli.NewVersionCmd(cli.VersionInfo{
		Version:   version,
		Commit:    commit,
		Date:      date,
		GoVersion: runtime.Version(),
	}))
	cmd.AddCommand(cli.NewDocCmd(cmd))

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
