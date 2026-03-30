package cache

import (
	"ni81/fileutil"
	"ni81/serialization"
)

// FileCache represents a filesystem-based storage handler for key-value data.
type FileCache struct {
	Path string
}

// Read loads and parses JSON data from the underlying file path.
//
// It returns a map of strings or an error if the file is unreadable or malformed.
func (fc FileCache) Read() (map[string]string, error) {
	data, err := fileutil.ReadFile(fc.Path)
	if err != nil {
		return nil, err
	}

	return serialization.ParseJSON(data)
}

// Write serializes the provided map into JSON format and saves it to the file path.
//
// It returns an error if the write operation fails.
func (fc FileCache) Write(obj map[string]string) error {
	return serialization.WriteJSON(fc.Path, obj)
}
