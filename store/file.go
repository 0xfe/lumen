package store

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type fileEntry struct {
	Value     string    `json:"value"`
	NoExpire  bool      `json:"bool"`
	ExpiresOn time.Time `json:"expires_on"`
}

func (e fileEntry) expired() bool {
	return !e.NoExpire && time.Now().After(e.ExpiresOn)
}

type fileData struct {
	Version string               `json:"version"`
	Seq     uint64               `json:"seq"`
	Pairs   map[string]fileEntry `json:"pairs"`
}

func newFileData() *fileData {
	return &fileData{
		Version: "1",
		Seq:     0,
		Pairs:   make(map[string]fileEntry),
	}
}

// newFileDataFromFile tries to load data from fileName, creating a
// new file with empty data if it can't read it. Returns error if it
// can't parse an existing file, or reads invalid data.
func newFileDataFromFile(fileName string) (*fileData, error) {
	fileData := newFileData()

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Printf("Creating new file: %s", fileName)
		return fileData, fileData.sync(fileName)
	}

	err = json.Unmarshal(data, &fileData)

	if err != nil {
		return nil, errors.Errorf("invalid content in %s: %v", fileName, err)
	}

	return fileData, nil
}

func (data *fileData) sync(fileName string) error {
	jsonData, err := json.Marshal(*data)

	if err != nil {
		return errors.Errorf("could not marshall json: %v", err)
	}

	err = ioutil.WriteFile(fileName, jsonData, 0644)
	if err != nil {
		return errors.Errorf("could not write to file: %v", err)
	}

	return nil
}

// DataStore represents the conntection to the Google Cloud Datastore.
type FileStore struct {
	*Store
	path string
	mu   *sync.RWMutex // protects data
	data *fileData
}

func NewFileStore(path string) (*FileStore, error) {
	fileData, err := newFileDataFromFile(path)

	if err != nil {
		return nil, errors.Wrap(err, "can't create file store")
	}

	// Try to connect
	fileStore := &FileStore{
		Store: &Store{
			driver:     "file",
			parameters: path,
		},
		path: path,
		mu:   &sync.RWMutex{},
		data: fileData,
	}
	return fileStore, nil
}

// sync must be called under mu
func (fs *FileStore) sync() error {
	return fs.data.sync(fs.path)
}

func (fs *FileStore) Set(k string, v string, ttl time.Duration) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.data.Pairs[k] = fileEntry{
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
