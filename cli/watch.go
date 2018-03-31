package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/0xfe/microstellar"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func showEntry(logFields logrus.Fields, entry interface{}, format string) {
	if format == "json" {
		data, err := json.MarshalIndent(entry, "", "  ")

		if err != nil {
			logrus.WithFields(logFields).Errorf("skipping bad data: %v", err)
		} else {
			showSuccess("%v", string(data))
		}
	} else {
		showSuccess("%+v", entry)
	}
}

func showPayment(logFields logrus.Fields, payment *microstellar.Payment) {
	memo := ""
	if payment.Memo.Type != "none" {
		memo = fmt.Sprintf(" (memo: %v)", payment.Memo.Value)
	}

	if payment.Type == "create_account" {
		showSuccess("create_account: %v funded with %v lumens %v", payment.Account, payment.StartingBalance, memo)
	} else if payment.Type == "payment" {
		showSuccess("payment: %v %v from %v to %v %v", payment.Amount, payment.AssetCode, payment.From, payment.To, memo)
	}
}

func watch(ms *microstellar.MicroStellar, logFields logrus.Fields, entity string, address string, format string, stopFunc *func(), opts *microstellar.Options) error {
	var watcher interface{}
	var err error
	var streamErr *error

	for err == nil {
		switch entity {
		case "payments":
			watcher, err = ms.WatchPayments(address, opts)
			*stopFunc = watcher.(*microstellar.PaymentWatcher).Done
			streamErr = watcher.(*microstellar.PaymentWatcher).Err
			for entry := range watcher.(*microstellar.PaymentWatcher).Ch {
				if format == "line" {
					showPayment(logFields, entry)
				} else {
					showEntry(logFields, entry, format)
				}
			}
		case "transactions":
			watcher, err = ms.WatchTransactions(address, opts)
			*stopFunc = watcher.(*microstellar.TransactionWatcher).Done
			streamErr = watcher.(*microstellar.TransactionWatcher).Err
			for entry := range watcher.(*microstellar.TransactionWatcher).Ch {
				showEntry(logFields, entry, format)
			}
		case "ledger":
			watcher, err = ms.WatchLedgers(opts)
			*stopFunc = watcher.(*microstellar.LedgerWatcher).Done
			streamErr = watcher.(*microstellar.LedgerWatcher).Err
			for entry := range watcher.(*microstellar.LedgerWatcher).Ch {
				showEntry(logFields, entry, format)
			}
		default:
			return errors.Errorf("invalid watch entity: %s", entity)
		}

		if *streamErr != nil {
			debugf(logFields, "connection closed: %v", *streamErr)
		}

		if err != nil {
			return errors.Wrapf(err, "can't watch address: %v", microstellar.ErrorString(err))
		}

		debugf(logFields, "retrying in 2s...")
		time.Sleep(2 * time.Second)
	}

	return nil
}

func (cli *CLI) buildWatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch [payments|transactions|ledger] [account]",
		Short: "watch the account on the ledger",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			entity := args[0]
			address := ""

			logFields := logrus.Fields{"cmd": "watch"}

			if len(args) > 1 {
				name := args[1]
				var err error
				address, err = cli.ResolveAccount(logFields, name, "address")

				if err != nil {
					cli.error(logFields, "invalid address: %s", name)
					return
				}
			}

			opts := microstellar.Opts()
			cursor, _ := cmd.Flags().GetString("cursor")

			if cursor != "start" {
				opts = opts.WithCursor(cursor)
			}

			format, _ := cmd.Flags().GetString("format")
			err := watch(cli.ms, logFields, entity, address, format, &cli.stopWatcher, opts)

			if err != nil {
				cli.error(logFields, "can't watch stream: %v", microstellar.ErrorString(err))
				return
			}
		},
	}

	cmd.Flags().String("format", "line", "output format (json, yaml, struct)")
	cmd.Flags().String("cursor", "now", "start watching from (now, start, paging_token)")

	return cmd
}
