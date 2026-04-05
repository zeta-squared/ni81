package cache

import (
	"ni81/serialization"
	"os"
)

// Cache defines the behaviour for reading and writing
// flattened translation key-value pairs to a backing store.
//
// Implementations are responsible for persisting and retrieving
// the most recent state of translations, typically for change detection.
type Cache interface {
	Read() (map[string]string, error)
	Write(obj map[string]string) error
}

// fileCache is a file-based implementation of the Cache interface.
//
// It persists translation data to a JSON file at the specified path,
// delegating serialization concerns to a ReadWriter.
type fileCache struct {
	path       string
	readWriter serialization.ReadWriter
}

// Read loads and returns cached translation data from the file.
//
// It returns an error if the file cannot be opened or if the underlying
// ReadWriter fails to deserialize the contents.
func (fc fileCache) Read() (map[string]string, error) {
	file, err := os.Open(fc.path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	return fc.readWriter.Read(file)
}

// Write persists the provided translation map to the cache file.
//
// The file is created if it does not exist and truncated if it does.
// Serialization is delegated to the configured ReadWriter.
func (fc fileCache) Write(obj map[string]string) error {
	file, err := os.OpenFile(fc.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}

	defer file.Close()

	return fc.readWriter.Write(file, obj, false)
}

// NewFileCache returns a new fileCache instance for the given path.
//
// It uses JSONReadWriter for serialization by default.
func NewFileCache(path string) fileCache {
	return fileCache{
		path:       path,
		readWriter: serialization.JSONReadWriter{},
	}
}
