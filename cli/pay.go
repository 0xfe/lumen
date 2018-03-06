package cli

import (
	"strconv"

	"github.com/0xfe/microstellar"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (cli *CLI) getPayCmd() *cobra.Command {
	payCmd := &cobra.Command{
		Use:   "pay [amount] [asset] --from [source] --to [target]",
		Short: "send [amount] of [asset] from [source] to [target]",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fields := logrus.Fields{"cmd": "pay"}
			amount := args[0]
			to, _ := cmd.Flags().GetString("to")
			from, _ := cmd.Flags().GetString("from")

			source := cli.validateAddressOrSeed(fields, from, "seed")
			target := cli.validateAddressOrSeed(fields, to, "address")

			opts := microstellar.Opts()

			if memotext, err := cmd.Flags().GetString("memotext"); err == nil && memotext != "" {
				opts = opts.WithMemoText(memotext)
			}

			if memoid, err := cmd.Flags().GetString("memoid"); err == nil && memoid != "" {
				id, err := strconv.ParseUint(memoid, 10, 64)
				if err != nil {
					logrus.WithFields(fields).Debugf("memoid ParseUint: %v", err)
					showError(fields, "bad memoid: %v", memoid)
					return
				}
				opts = opts.WithMemoID(id)
			}

			fund, err := cmd.Flags().GetBool("fund")

			if fund {
				logrus.WithFields(fields).Debugf("initial fund from %s to %s, opts: %+v", source, target, opts)
				err = cli.ms.FundAccount(source, target, amount, opts)
			} else {

				logrus.WithFields(fields).Debugf("payment from %s to %s, opts: %+v", source, target, opts)
				err = cli.ms.PayNative(source, target, amount) //, opts)
			}

			if err != nil {
				showError(fields, "payment failed: %v", microstellar.ErrorString(err))
			}
		},
	}

	payCmd.Flags().String("from", "", "source account seed or name")
	payCmd.Flags().String("to", "", "target account address or name")
	payCmd.Flags().String("memotext", "", "memo text")
	payCmd.Flags().String("memoid", "", "memo ID")
	payCmd.Flags().Bool("fund", false, "fund a new account")

	payCmd.MarkFlagRequired("from")
	payCmd.MarkFlagRequired("to")
	return payCmd
}
