package cli

import (
	"testing"
)

// Note: add -v to any of these commands to enable verbose logging

func TestVariables(t *testing.T) {
	cli, _ := newTestCLI()
	cli.RunCommand("ns test")
	expectOutput(t, cli, "test", "ns")

	cli.RunCommand("set foo bar")
	expectOutput(t, cli, "bar", "get foo")

	cli.RunCommand("ns default")
	expectOutput(t, cli, "error", "get foo")

	cli.RunCommand("set foo haha")
	expectOutput(t, cli, "haha", "get foo")

	cli.RunCommand("ns test")
	expectOutput(t, cli, "bar", "get foo")
	cli.RunCommand("del foo")
	expectOutput(t, cli, "error", "get foo")
}
