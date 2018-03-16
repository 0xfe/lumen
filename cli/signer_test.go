package cli

import "testing"

// Note: add -v to any of these commands to enable verbose logging

func TestSigners(t *testing.T) {
	cli, _ := newTestCLI()
	cli.TestCommand("ns test")
	cli.TestCommand("set config:network fake")

	cli.TestCommand("account new master")
	cli.TestCommand("account new worker")
	cli.TestCommand("account new signer1")
	cli.TestCommand("account new signer2")

	expectOutput(t, cli, "error", "pay 4 --from nobody --to worker")
	expectOutput(t, cli, "", "signer add signer1 2 --to worker")
	expectOutput(t, cli, "", "signer add signer2 2 --to worker")

	expectOutput(t, cli, "", "pay 4 --from worker --to master --signers signer1,signer2 --memotext hello")
	expectOutput(t, cli, "error", "pay 4 --from master --to worker --signers nobody")

	expectOutput(t, cli, "", "signer remove signer1 --from worker")
	expectOutput(t, cli, "error", "signer remove nobody --from worker")

	expectOutput(t, cli, "error", "signer masterweight mo 400")
	expectOutput(t, cli, "", "signer masterweight master 400")

	expectOutput(t, cli, "address: weight:0", "signer list master")
}
