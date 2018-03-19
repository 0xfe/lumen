// Package microstellar is an easy-to-use Go client for the Stellar network.
//
//   go get github.com/0xfe/microstellar
//
// Author: Mohit Muthanna Cheppudira <mohit@muthanna.com>
//
// Usage notes
//
// In Stellar lingo, a private key is called a seed, and a public key is called an address. Seed
// strings start with "S", and address strings start with "G". (Not techincally accurate, but you
// get the picture.)
//
//   Seed:    S6H4HQPE6BRZKLK3QNV6LTD5BGS7S6SZPU3PUGMJDJ26V7YRG3FRNPGA
//   Address: GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM
//
// In most the methods below, the first parameter is usually "sourceSeed", which should be the
// seed of the account that signs the transaction.
//
// You can add a *Options struct to the end of many methods, which set extra parameters on the
// submitted transaction. If you add new signers via Options, then sourceSeed will not be used to sign
// the transaction -- and it's okay to use a public address instead of a seed for sourceSeed.
// See examples for how to use Options.
//
// Amounts in Microstellar are typically represented as strings, to protect users from accidentaly
// performing floating point operations on them. The convenience methods `ParseAmount` and `ToAmountString`
// to convert the strings to `int64` values that represent a `10e7` multiple of the numeric
// value. E.g.,
//
//   ParseAmount("2.5") == int64(25000000)
//   ToAmountString(1000000) == "1.000000"
//
// You can use ErrorString(...) to extract the Horizon error from a returned error.
package microstellar

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/federation"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/clients/stellartoml"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/xdr"
)

// MicroStellar is the user handle to the Stellar network. Use the New function
// to create a new instance.
type MicroStellar struct {
	networkName string
	params      Params
	fake        bool
	tx          *Tx
	lastTx      *Tx
}

// Error wraps underlying errors (e.g., horizon)
type Error struct {
	HorizonError horizon.Error
}

// Params lets you add optional parameters to the common microstellar methods.
type Params map[string]interface{}

// New returns a new MicroStellar client connected that operates on the network
// specified by networkName. The supported networks are:
//
//    public: the public horizon network
//    test: the public horizon testnet
//    fake: a fake network used for tests
//    custom: a custom network specified by the parameters
//
// If you're using "custom", provide the URL and Passphrase to your
// horizon network server in the parameters.
//
//    New("custom", Params{
//        "url": "https://my-horizon-server.com",
//        "passphrase": "foobar"})
func New(networkName string, params ...Params) *MicroStellar {
	var p Params

	if len(params) > 0 {
		p = params[0]
	}

	return &MicroStellar{
		networkName: networkName,
		params:      p,
		fake:        networkName == "fake",
		tx:          nil,
	}
}

func (ms *MicroStellar) getTx() *Tx {
	var tx *Tx

	// If this is a multi-op transaction, then tx
	// contains the
	if ms.tx != nil {
		tx = ms.tx
	} else {
		tx = NewTx(ms.networkName, ms.params)
	}

	return tx
}

func (ms *MicroStellar) signAndSubmit(tx *Tx, signers ...string) error {
	if !tx.isMultiOp {
		tx.Sign(signers...)
		tx.Submit()
	}

	// Save last tx to keep response and error
	ms.lastTx = tx
	return tx.Err()
}

// Start begins a new multi-op transaction. This lets you lump a set of operations into
// a single transaction, and submit them together in one atomic step.
//
// You can pass in the signers and envelope fields (such as memotext, memoid, etc.) for the
// transaction as options. The signers must have signing authority on all the operations
// in the transaction.
//
// The fee for the transaction is billed to sourceSeed, which is typically a seed, but can
// be an address if differnt signers are used.
//
// Call microstellar.Submit() on the instance to close the transaction and send it to
// the network.
//
//   ms = microstellar.New("test")
//   ms.Start("sourceSeed", microstellar.Opts().WithMemoText("big op").WithSigner("signerSeed"))
//   ms.Pay("marys_address", "bobs_address", "2000", INR)
//   ms.Pay("marys_address", "bills_address", "2000", USD)
//   ms.SetMasterWeight("bobs_address", 0)
//   ms.SetHomeDomain("bobs_address", "qubit.sh")
//   ms.Submit()
//
func (ms *MicroStellar) Start(sourceSeed string, options ...*Options) *MicroStellar {
	ms.tx = NewTx(ms.networkName, ms.params).WithOptions(mergeOptions(options).MultiOp(sourceSeed))
	return ms
}

