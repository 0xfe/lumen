package cli

import (
	"fmt"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildRootCmd() {
	if cli.rootCmd != nil {
		cli.rootCmd.ResetFlags()
		cli.rootCmd.ResetCommands()
	}

	rootCmd := &cobra.Command{
		Use:              "lumen",
		Short:            "Lumen is a commandline client for the Stellar blockchain",
		Run:              cli.help,
		PersistentPreRun: cli.setup,
	}
	cli.rootCmd = rootCmd

	home, _ := homedir.Dir()

	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output (false)")
	rootCmd.PersistentFlags().Bool("nosubmit", false, "display transaction without submitting")
	rootCmd.PersistentFlags().String("network", "test", "network to use (test)")
	rootCmd.PersistentFlags().String("ns", "default", "namespace to use (default)")
	rootCmd.PersistentFlags().String("store", fmt.Sprintf("file:%s/.lumen-data.yml", home), "namespace to use (default)")

	// Basic commands
	rootCmd.AddCommand(cli.buildVersionCmd()) // version
	rootCmd.AddCommand(cli.buildNSCmd())      // ns
	rootCmd.AddCommand(cli.buildSetCmd())     // set
	rootCmd.AddCommand(cli.buildGetCmd())     // get
	rootCmd.AddCommand(cli.buildDelCmd())     // del

	// Core commands
	rootCmd.AddCommand(cli.buildPayCmd())    // pay
	rootCmd.AddCommand(cli.buildTrustCmd())  // trust
	rootCmd.AddCommand(cli.buildSignerCmd()) // signer
	rootCmd.AddCommand(cli.buildDexCmd())    // dex
	rootCmd.AddCommand(cli.buildTxCmd())     // tx

	// Aux commands
	rootCmd.AddCommand(cli.buildFriendbotCmd()) // friendbot
	rootCmd.AddCommand(cli.buildInfoCmd())      // info
	rootCmd.AddCommand(cli.buildBalanceCmd())   // balance
	rootCmd.AddCommand(cli.buildWatchCmd())     // watch
	rootCmd.AddCommand(cli.buildFlagsCmd())     // flags
	rootCmd.AddCommand(cli.buildDataCmd())      // data

	// Alias commands
	rootCmd.AddCommand(cli.buildAccountCmd()) // account
	rootCmd.AddCommand(cli.buildAssetCmd())   // asset
}
