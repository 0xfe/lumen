package lumen

import (
	"fmt"
	"os"

	"github.com/0xfe/lumen/cli"
	"github.com/0xfe/lumen/store"
	"github.com/0xfe/microstellar"
	"github.com/spf13/cobra"
)

func help(cmd *cobra.Command, args []string) {
	fmt.Fprint(os.Stderr, cmd.UsageString())
}

var rootCmd = &cobra.Command{
	Use:   "lumen",
	Short: "Lumen is a commandline client for the Stellar blockchain",
	Run:   help,
}

func Start() {
	store, _ := store.NewStore("internal", "")
	cli := cli.NewCLI(store, microstellar.New("test"))

	// Install the CLI
	cli.Install(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		// fmt.Fprintf(os.Stderr, "error: %v", err)
	}
}
