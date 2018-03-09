package lumen

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/0xfe/lumen/cli"
	"github.com/sirupsen/logrus"
)

// Send some funds here if friendbot works
const payBack = "GAVQK5QMS6TNUZNS3TXDV5RKFIHLSMWDYZQR5ZV72VJT2SL4JMJXWQZE"

// Suck funds from here if friendbot fails
const fundSource = "SDGRPZDOPXDZPBY5LMCY7HE7ISSPMG2IIAXFIKF4TQW7RZHCUH5O7SJK"

func getTempFile() (string, func()) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		log.Fatal(err)
	}

	file := dir + string(os.PathSeparator) + "lumen-integration-test.json"

	return file, func() {
		logrus.Debugf("cleaning up temp file: %s", file)
		os.RemoveAll(dir)
	}
}

func run(cli *cli.CLI, command string) string {
	fmt.Printf("$ lumen %s\n", command)
	got := cli.TestCommand(command)
	fmt.Printf("%s\n", got)
	return strings.TrimSpace(got)
}

func expectOutput(t *testing.T, cli *cli.CLI, want string, command string) {
	got := run(cli, command)

	if got != want {
		t.Errorf("(%s) wrong output: want %v, got %v", command, want, got)
	}
}

func newCLI() (*cli.CLI, func()) {
	file, cleanupFunc := getTempFile()
	os.Setenv("LUMEN_STORE", "file,"+file)

	lumen := cli.NewCLI()
	lumen.TestCommand("version")
	lumen.TestCommand("ns test")
	lumen.TestCommand("set config:network test")
	run(lumen, fmt.Sprintf("account set landlord %s", payBack))
	run(lumen, fmt.Sprintf("account set sugar %s", fundSource))

	return lumen, cleanupFunc
}

func getBalance(cli *cli.CLI, account string) float64 {
	balanceString := run(cli, "balance mo")

	balance, err := strconv.ParseFloat(balanceString, 64)

	if err != nil {
		return 0
	}

	return balance
}

// Create new funded test account
func createFundedAccount(t *testing.T, cli *cli.CLI, name string) {
	run(cli, "account new "+name)
	run(cli, "friendbot "+name)

	balance := getBalance(cli, name)
	if balance > 100 {
		// return some balance to landlord
		run(cli, fmt.Sprintf("pay 5000 --from %s --to landlord", name))
	} else {
		// friendbot failed, fund via sugar daddy
		run(cli, fmt.Sprintf("pay 5000 --from sugar --to %s", name))
	}

	balance = getBalance(cli, name)
	if balance < 4999 {
		t.Fatalf("could not fund account: %s", name)
	}
}

func TestPayments(t *testing.T) {
	cli, cleanupFunc := newCLI()
	defer cleanupFunc()

	createFundedAccount(t, cli, "mo")
	run(cli, "account new kelly")

	expectOutput(t, cli, "", "pay 100 --from mo --to kelly --memotext hi --fund")
	expectOutput(t, cli, "", "pay 1 --from kelly --to mo --memotext yo -v")

	balance := getBalance(cli, "kelly")

	if balance > 99 {
		t.Fatalf("expected balance <= 99 got %v", balance)
	}
}

func TestAssets(t *testing.T) {
	cli, cleanupFunc := newCLI()
	defer cleanupFunc()

	createFundedAccount(t, cli, "mo")

	// Create a USD asset issued by citibank
	run(cli, "account new citibank")
	expectOutput(t, cli, "", "pay 100 --from mo --to citibank --memoid 1 --fund")
	run(cli, "asset set USD citibank")

	// Create a trustline between kelly and citibank with a $1000 limit, then
	// send her $100
	run(cli, "account new kelly")
	expectOutput(t, cli, "", "pay 10 --from mo --to kelly --memotext initial --fund")

	expectOutput(t, cli, "", "trust create kelly USD 1000")
	expectOutput(t, cli, "", "pay 100 USD --from citibank --to kelly")
}
