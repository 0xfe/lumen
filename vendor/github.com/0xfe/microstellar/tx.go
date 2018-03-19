package microstellar

import (
	"net/http"

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
	client        *horizon.Client
	networkName   string
	network       build.Network
	fake          bool
	options       *Options
	builder       *build.TransactionBuilder
	payload       string
	submitted     bool
	response      *horizon.TransactionSuccess
	isMultiOp     bool                       // is this a multi-op transaction
	ops           []build.TransactionMutator // all ops for multi-op
	sourceAccount string
	err           error
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
		network = build.PublicNetwork
		client = horizon.DefaultPublicNetClient
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
		isMultiOp:   false,
		ops:         []build.TransactionMutator{},
		err:         nil,
	}
}

// SetOptions sets the Tx options
func (tx *Tx) SetOptions(options *Options) {
	tx.options = options
	if options.isMultiOp {
		tx.Start(options.multiOpSource)
	}
}

// WithOptions sets the Tx options and returns the Tx
func (tx *Tx) WithOptions(options *Options) *Tx {
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

// TxResponse is returned by the horizon server for a successful transaction.
type TxResponse horizon.TransactionSuccess

// Response returns the horison response for the submitted operation.
func (tx *Tx) Response() *TxResponse {
	response := TxResponse(*tx.response)
	return &response
}

// Reset clears all internal sate, so you can run a new operation.
func (tx *Tx) Reset() {
	tx.options = nil
	tx.builder = nil
	tx.payload = ""
	tx.submitted = false
	tx.response = nil
	tx.isMultiOp = false
	tx.err = nil
}

func sourceAccount(addressOrSeed string) build.SourceAccount {
	return build.SourceAccount{AddressOrSeed: addressOrSeed}
}

// Start begins a new multi-op transaction with fees billed to account
func (tx *Tx) Start(account string) *Tx {
	tx.sourceAccount = account
	sourceAccount := sourceAccount(account)
	tx.ops = []build.TransactionMutator{
		build.TransactionMutator(sourceAccount),
		tx.network,
		build.AutoSequence{SequenceProvider: tx.client},
	}
	tx.isMultiOp = true

	return tx
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

	if tx.fake && !tx.isMultiOp {
		tx.builder = &build.TransactionBuilder{}
		return nil
	}

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

	if tx.isMultiOp {
		tx.ops = append(tx.ops, muts...)
	} else {
		muts = append([]build.TransactionMutator{
			sourceAccount,
			tx.network,
			build.AutoSequence{SequenceProvider: tx.client},
		}, muts...)

		builder, err := build.Transaction(muts...)
		tx.builder = builder
		tx.err = errors.Wrap(err, "could not build transaction")
	}
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

	if tx.builder == nil && !tx.isMultiOp {
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

	if tx.isMultiOp {
		tx.builder, err = build.Transaction(tx.ops...)
	}

	debugf("Tx.Sign", "signing transaction, seq: %v", tx.builder.TX.SeqNum)
	if tx.options != nil && len(tx.options.signerSeeds) > 0 {
		txe, err = tx.builder.Sign(tx.options.signerSeeds...)
	} else {
		if len(keys) == 0 {
			keys = []string{tx.sourceAccount}
		}
		txe, err = tx.builder.Sign(keys...)
	}

	if err != nil {
		tx.err = errors.Wrap(err, "signing error")
		return tx.err
	}

	tx.payload, err = txe.Base64()
	debugf("Tx.Sign", "signed transaction, payload: %s", tx.payload)

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

	debugf("Tx.Submit", "submitting transaction to network %s", tx.networkName)
	resp, err := tx.client.SubmitTransaction(tx.payload)

	if err != nil {
		debugf("Tx.Submit", "submit failed: %s", ErrorString(err))
		tx.err = errors.Wrap(err, "could not submit transaction")
		return tx.err
	}

	debugf("Tx.Submit", "transaction submitted to ledger %d with hash %s", int32(resp.Ledger), resp.Hash)
	tx.response = &resp
	tx.submitted = true
	return nil
}
