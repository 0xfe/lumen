package microstellar

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/stellar/go/amount"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/strkey"
	"github.com/stellar/go/xdr"
)

func debugf(method string, msg string, args ...interface{}) {
	logrus.WithFields(logrus.Fields{"lib": "microstellar", "method": method}).Debugf(msg, args...)
}

// ParseAmount converts a currency amount string to an int64
func ParseAmount(v string) (int64, error) {
	return amount.ParseInt64(v)
}

// ToAmountString converts an int64 amount to a string
func ToAmountString(v int64) string {
	return amount.StringFromInt64(v)
}

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
		errorString += fmt.Sprintf("%v: %v", herr.Problem.Status, herr.Problem.Title)

		resultCodes, err := herr.ResultCodes()
		if err == nil {
			errorString += fmt.Sprintf(" (%v)", resultCodes)
		}
	} else {
		errorString = fmt.Sprintf("%v", err)
	}

	if len(showStackTrace) > 0 {
		if isHorizonError {
			errorString += fmt.Sprintf("\nDetail: %s\nType: %s\n", herr.Problem.Detail, herr.Problem.Type)
		}
		errorString += fmt.Sprintf("\nStack trace:\n%+v\n", err)
	}

	return errorString
}

// FundWithFriendBot funds address on the test network with some initial funds.
func FundWithFriendBot(address string) (string, error) {
	debugf("FundWithFriendBot", "funding address: %s", address)
	resp, err := http.Get("https://friendbot.stellar.org/?addr=" + address)
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

// DecodeTx extracts a TransactionEnvelope out of the base64-encoded string.
func DecodeTx(base64tx string) (*xdr.TransactionEnvelope, error) {
	var tx xdr.TransactionEnvelope

	if base64tx[len(base64tx)-1] != '=' {
		base64tx = base64tx + "=="
	}

	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(base64tx))
	_, err := xdr.Unmarshal(reader, &tx)

	if err != nil {
		return nil, errors.Wrapf(err, "error decoding base64 transaction")
	}

	return &tx, nil
}

// DecodeTxToJSON converts the base-64 TX to a JSON string.
func DecodeTxToJSON(base64tx string, pretty bool) (string, error) {
	xdr, err := DecodeTx(base64tx)
	if err != nil {
		return "", errors.Wrap(err, "DecodeTxToJSON")
	}

	var xdrJSON []byte
	if pretty {
		xdrJSON, err = json.MarshalIndent(xdr, "", "  ")
	} else {
		xdrJSON, err = json.Marshal(xdr)
	}

	if err != nil {
		return "", errors.Wrap(err, "json.MarshalIndent")
	}

	return string(xdrJSON), nil
}
