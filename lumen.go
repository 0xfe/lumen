package lumen

import (
	"fmt"
	"os"

	"github.com/0xfe/microstellar"
	"github.com/spf13/cobra"
)

type Lumen struct {
	store StoreAPI
}

// NewState returns a new initialized state.
func NewLumen(store StoreAPI) *Lumen {
	return &Lumen{
		store: store,
	}
}

func (l *Lumen) SetVar(key string, value string) error {
	return l.store.Set(key, value, 0)
}

func (l *Lumen) GetVar(key string) (string, error) {
	return l.store.Get(key)
}

var rootCmd = &cobra.Command{
	Use:   "lumen",
	Short: "Lumen is a MicroStellar-based commandline client for the Stellar blockchain based",
	Run:   root,
}

func root(cmd *cobra.Command, args []string) {
	fmt.Println(cmd)
}

func showSuccess(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
}

func showError(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, args)
}

func Start() {
	store, _ := NewStore("internal", "")
	lumen := NewLumen(store)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Get version of lumen CLI",
		Run: func(cmd *cobra.Command, args []string) {
			showSuccess("v0.1")
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "set [key] [value]",
		Short: "set variable",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			showSuccess("setting %s to %s\n", args[0], args[1])
			lumen.SetVar(args[0], args[1])
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "get [key]",
		Short: "get variable",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			val, err := lumen.GetVar(args[0])
			if err == nil {
				showSuccess(val)
			} else {
				showError("no such variable: %s\n", args[0])
			}

		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "watch [address]",
		Short: "watch the address on the ledger",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			address := args[0]

			ms := microstellar.New("test")
			watcher, err := ms.WatchPayments(address)

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
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "pay [source] [target] [amount]",
		Short: "pay [amount] lumens from [source] to [target]",
		Args:  cobra.MinimumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			source := args[0]
			target := args[1]
			amount := args[2]

			ms := microstellar.New("test")
			err := ms.PayNative(source, target, amount)

			if err != nil {
				showError("payment failed: %v\n", err)
			} else {
				showSuccess("paid\n")
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "balance [address]",
		Short: "get the balance of [address] in lumens",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			address := args[0]

			ms := microstellar.New("test")
			account, err := ms.LoadAccount(address)

			if err != nil {
				showError("payment failed: %v\n", err)
			} else {
				showSuccess("%v\n", account.GetNativeBalance())
			}
		},
	})

	rootCmd.Execute()
}
