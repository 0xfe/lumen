package cli

import "testing"

// Note: add -v to any of these commands to enable verbose logging

func TestDex(t *testing.T) {
	cli, _ := newTestCLI()
	cli.TestCommand("ns test")
	cli.TestCommand("set config:network fake")

	cli.TestCommand("account new mo")
	cli.TestCommand("account new kelly")

	cli.TestCommand("account new issuer-chase")
	cli.TestCommand("account new issuer-citi")

	cli.TestCommand("asset set USD issuer-chase")
	cli.TestCommand("asset set INR issuer-chase")
	cli.TestCommand("asset set EUR issuer-chase")
	cli.TestCommand("asset set USD-citi issuer-citi --code USD")

	expectOutput(t, cli, "", "dex trade mo --buy USD --sell INR --amount 20 --price 2")
	expectOutput(t, cli, "", "dex trade mo --buy INR --sell USD --amount 20 --price 2 --passive")
	expectOutput(t, cli, "", "dex trade mo --buy USD --sell EUR --amount 20 --price 2")
	expectOutput(t, cli, "", "dex trade mo --buy INR --sell USD --amount 20 --price 2 --update 23112")
	expectOutput(t, cli, "", "dex trade mo --buy INR --sell USD --amount 20 --price 2 --delete 23112")
	expectOutput(t, cli, "", "dex list mo --cursor 23443 --limit 3 --desc")
}
