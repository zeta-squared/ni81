package fileutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_Exists(t *testing.T) {
	tmp := t.TempDir()

	realFile := filepath.Join(tmp, "test.txt")
	fakeFile := filepath.Join(tmp, "test.file")

	if err := os.WriteFile(realFile, []byte(""), 0o664); err != nil {
		t.Fatal(err)
	}

	if !Exists(realFile) {
		t.Fatal("expected file to exist")
	}

	if Exists(fakeFile) {
		t.Fatal("expected file to not exist")
	}
}

func Test_NotExists(t *testing.T) {
	tmp := t.TempDir()

	realFile := filepath.Join(tmp, "test.txt")
	fakeFile := filepath.Join(tmp, "test.file")

	if err := os.WriteFile(realFile, []byte(""), 0o664); err != nil {
		t.Fatal(err)
	}

	if NotExists(realFile) {
		t.Fatal("expected file to exist")
	}

	if !NotExists(fakeFile) {
		t.Fatal("expected file to not exist")
	}
}

func Test_FindNearestConfigDir(t *testing.T) {
	configName := "config.toml"

	t.Run("Config in current dir", func(t *testing.T) {
		tmp := t.TempDir()

		if err := os.WriteFile(filepath.Join(tmp, configName), []byte(""), 0o664); err != nil {
			t.Fatal(err)
		}

		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)

		if err := os.Chdir(tmp); err != nil {
			t.Fatal(err)
		}

		dir, err := FindNearestConfigDir(configName)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if dir != tmp {
			t.Fatalf("expected %s, got %s", tmp, dir)
		}
	})

	t.Run("Config in parent dir", func(t *testing.T) {
		root := t.TempDir()
		child := filepath.Join(root, "child")

		if err := os.Mkdir(child, 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(root, configName), []byte(""), 0o664); err != nil {
			t.Fatal(err)
		}

		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)

		if err := os.Chdir(child); err != nil {
			t.Fatal(err)
		}

		dir, err := FindNearestConfigDir(configName)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if dir != root {
			t.Fatalf("expected %s, got %s", root, dir)
		}
	})

	t.Run("Config not found", func(t *testing.T) {
		tmp := t.TempDir()

		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)

		if err := os.Chdir(tmp); err != nil {
			t.Fatal(err)
		}

		dir, err := FindNearestConfigDir(configName)

		if dir != "" {
			t.Fatalf("expected empty dir, got %s", dir)
		}

		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("expected os.ErrNotExist, got %v", err)
		}
	})
}
