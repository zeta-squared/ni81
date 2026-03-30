package project

import (
	"fmt"
	"ni81/cache"
	"ni81/config"
	"ni81/fileutil"
	"ni81/serialization"
	"ni81/translate"
	"path/filepath"
)

// Project holds the configuration and dependencies required to manage
// translations, including
//   - locale settings,
//   - model configurations,
//   - and caching mechanisms.
type Project struct {
	DefaultLocale string
	TargetLocales []string
	LocaleDir     string
	Model         config.ModelConfig
	Cache         cache.Cache
	Translator    translate.Translator
}

// LoadLocale reads a JSON locale file from the project's locale directory
// and returns a flattened map of key-value translation pairs.
func (p Project) LoadLocale(locale string) (map[string]string, error) {
	localePath := filepath.Join(p.LocaleDir, locale+".json")

	data, err := fileutil.ReadFile(localePath)
	if err != nil {
		return nil, err
	}

	return serialization.ParseJSON(data)
}

// CreateCache populates the project's cache by:
//   - loading the current default locale file,
//   - and writing its contents to the cache store.
func (p Project) CreateCache() error {
	obj, err := p.LoadLocale(p.DefaultLocale)
	if err != nil {
		return err
	}

	return p.Cache.Write(obj)
}

// Translate synchronises target locale files with the default locale.
// It identifies new, updated, or deleted keys by
//   - comparing the current default locale against the cache,
//   - performs translations using the configured translator,
//   - and updates the cache upon completion.
func (p Project) Translate() error {
	newDefault, err := p.LoadLocale(p.DefaultLocale)
	if err != nil {
		return err
	}

	// Load oldDefault
	oldDefault, err := p.Cache.Read()
	if err != nil {
		return err
	}

	failed := false

localeLoop:
	for _, locale := range p.TargetLocales {
		// Load and flatten target JSON
		target, err := p.LoadLocale(locale)
		if err != nil {
			fmt.Println(err)
			continue localeLoop
		}

		// Find new entries
		for k := range newDefault {
			if _, ok := oldDefault[k]; !ok {
				// New key-value pair added
				target[k], err = p.Translator.Translate(newDefault[k], p.DefaultLocale, locale)
			} else if oldDefault[k] != newDefault[k] {
				// Previous key-value pair updated
				target[k], err = p.Translator.Translate(newDefault[k], p.DefaultLocale, locale)
			}
			if err != nil {
				failed = true
				fmt.Println(err)
				continue localeLoop
			}
		}

		// Remove deleted key-value pairs
		for k := range oldDefault {
			if _, ok := newDefault[k]; !ok {
				delete(target, k)
			}
		}

		// Write new translations
		err = writeTranslations(filepath.Join(p.LocaleDir, locale+".json"), target)
		if err != nil {
			fmt.Println(err)
			continue localeLoop
		}
	}

	if failed {
		return nil
	}

	// Rewrite cache
	return p.Cache.Write(newDefault)
}

// writeTranslations unflattens a map of translations and saves them
// as a structured JSON file at the specified path.
func writeTranslations(path string, translation map[string]string) error {
	out := serialization.Unflatten(translation)

	return serialization.WriteJSON(path, out)
}

// NewProject creates and returns a Project instance by:
//   - loading configuration from the given file path,
//   - and initialising the necessary cache and translator components.
func NewProject(path string) (Project, error) {
	cfg := config.Config{}
	err := cfg.Load(path)
	if err != nil {
		return Project{}, err
	}

	cachePath := filepath.Join(cfg.I18n.LocaleDir, cfg.I18n.DefaultLocale+".cache.json")
	targetLocales := make([]string, 0, len(cfg.I18n.Locales))
	for i := range cfg.I18n.Locales {
		if cfg.I18n.Locales[i] == cfg.I18n.DefaultLocale {
			continue
		}

		targetLocales = append(targetLocales, cfg.I18n.Locales[i])
	}

	return Project{
		DefaultLocale: cfg.I18n.DefaultLocale,
		TargetLocales: targetLocales,
		LocaleDir:     cfg.I18n.LocaleDir,
		Model:         cfg.Model,
		Cache:         cache.FileCache{Path: cachePath},
		Translator:    translate.Ollama{Model: cfg.Model.Name, Url: cfg.Model.Url},
	}, nil
}
