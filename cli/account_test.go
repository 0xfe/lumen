package cli

import "testing"

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

	result = cli.RunCommand("set config:network fake")
	expectOutput(t, cli, "error", "pay 4 --from nobody --to worker")
	expectOutput(t, cli, "", "pay 4 --from master --to worker")

	result = cli.RunCommand("set config:network fake")
	cli.RunCommand("ns other")
	expectOutput(t, cli, "error", "pay 4 --from master --to worker")
}
