package store

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// StoreAPI abstracts the storage backend.
type StoreAPI interface {
	Set(k string, v string, ttl time.Duration) error
	Get(k string) (string, error)
	Delete(k string) error
}

// Backend represents the storage backend. Currently, only "internal" and "redis" are supported.
type Store struct {
	driver     string
	parameters string
}

func NewStore(driver, parameters string) (StoreAPI, error) {
	if driver == "redis" {
		store, err := NewRedisStore(parameters)
		return store, err
	} else if driver == "internal" {
		return NewInternalStore()
	} else if driver == "none" || driver == "" {
		return &DummyStore{&Store{"dummy", "dummy"}}, nil
	}

	log.WithFields(log.Fields{"type": "store"}).Errorf("Driver not found: %s", driver)
	return nil, fmt.Errorf("Driver not found: %s", driver)
}

type DummyStore struct {
	store *Store
}

func (store *DummyStore) Set(k string, v string, ttl time.Duration) error {
	return nil
}

func (store *DummyStore) Get(k string) (string, error) {
	return "", fmt.Errorf("Dummy store stores nothing!")
}

func (store *DummyStore) Delete(k string) error {
	return nil
}
