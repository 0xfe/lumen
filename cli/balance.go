package cli

import (
	"encoding/json"

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

			if len(args) > 1 {
				var err error
				assetName := args[1]
				asset, err = cli.ResolveAsset(assetName)

				if err != nil {
					cli.error(logFields, "bad asset: %s", assetName)
					return
				}
			}

			account := cli.LoadAccount(logFields, name)
			if account == nil {
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

func (cli *CLI) buildInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info [account]",
		Short: "get account info",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			logFields := logrus.Fields{"cmd": "balance"}
			account := cli.LoadAccount(logFields, name)
			if account == nil {
				return
			}

			info, _ := json.MarshalIndent(*account, "", "  ")
			showSuccess(string(info))
		},
	}

	return cmd
}
