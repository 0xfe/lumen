package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/0xfe/lumen/store"
	"github.com/0xfe/microstellar"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// CLI represents a command-line interface. This class is
// not threadsafe.
type CLI struct {
	store   store.API
	ms      *microstellar.MicroStellar
	ns      string // namespace
	rootCmd *cobra.Command
	version string
	testing bool
}

// NewCLI returns an initialized CLI
func NewCLI() *CLI {
	cli := &CLI{
		store:   nil,
		ms:      nil,
		ns:      "",
		rootCmd: nil,
		version: "v0.0",
		testing: false,
	}

	cli.init()
	return cli
}

// Set the data store (used for testing.)
func (cli *CLI) SetStore(store store.API) {
	cli.store = store
}

func (cli *CLI) help(cmd *cobra.Command, args []string) {
	fmt.Fprint(os.Stderr, cmd.UsageString())
}

func (cli *CLI) error(logFields logrus.Fields, msg string, args ...interface{}) {
	showError(logFields, msg, args...)

	if !cli.testing {
		os.Exit(-1)
	} else {
		fmt.Println("error")
	}
}

func (cli *CLI) setup(cmd *cobra.Command, args []string) {
	if cli.testing {
		buf := new(bytes.Buffer)
		logrus.SetOutput(buf)
	}

	if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
		logrus.SetOutput(os.Stderr)
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	env := os.Getenv("LUMEN_ENV")
	if env != "" {
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("env LUMEN_ENV: %s", env)
	} else {
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("LUMEN_ENV not set")
	}

	config := readConfig(env)

	// Do this again if the configuration file says so
	if config.verbose {
		logrus.SetOutput(os.Stderr)
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("using storage driver %s with %s", config.storageDriver, config.storageParams)

	cli.setupStore(config.storageDriver, config.storageParams)
	cli.setupNameSpace()
	cli.setupNetwork()
}

func (cli *CLI) setupStore(driver, params string) {
	if cli.store != nil {
		// Custom store takes precedence
		return
	}

	if cli.rootCmd.Flag("store").Changed {
		store, _ := cli.rootCmd.Flags().GetString("store")
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("using store %s", store)

		parts := strings.Split(store, ":")
		driver = parts[0]
		if len(parts) > 1 {
			params = parts[1]
		} else {
			params = ""
		}
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("selecting store driver: %s params: %s", driver, params)
	} else {
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("using default store")
	}

	var err error
	cli.store, err = store.NewStore(driver, params)

	if err != nil {
		showError(logrus.Fields{"type": "setup"}, "could not initialize filestore: %s:%s", driver, params)
		return
	}
}

func (cli *CLI) setupNameSpace() {
	if cli.ns != "" {
		return
	}

	if cli.rootCmd.Flag("ns").Changed {
		ns, _ := cli.rootCmd.Flags().GetString("ns")
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("using namespace %s", ns)
		cli.ns = ns
	} else {
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("using default namespace")
		ns, err := cli.GetGlobalVar("ns")
		if err != nil {
			logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("using default namespace")
			cli.ns = "default"
		} else {
			cli.ns = ns
		}
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

// Run executes CLI with the given arguments. Used for testing. Not thread safe.
func (cli *CLI) Run(args ...string) string {
	oldStdout := os.Stdout

	r, w, _ := os.Pipe()

	os.Stdout = w

	cli.testing = true
	cli.rootCmd.SetArgs(args)
	cli.rootCmd.Execute()
	cli.testing = false

	w.Close()

	os.Stdout = oldStdout

	var stdOut bytes.Buffer
	io.Copy(&stdOut, r)
	return stdOut.String()
}

func (cli *CLI) RunCommand(command string) string {
	return cli.Run(strings.Fields(command)...)
}

// RootCmd returns the cobra root comman for this instance
func (cli *CLI) RootCmd() *cobra.Command {
	return cli.rootCmd
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
	home := os.Getenv("HOME")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output (false)")
	rootCmd.PersistentFlags().String("network", "test", "network to use (test)")
	rootCmd.PersistentFlags().String("ns", "default", "namespace to use (default)")
	rootCmd.PersistentFlags().String("store", fmt.Sprintf("file:%s/.lumen-data.yml", home), "namespace to use (default)")

	rootCmd.AddCommand(cli.getPayCmd())
	rootCmd.AddCommand(cli.getAccountCmd())
	rootCmd.AddCommand(cli.getAssetCmd())
	rootCmd.AddCommand(cli.getTrustCmd())

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "get version of lumen CLI",
		Run:   cli.cmdVersion,
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "ns [namespace]",
		Short: "set namespace to [namespace]",
		Args:  cobra.MinimumNArgs(0),
		Run:   cli.cmdNS,
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
}

// SetGlobalVar writes the kv pair to the global namespace in the storage backend
func (cli *CLI) SetGlobalVar(key string, value string) error {
	key = fmt.Sprintf("global:%s", key)
	logrus.WithFields(logrus.Fields{"type": "cli", "method": "SetGlobalVar"}).Debugf("setting %s: %s", key, value)
	return cli.store.Set(key, value, 0)
}

// GetGlobalVar reads global var "key"
func (cli *CLI) GetGlobalVar(key string) (string, error) {
	key = fmt.Sprintf("global:%s", key)
	logrus.WithFields(logrus.Fields{"type": "cli", "method": "GetGlobalVar"}).Debugf("getting %s", key)
	return cli.store.Get(key)
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
