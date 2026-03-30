package fileutil

import (
	"errors"
	"fmt"
	"os"
)

// ReadFile reads the contents of a file located at path.
//
// If the file does not exist, it returns an empty slice of bytes and nil error.
// If the file exists but cannot be read due to some other reason, it returns
// nil for the byte slice and an error.
func ReadFile(path string) ([]byte, error) {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		// File does not exist - create it
		return []byte{}, nil
	} else if err != nil {
		return nil, err
	}

	return os.ReadFile(path)
}

// CreateFile creates a new file at path.
//
// It returns an error if the file creation fails or if there is another issue.
func CreateFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	return file.Close()
}

// Exists checks if the file at path exists.
//
// If path exists, it returns an error.
func Exists(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists", path)
	}

	return nil
}
