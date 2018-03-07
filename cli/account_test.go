package cli

import "testing"

// Note: add -v to any of these commands to enable verbose logging

func TestAccounts(t *testing.T) {
	cli, _ := newTestCLI()
	cli.RunCommand("ns test")
	cli.RunCommand("account new master")
	cli.RunCommand("account new worker")

	result := cli.RunCommand("account address master")

	if result[0] != 'G' {
		t.Error("not an address: ", result)
	}

	result = cli.RunCommand("account seed master")

	if result[0] != 'S' {
		t.Error("not a seed: ", result)
	}

	cli.RunCommand("ns other")
	expectOutput(t, cli, "error", "account address master")

	cli.RunCommand("ns test")
	cli.RunCommand("account del master")
	expectOutput(t, cli, "error", "account address master")
}
