package cli

import "testing"

// Note: add -v to any of these commands to enable verbose logging

func TestAccounts(t *testing.T) {
	cli, _ := newTestCLI()
	cli.TestCommand("ns test")
	cli.TestCommand("account new master")
	cli.TestCommand("account new worker")

	result := cli.TestCommand("account address master")

	if result[0] != 'G' {
		t.Error("not an address: ", result)
	}

	result = cli.TestCommand("account seed master")

	if result[0] != 'S' {
		t.Error("not a seed: ", result)
	}

	cli.TestCommand("ns other")
	expectOutput(t, cli, "error", "account address master")

	cli.TestCommand("ns test")
	cli.TestCommand("account del master")
	expectOutput(t, cli, "error", "account address master")
}
