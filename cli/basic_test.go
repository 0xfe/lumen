package cli

import (
	"testing"
)

// Note: add -v to any of these commands to enable verbose logging

func TestVariables(t *testing.T) {
	cli, _ := newTestCLI()
	cli.TestCommand("ns test")
	expectOutput(t, cli, "test", "ns")

	cli.TestCommand("set foo bar")
	expectOutput(t, cli, "bar", "get foo")

	cli.TestCommand("ns default")
	expectOutput(t, cli, "error", "get foo")

	cli.TestCommand("set foo haha")
	expectOutput(t, cli, "haha", "get foo")

	cli.TestCommand("ns test")
	expectOutput(t, cli, "bar", "get foo")

	cli.TestCommand("del foo")
	expectOutput(t, cli, "error", "get foo")

	expectOutput(t, cli, "error", "flags nobody nothing")

	cli.TestCommand("set config:network fake")
	cli.TestCommand("account new mo")
	expectOutput(t, cli, "", "flags mo none")
	expectOutput(t, cli, "", "flags mo auth_revocable auth_immutable")
}