// Submit signs and submits a multi-op transaction to the network. See microstellar.Start() for
// details.
func (ms *MicroStellar) Submit() error {
	tx := ms.getTx()

	if !tx.isMultiOp {
		return errors.Errorf("can't submit, not a multi-op transaction")
	}

	ms.tx.Sign()
	ms.tx.Submit()

	// Save last tx to keep response and error
	ms.lastTx = tx
	tx = nil
	return ms.lastTx.Err()
}

// TxError returns the error from the last submission attempt.
func (ms *MicroStellar) TxError() error {
	return ms.lastTx.Err()
}

// TxResponse returns the response from the last submission.
func (ms *MicroStellar) TxResponse() *TxResponse {
	return ms.lastTx.Response()
}

// CreateKeyPair generates a new random key pair.
func (ms *MicroStellar) CreateKeyPair() (*KeyPair, error) {
	pair, err := keypair.Random()
	if err != nil {
		return nil, err
	}

	debugf("CreateKeyPair", "created address: %s, seed: <redacted>", pair.Address())
	return &KeyPair{pair.Seed(), pair.Address()}, nil
}

// FundAccount creates a new account out of addressOrSeed by funding it with lumens
// from sourceSeed. The minimum funding amount today is 0.5 XLM.
func (ms *MicroStellar) FundAccount(sourceSeed string, addressOrSeed string, amount string, options ...*Options) error {
	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("invalid source address or seed: %s", sourceSeed)
	}

	if !ValidAddressOrSeed(addressOrSeed) {
		return errors.Errorf("invalid target address or seed: %s", addressOrSeed)
	}

	payment := build.CreateAccount(
		build.Destination{AddressOrSeed: addressOrSeed},
		build.NativeAmount{Amount: amount})

	tx := ms.getTx()

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	tx.Build(sourceAccount(sourceSeed), payment)
	return ms.signAndSubmit(tx, sourceSeed)
}

// LoadAccount loads the account information for the given address.
func (ms *MicroStellar) LoadAccount(address string) (*Account, error) {
	if !ValidAddressOrSeed(address) {
		return nil, errors.Errorf("can't load account: invalid address or seed: %v", address)
	}

	if ms.fake {
		return newAccount(), nil
	}

	debugf("LoadAccount", "loading account: %s", address)
	tx := NewTx(ms.networkName, ms.params)
	account, err := tx.GetClient().LoadAccount(address)

	if err != nil {
		return nil, errors.Wrap(err, "could not load account")
	}

	return newAccountFromHorizon(account), nil
}

// Resolve looks up a federated address
func (ms *MicroStellar) Resolve(address string) (string, error) {
	debugf("Resolve", "looking up: %s", address)
	if !strings.Contains(address, "*") {
		return "", errors.Errorf("not a fedaration address: %s", address)
	}

	// Create a new federation client and lookup address
	var fedClient = &federation.Client{
		HTTP:        http.DefaultClient,
		Horizon:     NewTx(ms.networkName, ms.params).GetClient(),
		StellarTOML: stellartoml.DefaultClient,
	}

	resp, err := fedClient.LookupByAddress(address)

	if err != nil {
		return "", errors.Wrapf(err, "resolve error")
	}

	return resp.AccountID, nil
}

// PayNative makes a native asset payment of amount from source to target.
func (ms *MicroStellar) PayNative(sourceSeed string, targetAddress string, amount string, options ...*Options) error {
	return ms.Pay(sourceSeed, targetAddress, amount, NativeAsset, options...)
}

