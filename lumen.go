package lumen

/*
Lumen is a batteries-included command-line utility for working with the
Stellar Network.

See github.com/0xfe/lumen for more.

Copyright Mohit Muthanna Cheppudira 2018
*/

import (
	"github.com/0xfe/lumen/cli"
)

// Start turns up the CLI environment and runs lumen.
func Start() {
	cli.NewCLI().Execute()
}
