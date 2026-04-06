package project

import (
	"fmt"
	"net/http"
	"net/url"
	"ni81/cache"
	"ni81/config"
	"ni81/fileutil"
	"ni81/serialization"
	"ni81/translate"
	"os"
	"path/filepath"
)

// project represents a translation project with all required configuration
// and dependencies, including locale settings, model configuration,
// cache storage, and translation mechanisms.
type project struct {
	defaultLocale  string
	targetLocales  []string
	localeDir      string
	model          config.ModelConfig
	cache          cache.Cache
	jsonReadWriter serialization.ReadWriter
	translator     translate.Translator
}

// CreateCache reads the default locale file and writes its contents
// to the project's cache. This establishes the baseline state used
// for detecting future translation changes.
func (p project) CreateCache() error {
	obj, err := p.readTranslations(p.defaultLocale)
	if err != nil {
		return err
	}

	return p.cache.Write(obj)
}

// Translate synchronises all target locale files with the default locale.
//
// It compares the current default locale against the cached version to detect
// added, updated, or removed keys. New and updated keys are translated using
// the configured translator, while removed keys are deleted from targets.
//
// If any translation or write operation fails, processing continues for
// remaining locales, but the cache is not updated.
func (p project) Translate() error {
	newDefault, err := p.readTranslations(p.defaultLocale)
	if err != nil {
		return err
	}

	// Load oldDefault
	oldDefault, err := p.cache.Read()
	if err != nil {
		return err
	}

	failed := false

localeLoop:
	for _, locale := range p.targetLocales {
		// Load and flatten target JSON
		target, err := p.readTranslations(locale)
		if err != nil {
			fmt.Println(err)
			continue localeLoop
		}

		// Find new entries
		for k := range newDefault {
			if _, ok := oldDefault[k]; !ok {
				// New key-value pair added
				target[k], err = p.translator.Translate(newDefault[k], p.defaultLocale, locale)
			} else if oldDefault[k] != newDefault[k] {
				// Previous key-value pair updated
				target[k], err = p.translator.Translate(newDefault[k], p.defaultLocale, locale)
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
		if err = p.writeTranslations(locale, target); err != nil {
			failed = true
			fmt.Println(err)
			continue localeLoop
		}
	}

	if failed {
		return nil
	}

	// Rewrite cache
	return p.cache.Write(newDefault)
}

// readTranslations reads a locale JSON file from the project's locale directory
// and returns its contents as a flattened map of key-value pairs.
func (p project) readTranslations(locale string) (map[string]string, error) {
	file, err := os.Open(filepath.Join(p.localeDir, locale+".json"))
	if err != nil {
		return nil, err
	}

	defer file.Close()

	return p.jsonReadWriter.Read(file)
}

// writeTranslations writes the provided translations to the specified locale file.
// The input map is unflattened into a nested JSON structure before being written.
func (p project) writeTranslations(locale string, translation map[string]string) error {
	file, err := os.OpenFile(filepath.Join(p.localeDir, locale+".json"), os.O_WRONLY|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}

	defer file.Close()

	return p.jsonReadWriter.Write(file, translation, true)
}

// NewProject constructs a project from the configuration file with the given name.
//
// It locates the nearest configuration file, loads its contents, and initialises
// the cache and translator. The default locale is excluded from the target locales.
func NewProject(name string) (project, error) {
	dir, err := fileutil.FindNearestConfigDir(name)
	if dir == "" {
		return project{}, err
	}

	cfg, err := config.NewConfigFromFile(filepath.Join(dir, name))
	if err != nil {
		return project{}, err
	}

	targetLocales := make([]string, 0, len(cfg.I18n.Locales))
	for i := range cfg.I18n.Locales {
		if cfg.I18n.Locales[i] == cfg.I18n.DefaultLocale {
			continue
		}

		targetLocales = append(targetLocales, cfg.I18n.Locales[i])
	}

	localeDir := filepath.Join(dir, cfg.I18n.LocaleDir)
	cachePath := filepath.Join(localeDir, cfg.I18n.DefaultLocale+".cache.json")
	cache := cache.NewFileCache(cachePath)

	modelClient := http.DefaultClient
	modelBase, err := url.ParseRequestURI(cfg.Model.Url)
	if err != nil {
		return project{}, err
	}

	translator, err := translate.NewOllamaFromConfig(cfg.Model.Name, modelBase, modelClient)
	if err != nil {
		return project{}, err
	}

	return project{
		defaultLocale:  cfg.I18n.DefaultLocale,
		targetLocales:  targetLocales,
		localeDir:      localeDir,
		model:          cfg.Model,
		cache:          cache,
		jsonReadWriter: serialization.JSONReadWriter{},
		translator:     translator,
	}, nil
}