// Pay lets you make payments with credit assets.
//
//   USD := microstellar.NewAsset("USD", "ISSUERSEED", microstellar.Credit4Type)
//   ms.Pay("source_seed", "target_address", "3", USD, microstellar.Opts().WithMemoText("for shelter"))
//
// Pay also lets you make path payments. E.g., Mary pays Bob 2000 INR with XLM (lumens), using the
// path XLM -> USD -> EUR -> INR, spending no more than 20 XLM (lumens.)
//
//   XLM := microstellar.NativeAsset
//   USD := microstellar.NewAsset("USD", "ISSUERSEED", microstellar.Credit4Type)
//   EUR := microstellar.NewAsset("EUR", "ISSUERSEED", microstellar.Credit4Type)
//   INR := microstellar.NewAsset("INR", "ISSUERSEED", microstellar.Credit4Type)
//
//   ms.Pay("marys_seed", "bobs_address", "2000", INR,
//       microstellar.Opts().WithAsset(XLM, "20").Through(USD, EUR).WithMemoText("take your rupees!"))
//
// If you don't know what path to take ahead of time, use Options.FindPathFrom(sourceAddress) to
// find a path for you.
//
//   ms.Pay("marys_seed", "bobs_address", "2000", INR,
//       microstellar.Opts().WithAsset(XLM, "20").Through(USD, EUR).FindPathFrom("marys_address"))
func (ms *MicroStellar) Pay(sourceAddressOrSeed string, targetAddress string, amount string, asset *Asset, options ...*Options) error {
	if err := asset.Validate(); err != nil {
		return errors.Wrap(err, "can't pay")
	}

	if !ValidAddressOrSeed(sourceAddressOrSeed) {
		return errors.Errorf("can't pay: invalid source address or seed: %s", sourceAddressOrSeed)
	}

	if !ValidAddressOrSeed(targetAddress) {
		return errors.Errorf("can't pay: invalid address: %v", targetAddress)
	}

	paymentMuts := []interface{}{
		build.Destination{AddressOrSeed: targetAddress},
	}

	if asset.IsNative() {
		paymentMuts = append(paymentMuts, build.NativeAmount{Amount: amount})
	} else {
		paymentMuts = append(paymentMuts,
			build.CreditAmount{Code: asset.Code, Issuer: asset.Issuer, Amount: amount})
	}

	tx := ms.getTx()

	if len(options) > 0 {
		opts := options[0]
		tx.SetOptions(opts)

		// Is this a path payment?
		if opts.sendAsset != nil {
			debugf("Pay", "path payment: deposit %s with %s", asset.Code, opts.sendAsset.Code)
			payPath := build.PayWith(opts.sendAsset.ToStellarAsset(), opts.maxAmount)

			if len(opts.path) > 0 {
				for _, through := range opts.path {
					debugf("Pay", "path payment: through %s", through.Code)
					payPath = payPath.Through(through.ToStellarAsset())
				}
			} else {
				debugf("Pay", "no path specified, searching for paths from: %s", opts.sourceAddress)
				if err := ValidAddress(opts.sourceAddress); err != nil {
					return errors.Wrapf(err, "not a valid source address: %s", opts.sourceAddress)
				}

				paths, err := ms.FindPaths(opts.sourceAddress, targetAddress, asset, amount, Opts().WithAsset(opts.sendAsset, opts.maxAmount))
				if err != nil {
					return errors.Wrap(err, "path finding error")
				}

				if len(paths) < 1 {
					return errors.Errorf("no paths found from %s to %s", opts.sendAsset.Code, asset.Code)
				}

				for _, hop := range paths[0].Hops {
					debugf("Pay", "path payment: through %s", hop.Code)
					payPath = payPath.Through(hop.ToStellarAsset())
				}
			}

			paymentMuts = append(paymentMuts, payPath)
		}
	}

	tx.Build(sourceAccount(sourceAddressOrSeed), build.Payment(paymentMuts...))
	return ms.signAndSubmit(tx, sourceAddressOrSeed)
}

