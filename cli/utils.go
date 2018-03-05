package cli

import (
	"fmt"
	"os"

	"github.com/0xfe/microstellar"
	"github.com/sirupsen/logrus"
)

func (cli *CLI) validateAddressOrSeed(fields logrus.Fields, addressOrSeed string, keyType string) string {
	var err error

	if !microstellar.ValidAddressOrSeed(addressOrSeed) {
		addressOrSeed, err = cli.GetAccountOrSeed(addressOrSeed, keyType)
		if err != nil {
			showError(fields, "invalid address, seed, or account name: %s", addressOrSeed)
			os.Exit(-1)
		}
	}

	return addressOrSeed
}

func showSuccess(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
}

func showError(fields logrus.Fields, msg string, args ...interface{}) {
	logrus.WithFields(fields).Errorf(msg, args...)
	os.Exit(-1)
}
