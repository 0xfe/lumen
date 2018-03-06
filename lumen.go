package lumen

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/0xfe/lumen/cli"
	"github.com/0xfe/lumen/store"
	"github.com/0xfe/microstellar"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type config struct {
	storageDriver string
	storageParams string
	verbose       bool
}

func readConfig(env string) config {
	homeDir := os.Getenv("HOME")
	filePath := fmt.Sprintf("%s%s%s", homeDir, string(os.PathSeparator), ".lumen-data.json")

	config := config{
		storageDriver: "file",
		storageParams: filePath,
		verbose:       false,
	}

	switch env {
	case "dev":
		viper.SetConfigName(".lumen-config-dev")
	case "test":
		viper.SetConfigName(".lumen-config-test")
	default: // also "prod"
		viper.SetConfigName(".lumen-config")
	}

	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AddConfigPath("$HOME/.lumen")
	viper.AddConfigPath("/etc/lumen/")

	err := viper.ReadInConfig() // Find and read the config file

	if err == nil {
		config.storageDriver = viper.GetString("storage.driver")
		config.storageParams = viper.GetString("storage.params")
		config.verbose = viper.GetBool("verbose")
	}

	return config
}

func help(cmd *cobra.Command, args []string) {
	fmt.Fprint(os.Stderr, cmd.UsageString())
}

func setup(cmd *cobra.Command, args []string) {
	if cmd.Flags().Lookup("verbose").Value.String() == "true" {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func Start() {
	// Load config and setup CLI
	config := readConfig(os.Getenv("LUMEN_ENV"))

	if config.verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	store, _ := store.NewStore(config.storageDriver, config.storageParams)

	var rootCmd = &cobra.Command{
		Use:              "lumen",
		Short:            "Lumen is a commandline client for the Stellar blockchain",
		Run:              help,
		PersistentPreRun: setup,
	}
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	cli := cli.NewCLI(rootCmd, store, microstellar.New("test"))
	cli.Execute()
}
