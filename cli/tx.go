package cli

import (
	"encoding/json"

	"github.com/0xfe/microstellar"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx [sign|submit] [base64-encoded string] --signers seed1,seed2...",
		Short: "handle base64 encoded transactions",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				showError(logrus.Fields{"cmd": "tx"}, "unrecognized tx command: %s, expecting: sign|submit", args[0])
				return
			}
		},
	}

	cmd.AddCommand(cli.buildTxSignCmd())
	cmd.AddCommand(cli.buildTxSubmitCmd())
	cmd.AddCommand(cli.buildTxDecodeCmd())

	return cmd
}

func (cli *CLI) buildTxSignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign [base64-encoded transaction] --signers seed1,seed2...",
		Short: "sign the supplied transaction (on the current network) with the given seeds (or accounts)",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			b64tx := args[0]

			logFields := logrus.Fields{"cmd": "sign"}
			signers, err := cmd.Flags().GetStringSlice("signers")

			if err != nil {
				cli.error(logFields, "can't get signers: %v", err)
				return
			}

			if len(signers) < 1 {
				cli.error(logFields, "need at least one seed in --signers")
				return
			}

			var seeds []string

			for _, signer := range signers {
				seed, err := cli.ResolveAccount(logFields, signer, "seed")

				if err != nil {
					cli.error(logFields, "bad signer account: %v", signer)
					return
				}

				if microstellar.ValidSeed(seed) != nil {
					cli.error(logFields, "no seed found in %v", signer)
					return
				}

				seeds = append(seeds, seed)
			}

			signedTx, err := cli.ms.SignTransaction(b64tx, seeds...)

			if err != nil {
				cli.error(logFields, "signing error: %v", err)
				return
			}

			showSuccess(signedTx)
		},
	}

	buildFlagsForTxOptions(cmd)
	return cmd
}

func (cli *CLI) buildTxSubmitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit [base64-encoded transaction]",
		Short: "submit the supplied transaction to the current network",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			b64tx := args[0]

			logFields := logrus.Fields{"cmd": "submit"}
			resp, err := cli.ms.SubmitTransaction(b64tx)

			if err != nil {
				cli.error(logFields, "submit error: %v", microstellar.ErrorString(err))
				return
			}

			respJSON, _ := json.MarshalIndent(*resp, "", "  ")
			showSuccess(string(respJSON))
		},
	}

	return cmd
}

func (cli *CLI) buildTxDecodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decode [base64-encoded transaction] [--pretty]",
		Short: "display the base64-encoded transaction in JSON",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			b64tx := args[0]

			logFields := logrus.Fields{"cmd": "decode"}
			pretty, _ := cmd.Flags().GetBool("pretty")
			txe, err := microstellar.DecodeTxToJSON(b64tx, pretty)

			if err != nil {
				cli.error(logFields, "decode error: %v", microstellar.ErrorString(err))
				return
			}

			showSuccess(txe)
		},
	}

	cmd.Flags().Bool("pretty", false, "format JSON output")
	return cmd
}
