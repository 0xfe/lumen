package lumen

import (
	"fmt"
	"os"

	"github.com/0xfe/lumen/cli"
	"github.com/0xfe/lumen/store"
	"github.com/0xfe/microstellar"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	StorageDriver string
	StorageParams string
}

func ReadConfig(env string) Config {
	homeDir := os.Getenv("HOME")
	filePath := fmt.Sprintf("%s%s%s", homeDir, string(os.PathSeparator), ".lumen-data.json")

	config := Config{
		StorageDriver: "file",
		StorageParams: filePath,
	}

	switch env {
	case "dev":
		viper.SetConfigName("lumen-config-dev")
	case "test":
		viper.SetConfigName("lumen-config-test")
	default: // also "prod"
		viper.SetConfigName("lumen-config")
	}

	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AddConfigPath("$HOME/.lumen")
	viper.AddConfigPath("/etc/lumen/")

	err := viper.ReadInConfig() // Find and read the config file

	if err == nil {
		config.StorageDriver = viper.GetString("storage.driver")
		config.StorageParams = viper.GetString("storage.params")
	}

	return config
}

func help(cmd *cobra.Command, args []string) {
	fmt.Fprint(os.Stderr, cmd.UsageString())
}

var rootCmd = &cobra.Command{
	Use:   "lumen",
	Short: "Lumen is a commandline client for the Stellar blockchain",
	Run:   help,
}

func Start() {
	config := ReadConfig(os.Getenv("LUMEN_ENV"))
	store, _ := store.NewStore(config.StorageDriver, config.StorageParams)

	cli := cli.NewCLI(store, microstellar.New("test"))

	// Install the CLI
	cli.Install(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		// fmt.Fprintf(os.Stderr, "error: %v", err)
	}
}
