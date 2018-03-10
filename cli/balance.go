package cli

import (
	"github.com/0xfe/microstellar"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance [account] [asset]",
		Short: "check the balance of [asset] on [account]",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			asset := microstellar.NativeAsset

			logFields := logrus.Fields{"cmd": "balance"}
			address, err := cli.validateAddressOrSeed(logFields, name, "address")

			if err != nil {
				cli.error(logFields, "invalid address: %s", name)
				return
			}

			if len(args) > 1 {
				assetName := args[1]
				asset, err = cli.GetAsset(assetName)

				if err != nil {
					cli.error(logFields, "bad asset: %s", assetName)
					return
				}
			}

			account, err := cli.ms.LoadAccount(address)

			if err != nil {
				cli.error(logFields, "can't load account: %v", microstellar.ErrorString(err))
				return
			}

			balance := account.GetBalance(asset)

			if balance == "" {
				showSuccess("0")
			} else {
				showSuccess(balance)
			}
		},
	}

	return cmd
}
