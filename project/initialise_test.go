package project

import (
	"ni81/config"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Initialise(t *testing.T) {
	t.Run("Creates config when none exists", func(t *testing.T) {
		tmp := t.TempDir()

		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)

		if err := os.Chdir(tmp); err != nil {
			t.Fatal(err)
		}

		// Create required directory for locale input
		if err := os.Mkdir("i18n", 0755); err != nil {
			t.Fatal(err)
		}

		input := strings.Join([]string{
			"i18n",
			"en",
			"en",
			"ollama",
			"",
			"",
		}, "\n")

		oldStdin := os.Stdin
		r, w, _ := os.Pipe()
		defer func() { os.Stdin = oldStdin }()

		os.Stdin = r

		go func() {
			defer w.Close()
			_, _ = w.Write([]byte(input))
		}()

		if err := Initialise(); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Config file created
		if _, err := os.Stat(config.ConfigName); err != nil {
			t.Fatalf("Expected config file to be created: %v", err)
		}

		// Locale file created + normalised
		localeFile := filepath.Join("i18n", "en.json")
		if _, err := os.Stat(localeFile); err != nil {
			t.Fatalf("Expected locale file to be created: %v", err)
		}

		// Cache file created
		cacheFile := filepath.Join("i18n", "en.cache.json")
		if _, err := os.Stat(cacheFile); err != nil {
			t.Fatalf("Expected cache file to be created: %v", err)
		}
	})

	t.Run("Fails if config already exists", func(t *testing.T) {
		tmp := t.TempDir()

		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)

		if err := os.Chdir(tmp); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(config.ConfigName, []byte(""), 0664); err != nil {
			t.Fatal(err)
		}

		err := Initialise()
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})
}
