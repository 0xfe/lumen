package cli

// Implement a simple TTL based store.

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// A store entry
type entry struct {
	value     string
	expiresOn time.Time
	noexpire  bool
}

// Returns true if entry has expired
func (e entry) expired() bool {
	if e.noexpire {
		return false
	}

	if time.Now().After(e.expiresOn) {
		return true
	}

	return false
}

// Internal represents the internal memstore
type Internal struct {
	store      *Store
	entries    map[string]*entry
	ticker     *time.Ticker
	gcInterval time.Duration
	mu         *sync.RWMutex
}

// NewInternalStore returns a new initialized store with a 60s TTL
func NewInternalStore() (*Internal, error) {
	// Try to connect
	return &Internal{
		store: &Store{
			driver:     "internal",
			parameters: "",
		},
		entries:    make(map[string]*entry),
		ticker:     time.NewTicker(time.Second),
		gcInterval: 60 * time.Second,
		mu:         &sync.RWMutex{},
	}, nil
}

// start starts the garbage collector for expired entires
func (store *Internal) start() {
	go func() {
		for range store.ticker.C {
			store.expire()
		}
	}()
}

// expire deletes entries that have expired
func (store *Internal) expire() {
	log.WithFields(log.Fields{"type": "store", "store": "internal"}).Infof("running garbage collector")
	count := 0
	for k, v := range store.entries {
		store.mu.Lock()
		if v.expired() {
			delete(store.entries, k)
			count++
		}
		store.mu.Unlock()
	}
	log.WithFields(log.Fields{"type": "store", "store": "internal"}).Infof("garbage collector deleted %d entries", count)
}

// NumEntries returns the number of entries currently in the store
func (store *Internal) NumEntries() int {
	return len(store.entries)
}

// Set adds (or updates) an entry in the store. If 'ttl' is 0, the entry never expires
func (store *Internal) Set(k string, v string, ttl time.Duration) error {
	expiresOn := time.Now().Add(ttl)

	noexpire := false
	if ttl == 0 {
		noexpire = true
	}

	store.mu.Lock()
	store.entries[k] = &entry{v, expiresOn, noexpire}
	store.mu.Unlock()
	return nil
}

// Get looks up an entry in the store.
func (store *Internal) Get(k string) (string, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	v, ok := store.entries[k]
	if ok && !v.expired() {
		return v.value, nil
	}

	return "", fmt.Errorf("No value in store for key: %v", k)
}

// Delete removes an entry in the store.
func (store *Internal) Delete(k string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	if _, ok := store.entries[k]; ok {
		delete(store.entries, k)
		return nil
	}

	return fmt.Errorf("No value in store for key: %v", k)
}
