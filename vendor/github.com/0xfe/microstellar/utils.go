package microstellar

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/strkey"
)

// ValidAddress returns error if address is an invalid stellar address
func ValidAddress(address string) error {
	_, err := strkey.Decode(strkey.VersionByteAccountID, address)
	return errors.Wrap(err, "invalid address")
}

// ValidSeed returns error if the seed is invalid
func ValidSeed(seed string) error {
	_, err := strkey.Decode(strkey.VersionByteSeed, seed)
	return errors.Wrap(err, "invalid seed")
}

// ValidAddressOrSeed returns true if the string is a valid address or seed
func ValidAddressOrSeed(addressOrSeed string) bool {
	err := ValidAddress(addressOrSeed)
	if err == nil {
		return true
	}

	err = ValidSeed(addressOrSeed)
	return err == nil
}

// ErrorString parses the horizon error out of err.
func ErrorString(err error, showStackTrace ...bool) string {
	var errorString string
	herr, isHorizonError := errors.Cause(err).(*horizon.Error)

	if isHorizonError {
		errorString += fmt.Sprintf("Horizon error: \n")

		resultCodes, err := herr.ResultCodes()
		if err == nil {
			errorString += fmt.Sprintf("  error code: %v\n", resultCodes)
		}

		errorString += fmt.Sprintf("  %v: %v\n  detail: %v\n  type: %v\n",
			herr.Problem.Status,
			herr.Problem.Title,
			herr.Problem.Detail,
			herr.Problem.Type)
	}

	if len(showStackTrace) > 0 {
		errorString += fmt.Sprintf("\nStack trace:\n%+v\n", err)
	}

	return errorString
}

// FundWithFriendBot funds address on the test network with some initial funds.
func FundWithFriendBot(address string) (string, error) {
	resp, err := http.Get("https://horizon-testnet.stellar.org/friendbot?addr=" + address)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
