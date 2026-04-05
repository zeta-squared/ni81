package project

import (
	"bufio"
	"errors"
	"ni81/config"
	"ni81/fileutil"
	"os"
)

// Initialise sets up a new project in the current directory.
//
// It ensures that no existing configuration file is present in the current
// or parent directories, then interactively collects configuration input
// from the user and writes it to disk.
//
// After creating the configuration, it initialises the project by:
//   - normalising all locale JSON files (read + rewrite),
//   - and creating the initial cache from the default locale.
//
// An error is returned if a configuration already exists, user input fails,
// or any filesystem or initialisation step encounters an error.
func Initialise() error {
	_, err := fileutil.FindNearestConfigDir(config.ConfigName)
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	reader := bufio.NewReader(os.Stdin)
	cfg, err := config.NewConfigFromReader(reader)
	if err != nil {
		return err
	}

	err = cfg.Save(config.ConfigName)
	if err != nil {
		return err
	}

	proj, err := NewProject(config.ConfigName)
	if err != nil {
		return err
	}

	// Read and write locale files to order the JSON and avoid confusion
	// with changes in future translations
	for _, locale := range cfg.I18n.Locales {
		obj, err := proj.readTranslations(locale)
		if err != nil {
			return err
		}

		err = proj.writeTranslations(locale, obj)
		if err != nil {
			return err
		}
	}

	// Create cache on initialisation
	return proj.CreateCache()
}
