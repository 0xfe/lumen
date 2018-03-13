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
// You can use ErrorString(...) to extract the Horizon error from a returned error.
package microstellar

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/federation"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/clients/stellartoml"
	"github.com/stellar/go/keypair"
)

// MicroStellar is the user handle to the Stellar network. Use the New function
// to create a new instance.
type MicroStellar struct {
	networkName string
	params      Params
	fake        bool
}

// Error wraps underlying errors (e.g., horizon)
type Error struct {
	HorizonError horizon.Error
}

// Params lets you add optional parameters to New and NewTx.
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
//    NewTx("custom", Params{
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
	}
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

	tx := NewTx(ms.networkName, ms.params)

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	tx.Build(sourceAccount(sourceSeed), payment)
	tx.Sign(sourceSeed)
	tx.Submit()
	return tx.Err()
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
func (ms *MicroStellar) Pay(sourceSeed string, targetAddress string, amount string, asset *Asset, options ...*Options) error {
	if err := asset.Validate(); err != nil {
		return errors.Wrap(err, "can't pay")
	}

	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("can't pay: invalid source address or seed: %s", sourceSeed)
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

	tx := NewTx(ms.networkName, ms.params)

	if len(options) > 0 {
		opts := options[0]
		tx.SetOptions(opts)

		// Is this a path payment?
		if opts.sendAsset != nil {
			debugf("Pay", "path payment: deposit %s with %s", asset.Code, opts.sendAsset.Code)
			payPath := build.PayWith(opts.sendAsset.ToStellarAsset(), opts.maxAmount)

			for _, through := range opts.path {
				debugf("Pay", "path payment: through %s", through.Code)
				payPath = payPath.Through(through.ToStellarAsset())
			}

			paymentMuts = append(paymentMuts, payPath)
		}
	}

	tx.Build(sourceAccount(sourceSeed), build.Payment(paymentMuts...))
	tx.Sign(sourceSeed)
	tx.Submit()
	return tx.Err()
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

	tx := NewTx(ms.networkName, ms.params)

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	if limit == "" {
		tx.Build(sourceAccount(sourceSeed), build.Trust(asset.Code, asset.Issuer))
	} else {
		tx.Build(sourceAccount(sourceSeed), build.Trust(asset.Code, asset.Issuer, build.Limit(limit)))
	}

	tx.Sign(sourceSeed)
	tx.Submit()
	return tx.Err()
}

// RemoveTrustLine removes an trustline from sourceSeed to an asset.
func (ms *MicroStellar) RemoveTrustLine(sourceSeed string, asset *Asset, options ...*Options) error {
	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("can't remove trust line: invalid source address or seed: %s", sourceSeed)
	}

	if err := asset.Validate(); err != nil {
		return errors.Wrapf(err, "can't remove trust line")
	}

	tx := NewTx(ms.networkName, ms.params)

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	tx.Build(sourceAccount(sourceSeed), build.RemoveTrust(asset.Code, asset.Issuer))
	tx.Sign(sourceSeed)
	tx.Submit()
	return tx.Err()
}

// SetMasterWeight changes the master weight of sourceSeed.
func (ms *MicroStellar) SetMasterWeight(sourceSeed string, weight uint32, options ...*Options) error {
	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("can't set master weight: invalid source address or seed: %s", sourceSeed)
	}

	tx := NewTx(ms.networkName, ms.params)

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	tx.Build(sourceAccount(sourceSeed), build.MasterWeight(weight))
	tx.Sign(sourceSeed)
	tx.Submit()
	return tx.Err()
}

// AccountFlags are used by issuers of assets.
type AccountFlags int32

const (
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

// SetFlags changes the flags for the account.
func (ms *MicroStellar) SetFlags(sourceSeed string, flags AccountFlags, options ...*Options) error {
	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("can't set flags: invalid source address or seed: %s", sourceSeed)
	}

	tx := NewTx(ms.networkName, ms.params)

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	tx.Build(sourceAccount(sourceSeed), build.SetFlag(int32(flags)))
	tx.Sign(sourceSeed)
	tx.Submit()
	return tx.Err()
}

// SetHomeDomain changes the home domain of sourceSeed.
func (ms *MicroStellar) SetHomeDomain(sourceSeed string, domain string, options ...*Options) error {
	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("can't set home domain: invalid source address or seed: %s", sourceSeed)
	}

	tx := NewTx(ms.networkName, ms.params)

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	tx.Build(sourceAccount(sourceSeed), build.HomeDomain(domain))
	tx.Sign(sourceSeed)
	tx.Submit()
	return tx.Err()
}

