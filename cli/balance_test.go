package cli

import "testing"

// Note: add -v to any of these commands to enable verbose logging

func TestBalance(t *testing.T) {
	cli, _ := newTestCLI()
	cli.TestCommand("ns test")
	cli.TestCommand("set config:network fake")

	cli.TestCommand("account new master")
	cli.TestCommand("account new worker")

	expectOutput(t, cli, "", "pay 100 --from master --to worker --memotext hello")

	cli.TestCommand("account new issuer-chase")
	cli.TestCommand("asset set USD issuer-chase")

	expectOutput(t, cli, "0", "balance worker")
	expectOutput(t, cli, "0", "balance worker USD")
}
