package cli

import (
	"fmt"
	"os"

	"github.com/0xfe/microstellar"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func showSuccess(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
}

func showError(fields logrus.Fields, msg string, args ...interface{}) {
	logrus.WithFields(fields).Errorf(msg, args...)
	os.Exit(-1)
}

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

// GetAccount returns the account address or seed for "name". Set keyType
// to "address" or "seed" to specify the return value.
func (cli *CLI) GetAccount(name, keyType string) (string, error) {
	if keyType != "address" && keyType != "seed" {
		return "", errors.Errorf("invalid key type: %s", keyType)
	}

	key := fmt.Sprintf("account:%s:%s", name, keyType)

	code, err := cli.GetVar(key)

	if err != nil {
		return "", err
	}

	return code, nil
}

// GetAccountOrSeed returns the account address or seed for "name". It prefers
// keyType ("address" or "seed")
func (cli *CLI) GetAccountOrSeed(name, keyType string) (string, error) {
	code, err := cli.GetAccount(name, keyType)

	if err != nil {
		if keyType == "address" {
			keyType = "seed"
		} else {
			keyType = "address"
		}

		code, err = cli.GetAccount(name, keyType)
	}

	return code, err
}
