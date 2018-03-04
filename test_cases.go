package cli

import (
	"testing"
	"time"
)

func testBasicLookup(t *testing.T, store StoreAPI) {
	store.Set("foo", "bar", 0)
	v, err := store.Get("foo")

	if err != nil {
		t.Errorf("couldn't get value for 'foo': %v", err)
	}

	if v != "bar" {
		t.Errorf("incorrect value in store: want %v, got %v", "bar", v)
	}

	store.Delete("foo")
}

// Test store expiry. This is hacky because we're using time.Sleep(), but an okay
// tradeoff given the simplicity.
func testTTL(t *testing.T, store StoreAPI) {
	store.Set("mo", "bar", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	v, err := store.Get("mo")
	if err == nil {
		t.Errorf("entry not expired: want error, got %v", v)
	}

	store.Set("mo", "bar", 1*time.Second)
	time.Sleep(20 * time.Millisecond)

	v, err = store.Get("mo")
	if err != nil {
		t.Errorf("couldn't get value for 'mo': %v", err)
	}

	if v != "bar" {
		t.Errorf("incorrect value in store: want %v, got %v", "bar", v)
	}

	time.Sleep(1 * time.Second)
	v, err = store.Get("mo")
	if err == nil {
		t.Errorf("entry not expired: want error, got %v", v)
	}

	store.Delete("mo")
}
