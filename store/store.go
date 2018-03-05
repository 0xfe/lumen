package store

import (
	"time"

	"github.com/pkg/errors"
	"github.com/stellar/go/support/log"
)

// StoreAPI abstracts the storage backend.
type API interface {
	Set(k string, v string, ttl time.Duration) error
	Get(k string) (string, error)
	Delete(k string) error
}

// Store represents the storage backend. Currently, only "internal" and "redis" are supported.
type Store struct {
	driver     string
	parameters string
}

func NewStore(driver, parameters string) (API, error) {
	switch driver {
	case "redis":
		store, err := NewRedisStore(parameters)
		return store, err
	case "internal":
		return NewInternalStore()
	case "file":
		return NewFileStore(parameters)
	case "dummy":
		return &DummyStore{&Store{"dummy", "dummy"}}, nil
	}

	return nil, errors.Errorf("Driver not found: %s", driver)
}

type DummyStore struct {
	store *Store
}

func (store *DummyStore) Set(k string, v string, ttl time.Duration) error {
	log.Errorf("Calling Set() on dummy store does nothing")
	return errors.Errorf("Dummy store stores nothing!")
}

func (store *DummyStore) Get(k string) (string, error) {
	log.Errorf("Calling Set() on dummy store does nothing")
	return "", errors.Errorf("Dummy store stores nothing!")
}

func (store *DummyStore) Delete(k string) error {
	return errors.Errorf("Dummy store stores nothing!")
}
