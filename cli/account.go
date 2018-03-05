package cli

import (
	"fmt"

	"github.com/0xfe/microstellar"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (cli *CLI) getAccountsCmd() *cobra.Command {
	accountsCmd := &cobra.Command{
		Use:   "account [new|set|get|del]",
		Short: "manage stellar keypairs and accounts",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				logrus.WithFields(logrus.Fields{"cmd": "accounts"}).Errorf("unrecognized command: %s, expecting: new|set|get|del", args[0])
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
		Use:   "set [name] [address|seed]...",
		Short: "create a keypair named [name] with [address] and/and [seed]",
		Args:  cobra.MinimumNArgs(2),
		Run:   cli.cmdSetAccount,
	})

	accountsCmd.AddCommand(&cobra.Command{
		Use:   "get [name] [address|seed]",
		Short: "get the address or seed of keypair [name]",
		Args:  cobra.MinimumNArgs(2),
		Run:   cli.cmdGetAccount,
	})

	accountsCmd.AddCommand(&cobra.Command{
		Use:   "del [name] [name]...",
		Short: "delete keypair",
		Args:  cobra.MinimumNArgs(1),
		Run:   cli.cmdNewAccount,
	})

	return accountsCmd
}

func (cli *CLI) cmdNewAccount(cmd *cobra.Command, args []string) {
	name := args[0]
	pair, err := cli.ms.CreateKeyPair()

	if err != nil {
		logrus.WithFields(logrus.Fields{"cmd": "account", "subcmd": "new"}).Errorf("could not create keypair: %s", name)
		showError("could not create key pair: %v\n", err)
		return
	}

	err = cli.SetVar(fmt.Sprintf("account:%s:address", name), pair.Address)

	if err != nil {
		logrus.WithFields(logrus.Fields{"cmd": "account", "subcmd": "new"}).Errorf("could not save keypair: %s", name)
		showError("could not save key pair: %v\n", err)
		return
	}

	showSuccess("added account %s %s:%s\n", name, pair.Address, pair.Seed)
}

func (cli *CLI) cmdSetAccount(cmd *cobra.Command, args []string) {
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
			showError("skipping invalid seed or address: %v\n", code)
			continue
		}

		err := cli.SetVar(key+keyType, code)

		if err != nil {
			logrus.WithFields(logrus.Fields{"cmd": "account", "subcmd": "set"}).Errorf("could not save account: %s", name)
			showError("could not save account: %v\n", err)
			return
		}

		showSuccess("set account %s %s %s\n", name, keyType, code)
	}
}

func (cli *CLI) cmdGetAccount(cmd *cobra.Command, args []string) {
	name := args[0]
	keyType := args[1]
	key := fmt.Sprintf("account:%s:%s", name, keyType)

	code, err := cli.GetVar(key)

	if err != nil {
		logrus.WithFields(logrus.Fields{"cmd": "account", "subcmd": "get"}).Errorf("could not load account: %s", name)
		showError("could not load account: %v\n", err)
		return
	}

	showSuccess("%s\n", code)
}
