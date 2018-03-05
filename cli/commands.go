package cli

import (
	"github.com/0xfe/lumen/store"
	"github.com/0xfe/microstellar"
	"github.com/spf13/cobra"
)

type CLI struct {
	store store.API
	ms    *microstellar.MicroStellar
}

// NewCLI
func NewCLI(store store.API, ms *microstellar.MicroStellar) *CLI {
	return &CLI{
		store: store,
		ms:    ms,
	}
}

func (cli *CLI) Install(rootCmd *cobra.Command) {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Get version of lumen CLI",
		Run:   cli.cmdVersion,
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
		Use:   "pay [source] [target] [amount]",
		Short: "pay [amount] lumens from [source] to [target]",
		Args:  cobra.MinimumNArgs(3),
		Run:   cli.cmdPay,
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "balance [address]",
		Short: "get the balance of [address] in lumens",
		Args:  cobra.MinimumNArgs(1),
		Run:   cli.cmdBalance,
	})

	rootCmd.AddCommand(cli.getAccountsCmd())
}

func (cli *CLI) SetVar(key string, value string) error {
	return cli.store.Set(key, value, 0)
}

func (cli *CLI) GetVar(key string) (string, error) {
	return cli.store.Get(key)
}

func (cli *CLI) DelVar(key string) error {
	return cli.store.Delete(key)
}

func (cli *CLI) cmdVersion(cmd *cobra.Command, args []string) {
	showSuccess("v0.1")
}

func (cli *CLI) cmdSet(cmd *cobra.Command, args []string) {
	showSuccess("setting %s to %s\n", args[0], args[1])
	cli.SetVar(args[0], args[1])
}

func (cli *CLI) cmdDel(cmd *cobra.Command, args []string) {
	err := cli.DelVar(args[0])
	if err != nil {
		showError("del failed: %s\n", err)
	}
}

func (cli *CLI) cmdGet(cmd *cobra.Command, args []string) {
	val, err := cli.GetVar(args[0])
	if err == nil {
		showSuccess(val + "\n")
	} else {
		showError("no such variable: %s\n", args[0])
	}
}

func (cli *CLI) cmdWatch(cmd *cobra.Command, args []string) {
	address := args[0]

	watcher, err := cli.ms.WatchPayments(address)

	if err != nil {
		showError("can't watch address: %+v\n", err)
		return
	}

	for p := range watcher.Ch {
		showSuccess("%v %v from %v to %v\n", p.Amount, p.AssetCode, p.From, p.To)
	}

	if watcher.Err != nil {
		showError("%+v\n", *watcher.Err)
	}
}

func (cli *CLI) cmdPay(cmd *cobra.Command, args []string) {
	source := args[0]
	target := args[1]
	amount := args[2]

	err := cli.ms.PayNative(source, target, amount)

	if err != nil {
		showError("payment failed: %v\n", err)
	} else {
		showSuccess("paid\n")
	}
}

func (cli *CLI) cmdBalance(cmd *cobra.Command, args []string) {
	address := args[0]

	account, err := cli.ms.LoadAccount(address)

	if err != nil {
		showError("payment failed: %v\n", err)
	} else {
		showSuccess("%v\n", account.GetNativeBalance())
	}
}
