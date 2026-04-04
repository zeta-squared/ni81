package project

import (
	"errors"
	"ni81/config"
	"ni81/fileutil"
	"os"
	"path/filepath"
)

// Initialise sets up the application by
//   - checking for an existing configuration file,
//   - prompting the user for new configuration settings,
//   - saving those settings to "ni81.toml",
//   - and creating an initial locale cache file.
func Initialise() error {
	_, err := fileutil.FindNearestConfigDir("ni81.toml")
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	cfg, err := config.NewConfigFromPrompt()
	if err != nil {
		return err
	}

	err = cfg.Save("ni81.toml")
	if err != nil {
		return err
	}

	return fileutil.CreateFile(filepath.Join(cfg.I18n.LocaleDir, cfg.I18n.DefaultLocale+".cache.json"))
}
