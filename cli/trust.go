package cli

import (
	"github.com/0xfe/microstellar"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildTrustCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trust [create|remove|allow]",
		Short: "manage trustlines between accounts and assets",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				showError(logrus.Fields{"cmd": "trust"}, "unrecognized trust command: %s, expecting: create|remove", args[0])
				return
			}
		},
	}

	cmd.AddCommand(cli.buildTrustCreateCmd())
	cmd.AddCommand(cli.buildTrustRemoveCmd())
	cmd.AddCommand(cli.buildTrustAllowCmd())

	return cmd
}

func (cli *CLI) buildTrustCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [account] [asset] [limit]",
		Short: "create a new trustline to the asset for [account]",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			assetName := args[1]

			limit := ""
			if len(args) > 2 {
				limit = args[2]
			}

			logFields := logrus.Fields{"cmd": "trust", "subcmd": "create"}
			source, err := cli.ResolveAccount(logFields, name, "seed")

			if err != nil {
				cli.error(logFields, "invalid account: %s", name)
				return
			}

			asset, err := cli.ResolveAsset(assetName)
			if err != nil {
				cli.error(logFields, "invalid asset: %s", assetName)
				return
			}

			opts, err := cli.genTxOptions(cmd, logFields)
			if err != nil {
				cli.error(logFields, "can't generate trustline transaction: %v", err)
				return
			}

			err = cli.ms.CreateTrustLine(source, asset, limit, opts)
			if err != nil {
				cli.error(logFields, "failed to create trustline from %s to %s: %v", name, assetName, microstellar.ErrorString(err))
				return
			}
		},
	}

	buildFlagsForTxOptions(cmd)
	return cmd
}

func (cli *CLI) buildTrustRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [account] [asset]",
		Short: "remove the trustline between [account] and [asset]",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			assetName := args[1]

			logFields := logrus.Fields{"cmd": "trust", "subcmd": "remove"}
			source, err := cli.ResolveAccount(logFields, name, "seed")

			if err != nil {
				cli.error(logFields, "invalid account: %s", name)
				return
			}

			asset, err := cli.ResolveAsset(assetName)
			if err != nil {
				cli.error(logFields, "invalid asset: %s", assetName)
				return
			}

			opts, err := cli.genTxOptions(cmd, logFields)
			if err != nil {
				cli.error(logFields, "can't generate trustline transaction: %v", err)
				return
			}

			err = cli.ms.RemoveTrustLine(source, asset, opts)
			if err != nil {
				cli.error(logFields, "error: %v", name, assetName, microstellar.ErrorString(err))
				return
			}
		},
	}

	buildFlagsForTxOptions(cmd)
	return cmd
}

func (cli *CLI) buildTrustAllowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "allow [account] [asset] [--revoke] [--signers seed1,seed2...]",
		Short: "allow/revoke a trustline between account and asset",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			assetName := args[1]

			logFields := logrus.Fields{"cmd": "trust", "subcmd": "allow"}
			address, err := cli.ResolveAccount(logFields, name, "address")

			if err != nil {
				cli.error(logFields, "invalid account: %s", name)
				return
			}

			asset, err := cli.ResolveAsset(assetName)
			if err != nil {
				cli.error(logFields, "invalid asset: %s", assetName)
				return
			}

			opts, err := cli.genTxOptions(cmd, logFields)
			if err != nil {
				cli.error(logFields, "can't generate allowtrust transaction: %v", err)
				return
			}

			revoke, _ := cmd.Flags().GetBool("revoke")
			err = cli.ms.AllowTrust(asset.Issuer, address, asset.Code, !revoke, opts)
			if err != nil {
				cli.error(logFields, "failed to create trustline from %s to %s: %v", name, assetName, microstellar.ErrorString(err))
				return
			}
		},
	}

	cmd.Flags().Bool("revoke", false, "revoke authorization")
	buildFlagsForTxOptions(cmd)
	return cmd
}