// CreateTrustLine creates a trustline from sourceSeed to asset, with the specified trust limit. An empty
// limit string indicates no limit.
func (ms *MicroStellar) CreateTrustLine(sourceSeed string, asset *Asset, limit string, options ...*Options) error {
	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("can't create trust line: invalid source address or seed: %s", sourceSeed)
	}

	if err := asset.Validate(); err != nil {
		return errors.Wrap(err, "can't create trust line")
	}

	tx := ms.getTx()

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	if limit == "" {
		tx.Build(sourceAccount(sourceSeed), build.Trust(asset.Code, asset.Issuer))
	} else {
		tx.Build(sourceAccount(sourceSeed), build.Trust(asset.Code, asset.Issuer, build.Limit(limit)))
	}

	return ms.signAndSubmit(tx, sourceSeed)
}

// RemoveTrustLine removes an trustline from sourceSeed to an asset.
func (ms *MicroStellar) RemoveTrustLine(sourceSeed string, asset *Asset, options ...*Options) error {
	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("can't remove trust line: invalid source address or seed: %s", sourceSeed)
	}

	if err := asset.Validate(); err != nil {
		return errors.Wrapf(err, "can't remove trust line")
	}

	tx := ms.getTx()

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	tx.Build(sourceAccount(sourceSeed), build.RemoveTrust(asset.Code, asset.Issuer))
	return ms.signAndSubmit(tx, sourceSeed)
}

// SetMasterWeight changes the master weight of sourceSeed.
func (ms *MicroStellar) SetMasterWeight(sourceSeed string, weight uint32, options ...*Options) error {
	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("can't set master weight: invalid source address or seed: %s", sourceSeed)
	}

	tx := ms.getTx()

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	tx.Build(sourceAccount(sourceSeed), build.MasterWeight(weight))
	return ms.signAndSubmit(tx, sourceSeed)
}

// AccountFlags are used by issuers of assets.
type AccountFlags int32

const (
	// FlagsNone disables all flags (can only be used alone.)
	FlagsNone = AccountFlags(0)

	// FlagAuthRequired requires the issuing account to give the receiving
	// account permission to hold the asset.
	FlagAuthRequired = AccountFlags(1)

	// FlagAuthRevocable allows the issuer to revoke the credit held by another
	// account.
	FlagAuthRevocable = AccountFlags(2)

	// FlagAuthImmutable means that the other auth parameters can never be set
	// and the issuer's account can never be deleted.
	FlagAuthImmutable = AccountFlags(4)
)

// SetFlags sets flags on the account.
func (ms *MicroStellar) SetFlags(sourceSeed string, flags AccountFlags, options ...*Options) error {
	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("can't set flags: invalid source address or seed: %s", sourceSeed)
	}

	tx := ms.getTx()

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	tx.Build(sourceAccount(sourceSeed), build.SetFlag(int32(flags)))
	return ms.signAndSubmit(tx, sourceSeed)
}

// ClearFlags clears the specified flags for the account.
func (ms *MicroStellar) ClearFlags(sourceSeed string, flags AccountFlags, options ...*Options) error {
	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("can't clear flags: invalid source address or seed: %s", sourceSeed)
	}

	tx := ms.getTx()

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	tx.Build(sourceAccount(sourceSeed), build.ClearFlag(int32(flags)))
	return ms.signAndSubmit(tx, sourceSeed)
}

// SetHomeDomain changes the home domain of sourceSeed.
func (ms *MicroStellar) SetHomeDomain(sourceSeed string, domain string, options ...*Options) error {
	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("can't set home domain: invalid source address or seed: %s", sourceSeed)
	}

	tx := ms.getTx()

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	tx.Build(sourceAccount(sourceSeed), build.HomeDomain(domain))
	return ms.signAndSubmit(tx, sourceSeed)
}

// AddSigner adds signerAddress as a signer to sourceSeed's account with weight signerWeight.
func (ms *MicroStellar) AddSigner(sourceSeed string, signerAddress string, signerWeight uint32, options ...*Options) error {
	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("can't add signer: invalid source address or seed: %s", sourceSeed)
	}

	if !ValidAddressOrSeed(signerAddress) {
		return errors.Errorf("can't add signer: invalid signer address or seed: %s", signerAddress)
	}

	tx := ms.getTx()

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	tx.Build(sourceAccount(sourceSeed), build.AddSigner(signerAddress, signerWeight))
	return ms.signAndSubmit(tx, sourceSeed)
}

