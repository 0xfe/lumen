package cli

import (
	"github.com/0xfe/microstellar"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (cli *CLI) getPayCmd() *cobra.Command {
	payCmd := &cobra.Command{
		Use:   "pay [amount] [asset] --from [source] --to [target]",
		Short: "pay [amount] lumens from [source] to [target]",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fields := logrus.Fields{"cmd": "pay"}
			amount := args[0]
			to, _ := cmd.Flags().GetString("to")
			from, _ := cmd.Flags().GetString("from")

			source := cli.validateAddressOrSeed(fields, from, "seed")
			target := cli.validateAddressOrSeed(fields, to, "address")

			opts := microstellar.Opts()
			memotext, err := cmd.Flags().GetString("memotext")

			if err != nil {
				opts = opts.WithMemoText(memotext)
			}

			err = cli.ms.PayNative(source, target, amount, opts)

			if err != nil {
				showError(fields, "payment failed: %v", microstellar.ErrorString(err))
			}
		},
	}

	payCmd.Flags().String("from", "", "source account seed or name")
	payCmd.Flags().String("to", "", "target account address or name")
	payCmd.Flags().String("memotext", "", "memo text")

	payCmd.MarkFlagRequired("from")
	payCmd.MarkFlagRequired("to")
	return payCmd
}
