package store

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func getTempFile() (string, string) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		log.Fatal(err)
	}

	return dir, (dir + string(os.PathSeparator) + "lumen_file_test")
}

func TestFileStore_BasicLookup(t *testing.T) {
	tmpDir, tmpFile := getTempFile()
	defer os.RemoveAll(tmpDir)

	store, err := NewStore("file", tmpFile)

	if err != nil {
		t.Errorf("couldn't setup internal store, want %v, got %v", nil, err)
	}

	testBasicLookup(t, store)
}

func TestFileStore_TTL(t *testing.T) {
	tmpDir, tmpFile := getTempFile()
	defer os.RemoveAll(tmpDir)

	store, err := NewStore("file", tmpFile)

	if err != nil {
		t.Errorf("couldn't setup internal store, want %v, got %v", nil, err)
	}

	testTTL(t, store)
}
