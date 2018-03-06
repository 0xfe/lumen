package cli

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
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
		logrus.WithFields(logrus.Fields{"type": "config"}).Debugf("loaded config from file %s", viper.ConfigFileUsed())
		config.storageDriver = viper.GetString("storage.driver")
		config.storageParams = viper.GetString("storage.params")
		config.verbose = viper.GetBool("verbose")
	}

	return config
}
