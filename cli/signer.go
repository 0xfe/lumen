package cli

import (
	"strconv"

	"github.com/0xfe/microstellar"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildSignerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signer [add|remove|thresholds]",
		Short: "manage signers on account",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				cli.error(logrus.Fields{"cmd": "signer"}, "unrecognized signer command: %s, expecting: add|remove", args[0])
				return
			}
		},
	}

	cmd.AddCommand(cli.buildSignerAddCmd())
	cmd.AddCommand(cli.buildSignerRemoveCmd())
	cmd.AddCommand(cli.buildSignerThresholdsCmd())

	return cmd
}

func (cli *CLI) buildSignerAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [signer_address] [weight] --to [account]",
		Short: "add signer_address as a signer on [account] with key weight [weight]",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			signerAddress := args[0]
			weight := args[1]

			logFields := logrus.Fields{"cmd": "signer", "subcmd": "add"}
			signer, err := cli.validateAddressOrSeed(logFields, signerAddress, "address")

			if err != nil {
				cli.error(logFields, "invalid account: %s", signerAddress)
				return
			}

			to, _ := cmd.Flags().GetString("to")

			signee, err := cli.validateAddressOrSeed(logFields, to, "seed")

			if err != nil {
				cli.error(logFields, "invalid signee: %s", to)
				return
			}

			opts, err := cli.genTxOptions(cmd, logFields)
			if err != nil {
				cli.error(logFields, "can't generate transaction: %v", err)
				return
			}

			intWeight, err := strconv.ParseUint(weight, 10, 32)
			if err != nil {
				cli.error(logFields, "invalid weight: %s", weight)
				return
			}

			err = cli.ms.AddSigner(signee, signer, uint32(intWeight), opts)
			if err != nil {
				cli.error(logFields, "failed to add signer %s to %s: %v", signerAddress, to, microstellar.ErrorString(err))
				return
			}
		},
	}

	cmd.Flags().String("to", "", "account seed of signee")
	cmd.MarkFlagRequired("to")

	buildFlagsForTxOptions(cmd)
	return cmd
}

func (cli *CLI) buildSignerRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [signer_address] --from [account]",
		Short: "remove signer_address as a signer from [account]",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			signerAddress := args[0]

			logFields := logrus.Fields{"cmd": "signer", "subcmd": "remove"}
			signer, err := cli.validateAddressOrSeed(logFields, signerAddress, "address")

			if err != nil {
				cli.error(logFields, "invalid account: %s", signerAddress)
				return
			}

			from, _ := cmd.Flags().GetString("from")

			signee, err := cli.validateAddressOrSeed(logFields, from, "seed")

			if err != nil {
				cli.error(logFields, "invalid signee: %s", from)
				return
			}

			opts, err := cli.genTxOptions(cmd, logFields)
			if err != nil {
				cli.error(logFields, "can't generate transaction: %v", err)
				return
			}

			err = cli.ms.RemoveSigner(signee, signer, opts)
			if err != nil {
				cli.error(logFields, "failed to remove signer %s from %s: %v", signerAddress, from, microstellar.ErrorString(err))
				return
			}
		},
	}

	cmd.Flags().String("from", "", "account seed of signee")
	cmd.MarkFlagRequired("from")

	buildFlagsForTxOptions(cmd)
	return cmd
}

func (cli *CLI) buildSignerThresholdsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "thresholds [account] [low] [medium] [high]",
		Short: "set low, medium, and high thresholds for [account]",
		Args:  cobra.ExactArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			account := args[0]
			lowString := args[1]
			mediumString := args[2]
			highString := args[3]

			logFields := logrus.Fields{"cmd": "signer", "subcmd": "thresholds"}
			address, err := cli.validateAddressOrSeed(logFields, account, "seed")

			if err != nil {
				cli.error(logFields, "invalid account: %s", account)
				return
			}

			low, err := strconv.ParseUint(lowString, 10, 32)
			if err != nil {
				logrus.WithFields(logFields).Errorf("threshold parse error: %v", err)
				cli.error(logFields, "bad threshold (low): %s", lowString)
				return
			}

			medium, err := strconv.ParseUint(mediumString, 10, 32)
			if err != nil {
				logrus.WithFields(logFields).Errorf("threshold parse error: %v", err)
				cli.error(logFields, "bad threshold (medium): %s", mediumString)
				return
			}

			high, err := strconv.ParseUint(highString, 10, 32)
			if err != nil {
				logrus.WithFields(logFields).Errorf("threshold parse error: %v", err)
				cli.error(logFields, "bad threshold (high): %s", highString)
				return
			}

			opts, err := cli.genTxOptions(cmd, logFields)
			if err != nil {
				cli.error(logFields, "can't generate transaction: %v", err)
				return
			}

			err = cli.ms.SetThresholds(address, uint32(low), uint32(medium), uint32(high), opts)
			if err != nil {
				cli.error(logFields, "failed to set thresholds for %s: %v", account, microstellar.ErrorString(err))
				return
			}
		},
	}

	buildFlagsForTxOptions(cmd)
	return cmd
}
