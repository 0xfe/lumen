package cli

import (
	"strings"
	"testing"

	"github.com/0xfe/lumen/store"
)

func expectOutput(t *testing.T, cli *CLI, want string, command string) {
	got := cli.TestCommand(command)

	if strings.TrimSpace(got) != want {
		t.Errorf("(%s) wrong output: want %v, got %v", command, want, got)
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
