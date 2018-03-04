package store

// Test Redis Store
//
// Make sure redis is running:
// $ docker run -it -p 6379:6379 redis:alpine

import (
	"log"
	"testing"
)

func TestRedisStore_BasicLookup(t *testing.T) {
	store, err := NewStore("redis", "localhost:6379")

	if err != nil {
		log.Printf("skipping tests: couldn't setup internal store, want %v, got %v", nil, err)
		return
	}

	testBasicLookup(t, store)
}

func TestRedisStore_TTL(t *testing.T) {
	store, err := NewStore("redis", "localhost:6379")

	if err != nil {
		log.Printf("skipping tests: couldn't setup internal store, want %v, got %v", nil, err)
		return
	}

	testTTL(t, store)
}
