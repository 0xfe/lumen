package cli

import (
	"github.com/0xfe/microstellar"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildDexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dex [trade|list]",
		Short: "trade assets on the DEX",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				cli.error(logrus.Fields{"cmd": "signer"}, "unrecognized trade command: %s, expecting: trade|list", args[0])
				return
			}
		},
	}

	cmd.AddCommand(cli.buildDexTradeCmd())
	// cmd.AddCommand(cli.buildDexListCmd())

	return cmd
}

func (cli *CLI) buildDexTradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trade [account] --buy [asset1] --sell [asset2] --amount [sellAmount] --price [rate]",
		Short: "offer to sell [sellAmount] quantity of asset2 for asset1 at price [rate]",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			logFields := logrus.Fields{"cmd": "dex", "subcmd": "trade"}

			account := args[0]
			buy, _ := cmd.Flags().GetString("buy")
			sell, _ := cmd.Flags().GetString("sell")
			amount, _ := cmd.Flags().GetString("amount")
			price, _ := cmd.Flags().GetString("price")
			update, _ := cmd.Flags().GetString("update")
			delete, _ := cmd.Flags().GetString("delete")
			isPassive, _ := cmd.Flags().GetBool("passive")

			source, err := cli.validateAddressOrSeed(logFields, account, "seed")
			if err != nil {
				cli.error(logFields, "invalid account: %s", account)
				return
			}

			buyAsset, err := cli.GetAsset(buy)
			if err != nil {
				cli.error(logFields, "invalid buy asset: %s", buy)
			}

			sellAsset, err := cli.GetAsset(sell)
			if err != nil {
				cli.error(logFields, "invalid sell asset: %s", sell)
			}

			offerType := microstellar.OfferCreate
			offerID := ""

			if update != "" {
				offerType = microstellar.OfferUpdate
				offerID = update
			} else if delete != "" {
				offerType = microstellar.OfferDelete
				offerID = delete
			} else if isPassive {
				offerType = microstellar.OfferCreatePassive
			}

			opts, err := cli.genTxOptions(cmd, logFields)
			if err != nil {
				cli.error(logFields, "can't generate offer: %v", err)
				return
			}

			err = cli.ms.ManageOffer(source, &microstellar.OfferParams{
				OfferType:  offerType,
				SellAsset:  sellAsset,
				SellAmount: amount,
				BuyAsset:   buyAsset,
				Price:      price,
				OfferID:    offerID,
			}, opts)

			if err != nil {
				cli.error(logFields, "failed to submit offer: %v", microstellar.ErrorString(err))
				return
			}
		},
	}

	cmd.Flags().String("buy", "", "asset to buy")
	cmd.Flags().String("sell", "", "asset to sell")
	cmd.Flags().String("amount", "", "amount to sell")
	cmd.Flags().String("price", "", "price in units-of-buy per unit-of-sell")
	cmd.Flags().String("update", "", "Offer ID to update")
	cmd.Flags().String("delete", "", "Offer ID to delete")
	cmd.Flags().Bool("passive", false, "make this a passive offer")

	cmd.MarkFlagRequired("buy")
	cmd.MarkFlagRequired("sell")
	cmd.MarkFlagRequired("amount")
	cmd.MarkFlagRequired("price")

	buildFlagsForTxOptions(cmd)
	return cmd
}