// RemoveSigner removes signerAddress as a signer from sourceSeed's account.
func (ms *MicroStellar) RemoveSigner(sourceSeed string, signerAddress string, options ...*Options) error {
	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("can't remove signer: invalid source address or seed: %s", sourceSeed)
	}

	if !ValidAddressOrSeed(signerAddress) {
		return errors.Errorf("can't remove signer: invalid signer address or seed: %s", signerAddress)
	}

	tx := ms.getTx()

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	tx.Build(sourceAccount(sourceSeed), build.RemoveSigner(signerAddress))
	return ms.signAndSubmit(tx, sourceSeed)
}

// SetThresholds sets the signing thresholds for the account.
func (ms *MicroStellar) SetThresholds(sourceSeed string, low, medium, high uint32, options ...*Options) error {
	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("can't set thresholds: invalid source address or seed: %s", sourceSeed)
	}

	tx := ms.getTx()

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	tx.Build(sourceAccount(sourceSeed), build.SetThresholds(low, medium, high))
	return ms.signAndSubmit(tx, sourceSeed)
}

// SetData lets you attach (or update) arbitrary data to an account. The lengths of the key and value must each be
// less than 64 bytes.
func (ms *MicroStellar) SetData(sourceSeed string, key string, val []byte, options ...*Options) error {
	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("can't set thresholds: invalid source address or seed: %s", sourceSeed)
	}

	tx := ms.getTx()

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	if key == "" {
		return errors.Errorf("data key must not be empty")
	}

	if len(key) > 64 {
		return errors.Errorf("data key must be under 64 bytes: %s", key)
	}

	if len(val) > 64 {
		return errors.Errorf("data value must be under 64 bytes: %s", string(val))
	}

	tx.Build(sourceAccount(sourceSeed), build.SetData(key, val))
	return ms.signAndSubmit(tx, sourceSeed)
}

// ClearData removes attached data from an account.
func (ms *MicroStellar) ClearData(sourceSeed string, key string, options ...*Options) error {
	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("can't set thresholds: invalid source address or seed: %s", sourceSeed)
	}

	tx := ms.getTx()

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	if len(key) > 64 {
		return errors.Errorf("data key must be under 64 bytes: %s", key)
	}

	tx.Build(sourceAccount(sourceSeed), build.ClearData(key))
	return ms.signAndSubmit(tx, sourceSeed)
}

// SignTransaction signs a base64-encoded transaction envelope with the specified seeds
// for the current network.
func (ms *MicroStellar) SignTransaction(b64Tx string, seeds ...string) (string, error) {
	tx := ms.getTx()
	xdrTxe, err := DecodeTx(b64Tx)

	if err != nil {
		return "", errors.Wrap(err, "DecodeTx")
	}

	debugf("SignTransaction", "decoded transaction: %+v", xdrTxe)
	hash, err := network.HashTransaction(&xdrTxe.Tx, tx.network.Passphrase)

	if err != nil {
		return "", errors.Wrap(err, "hash failed")
	}

	for _, seed := range seeds {
		kp, err := keypair.Parse(seed)
		if err != nil {
			return "", errors.Wrap(err, "parse failed")
		}

		sig, err := kp.SignDecorated(hash[:])
		if err != nil {
			return "", errors.Wrap(err, "sign failed")
		}

		debugf("SignTransaction", "adding signature: %+v", sig)
		xdrTxe.Signatures = append(xdrTxe.Signatures, sig)
	}

	signedTx, err := xdr.MarshalBase64(xdrTxe)
	if err != nil {
		return "", errors.Wrap(err, "could not marshal transaction")
	}

	return signedTx, nil
}

// SubmitTransaction submits a base64-encoded transaction envelope to the Stellar network
func (ms *MicroStellar) SubmitTransaction(b64Tx string) (*TxResponse, error) {
	tx := ms.getTx()
	resp, err := tx.GetClient().SubmitTransaction(b64Tx)
	txResponse := TxResponse(resp)
	return &txResponse, err
}
