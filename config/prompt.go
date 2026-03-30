package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// NewConfigFromPrompt initiates an interactive CLI session to gather user input
// and construct a new Config instance.
func NewConfigFromPrompt() (Config, error) {
	reader := bufio.NewReader(os.Stdin)

	localeDir, err := getLocaleDir(reader)
	if err != nil {
		return Config{}, err
	}

	locales, err := getProjectLocales(reader, localeDir)
	if err != nil {
		return Config{}, err
	}

	defaultLocale, err := getDefaultLocale(reader)
	if err != nil {
		return Config{}, err
	}

	modelName, err := getModelName(reader)
	if err != nil {
		return Config{}, err
	}

	return Config{
		I18n: I18nConfig{
			DefaultLocale: defaultLocale,
			Locales:       locales,
			LocaleDir:     localeDir,
		},
		Model: ModelConfig{
			Name: modelName,
			Url:  "http://localhost:11434", // TODO: read from CLI when other models are supported
		},
	}, nil
}

// getLocaleDir prompts the user for the directory containing translation files
// and returns the absolute path.
func getLocaleDir(reader *bufio.Reader) (string, error) {
	// TODO: add validation to user input
	fmt.Print("What (relative) directory are the project i18n translation files stored in? ")
	localeDir, err := readLine(reader)
	if err != nil {
		return "", err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if localeDir == "." {
		localeDir = ""
	}

	return filepath.Join(cwd, localeDir), nil
}

// getProjectLocales identifies existing translation files in the localeDir and
// allows the user to manually add or confirm supported language codes.
func getProjectLocales(reader *bufio.Reader, localeDir string) ([]string, error) {
	// TODO: add validation to user input
	var ans string
	existingLocales, err := findExistingLocales(localeDir)
	if err != nil {
		return nil, err
	}

	if len(existingLocales) > 0 {
		foundLocales := ""
		for i := range existingLocales {
			foundLocales += "\n" + existingLocales[i]
		}

		fmt.Printf("The following locales have been found\n%s\n\nDo you want to add them to the project? [y/N] ", foundLocales)
		ans, err = readLine(reader)
		if err != nil {
			return nil, err
		}
	}

	if strings.ToLower(ans) == "y" {
		fmt.Print("Enter a comma separated list of other locales to include in this project, e.g. en-US, en-GB (optional): ")
	} else {
		fmt.Print("Enter a comma separated list of locales in this project, e.g. en-US, en-GB: ")
	}

	locales, err := readLine(reader)
	if err != nil {
		return nil, err
	}

	if locales != "" {
		dedupedLocales := map[string]bool{}
		for i := range existingLocales {
			dedupedLocales[existingLocales[i]] = true
		}

		splitLocales := strings.Split(locales, ",")
		for i := range splitLocales {
			dedupedLocales[strings.TrimSpace(splitLocales[i])] = true
		}

		tomlLocales := make([]string, 0, len(dedupedLocales))
		for k := range dedupedLocales {
			tomlLocales = append(tomlLocales, k)
		}

		return tomlLocales, nil
	}

	return existingLocales, nil
}

// findExistingLocales scans the specified directory for .json files and
// extracts the locale names from the filenames.
func findExistingLocales(localeDir string) ([]string, error) {
	files, err := os.ReadDir(localeDir)
	if err != nil {
		return nil, err
	}

	existingLocales := make([]string, 0, len(files))
	for i := range files {
		if strings.HasSuffix(files[i].Name(), ".json") && len(strings.Split(files[i].Name(), ".")) == 2 {
			existingLocales = append(existingLocales, strings.Split(files[i].Name(), ".")[0])
		}
	}

	return existingLocales, nil
}

// getDefaultLocale prompts the user to specify the primary fallback language for the project.
func getDefaultLocale(reader *bufio.Reader) (string, error) {
	// TODO: add validation to user input

	fmt.Print("What is the default locale of this project? ")
	defaultLocale, err := readLine(reader)
	if err != nil {
		return "", err
	}

	return defaultLocale, nil
}

// getModelName prompts the user for the name of the Ollama model to be used for processing.
func getModelName(reader *bufio.Reader) (string, error) {
	// TODO: add validation to user input

	fmt.Print("Provide a (ollama) model name to use for translations: ")
	modelName, err := readLine(reader)
	if err != nil {
		return "", err
	}

	return modelName, nil
}

// readLine captures a single line of input from the provided reader and trims leading/trailing whitespace.
func readLine(reader *bufio.Reader) (string, error) {
	res, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(res), nil
}
