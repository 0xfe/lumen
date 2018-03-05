package microstellar

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
)

// Tx represents a unique stellar transaction. This is used by the MicroStellar
// library to abstract away the Horizon API and transport. To reuse Tx
// instances, you must call Tx.Reset() between operations.
//
// This struct is not thread-safe by design -- you must use separate instances
// in different goroutines.
//
// Unless you're hacking around in the guts, you should not need to use Tx.
type Tx struct {
	client      *horizon.Client
	networkName string
	network     build.Network
	fake        bool
	options     *TxOptions
	builder     *build.TransactionBuilder
	payload     string
	submitted   bool
	response    *horizon.TransactionSuccess
	err         error
}

// NewTx returns a new Tx that operates on the network specified by
// networkName. The supported networks are:
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
func NewTx(networkName string, params ...Params) *Tx {
	var network build.Network
	var client *horizon.Client

	fake := false

	switch networkName {
	case "public":
		network = build.TestNetwork
		client = horizon.DefaultTestNetClient
	case "test":
		network = build.TestNetwork
		client = horizon.DefaultTestNetClient
	case "fake":
		network = build.TestNetwork
		client = horizon.DefaultTestNetClient
		fake = true
	case "custom":
		network = build.Network{Passphrase: params[0]["passphrase"].(string)}
		client = &horizon.Client{
			URL:  params[0]["url"].(string),
			HTTP: http.DefaultClient,
		}
	default:
		// use the test network
		network = build.TestNetwork
		client = horizon.DefaultTestNetClient
	}

	return &Tx{
		networkName: networkName,
		client:      client,
		network:     network,
		fake:        fake,
		options:     nil,
		builder:     nil,
		payload:     "",
		submitted:   false,
		response:    nil,
		err:         nil,
	}
}

// SetOptions sets the Tx options
func (tx *Tx) SetOptions(options *TxOptions) {
	tx.options = options
}

// WithOptions sets the Tx options and returns the Tx
func (tx *Tx) WithOptions(options *TxOptions) *Tx {
	tx.SetOptions(options)
	return tx
}

// GetClient returns the underlying horizon client handle.
func (tx *Tx) GetClient() *horizon.Client {
	return tx.client
}

// Err returns the error from the most recent failed operation.
func (tx *Tx) Err() error {
	return tx.err
}

// Response returns the horison response for the submitted operation.
func (tx *Tx) Response() string {
	return fmt.Sprintf("%v", tx.response)
}

// Reset clears all internal state, so you can run a new operation.
func (tx *Tx) Reset() {
	tx.options = nil
	tx.builder = nil
	tx.payload = ""
	tx.submitted = false
	tx.response = nil
	tx.err = nil
}

func sourceAccount(addressOrSeed string) build.SourceAccount {
	return build.SourceAccount{AddressOrSeed: addressOrSeed}
}

// Build creates a new operation out of the provided mutators.
func (tx *Tx) Build(sourceAccount build.TransactionMutator, muts ...build.TransactionMutator) error {
	if tx.err != nil {
		return tx.err
	}

	if tx.builder != nil {
		tx.err = errors.Errorf("transaction already built")
		return tx.err
	}

	if tx.fake {
		tx.builder = &build.TransactionBuilder{}
		return nil
	}

	muts = append([]build.TransactionMutator{
		sourceAccount,
		tx.network,
		build.AutoSequence{SequenceProvider: tx.client},
	}, muts...)

	if tx.options != nil {
		switch tx.options.memoType {
		case MemoText:
			if len(tx.options.memoText) > 28 {
				return errors.Errorf("memo text >28 bytes: %v", tx.options.memoText)
			}
			muts = append(muts, build.MemoText{Value: tx.options.memoText})
		case MemoID:
			muts = append(muts, build.MemoID{Value: tx.options.memoID})
		}
	}

	builder, err := build.Transaction(muts...)
	tx.builder = builder
	tx.err = errors.Wrap(err, "could not build transaction")
	return tx.err
}

