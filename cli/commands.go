package cli

import (
	"fmt"
	"os"

	"github.com/0xfe/microstellar"
	"github.com/sirupsen/logrus"
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

	home := os.Getenv("HOME")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output (false)")
	rootCmd.PersistentFlags().String("network", "test", "network to use (test)")
	rootCmd.PersistentFlags().String("ns", "default", "namespace to use (default)")
	rootCmd.PersistentFlags().String("store", fmt.Sprintf("file:%s/.lumen-data.yml", home), "namespace to use (default)")

	rootCmd.AddCommand(cli.buildPayCmd())
	rootCmd.AddCommand(cli.buildAccountCmd())
	rootCmd.AddCommand(cli.buildAssetCmd())
	rootCmd.AddCommand(cli.buildTrustCmd())

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "get version of lumen CLI",
		Run:   cli.cmdVersion,
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "ns [namespace]",
		Short: "set namespace to [namespace]",
		Args:  cobra.MinimumNArgs(0),
		Run:   cli.cmdNS,
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "set [key] [value]",
		Short: "set variable",
		Args:  cobra.MinimumNArgs(2),
		Run:   cli.cmdSet,
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "get [key]",
		Short: "get variable",
		Args:  cobra.MinimumNArgs(1),
		Run:   cli.cmdGet,
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "del [key]",
		Short: "delete variable",
		Args:  cobra.MinimumNArgs(1),
		Run:   cli.cmdDel,
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "watch [address]",
		Short: "watch the address on the ledger",
		Args:  cobra.MinimumNArgs(1),
		Run:   cli.cmdWatch,
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "balance [address]",
		Short: "get the balance of [address] in lumens",
		Args:  cobra.MinimumNArgs(1),
		Run:   cli.cmdBalance,
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "friendbot [address]",
		Short: "fund [address] on the test network with friendbot",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			logFields := logrus.Fields{"cmd": "trust", "subcmd": "create"}
			address, err := cli.validateAddressOrSeed(logFields, name, "address")

			if err != nil {
				cli.error(logFields, "invalid account: %s", name)
				return
			}

			response, err := microstellar.FundWithFriendBot(address)

			if err != nil {
				cli.error(logFields, "friendbot error: %v", err)
				return
			}

			showSuccess("friendbot says:\n %v", response)
		},
	})
}

func (cli *CLI) cmdVersion(cmd *cobra.Command, args []string) {
	showSuccess(cli.version)
}

func (cli *CLI) cmdNS(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		ns := args[0]

		err := cli.SetGlobalVar("ns", ns)
		if err != nil {
			cli.error(logrus.Fields{"cmd": "setNS"}, "set failed: ", err)
			return
		}

		cli.ns = ns
	} else {
		showSuccess(cli.ns)
	}
}

func (cli *CLI) cmdSet(cmd *cobra.Command, args []string) {
	key := fmt.Sprintf("vars:%s", args[0])
	val := args[1]

	err := cli.SetVar(key, val)
	if err != nil {
		cli.error(logrus.Fields{"cmd": "set"}, "set failed: ", err)
		return
	}
}

func (cli *CLI) cmdDel(cmd *cobra.Command, args []string) {
	key := fmt.Sprintf("vars:%s", args[0])

	err := cli.DelVar(key)
	if err != nil {
		cli.error(logrus.Fields{"cmd": "del"}, "del failed: %s\n", err)
		return
	}
}

func (cli *CLI) cmdGet(cmd *cobra.Command, args []string) {
	key := fmt.Sprintf("vars:%s", args[0])

	val, err := cli.GetVar(key)
	if err == nil {
		showSuccess(val)
	} else {
		cli.error(logrus.Fields{"cmd": "get"}, "no such variable: %s\n", args[0])
		return
	}
}

func (cli *CLI) cmdWatch(cmd *cobra.Command, args []string) {
	address := args[0]

	watcher, err := cli.ms.WatchPayments(address)

	if err != nil {
		cli.error(logrus.Fields{"cmd": "watch"}, "can't watch address: %+v\n", err)
		return
	}

	for p := range watcher.Ch {
		showSuccess("%v %v from %v to %v", p.Amount, p.AssetCode, p.From, p.To)
	}

	if watcher.Err != nil {
		cli.error(logrus.Fields{"cmd": "watch"}, "%+v\n", *watcher.Err)
		return
	}
}

func (cli *CLI) cmdBalance(cmd *cobra.Command, args []string) {
	fields := logrus.Fields{"cmd": "balance"}
	address, err := cli.validateAddressOrSeed(fields, args[0], "address")

	if err != nil {
		return
	}

	account, err := cli.ms.LoadAccount(address)

	if err != nil {
		cli.error(logrus.Fields{"cmd": "balance"}, "payment failed: %v", microstellar.ErrorString(err))
		// must return
	} else {
		showSuccess(account.GetNativeBalance())
	}
}
