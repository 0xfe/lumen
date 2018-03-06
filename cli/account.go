package cli

import (
	"fmt"

	"github.com/0xfe/microstellar"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (cli *CLI) getAccountCmd() *cobra.Command {
	accountsCmd := &cobra.Command{
		Use:   "account [new|set|get|del]",
		Short: "manage stellar keypairs and accounts",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				showError(logrus.Fields{"cmd": "accounts"}, "unrecognized command: %s, expecting: new|set|get|del", args[0])
				return
			}
		},
	}

	accountsCmd.AddCommand(cli.getAccountNewCmd())
	accountsCmd.AddCommand(cli.getAccountSetCmd())
	accountsCmd.AddCommand(cli.getAccountGetCmd())
	accountsCmd.AddCommand(cli.getAccountDelCmd())

	return accountsCmd
}

func (cli *CLI) getAccountNewCmd() *cobra.Command {
	accountNewCmd := &cobra.Command{
		Use:   "new [name]",
		Short: "create a new random keypair named [name]",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			pair, err := cli.ms.CreateKeyPair()

			if err != nil {
				showError(logrus.Fields{"cmd": "account", "subcmd": "new"}, "could not create keypair: %s", name)
				return
			}

			err = cli.SetVar(fmt.Sprintf("account:%s:address", name), pair.Address)

			if err != nil {
				showError(logrus.Fields{"cmd": "account", "subcmd": "new"}, "could not save keypair: %s", name)
				return
			}

			err = cli.SetVar(fmt.Sprintf("account:%s:seed", name), pair.Seed)

			if err != nil {
				showError(logrus.Fields{"cmd": "account", "subcmd": "new"}, "could not save keypair: %s", name)
				return
			}

			showSuccess("%s %s\n", pair.Address, pair.Seed)
		},
	}

	accountNewCmd.Flags().String("name", "", "give the account a name")
	return accountNewCmd
}

func (cli *CLI) getAccountSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set [name] [address|seed]...",
		Short: "create a keypair named [name] with [address] and/and [seed]",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			for i := range args {
				if i == 0 {
					continue
				}

				code := args[i]
				keyType := ""

				key := fmt.Sprintf("account:%s:", name)
				if microstellar.ValidAddress(code) == nil {
					keyType = "address"
				} else if microstellar.ValidSeed(code) == nil {
					keyType = "seed"
				} else {
					logrus.WithFields(logrus.Fields{"cmd": "account", "subcmd": "sed"}).Errorf("skipping invalid seed or address: %v\n", code)
					continue
				}

				err := cli.SetVar(key+keyType, code)

				if err != nil {
					showError(logrus.Fields{"cmd": "account", "subcmd": "set"}, "could not save account: %s", name)
					return
				}
			}
		},
	}
}

func (cli *CLI) getAccountGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [name] [address|seed]",
		Short: "get the address or seed of keypair [name]",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			keyType := args[1]
			key := fmt.Sprintf("account:%s:%s", name, keyType)

			code, err := cli.GetVar(key)

			if err != nil {
				showError(logrus.Fields{"cmd": "account", "subcmd": "get"}, "could not get %s for account: %s", keyType, name)
				return
			}

			showSuccess("%s\n", code)
		},
	}
}

func (cli *CLI) getAccountDelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "del [name]",
		Short: "delete keypair",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			err := cli.DelVar(fmt.Sprintf("account:%s:seed", name))
			err = cli.DelVar(fmt.Sprintf("account:%s:address", name))

			if err != nil {
				showError(logrus.Fields{"cmd": "account", "subcmd": "del"}, "could not delete account: %s", name)
				return
			}
		},
	}
}
