package cli

import (
	"strings"
	"testing"

	"github.com/0xfe/lumen/store"
)

func expectOutput(t *testing.T, cli *CLI, want string, command string) {
	cli.Run("ns", "test")
	got, _ := cli.RunCommand(command)

	if strings.TrimSpace(got) != want {
		t.Errorf("wrong output: want %v, got %v", want, got)
	}
}

func newTestCLI() (*CLI, store.API) {
	cli := NewCLI()
	memStore, _ := store.NewStore("internal", "")
	cli.SetStore(memStore)

	return cli, memStore
}

func TestCLIVersion(t *testing.T) {
	cli, _ := newTestCLI()
	expectOutput(t, cli, cli.version, "version")
}

func TestNameSpaces(t *testing.T) {
	cli, _ := newTestCLI()
	cli.RunCommand("ns test")
	expectOutput(t, cli, "test", "ns")

	cli.RunCommand("set foo bar")
	expectOutput(t, cli, "bar", "get foo")

	cli.RunCommand("ns default -v")
	expectOutput(t, cli, "", "get foo")
}
