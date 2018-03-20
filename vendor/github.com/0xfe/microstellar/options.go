package microstellar

import (
	"context"
	"time"
)

// SortOrder is used with WithSortOrder
type SortOrder int

// For use with WithSortOrder
const (
	SortAscending  = SortOrder(0)
	SortDescending = SortOrder(1)
)

// MemoType sets the memotype field on the payment request.
type MemoType int

// Supported memo types.
const (
	MemoNone   = MemoType(0) // No memo
	MemoID     = MemoType(1) // ID memo
	MemoText   = MemoType(2) // Text memo (max 28 chars)
	MemoHash   = MemoType(3) // Hash memo
	MemoReturn = MemoType(4) // Return hash memo
)

// Event denotes a specific event for a handler.
type Event int

// Supported events.
const (
	EvBeforeSubmit = Event(0)
)

// TxHandler is a custom function that can be called when certain events occur. If
// false is returned or if the method returns an error, the caller stops processing the
// event immediately.
type TxHandler func(data ...interface{}) (bool, error)

// Options are additional parameters for a transaction. Use Opts() or NewOptions()
// to create a new instance.
type Options struct {
	// Defaults to context.Background if unset.
	ctx      context.Context
	handlers map[Event]*TxHandler

	// Use With* methods to set these options
	hasFee        bool
	fee           uint32
	hasTimeBounds bool
	timeBounds    time.Duration

	// Used by all transactions.
	memoType MemoType // defaults to no memo
	memoText string   // additional memo text
	memoID   uint64   // additional memo ID

	skipSignatures bool
	signerSeeds    []string

	// Options for query methods (Watch*, Load*)
	hasCursor      bool
	cursor         string
	hasLimit       bool
	limit          uint
	sortDescending bool

	// For offer management.
	passiveOffer bool

	// for Path payments.
	sourceAddress string
	sendAsset     *Asset
	maxAmount     string
	path          []*Asset

	// for multi-op transactions
	isMultiOp     bool
	multiOpSource string
}

// NewOptions creates a new options structure for Tx.
func NewOptions() *Options {
	return &Options{
		ctx:            nil,
		handlers:       map[Event]*TxHandler{},
		hasFee:         false,
		hasTimeBounds:  false,
		memoType:       MemoNone,
		hasCursor:      false,
		hasLimit:       false,
		sortDescending: false,
		passiveOffer:   false,
		sourceAddress:  "",
		isMultiOp:      false,
		multiOpSource:  "",
	}
}

// Opts is just an alias for NewOptions
func Opts() *Options {
	return NewOptions()
}

// mergeOptions takes a slice of Options and merges them.
func mergeOptions(opts []*Options) *Options {
	// for now, just return the first option
	if len(opts) > 0 {
		return opts[0]
	}

	return NewOptions()
}

// WithMemoText sets the memoType and memoText fields on a Payment. Used
// with all transactions.
func (o *Options) WithMemoText(text string) *Options {
	o.memoType = MemoText
	o.memoText = text
	return o
}

// WithMemoID sets the memoType and memoID fields on a Payment. Used
// with all transactions.
func (o *Options) WithMemoID(id uint64) *Options {
	o.memoType = MemoID
	o.memoID = id
	return o
}

// WithSigner adds a signer to Payment. Used with all transactions.
func (o *Options) WithSigner(signerSeed string) *Options {
	o.signerSeeds = append(o.signerSeeds, signerSeed)
	return o
}

// WithContext sets the context.Context for the connection. Used with
// Watch* methods.
func (o *Options) WithContext(context context.Context) *Options {
	o.ctx = context
	return o
}

// WithCursor sets the cursor for watchers and queries. Used with Watch*
// methods and LoadOffers.
func (o *Options) WithCursor(cursor string) *Options {
	o.hasCursor = true
	o.cursor = cursor
	return o
}

// WithLimit sets the limit for queries. Used with LoadOffers.
func (o *Options) WithLimit(limit uint) *Options {
	o.hasLimit = true
	o.limit = limit
	return o
}

// WithSortOrder sets the sort order of the results. Used with LoadOffers.
func (o *Options) WithSortOrder(order SortOrder) *Options {
	if order == SortDescending {
		o.sortDescending = true
	}
	return o
}

// MakePassive turns this into a passive offer. Used with LoadOffers.
func (o *Options) MakePassive() *Options {
	o.passiveOffer = true
	return o
}

// WithAsset is used to setup a path payment. This makes the Pay method
// use "asset" as the sending asset, and sends no more than maxAmount units
// of the asset. Used with Pay and FindPaths.
//
// E.g.,
//   ms.Pay(sourceSeed, address, "20", INR, Opts().WithAsset(NativeAsset, "20").Through(USD, EUR)
func (o *Options) WithAsset(asset *Asset, maxAmount string) *Options {
	o.sendAsset = asset
	o.maxAmount = maxAmount
	return o
}

// Through adds "asset" as a routing point in the payment path.
//
// E.g.,
//   ms.Pay(sourceSeed, address, "20", INR, Opts().WithAsset(NativeAsset, "20").Through(USD, EUR)
func (o *Options) Through(asset ...*Asset) *Options {
	o.path = append(o.path, asset...)
	return o
}

// FindPathFrom enables automatic path finding for path payments. Use sourceAddress
// to specify the address (not seed) for the source account.
func (o *Options) FindPathFrom(sourceAddress string) *Options {
	o.sourceAddress = sourceAddress
	return o
}

// MultiOp specifies that this is a multi-op transactions, and sets the fund source account
// to sourceAccount.
func (o *Options) MultiOp(sourceAccount string) *Options {
	o.isMultiOp = true
	o.multiOpSource = sourceAccount
	return o
}

// On attaches a handler to an event. E.g.,
//
//   Opts().On(microstellar.EvBeforeSubmit, func(tx) { log.Print(tx); return nil })
func (o *Options) On(event Event, handler *TxHandler) *Options {
	o.handlers[event] = handler
	return o
}

// SkipSignatures prevents Tx from signing transactions. This is typically done if the
// transaction is not meant to be submitted.
func (o *Options) SkipSignatures() *Options {
	o.skipSignatures = true
	return o
}

// TxOptions is a deprecated alias for TxOptoins
type TxOptions Options
