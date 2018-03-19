package microstellar

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/stellar/go/clients/horizon"
)

// Watcher is an abstract watcher struct.
type Watcher struct {
	// Call Done to stop watching the ledger. This closes Ch.
	Done func()

	// This is set if the stream terminates unexpectedly. Safe to check
	// after Ch is closed.
	Err *error
}

// Ledger represents an entry in the ledger. You can subscribe a continuous stream of ledger
// updates on the Stellar network via the WatchLedgers call.
type Ledger horizon.Ledger

// LedgerWatcher is returned by WatchLedgers, which watches the stellar network for ledger updates.
type LedgerWatcher struct {
	Watcher

	// Ch gets a *Ledger everytime there's a new entry.
	Ch chan *Ledger
}

// WatchLedgers watches the the stellar network for entries and streams them to LedgerWatcher.Ch. Use
// Options.WithContext to set a context.Context, and Options.WithCursor to set a cursor.
func (ms *MicroStellar) WatchLedgers(options ...*Options) (*LedgerWatcher, error) {
	var streamError error
	w := &LedgerWatcher{
		Ch:      make(chan *Ledger),
		Watcher: Watcher{Err: &streamError, Done: func() {}},
	}

	watcherFunc := func(params streamParams) {
		if params.tx.fake {
			w.Ch <- &Ledger{ID: "fake", TotalCoins: "0"}
			return
		}

		err := params.tx.GetClient().StreamLedgers(params.ctx, params.cursor, func(ledger horizon.Ledger) {
			debugf("WatchLedger", "entry (%s) total_coins: %s, tx_count: %v, op_count: %v", ledger.ID, ledger.TotalCoins, ledger.TransactionCount, ledger.OperationCount)
			l := Ledger(ledger)
			w.Ch <- &l
		})

		if err != nil {
			debugf("WatchLedger", "stream unexpectedly disconnected", err)
			*w.Err = errors.Wrapf(err, "stream disconnected")
			w.Done()
		}

		close(w.Ch)
	}

	cancelFunc, err := ms.watch("ledger", "", watcherFunc, options...)
	w.Done = cancelFunc

	return w, err
}

// Transaction represents a finalized transaction in the ledger. You can subscribe to transactions
// on the stellar network via the WatchTransactions call.
type Transaction horizon.Transaction

// TransactionWatcher is returned by WatchTransactions, which watches the ledger for transactions
// to and from an address.
type TransactionWatcher struct {
	Watcher

	// Ch gets a *Transaction everytime there's a new entry in the ledger.
	Ch chan *Transaction
}

// WatchTransactions watches the ledger for transactions to and from address and streams them on a channel . Use
// Options.WithContext to set a context.Context, and Options.WithCursor to set a cursor.
func (ms *MicroStellar) WatchTransactions(address string, options ...*Options) (*TransactionWatcher, error) {
	var streamError error
	w := &TransactionWatcher{
		Ch:      make(chan *Transaction),
		Watcher: Watcher{Err: &streamError, Done: func() {}},
	}

	watcherFunc := func(params streamParams) {
		if params.tx.fake {
			w.Ch <- &Transaction{Account: "FAKE"}
			return
		}

		err := params.tx.GetClient().StreamTransactions(params.ctx, params.address, params.cursor, func(transaction horizon.Transaction) {
			debugf("WatchTransaction", "found transaction (%s) on %s", transaction.ID, transaction.Account)
			t := Transaction(transaction)
			w.Ch <- &t
		})

		if err != nil {
			debugf("WatchTransaction", "stream unexpectedly disconnected", err)
			*w.Err = errors.Wrapf(err, "stream disconnected")
			w.Done()
		}

		close(w.Ch)
	}

	cancelFunc, err := ms.watch("transaction", address, watcherFunc, options...)
	w.Done = cancelFunc

	return w, err
}

// Payment represents a finalized payment in the ledger. You can subscribe to payments
// on the stellar network via the WatchPayments call.
type Payment horizon.Payment

// PaymentWatcher is returned by WatchPayments, which watches the ledger for payments
// to and from an address.
type PaymentWatcher struct {
	Watcher

	// Ch gets a *Payment everytime there's a new entry in the ledger.
	Ch chan *Payment
}

// WatchPayments watches the ledger for payments to and from address and streams them on a channel . Use
// Options.WithContext to set a context.Context, and Options.WithCursor to set a cursor.
func (ms *MicroStellar) WatchPayments(address string, options ...*Options) (*PaymentWatcher, error) {
	var streamError error
	w := &PaymentWatcher{
		Ch:      make(chan *Payment),
		Watcher: Watcher{Err: &streamError, Done: func() {}},
	}

	watcherFunc := func(params streamParams) {
		if params.tx.fake {
			w.Ch <- &Payment{Type: "fake"}
			return
		}

		err := params.tx.GetClient().StreamPayments(params.ctx, params.address, params.cursor, func(payment horizon.Payment) {
			debugf("WatchPayments", "found payment (%s) at %s, loading memo", payment.Type, address)
			params.tx.GetClient().LoadMemo(&payment)
			p := Payment(payment)
			w.Ch <- &p
		})

		if err != nil {
			debugf("WatchPayment", "stream unexpectedly disconnected", err)
			*w.Err = errors.Wrapf(err, "stream disconnected")
			w.Done()
		}

		close(w.Ch)
	}

	cancelFunc, err := ms.watch("payment", address, watcherFunc, options...)
	w.Done = cancelFunc

	return w, err
}

// streamParams is sent to streamFunc with the parameters for a horizon stream.
type streamParams struct {
	ctx        context.Context
	tx         *Tx
	cursor     *horizon.Cursor
	address    string
	cancelFunc func()
	err        *error
}

// streamFunc starts a horizon stream with the specified parameters.
type streamFunc func(streamParams)

// watch is a helper method to work with the Horizon Stream* methods. Returns a cancelFunc and error.
func (ms *MicroStellar) watch(entity string, address string, streamer streamFunc, options ...*Options) (func(), error) {
	logField := fmt.Sprintf("watch:%s", entity)
	debugf(logField, "watching address: %s", address)

	if err := ValidAddress(address); address != "" && err != nil {
		return nil, errors.Errorf("can't watch %s, invalid address: %s", entity, address)
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
			debugf(logField, "starting stream at cursor: %s", string(*cursor))
		}
		ctx = options[0].ctx
	}

	if ctx == nil {
		ctx, cancelFunc = context.WithCancel(context.Background())
	} else {
		ctx, cancelFunc = context.WithCancel(ctx)
	}

	go func() {
		if tx.fake {
		out:
			for {
				select {
				case <-ctx.Done():
					break out
				default:
					// continue
				}
				streamer(streamParams{ctx: ctx, tx: tx, cursor: cursor, address: address, cancelFunc: cancelFunc})
				time.Sleep(200 * time.Millisecond)
			}
		} else {
			streamer(streamParams{
				ctx:        ctx,
				tx:         tx,
				cursor:     cursor,
				address:    address,
				cancelFunc: cancelFunc,
			})
		}
	}()

	return cancelFunc, nil
}
