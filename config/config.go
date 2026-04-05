package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

const ConfigName = "ni81.toml"

// i18nConfig defines internationalisation settings for a project.
//
// It includes the set of supported locales, the default locale,
// and the directory where translation files are stored.
type i18nConfig struct {
	Locales       []string `toml:"locales"`
	DefaultLocale string   `toml:"default_locale"`
	LocaleDir     string   `toml:"locale_dir"`
}

// ModelConfig defines the configuration for the translation model.
//
// It specifies the model name and the base URL of the model server.
type ModelConfig struct {
	Name string `toml:"name"`
	Url  string `toml:"url"`
}

// config represents the full application configuration.
//
// It combines internationalisation settings with model configuration
// and is typically loaded from or saved to a TOML file.
type config struct {
	I18n  i18nConfig  `toml:"i18n"`
	Model ModelConfig `toml:"model"`
}

// Save writes the configuration to the specified file path in TOML format.
//
// It creates or truncates the file and returns an error if writing fails.
func (c config) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return toml.NewEncoder(file).Encode(c)
}

// NewConfigFromFile loads and parses a configuration from a TOML file.
//
// It returns the populated config and any error encountered during decoding.
func NewConfigFromFile(path string) (config, error) {
	cfg := config{}
	_, err := toml.DecodeFile(path, &cfg)

	return cfg, err
}
