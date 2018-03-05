package cli

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (cli *CLI) getAccountsCmd() *cobra.Command {
	accountsCmd := &cobra.Command{
		Use:   "account [new|add|del|address|seed]",
		Short: "look up accounts",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				logrus.WithFields(logrus.Fields{"cmd": "accounts"}).Errorf("unrecognized command: %s, expecting: new|add|del|address|seed", args[0])
			}
		},
	}

	accountsCmd.AddCommand(&cobra.Command{
		Use:   "new [name]",
		Short: "create a new random keypair named [name]",
		Args:  cobra.MinimumNArgs(1),
		Run:   cli.cmdNewAccount,
	})

	accountsCmd.AddCommand(&cobra.Command{
		Use:   "add [name] [address] [seed]",
		Short: "add a keypayr named [name] with [address] and [seed]",
		Args:  cobra.MinimumNArgs(2),
		Run:   cli.cmdNewAccount,
	})

	return accountsCmd
}

func (cli *CLI) cmdNewAccount(cmd *cobra.Command, args []string) {
	showSuccess("adding account: %s\n", args[0])
}
