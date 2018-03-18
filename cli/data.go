package cli

import (
	"github.com/0xfe/microstellar"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildDataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data [account] [key] [value] [--clear]",
		Short: "get, set, or remove data records on an account",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			logFields := logrus.Fields{"cmd": "data"}
			account := args[0]
			key := args[1]
			val := ""

			seed, err := cli.ResolveAccount(logFields, account, "seed")

			if err != nil {
				cli.error(logFields, "invalid account: %s", account)
				return
			}

			opts, err := cli.genTxOptions(cmd, logFields)
			if err != nil {
				cli.error(logFields, "can't generate transaction: %v", err)
				return
			}

			if len(args) > 2 {
				val = args[2]
			}

			clear, _ := cmd.Flags().GetBool("clear")

			if clear {
				err = cli.ms.ClearData(seed, key, opts)
			} else if val != "" {
				err = cli.ms.SetData(seed, key, []byte(val), opts)
			} else {
				address, err := cli.ResolveAccount(logFields, account, "address")
				if err != nil {
					cli.error(logFields, "invalid account: %s", account)
					return
				}

				a, err := cli.ms.LoadAccount(address)
				if err != nil {
					cli.error(logFields, "could not load account %s: %v", account, microstellar.ErrorString(err))
					return
				}

				val, ok := a.GetData(key)
				if !ok {
					cli.error(logFields, "key not found: %s", key)
					return
				} else {
					showSuccess(string(val))
				}
			}

			if err != nil {
				cli.error(logFields, "failed to update data for %s (%s): %v", account, key, microstellar.ErrorString(err))
				return
			}
		},
	}

	cmd.Flags().Bool("clear", false, "remove data associated with key")

	buildFlagsForTxOptions(cmd)
	return cmd
}
