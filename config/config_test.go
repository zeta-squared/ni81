package config

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_Config(t *testing.T) {
	cfg := config{
		I18n: i18nConfig{
			Locales:       []string{"en-AU", "es-ES"},
			DefaultLocale: "en-AU",
			LocaleDir:     "src/lib/i18n/translation",
		},
		Model: ModelConfig{
			Name: "translategemma-4k:latest",
			Url:  "http://localhost:11434",
		},
	}

	t.Run("Valid config path", func(t *testing.T) {
		tmp := t.TempDir()

		dir := filepath.Join(tmp, cfg.I18n.LocaleDir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}

		path := filepath.Join(dir, cfg.I18n.DefaultLocale+".cache.json")

		if err := cfg.Save(path); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected file to exist: %v", err)
		}
	})

	t.Run("Invalid config path", func(t *testing.T) {
		tmp := t.TempDir()

		path := filepath.Join(tmp, cfg.I18n.LocaleDir, cfg.I18n.DefaultLocale+".cache.json")

		if err := cfg.Save(path); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
