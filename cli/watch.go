package cli

import (
	"encoding/json"

	"github.com/0xfe/microstellar"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildWatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch [account]",
		Short: "watch the account on the ledger",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			logFields := logrus.Fields{"cmd": "watch"}

			address, err := cli.ResolveAccount(logFields, name, "address")

			if err != nil {
				cli.error(logFields, "invalid address: %s", name)
				return
			}

			opts := microstellar.Opts()
			cursor, _ := cmd.Flags().GetString("cursor")

			if cursor != "start" {
				opts = opts.WithCursor(cursor)
			}

			watcher, err := cli.ms.WatchPayments(address, opts)

			if err != nil {
				cli.error(logFields, "can't watch address: %v", microstellar.ErrorString(err))
				return
			}

			cli.stopWatcher = watcher.Done
			format, err := cmd.Flags().GetString("format")

			for p := range watcher.Ch {
				if format == "json" {
					data, err := json.MarshalIndent(p, "", "  ")

					if err != nil {
						logrus.WithFields(logFields).Errorf("skipping bad data: %v", err)
					} else {
						showSuccess("%v", string(data))
					}
				} else {
					showSuccess("%+v", p)
				}
			}

			if watcher.Err != nil {
				cli.error(logrus.Fields{"cmd": "watch"}, "%v", *watcher.Err)
				return
			}
		},
	}

	cmd.Flags().String("format", "json", "output format (json, yaml, struct)")
	cmd.Flags().String("cursor", "now", "start watching from (now, start, paging_token)")

	return cmd
}
