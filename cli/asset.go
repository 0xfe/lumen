package cli

import (
	"fmt"

	"github.com/0xfe/microstellar"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildAssetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset [set|del|code|issuer|type]",
		Short: "manage stellar assets",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				cli.error(logrus.Fields{"cmd": "asset"}, "unrecognized asset command: %s, expecting: set|del|code|issuer|type", args[0])
				return
			}
		},
	}

	cmd.AddCommand(cli.buildAssetSetCmd())
	cmd.AddCommand(cli.buildAssetCodeCmd())
	cmd.AddCommand(cli.buildAssetIssuerCmd())
	cmd.AddCommand(cli.buildAssetTypeCmd())
	cmd.AddCommand(cli.buildAssetDelCmd())

	return cmd
}

func (cli *CLI) buildAssetSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [name] [issuer]",
		Short: "set asset issuer of asset [name]",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			issuer := args[1]
			code := name
			assetType := string(microstellar.NativeType)

			for _, part := range []string{"issuer", "code", "type"} {
				key := fmt.Sprintf("asset:%s:%s", name, part)
				value := ""

				if part == "issuer" {
					value = issuer
					if microstellar.ValidAddress(issuer) != nil {
						var err error
						value, err = cli.GetAccount(issuer, "address")
						if err != nil {
							cli.error(logrus.Fields{"cmd": "asset", "subcmd": "set"}, "invalid issuer: %s", issuer)
						}
					}
				}

				if part == "code" {
					if cmd.Flag("code").Changed {
						code, _ = cmd.Flags().GetString("code")
					}

					value = code
				}

				if part == "type" {
					if cmd.Flag("type").Changed {
						assetType, _ = cmd.Flags().GetString("type")
						switch assetType {
						case
							string(microstellar.Credit12Type),
							string(microstellar.Credit4Type),
							string(microstellar.NativeType):
							break
						default:
							cli.error(logrus.Fields{"cmd": "asset", "subcmd": "set"}, "bad asset type: %s", assetType)
							return
						}
					} else {
						if len(code) > 4 {
							assetType = string(microstellar.Credit12Type)
						} else {
							assetType = string(microstellar.Credit4Type)
						}
					}

					value = assetType
				}

				logrus.WithFields(logrus.Fields{"cmd": "asset", "subcmd": "set"}).Debugf("saving asset %s: %s %s", name, part, value)
				err := cli.SetVar(key, value)

				if err != nil {
					logrus.WithFields(logrus.Fields{"cmd": "asset", "subcmd": "set"}).Debugf("%v", err)
					cli.error(logrus.Fields{"cmd": "asset", "subcmd": "set"}, "could not save asset: %s", name)
					return
				}
			}
		},
	}

	cmd.Flags().String("code", "XLM", "specify asset code")
	cmd.Flags().String("type", string(microstellar.NativeType), "specify asset type (credit_alphanum4, credit_alphanum12, native)")

	return cmd
}

func (cli *CLI) buildAssetCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "code [name]",
		Short: "get asset code of [name]",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			if asset, err := cli.GetAsset(name); err != nil {
				logrus.WithFields(logrus.Fields{"cmd": "asset", "subcmd": "code"}).Debugf("%v", err)
				cli.error(logrus.Fields{"cmd": "asset", "subcmd": "code"}, "could not load asset: %s", name)
				return
			} else {
				showSuccess(asset.Code)
			}
		},
	}

	return cmd
}

func (cli *CLI) buildAssetIssuerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issuer [name]",
		Short: "get asset issuer of [name]",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			if asset, err := cli.GetAsset(name); err != nil {
				logrus.WithFields(logrus.Fields{"cmd": "asset", "subcmd": "issuer"}).Debugf("%v", err)
				cli.error(logrus.Fields{"cmd": "asset", "subcmd": "issuer"}, "could not load asset: %s", name)
				return
			} else {
				showSuccess(asset.Issuer)
			}
		},
	}

	return cmd
}

func (cli *CLI) buildAssetTypeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "type [name]",
		Short: "get asset type of [name]",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			if asset, err := cli.GetAsset(name); err != nil {
				logrus.WithFields(logrus.Fields{"cmd": "asset", "subcmd": "type"}).Debugf("%v", err)
				cli.error(logrus.Fields{"cmd": "asset", "subcmd": "type"}, "could not load asset: %s", name)
				return
			} else {
				assetType := microstellar.NativeType
				if asset.Type == microstellar.Credit4Type {
					assetType = microstellar.Credit4Type
				} else if asset.Type == microstellar.Credit12Type {
					assetType = microstellar.Credit12Type
				}
				showSuccess(string(assetType))
			}
		},
	}

	return cmd
}

func (cli *CLI) buildAssetDelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "del [name]",
		Short: "delete asset named [name]",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			for _, part := range []string{"issuer", "code", "type"} {
				key := fmt.Sprintf("asset:%s:%s", name, part)
				cli.DelVar(key)
			}
		},
	}

	return cmd
}
