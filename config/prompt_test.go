package config

import (
	"bufio"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_readLine(t *testing.T) {
	t.Run("Valid read", func(t *testing.T) {
		expected := "i18n"
		reader := bufio.NewReader(strings.NewReader(" " + expected + "\n"))

		s, err := readLine(reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if s != expected {
			t.Fatalf("expected %s, got %s", expected, s)
		}
	})

	t.Run("Invalid read", func(t *testing.T) {
		reader := bufio.NewReader(strings.NewReader(""))

		if _, err := readLine(reader); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func Test_validatedLocale(t *testing.T) {
	t.Run("Invalid locale", func(t *testing.T) {
		got, err := validateLocale("en-australia")

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if got != "" {
			t.Fatalf("expected empty string, got %s", got)
		}
	})

	t.Run("Valid locale", func(t *testing.T) {
		expected := "en-AU"

		got, err := validateLocale("en-au")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != expected {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	})
}

func Test_getLocaleDir(t *testing.T) {
	tests := []struct {
		name string
		dir  string
	}{
		{name: "Dot path", dir: "."},
		{name: "Subdirectory", dir: "src/i18n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			base := filepath.Join(tmp, "app")

			if err := os.MkdirAll(filepath.Join(base, tt.dir), 0o755); err != nil {
				t.Fatal(err)
			}

			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)

			if err := os.Chdir(base); err != nil {
				t.Fatal(err)
			}

			reader := bufio.NewReader(strings.NewReader(tt.dir + " \n"))

			got, err := getLocaleDir(reader)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.dir {
				t.Fatalf("expected %s, got %s", tt.dir, got)
			}
		})
	}

	t.Run("Retry loop", func(t *testing.T) {
		tmp := t.TempDir()
		base := filepath.Join(tmp, "app")

		if err := os.MkdirAll(base, 0o755); err != nil {
			t.Fatal(err)
		}

		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)

		if err := os.Chdir(base); err != nil {
			t.Fatal(err)
		}

		reader := bufio.NewReader(strings.NewReader(base + "\nsrc\n../.\n.\n"))

		got, err := getLocaleDir(reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != "." {
			t.Fatalf("expected ., got %s", got)
		}
	})
}

func Test_findExistingLocales(t *testing.T) {
	t.Run("No directory", func(t *testing.T) {
		tmp := t.TempDir()

		locales, err := findExistingLocales(filepath.Join(tmp, "src"))
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if locales != nil {
			t.Fatalf("expected nil, got %v", locales)
		}
	})

	t.Run("Valid directory", func(t *testing.T) {
		tmp := t.TempDir()
		expected := []string{"en", "es"}

		for _, locale := range expected {
			if err := os.WriteFile(filepath.Join(tmp, locale+".json"), []byte(""), 0o664); err != nil {
				t.Fatal(err)
			}
		}

		if err := os.Mkdir(filepath.Join(tmp, "test.json"), 0o755); err != nil {
			t.Fatal(err)
		}

		got, err := findExistingLocales(tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		slices.Sort(expected)
		slices.Sort(got)

		if diff := cmp.Diff(expected, got); diff != "" {
			t.Error(diff)
		}
	})
}

func Test_getProjectLocales(t *testing.T) {
	tmp := t.TempDir()

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	reader := bufio.NewReader(strings.NewReader("\nenglish, spanish\nen, es,pt \n"))

	got, err := getProjectLocales(reader, tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"en", "es", "pt"}

	slices.Sort(expected)
	slices.Sort(got)

	if diff := cmp.Diff(expected, got); diff != "" {
		t.Fatal(diff)
	}

	for _, locale := range expected {
		if _, err := os.Stat(filepath.Join(tmp, locale+".json")); err != nil {
			t.Fatalf("expected file for %s: %v", locale, err)
		}
	}
}

func Test_getDefaultLocale(t *testing.T) {
	expected := "en"
	reader := bufio.NewReader(strings.NewReader("english\npt\n" + expected + "\n"))

	locales := []string{"en", "es"}

	got, err := getDefaultLocale(reader, locales)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func Test_getModelName(t *testing.T) {
	expected := "ollama"
	reader := bufio.NewReader(strings.NewReader("\nollama\n"))

	got, err := getModelName(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func Test_getModelUrl(t *testing.T) {
	t.Run("Empty URL", func(t *testing.T) {
		reader := bufio.NewReader(strings.NewReader("\n"))

		expected := "http://localhost:11434"

		got, err := getModelUrl(reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != expected {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	})

	t.Run("Passed URL", func(t *testing.T) {
		expected := "http://localhost:11433"

		reader := bufio.NewReader(strings.NewReader(
			"localhost\nhtp://localhost:11434\n" + expected + "\n",
		))

		got, err := getModelUrl(reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != expected {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	})
}

func Test_NewConfigFromReader(t *testing.T) {
	tmp := t.TempDir()

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	reader := bufio.NewReader(strings.NewReader("\nen\nen\nollama\n\n"))

	expected := config{
		I18n: i18nConfig{
			Locales:       []string{"en"},
			LocaleDir:     ".",
			DefaultLocale: "en",
		},
		Model: ModelConfig{
			Name: "ollama",
			Url:  "http://localhost:11434",
		},
	}

	got, err := NewConfigFromReader(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if diff := cmp.Diff(expected, got); diff != "" {
		t.Fatal(diff)
	}
}
