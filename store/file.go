package store

import (
	"encoding/json"
	"io/ioutil"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type FileData struct {
	Version string            `json:"version"`
	Seq     uint64            `json:"seq"`
	Pairs   map[string]string `json:"pairs"`
}

func NewFileData() *FileData {
	return &FileData{
		Version: "1",
		Seq:     0,
		Pairs:   make(map[string]string),
	}
}

// DataStore represents the conntection to the Google Cloud Datastore.
type FileStore struct {
	store *Store
	path  string
	mu    *sync.RWMutex // protects data
	data  *FileData
}

func NewFileStore(path string) (*FileStore, error) {
	// Try to connect
	return &FileStore{
		store: &Store{
			driver:     "file",
			parameters: path,
		},
		path: path,
		data: NewFileData(),
	}, nil
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

	fs.data.Pairs[k] = v
	fs.data.Seq++

	return fs.sync()
}

func (fs *FileStore) Get(k string) (string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	val, ok := fs.data.Pairs[k]
	if !ok {
		return "", errors.Errorf("not found: %s", k)
	}
	return val, nil
}

func (fs *FileStore) Delete(k string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	delete(fs.data.Pairs, k)
	return fs.sync()
}
