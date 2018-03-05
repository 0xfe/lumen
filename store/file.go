package store

import (
	"encoding/json"
	"io/ioutil"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type FileEntry struct {
	Value     string    `json:"value"`
	NoExpire  bool      `json:"bool"`
	ExpiresOn time.Time `json:"expires_on"`
}

func (e FileEntry) expired() bool {
	return !e.NoExpire && time.Now().After(e.ExpiresOn)
}

type FileData struct {
	Version string               `json:"version"`
	Seq     uint64               `json:"seq"`
	Pairs   map[string]FileEntry `json:"pairs"`
}

func NewFileData() *FileData {
	return &FileData{
		Version: "1",
		Seq:     0,
		Pairs:   make(map[string]FileEntry),
	}
}

// DataStore represents the conntection to the Google Cloud Datastore.
type FileStore struct {
	*Store
	path string
	mu   *sync.RWMutex // protects data
	data *FileData
}

func NewFileStore(path string) (*FileStore, error) {
	// Try to connect
	fileStore := &FileStore{
		Store: &Store{
			driver:     "file",
			parameters: path,
		},
		path: path,
		data: NewFileData(),
	}

	return fileStore, nil
}

// sync must be called under mu
func (fs *FileStore) sync() error {
	jsonData, err := json.Marshal(*fs.data)
	if err != nil {
		return errors.Errorf("could not marshall json: %v", err)
	}

	err = ioutil.WriteFile(fs.path, jsonData, 0644)
	if err != nil {
		return errors.Errorf("could not write to file: %v", err)
	}

	return err
}

func (fs *FileStore) Set(k string, v string, ttl time.Duration) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.data.Pairs[k] = FileEntry{
		Value:     v,
		NoExpire:  ttl == 0,
		ExpiresOn: time.Now().Add(ttl),
	}

	fs.data.Seq++
	return fs.sync()
}

func (fs *FileStore) Get(k string) (string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	val, ok := fs.data.Pairs[k]
	if !ok || val.expired() {
		return "", errors.Errorf("not found: %s", k)
	}
	return val.Value, nil
}

func (fs *FileStore) Delete(k string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	delete(fs.data.Pairs, k)
	return fs.sync()
}