// AddSigner adds signerAddress as a signer to sourceSeed's account with weight signerWeight.
func (ms *MicroStellar) AddSigner(sourceSeed string, signerAddress string, signerWeight uint32, options ...*Options) error {
	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("can't add signer: invalid source address or seed: %s", sourceSeed)
	}

	if !ValidAddressOrSeed(signerAddress) {
		return errors.Errorf("can't add signer: invalid signer address or seed: %s", signerAddress)
	}

	tx := NewTx(ms.networkName, ms.params)

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	tx.Build(sourceAccount(sourceSeed), build.AddSigner(signerAddress, signerWeight))
	tx.Sign(sourceSeed)
	tx.Submit()
	return tx.Err()
}

// RemoveSigner removes signerAddress as a signer from sourceSeed's account.
func (ms *MicroStellar) RemoveSigner(sourceSeed string, signerAddress string, options ...*Options) error {
	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("can't remove signer: invalid source address or seed: %s", sourceSeed)
	}

	if !ValidAddressOrSeed(signerAddress) {
		return errors.Errorf("can't remove signer: invalid signer address or seed: %s", signerAddress)
	}

	tx := NewTx(ms.networkName, ms.params)

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	tx.Build(sourceAccount(sourceSeed), build.RemoveSigner(signerAddress))
	tx.Sign(sourceSeed)
	tx.Submit()
	return tx.Err()
}

// SetThresholds sets the signing thresholds for the account.
func (ms *MicroStellar) SetThresholds(sourceSeed string, low, medium, high uint32, options ...*Options) error {
	if !ValidAddressOrSeed(sourceSeed) {
		return errors.Errorf("can't set thresholds: invalid source address or seed: %s", sourceSeed)
	}

	tx := NewTx(ms.networkName, ms.params)

	if len(options) > 0 {
		tx.SetOptions(options[0])
	}

	tx.Build(sourceAccount(sourceSeed), build.SetThresholds(low, medium, high))
	tx.Sign(sourceSeed)
	tx.Submit()
	return tx.Err()
}

// Payment represents a finalized payment in the ledger. You can subscribe to payments
// on the stellar network via the WatchPayments call.
type Payment horizon.Payment

// NewPaymentFromHorizon converts a horizon JSON payment struct to Payment
func NewPaymentFromHorizon(p *horizon.Payment) *Payment {
	payment := Payment(*p)
	return &payment
}

// PaymentWatcher is returned by WatchPayments, which watches the ledger for payments
// to and from an address.
type PaymentWatcher struct {
	// Ch gets a *Payment everytime there's a new entry in the ledger.
	Ch chan *Payment

	// Call Done to stop watching the ledger. This closes Ch.
	Done func()

	// This is set if the stream terminates unexpectedly. Safe to check
	// after Ch is closed.
	Err *error
}

// WatchPayments watches the ledger for payments to and from address and streams them on a channel . Use
// Options.WithContext to set a context.Context, and Options.WithCursor to set a cursor.
func (ms *MicroStellar) WatchPayments(address string, options ...*Options) (*PaymentWatcher, error) {
	debugf("WatchPayments", "watching address: %s", address)
	if err := ValidAddress(address); err != nil {
		return nil, errors.Errorf("can't watch payments, invalid address: %s", address)
	}

	tx := NewTx(ms.networkName, ms.params)

	var cursor *horizon.Cursor
	var ctx context.Context
	var cancelFunc func()

	if len(options) > 0 {
		tx.SetOptions(options[0])
		if options[0].hasCursor {
			// Ugh! Why do I have to do this?
			c := horizon.Cursor(options[0].cursor)
			cursor = &c
			debugf("WatchPayments", "starting stream for address: %s at cursor: %s", address, string(*cursor))
		}
		ctx = options[0].ctx
	}

	if ctx == nil {
		ctx, cancelFunc = context.WithCancel(context.Background())
	} else {
		ctx, cancelFunc = context.WithCancel(ctx)
	}

	var streamError error
	ch := make(chan *Payment)

	go func(ch chan *Payment, streamError *error) {
		if tx.fake {
		out:
			for {
				select {
				case <-ctx.Done():
					break out
				default:
					// continue
				}
				ch <- &Payment{From: "FAKESOURCE", To: "FAKEDEST", Type: "payment", AssetCode: "QBIT", Amount: "5"}
				time.Sleep(200 * time.Millisecond)
			}
		} else {
			err := tx.GetClient().StreamPayments(ctx, address, cursor, func(payment horizon.Payment) {
				debugf("WatchPayments", "found payment (%s) at %s, loading memo", payment.Type, address)
				tx.GetClient().LoadMemo(&payment)
				ch <- NewPaymentFromHorizon(&payment)
			})

			if err != nil {
				debugf("WatchPayments", "stream unexpectedly disconnected", err)
				*streamError = errors.Wrapf(err, "payment stream disconnected")
				cancelFunc()
			}
		}

		close(ch)
	}(ch, &streamError)

	return &PaymentWatcher{ch, cancelFunc, &streamError}, nil
}
