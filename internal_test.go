package cli

import "testing"

func TestInternalStore_BasicLookup(t *testing.T) {
	store, err := NewStore("internal", "")

	if err != nil {
		t.Errorf("couldn't setup internal store, want %v, got %v", nil, err)
	}

	testBasicLookup(t, store)
}

func TestInternalStore_TTL(t *testing.T) {
	store, err := NewStore("internal", "")

	if err != nil {
		t.Errorf("couldn't setup internal store, want %v, got %v", nil, err)
	}

	testTTL(t, store)
}
