package cli

import "testing"

// Note: add -v to any of these commands to enable verbose logging

func TestPayments(t *testing.T) {
	cli, _ := newTestCLI()
	cli.RunCommand("ns test")
	cli.RunCommand("set config:network fake")

	cli.RunCommand("account new master")
	cli.RunCommand("account new worker")

	expectOutput(t, cli, "error", "pay 4 --from nobody --to worker")
	expectOutput(t, cli, "", "pay 4 --from SAFOI5YIH5MXO6HCICLBG3UYOER6PDYQXHP47JUB7XNWHNT2YISAOMAQ --to worker")
	expectOutput(t, cli, "", "pay 4 --from master --to GBH6GGAPBFH6IXCQBPJ7WSN2WMUFU7PO346BIVZXS6Q22YNFBUNVJS4U")
	expectOutput(t, cli, "", "pay 4 --from master --to worker --memotext hello")

	expectOutput(t, cli, "error", "pay 4 --from master --to worker --memoid hello -v")
	expectOutput(t, cli, "", "pay 4 --from master --to worker --memoid 234883")

	cli.RunCommand("ns other")
	cli.RunCommand("set config:network fake")
	expectOutput(t, cli, "error", "pay 4 --from master --to worker")

	cli.RunCommand("ns test")
	expectOutput(t, cli, "", "pay 4 --from master --to worker --fund")

	cli.RunCommand("account new issuer-chase")
	cli.RunCommand("account new issuer-citi")

	cli.RunCommand("asset set USD issuer-chase")
	cli.RunCommand("asset set USD-citi issuer-citi --code USD")

	expectOutput(t, cli, "", "pay 4 USD --from master --to worker")
	expectOutput(t, cli, "", "pay 4 USD-citi --from master --to worker")
}
