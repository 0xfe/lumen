package cli

import "testing"

// Note: add -v to any of these commands to enable verbose logging

func TestPayments(t *testing.T) {
	cli, _ := newTestCLI()
	cli.TestCommand("ns test")
	cli.TestCommand("set config:network fake")

	cli.TestCommand("account new master")
	cli.TestCommand("account new worker")

	expectOutput(t, cli, "error", "pay 4 --from nobody --to worker")
	expectOutput(t, cli, "", "pay 4 --from SAFOI5YIH5MXO6HCICLBG3UYOER6PDYQXHP47JUB7XNWHNT2YISAOMAQ --to worker")
	expectOutput(t, cli, "", "pay 4 --from master --to GBH6GGAPBFH6IXCQBPJ7WSN2WMUFU7PO346BIVZXS6Q22YNFBUNVJS4U")
	expectOutput(t, cli, "", "pay 4 --from master --to worker --memotext hello")

	expectOutput(t, cli, "error", "pay 4 --from master --to worker --memoid hello")
	expectOutput(t, cli, "", "pay 4 --from master --to worker --memoid 234883")

	cli.TestCommand("ns other")
	cli.TestCommand("set config:network fake")
	expectOutput(t, cli, "error", "pay 4 --from master --to worker")

	cli.TestCommand("ns test")
	expectOutput(t, cli, "", "pay 4 --from master --to worker --fund")

	cli.TestCommand("account new issuer-chase")
	cli.TestCommand("account new issuer-citi")

	cli.TestCommand("asset set USD issuer-chase")
	cli.TestCommand("asset set USD-citi issuer-citi --code USD")

	expectOutput(t, cli, "", "pay 4 USD --from master --to worker")
	expectOutput(t, cli, "", "pay 4 USD-citi --from master --to worker")
}
