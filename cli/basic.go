package cli

// This file contains implementations of some basic
// commands.

import (
	"fmt"

	"github.com/0xfe/microstellar"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "get version of lumen CLI",
		Run: func(cmd *cobra.Command, args []string) {
			showSuccess(cli.version)
		},
	}

	return cmd
}

func (cli *CLI) buildNSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ns [namespace]",
		Short: "set namespace to [namespace]",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				ns := args[0]

				err := cli.SetGlobalVar("ns", ns)
				if err != nil {
					cli.error(logrus.Fields{"cmd": "setNS"}, "set failed: ", err)
					return
				}

				cli.ns = ns
			} else {
				showSuccess(cli.ns)
			}
		},
	}
	return cmd
}

func (cli *CLI) buildSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "set variable",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			key := fmt.Sprintf("vars:%s", args[0])
			val := args[1]

			err := cli.SetVar(key, val)
			if err != nil {
				cli.error(logrus.Fields{"cmd": "set"}, "set failed: ", err)
				return
			}
		},
	}

	return cmd
}

func (cli *CLI) buildGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [key]",
		Short: "get variable",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			key := fmt.Sprintf("vars:%s", args[0])

			val, err := cli.GetVar(key)
			if err == nil {
				showSuccess(val)
			} else {
				cli.error(logrus.Fields{"cmd": "get"}, "no such variable: %s\n", args[0])
				return
			}
		},
	}

	return cmd
}

func (cli *CLI) buildDelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "del [key]",
		Short: "delete variable",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			key := fmt.Sprintf("vars:%s", args[0])

			err := cli.DelVar(key)
			if err != nil {
				cli.error(logrus.Fields{"cmd": "del"}, "del failed: %s\n", err)
				return
			}
		},
	}
	return cmd
}

func (cli *CLI) buildFriendbotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "friendbot [address]",
		Short: "fund [address] on the test network with friendbot",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			logFields := logrus.Fields{"cmd": "trust", "subcmd": "create"}
			address, err := cli.validateAddressOrSeed(logFields, name, "address")

			if err != nil {
				cli.error(logFields, "invalid account: %s", name)
				return
			}

			response, err := microstellar.FundWithFriendBot(address)

			if err != nil {
				cli.error(logFields, "friendbot error: %v", err)
				return
			}

			showSuccess("friendbot says:\n %v", response)
		},
	}

	return cmd
}
