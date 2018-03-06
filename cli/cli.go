package cli

import (
	"fmt"
	"os"

	"github.com/0xfe/lumen/store"
	"github.com/0xfe/microstellar"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// CLI represents a command-line interface
type CLI struct {
	store   store.API
	ms      *microstellar.MicroStellar
	ns      string // namespace
	rootCmd *cobra.Command
}

// NewCLI returns an initialized CLI
func NewCLI() *CLI {
	cli := &CLI{
		store:   nil,
		ms:      nil,
		ns:      "",
		rootCmd: nil,
	}

	cli.init()
	return cli
}

func (cli *CLI) help(cmd *cobra.Command, args []string) {
	fmt.Fprint(os.Stderr, cmd.UsageString())
}

func (cli *CLI) setup(cmd *cobra.Command, args []string) {
	if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	env := os.Getenv("LUMEN_ENV")
	if env != "" {
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("env LUMEN_ENV: %s", env)
	} else {
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("LUMEN_ENV not set")
	}

	config := readConfig(env)

	if config.verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("using storage driver %s with %s", config.storageDriver, config.storageParams)
	cli.store, _ = store.NewStore(config.storageDriver, config.storageParams)

	cli.setupNameSpace()
	cli.setupNetwork()
}

func (cli *CLI) setupNameSpace() {
	if cli.rootCmd.Flag("ns").Changed {
		ns, _ := cli.rootCmd.Flags().GetString("ns")
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("using namespace %s", ns)
		cli.ns = ns
	} else {
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("using default namespace")
		cli.ns = "default"
	}
}

func (cli *CLI) setupNetwork() {
	if cli.rootCmd.Flag("network").Changed {
		network, _ := cli.rootCmd.Flags().GetString("network")
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("using horizon network: %s", network)
		cli.ms = microstellar.New(network)
	} else {
		network, err := cli.GetVar("vars:config:network")
		if err != nil {
			cli.ms = microstellar.New("test")
		} else {
			cli.ms = microstellar.New(network)
		}
	}
}

// Execute parses the command line and processes it
func (cli *CLI) Execute() {
	cli.rootCmd.Execute()
}

func (cli *CLI) init() {
	rootCmd := &cobra.Command{
		Use:              "lumen",
		Short:            "Lumen is a commandline client for the Stellar blockchain",
		Run:              cli.help,
		PersistentPreRun: cli.setup,
	}

	cli.rootCmd = rootCmd

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output (false)")
	rootCmd.PersistentFlags().String("network", "test", "network to use (test)")
	rootCmd.PersistentFlags().String("ns", "default", "namespace to use (default)")

	rootCmd.AddCommand(cli.getPayCmd())

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Get version of lumen CLI",
		Run:   cli.cmdVersion,
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "set [key] [value]",
		Short: "set variable",
		Args:  cobra.MinimumNArgs(2),
		Run:   cli.cmdSet,
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "get [key]",
		Short: "get variable",
		Args:  cobra.MinimumNArgs(1),
		Run:   cli.cmdGet,
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "del [key]",
		Short: "delete variable",
		Args:  cobra.MinimumNArgs(1),
		Run:   cli.cmdDel,
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "watch [address]",
		Short: "watch the address on the ledger",
		Args:  cobra.MinimumNArgs(1),
		Run:   cli.cmdWatch,
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "balance [address]",
		Short: "get the balance of [address] in lumens",
		Args:  cobra.MinimumNArgs(1),
		Run:   cli.cmdBalance,
	})

	rootCmd.AddCommand(cli.getAccountCmd())
}

// SetVar writes the kv pair to the storage backend
func (cli *CLI) SetVar(key string, value string) error {
	key = fmt.Sprintf("%s:%s", cli.ns, key)
	logrus.WithFields(logrus.Fields{"type": "cli", "method": "SetVar"}).Debugf("setting %s: %s", key, value)
	return cli.store.Set(key, value, 0)
}

func (cli *CLI) GetVar(key string) (string, error) {
	key = fmt.Sprintf("%s:%s", cli.ns, key)
	logrus.WithFields(logrus.Fields{"type": "cli", "method": "GetVar"}).Debugf("getting %s", key)
	return cli.store.Get(key)
}

func (cli *CLI) DelVar(key string) error {
	key = fmt.Sprintf("%s:%s", cli.ns, key)
	logrus.WithFields(logrus.Fields{"type": "cli", "method": "DelVar"}).Debugf("deleting %s", key)
	return cli.store.Delete(key)
}
