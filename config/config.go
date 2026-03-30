package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

// I18nConfig holds the internationalisation settings for the application.
type I18nConfig struct {
	Locales       []string `toml:"locales"`
	DefaultLocale string   `toml:"default_locale"`
	LocaleDir     string   `toml:"locale_dir"`
}

// ModelConfig defines the parameters for the specifc LLM being used.
type ModelConfig struct {
	Name string `toml:"name"`
	Url  string `toml:"url"`
}

// Config represents the master configuration structure for the application
type Config struct {
	I18n  I18nConfig  `toml:"i18n"`
	Model ModelConfig `toml:"model"`
}

// Load reads the TOML configuration from the specified path and decodes it into the Config struct.
func (c *Config) Load(path string) error {
	_, err := toml.DecodeFile(path, c)

	return err
}

// Save encodes the current Config struct into TOML format and writes it to the specified path.
func (c *Config) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return toml.NewEncoder(file).Encode(c)
}
