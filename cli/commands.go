package cli

import (
	"fmt"

	"github.com/0xfe/microstellar"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (cli *CLI) cmdVersion(cmd *cobra.Command, args []string) {
	showSuccess("v0.1")
}

func (cli *CLI) cmdNS(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		ns := args[0]

		err := cli.SetGlobalVar("ns", ns)
		if err != nil {
			showError(logrus.Fields{"cmd": "setNS"}, "set failed: ", err)
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
		showError(logrus.Fields{"cmd": "set"}, "set failed: ", err)
		return
	}
}

func (cli *CLI) cmdDel(cmd *cobra.Command, args []string) {
	key := fmt.Sprintf("vars:%s", args[0])

	err := cli.DelVar(key)
	if err != nil {
		showError(logrus.Fields{"cmd": "del"}, "del failed: %s\n", err)
		return
	}
}

func (cli *CLI) cmdGet(cmd *cobra.Command, args []string) {
	key := fmt.Sprintf("vars:%s", args[0])

	val, err := cli.GetVar(key)
	if err == nil {
		showSuccess(val)
	} else {
		showError(logrus.Fields{"cmd": "get"}, "no such variable: %s\n", args[0])
		return
	}
}

func (cli *CLI) cmdWatch(cmd *cobra.Command, args []string) {
	address := args[0]

	watcher, err := cli.ms.WatchPayments(address)

	if err != nil {
		showError(logrus.Fields{"cmd": "watch"}, "can't watch address: %+v\n", err)
		return
	}

	for p := range watcher.Ch {
		showSuccess("%v %v from %v to %v", p.Amount, p.AssetCode, p.From, p.To)
	}

	if watcher.Err != nil {
		showError(logrus.Fields{"cmd": "watch"}, "%+v\n", *watcher.Err)
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
		showError(logrus.Fields{"cmd": "balance"}, "payment failed: %v", microstellar.ErrorString(err))
		// must return
	} else {
		showSuccess(account.GetNativeBalance())
	}
}
