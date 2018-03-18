package cli

import (
	"github.com/0xfe/microstellar"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildPayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pay [amount] [asset] --from [source] --to [target]",
		Short: "send [amount] of [asset] from [source] to [target]",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fields := logrus.Fields{"cmd": "pay"}
			amount := args[0]
			assetName := ""
			if len(args) > 1 {
				assetName = args[1]
			}

			asset, err := cli.ResolveAsset(assetName)
			if err != nil {
				logrus.WithFields(fields).Debugf("could not get asset %s: %v", assetName, err)
				cli.error(fields, "bad asset: %s", assetName)
				return
			}

			to, _ := cmd.Flags().GetString("to")
			from, _ := cmd.Flags().GetString("from")
			source, err := cli.ResolveAccount(fields, from, "seed")
			if err != nil {
				cli.error(fields, "bad --from address: %s", from)
				return
			}

			target, err := cli.ResolveAccount(fields, to, "address")
			if err != nil {
				cli.error(fields, "bad --to address: %s", to)
				return
			}

			opts, err := cli.genTxOptions(cmd, fields)
			if err != nil {
				cli.error(fields, "can't generate payment: %v", err)
				return
			}

			// Is this a fund request?
			fund, err := cmd.Flags().GetBool("fund")
			debugf(fields, "fund: %v, err %v", fund, err)

			// If --with is set, then this is a path payment
			with, _ := cmd.Flags().GetString("with")
			if with != "" {
				max, _ := cmd.Flags().GetString("max")
				path, _ := cmd.Flags().GetStringSlice("path")

				var withAsset *microstellar.Asset
				var assetPath []*microstellar.Asset

				withAsset, err = cli.ResolveAsset(with)
				if err != nil {
					cli.error(fields, "bad --with asset: %s", with)
					return
				}

				if max == "" {
					cli.error(fields, "--max is required for path payments")
					return
				}

				if len(path) > 0 {
					debugf(fields, "path payment with %s (max %s) through %+v", with, max, path)
					for _, a := range path {
						pathAsset, err := cli.ResolveAsset(a)
						if err != nil {
							cli.error(fields, "bad --path asset: %s", a)
							return
						}

						assetPath = append(assetPath, pathAsset)
					}

					opts = opts.WithAsset(withAsset, max).Through(assetPath...)
				} else {
					debugf(fields, "path payment with %s (max %s) using pathfinder", with, max)
					sourceAddress, err := cli.ResolveAccount(fields, from, "address")
					if err != nil {
						cli.error(fields, "no address in --from: %s", from)
						return
					}

					debugf(fields, "searching for paths from: %s", sourceAddress)
					opts = opts.WithAsset(withAsset, max).FindPathFrom(sourceAddress)
				}
			}

			if fund {
				logrus.WithFields(fields).Debugf("initial fund from %s to %s, opts: %+v", source, target, opts)
				err = cli.ms.FundAccount(source, target, amount, opts)
			} else {
				logrus.WithFields(fields).Debugf("paying %s %s/%s from %s to %s, opts: %+v", amount, asset.Code, asset.Issuer, source, target, opts)
				err = cli.ms.Pay(source, target, amount, asset, opts)
			}

			if err != nil {
				cli.error(fields, "payment failed: %v", microstellar.ErrorString(err))
				return
			}
		},
	}

	buildFlagsForTxOptions(cmd)
	cmd.Flags().String("from", "", "source account seed or name")
	cmd.Flags().String("to", "", "target account address or name")
	cmd.Flags().String("with", "", "make a path payment with this asset")
	cmd.Flags().String("max", "", "spend no more than this much during path payments")
	cmd.Flags().StringSlice("path", []string{}, "comma-separated list of paths, uses auto pathfinder if empty")

	cmd.Flags().Bool("fund", false, "fund a new account")
	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("to")

	return cmd
}
