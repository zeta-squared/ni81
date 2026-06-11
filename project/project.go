package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"ni81/cache"
	"ni81/config"
	"ni81/fileutil"
	"ni81/serialization"
	"ni81/translate"
	"os"
	"path/filepath"
	"slices"
)

const (
	red   = "\033[31m"
	green = "\033[32m"
	reset = "\033[0m"
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

type change struct {
	key  string
	text string
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
func (p project) Translate(clean bool) error {
	newDefault, err := p.readTranslations(p.defaultLocale)
	if err != nil {
		return err
	}

	oldDefault := map[string]string{}
	if !clean {
		// Load oldDefault
		oldDefault, err = p.cache.Read()
		if err != nil {
			return err
		}
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
		for k, newVal := range newDefault {
			oldVal, ok := oldDefault[k]

			if !ok {
				// New key-value pair added
				target[k], err = p.translator.Translate(newVal, p.defaultLocale, locale)
			} else if oldVal != newVal {
				// Previous key-value pair updated
				target[k], err = p.translator.Translate(newVal, p.defaultLocale, locale)
			}

			if err != nil {
				if errors.Is(err, translate.ConnError{}) {
					return err
				}

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

	// Order default locale JSON
	if err := p.writeTranslations(p.defaultLocale, newDefault); err != nil {
		return err
	}

	if failed {
		return nil
	}

	// Rewrite cache
	return p.cache.Write(newDefault)
}

// Diff compares the current default locale translations with the cached
// previous state and writes a human-readable, ANSI-colored diff to w.
//
// The output shows added, modified, and deleted translation keys:
//   - Added keys are shown with a green "+" line.
//   - Deleted keys are shown with a red "-" line.
//   - Modified keys are shown as a red "-" old value line followed by a green "+" new value line.
//
// Output is sorted by key to ensure deterministic ordering. Each change
// is separated by a blank line, except after the final entry.
//
// Diff returns an error if either the current or cached translations
// cannot be read.
func (p project) Diff(w io.Writer) error {
	newDefault, err := p.readTranslations(p.defaultLocale)
	if err != nil {
		return err
	}

	oldDefault, err := p.cache.Read()
	if err != nil {
		return err
	}

	changes := make([]change, 0)

	// Find new and modified entries
	for k, newVal := range newDefault {
		oldVal, ok := oldDefault[k]

		if !ok {
			// New key-value pair added
			changes = append(changes, change{key: k, text: fmt.Sprintf("%s+\"%s\": \"%s\"%s\n", green, k, newVal, reset)})
			continue
		} else if oldVal != newVal {
			// Previous key-value pair modified
			changes = append(changes, change{key: k, text: fmt.Sprintf("%[4]s-\"%[1]s\": \"%[2]s\"%[6]s\n%[5]s+\"%[1]s\": \"%[3]s\"%[6]s\n", k, oldVal, newVal, red, green, reset)})
		}

		delete(oldDefault, k)
	}

	// Remove deleted key-value pairs
	for k, oldVal := range oldDefault {
		changes = append(changes, change{key: k, text: fmt.Sprintf("%s-\"%s\": \"%s\"%s\n", red, k, oldVal, reset)})
	}

	slices.SortFunc(changes, func(a, b change) int {
		if a.key < b.key {
			return -1
		} else if a.key > b.key {
			return 1
		}

		return 0
	})

	for i := range changes {
		if i != len(changes)-1 {
			fmt.Fprintln(w, changes[i].text)
		} else {
			fmt.Fprint(w, changes[i].text)
		}
	}

	return nil
}

// readTranslations reads a locale JSON file from the project's locale directory
// and returns its contents as a flattened map of key-value pairs.
func (p project) readTranslations(locale string) (map[string]string, error) {
	file, err := os.Open(filepath.Join(p.localeDir, locale+".json"))
	if err != nil {
		return nil, err
	}

	defer file.Close()

	obj, err := p.jsonReadWriter.Read(file)
	var syntaxError *json.SyntaxError
	if errors.As(err, &syntaxError) {
		return obj, fmt.Errorf("%w | %s", err, file.Name())
	}

	return obj, err
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
	for _, locale := range cfg.I18n.Locales {
		if locale == cfg.I18n.DefaultLocale {
			continue
		}

		targetLocales = append(targetLocales, locale)
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
