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
	store       store.API
	ms          *microstellar.MicroStellar
	ns          string // namespace
	rootCmd     *cobra.Command
	version     string
	testing     bool
	stopWatcher func()
}

// NewCLI returns an initialized CLI
func NewCLI() *CLI {
	cli := &CLI{
		store:       nil,
		ms:          nil,
		ns:          "",
		rootCmd:     nil,
		version:     "v0.0",
		testing:     false,
		stopWatcher: func() {},
	}

	cli.buildRootCmd()
	return cli
}

// Execute parses the command line and processes it.
func (cli *CLI) Execute() {
	cli.rootCmd.Execute()
}

// SetStore lets you set the data store (used for testing.)
func (cli *CLI) SetStore(store store.API) {
	cli.store = store
}

// Embeddable returns a CLI that you can embed into your own Go programs. This
// is not thread-safe.
func (cli *CLI) Embeddable() *CLI {
	cli.testing = true
	return cli
}

// Run executes CLI with the given arguments. Used for testing. Not thread safe.
func (cli *CLI) Run(args ...string) string {
	oldStdout := os.Stdout

	r, w, _ := os.Pipe()

	os.Stdout = w

	cli.rootCmd.SetArgs(args)
	cli.rootCmd.Execute()
	cli.buildRootCmd()

	w.Close()

	os.Stdout = oldStdout

	var stdOut bytes.Buffer
	io.Copy(&stdOut, r)
	return stdOut.String()
}

// RunCommand is a helper that lets you send a full command line to Run, so you don't
// have to break up your arguments.
func (cli *CLI) RunCommand(command string) string {
	result := cli.Run(strings.Fields(command)...)
	return result
}

// TestCommand is a helper function that calls Run(...) in test mode. When running
// in test mode, os.Exit is not called on errors.
func (cli *CLI) TestCommand(command string) string {
	cli.testing = true
	result := cli.Run(strings.Fields(command)...)
	cli.testing = false
	return result
}

// Stop an existing watcher from streaming.
func (cli *CLI) StopWatcher() {
	cli.stopWatcher()
	cli.stopWatcher = func() {}
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

// setup turns up the CLI environment, and gets called by Cobra before
// a command is executed.
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

// setupStore sets up the storage backend.
func (cli *CLI) setupStore(driver, params string) {
	if cli.store != nil {
		// Custom store takes precedence
		return
	}

	parseStoreParams := func(store string) {
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("using store %s", store)
		parts := strings.Split(store, ",")
		driver = strings.TrimSpace(parts[0])
		if len(parts) > 1 {
			params = strings.TrimSpace(parts[1])
		} else {
			params = ""
		}
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("selecting store driver: %s params: %s", driver, params)
	}

	if cli.rootCmd.Flag("store").Changed {
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("using store from flag --store")
		store, _ := cli.rootCmd.Flags().GetString("store")
		parseStoreParams(store)
	} else if os.Getenv("LUMEN_STORE") != "" {
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("using store from env LUMEN_STORE")
		parseStoreParams(os.Getenv("LUMEN_STORE"))
	} else {
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("using default store")
	}

	var err error
	cli.store, err = store.NewStore(driver, params)

	if err != nil {
		logrus.WithFields(logrus.Fields{"type": "setup"}).Fatalf("could not initialize filestore: %s:%s", driver, params)
		return
	}
}

// setupNameSpace makes sure that storage commands used the correct namespace.
func (cli *CLI) setupNameSpace() {
	if cli.ns != "" {
		return
	}

	if cli.rootCmd.Flag("ns").Changed {
		ns, _ := cli.rootCmd.Flags().GetString("ns")
		cli.ns = ns
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("using namespace from flag --ns")
	} else if ns := os.Getenv("LUMEN_NS"); ns != "" {
		cli.ns = ns
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("using namespace from env LUMEN_NS")
	} else if ns, err := cli.GetGlobalVar("ns"); err == nil {
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("using namespace from store")
		cli.ns = ns
	} else {
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("using default namespace")
		cli.ns = "default"
	}

	logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("namespace: %s", cli.ns)
}

// setupNetwork ensures that lumen is operating on the correct network.
func (cli *CLI) setupNetwork() {
	if cli.rootCmd.Flag("network").Changed {
		network, _ := cli.rootCmd.Flags().GetString("network")
		logrus.WithFields(logrus.Fields{"type": "setup"}).Debugf("using horizon network: %s", network)
		cli.ms = microstellar.NewFromSpec(network)
	} else {
		network, err := cli.GetVar("vars:config:network")
		if err != nil {
			cli.ms = microstellar.NewFromSpec("test")
		} else {
			cli.ms = microstellar.NewFromSpec(network)
		}
	}
}
