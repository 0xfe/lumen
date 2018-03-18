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
	rootCmd.PersistentFlags().String("network", "test", "network to use (test)")
	rootCmd.PersistentFlags().String("ns", "default", "namespace to use (default)")
	rootCmd.PersistentFlags().String("store", fmt.Sprintf("file:%s/.lumen-data.yml", home), "namespace to use (default)")

	// Basic commands
	rootCmd.AddCommand(cli.buildVersionCmd())
	rootCmd.AddCommand(cli.buildNSCmd())
	rootCmd.AddCommand(cli.buildSetCmd())
	rootCmd.AddCommand(cli.buildGetCmd())
	rootCmd.AddCommand(cli.buildDelCmd())
	rootCmd.AddCommand(cli.buildFriendbotCmd())
	rootCmd.AddCommand(cli.buildFlagsCmd())
	rootCmd.AddCommand(cli.buildInfoCmd())

	// Higher level
	rootCmd.AddCommand(cli.buildBalanceCmd())
	rootCmd.AddCommand(cli.buildPayCmd())
	rootCmd.AddCommand(cli.buildAccountCmd())
	rootCmd.AddCommand(cli.buildAssetCmd())
	rootCmd.AddCommand(cli.buildTrustCmd())
	rootCmd.AddCommand(cli.buildSignerCmd())
	rootCmd.AddCommand(cli.buildWatchCmd())
	rootCmd.AddCommand(cli.buildDexCmd())
}