// IsSigned returns true of the transaction is signed.
func (tx *Tx) IsSigned() bool {
	return tx.payload != ""
}

// Sign signs the transaction with every key in keys.
func (tx *Tx) Sign(keys ...string) error {
	if tx.err != nil {
		return tx.err
	}

	if tx.builder == nil {
		tx.err = errors.Errorf("can't sign empty transaction")
		return tx.err
	}

	if tx.IsSigned() {
		tx.err = errors.Errorf("transaction already signed")
		return tx.err
	}

	if tx.fake {
		tx.payload = "FAKE"
		return nil
	}

	var txe build.TransactionEnvelopeBuilder
	var err error

	if tx.options != nil && len(tx.options.signerSeeds) > 0 {
		txe, err = tx.builder.Sign(tx.options.signerSeeds...)
	} else {
		txe, err = tx.builder.Sign(keys...)
	}

	if err != nil {
		tx.err = errors.Wrap(err, "signing error")
		return tx.err
	}

	tx.payload, err = txe.Base64()

	if err != nil {
		tx.err = errors.Wrap(err, "base64 conversion error")
		return tx.err
	}

	return nil
}

// Submit sends the transaction to the stellar network.
func (tx *Tx) Submit() error {
	if tx.err != nil {
		return tx.err
	}

	if !tx.IsSigned() {
		tx.err = errors.Errorf("transaction not signed")
		return tx.err
	}

	if tx.submitted {
		tx.err = errors.Errorf("transaction already submitted")
		return tx.err
	}

	if tx.fake {
		tx.response = &horizon.TransactionSuccess{Result: "fake_ok"}
		return nil
	}

	resp, err := tx.client.SubmitTransaction(tx.payload)

	if err != nil {
		tx.err = errors.Wrap(err, "could not submit transaction")
		return tx.err
	}

	tx.response = &resp
	tx.submitted = true
	return nil
}

// MemoType sets the memotype field on the payment request.
type MemoType int

const (
	MemoNone   = MemoType(0) // No memo
	MemoID     = MemoType(1) // ID memo
	MemoText   = MemoType(2) // Text memo (max 28 chars)
	MemoHash   = MemoType(3) // Hash memo
	MemoReturn = MemoType(4) // Return hash memo
)

// TxOptions are additional parameters for a transaction. Use Opts() or NewTxOptions()
// to create a new instance.
type TxOptions struct {
	// Use With* methods to set these options
	hasFee bool
	fee    uint32

	hasTimeBounds bool
	timeBounds    time.Duration

	memoType MemoType // defaults to no memo
	memoText string   // additional memo text
	memoID   uint64   // additional memo ID

	signerSeeds []string

	// Microstellar options
	hasCursor bool
	cursor    string
	ctx       context.Context
}

// NewTxOptions creates a new options structure for Tx.
func NewTxOptions() *TxOptions {
	return &TxOptions{
		hasFee:        false,
		hasTimeBounds: false,
		memoType:      MemoNone,
		hasCursor:     false,
		ctx:           nil,
	}
}

// Opts is just an alias for NewTxOptions
func Opts() *TxOptions {
	return NewTxOptions()
}

// WithMemoText sets the memoType and memoText fields on Payment p
func (o *TxOptions) WithMemoText(text string) *TxOptions {
	o.memoType = MemoText
	o.memoText = text
	return o
}

// WithMemoID sets the memoType and memoID fields on Payment p
func (o *TxOptions) WithMemoID(id uint64) *TxOptions {
	o.memoType = MemoID
	o.memoID = id
	return o
}

// WithSigner adds a signer to Payment p
func (o *TxOptions) WithSigner(signerSeed string) *TxOptions {
	o.signerSeeds = append(o.signerSeeds, signerSeed)
	return o
}

// WithContext sets the context.Context for the connection
func (o *TxOptions) WithContext(context context.Context) *TxOptions {
	o.ctx = context
	return o
}

// WithCursor sets the cursor for watchers
func (o *TxOptions) WithCursor(cursor string) *TxOptions {
	o.hasCursor = true
	o.cursor = cursor
	return o
}
