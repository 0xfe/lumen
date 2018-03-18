package cli

import "testing"

// Note: add -v to any of these commands to enable verbose logging

func TestData(t *testing.T) {
	cli, _ := newTestCLI()
	cli.TestCommand("ns test")
	cli.TestCommand("set config:network fake")

	cli.TestCommand("account new master")

	expectOutput(t, cli, "", "data master foo bar")
	expectOutput(t, cli, "", "data master foo --clear")
	expectOutput(t, cli, "error", "data worker foo --clear")
}
