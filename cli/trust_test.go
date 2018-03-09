package cli

import "testing"

// Note: add -v to any of these commands to enable verbose logging

func TestTrust(t *testing.T) {
	cli, _ := newTestCLI()
	cli.TestCommand("ns test")
	cli.TestCommand("set config:network fake")

	cli.TestCommand("account new mo")
	cli.TestCommand("account new kelly")

	cli.TestCommand("account new issuer-chase")
	cli.TestCommand("account new issuer-citi")

	cli.TestCommand("asset set USD issuer-chase")
	cli.TestCommand("asset set USD-citi issuer-citi --code USD")

	expectOutput(t, cli, "", "trust create mo USD --memotext ilikechase")
	expectOutput(t, cli, "", "trust create kelly USD-citi --memoid 43")
	expectOutput(t, cli, "", "trust remove mo USD --memotext ihatechase")
	expectOutput(t, cli, "", "trust remove kelly USD --memoid 748")
}
