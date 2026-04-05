package fileutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Exists reports whether a file or directory exists at the given path.
// It returns true if os.Stat succeeds, and false otherwise.
func Exists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}

	return false
}

// NotExists reports whether a file or directory does not exist at the given path.
// It returns true only if the error from os.Stat is os.ErrNotExist.
func NotExists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return true
	}

	return false
}

// FindNearestConfigDir searches upward from the current working directory
// for a file with the given name. It returns the directory containing the file.
//
// If the file is found, the directory is returned along with a non-nil error
// indicating the project already exists. If the file is not found after reaching
// the filesystem root, os.ErrNotExist is returned.
func FindNearestConfigDir(name string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if Exists(filepath.Join(dir, name)) {
			return dir, fmt.Errorf("Already in project %s", dir)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// reached filesystem root
			return "", os.ErrNotExist
		}

		dir = parent
	}
}
