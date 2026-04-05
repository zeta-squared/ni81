package config

import (
	"bufio"
	"fmt"
	"net/url"
	"ni81/fileutil"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/text/language"
)

// NewConfigFromReader interactively constructs a config by prompting
// the user for required inputs via the provided reader.
//
// It guides the user through setting up locale directories, supported locales,
// default locale, and model configuration.
func NewConfigFromReader(reader *bufio.Reader) (config, error) {
	localeDir, err := getLocaleDir(reader)
	if err != nil {
		return config{}, err
	}

	locales, err := getProjectLocales(reader, localeDir)
	if err != nil {
		return config{}, err
	}

	defaultLocale, err := getDefaultLocale(reader, locales)
	if err != nil {
		return config{}, err
	}

	modelName, err := getModelName(reader)
	if err != nil {
		return config{}, err
	}

	modelUrl, err := getModelUrl(reader)
	if err != nil {
		return config{}, err
	}

	return config{
		I18n: i18nConfig{
			DefaultLocale: defaultLocale,
			Locales:       locales,
			LocaleDir:     localeDir,
		},
		Model: ModelConfig{
			Name: modelName,
			Url:  modelUrl,
		},
	}, nil
}

// getLocaleDir prompts the user to provide a relative path to the
// project's translation directory.
//
// It validates that the path:
//   - is relative,
//   - does not escape the project directory,
//   - and exists on disk.
//
// The returned path is normalised to use forward slashes.
func getLocaleDir(reader *bufio.Reader) (string, error) {
	for {
		fmt.Print("What (relative) directory are the project i18n translation files stored in? ")
		localeDir, err := readLine(reader)
		if err != nil {
			return "", err
		}

		localeDir = filepath.Clean(localeDir)
		if filepath.IsAbs(localeDir) {
			fmt.Println("Please provide a relative path, not an absolute path.")
			continue
		}

		if strings.HasPrefix(localeDir, "..") {
			fmt.Println("Path must be within the project directory.")
			continue
		}

		if fileutil.NotExists(localeDir) {
			fmt.Printf("Directory %s does not exist. Please try again.\n", localeDir)
			continue
		}

		return filepath.ToSlash(localeDir), nil
	}
}

// getProjectLocales determines the set of locales used in the project.
//
// It detects existing locale files in the given directory and optionally
// allows the user to include them. The user may also input additional locales,
// which are validated and created as empty JSON files if necessary.
//
// The returned slice contains unique, validated locale identifiers.
func getProjectLocales(reader *bufio.Reader, localeDir string) ([]string, error) {
	var ans string
	existingLocales, err := findExistingLocales(localeDir)
	if err != nil {
		return nil, err
	}

	if len(existingLocales) > 0 {
		fmt.Printf(
			"The following locales have been found\n\n\t- %s\n\nDo you want to add them to the project? [Y/n] ",
			strings.Join(existingLocales, "\n\t- "),
		)
		ans, err = readLine(reader)
		if err != nil {
			return nil, err
		}
	}

localesLoop:
	for {
		if len(existingLocales) > 0 && strings.ToLower(strings.TrimSpace(ans)) != "n" {
			fmt.Print("Enter a comma separated list of other locales to include in this project, e.g. en-US, en-GB (optional): ")
		} else {
			fmt.Print("Enter a comma separated list of locales in this project, e.g. en-US, en-GB: ")
		}

		locales, err := readLine(reader)
		if err != nil {
			return nil, err
		}

		// If existing locales found and no additional entered by user
		// fall back to only the existing locales
		if len(existingLocales) > 0 && locales == "" {
			return existingLocales, nil
		} else if locales == "" {
			fmt.Println("Locales required. Please try again.")
			continue localesLoop
		}

		dedupedLocales := map[string]bool{}
		for i := range existingLocales {
			dedupedLocales[existingLocales[i]] = true
		}

		splitLocales := strings.Split(locales, ",")
		invalids := make([]string, 0, len(splitLocales))

		for i := range splitLocales {
			locale := strings.TrimSpace(splitLocales[i])
			if locale == "" {
				continue
			}

			validatedLocale, err := validateLocale(locale)
			if err != nil {
				invalids = append(invalids, locale)
				continue
			}

			if _, ok := dedupedLocales[validatedLocale]; !ok {
				os.WriteFile(filepath.Join(localeDir, validatedLocale+".json"), []byte("{}"), 0664)
				dedupedLocales[validatedLocale] = true
			}
		}

		if len(invalids) == 1 {
			fmt.Printf("Invalid locale: %s\n", strings.Join(invalids, ", "))
			continue
		} else if len(invalids) > 0 {
			fmt.Printf("Invalid locales: %s\n", strings.Join(invalids, ", "))
			continue
		}

		tomlLocales := make([]string, 0, len(dedupedLocales))
		for k := range dedupedLocales {
			tomlLocales = append(tomlLocales, k)
		}

		return tomlLocales, nil
	}
}

// findExistingLocales scans the given directory for valid locale JSON files.
//
// It extracts locale identifiers from filenames of the form "<locale>.json",
// validates them, and returns the list of recognised locales.
func findExistingLocales(localeDir string) ([]string, error) {
	files, err := os.ReadDir(localeDir)
	if err != nil {
		return nil, err
	}

	existingLocales := make([]string, 0, len(files))
	for i := range files {
		if strings.HasSuffix(files[i].Name(), ".json") && len(strings.Split(files[i].Name(), ".")) == 2 {
			locale := strings.Split(files[i].Name(), ".")[0]
			validatedLocale, err := validateLocale(locale)
			if err != nil {
				continue
			}

			existingLocales = append(existingLocales, validatedLocale)
		}
	}

	return existingLocales, nil
}

// getDefaultLocale prompts the user to select the default locale for the project.
//
// The selected locale must be valid and present in the provided options.
func getDefaultLocale(reader *bufio.Reader, options []string) (string, error) {
	for {
		fmt.Print("What is the default locale of this project? ")
		defaultLocale, err := readLine(reader)
		if err != nil {
			return "", err
		}

		validatedLocale, err := validateLocale(defaultLocale)
		if err != nil {
			fmt.Printf("Invalid locale: %s\n", defaultLocale)
			continue
		}

		if !slices.Contains(options, validatedLocale) {
			fmt.Printf("Default locale must be one of %s. Please try again.\n", strings.Join(options, ", "))
			continue
		}

		return validatedLocale, nil
	}
}

// getModelName prompts the user to provide the name of the translation model.
//
// The input must be non-empty.
func getModelName(reader *bufio.Reader) (string, error) {
	for {
		fmt.Print("Provide a (ollama) model name to use for translations: ")
		modelName, err := readLine(reader)
		if err != nil {
			return "", err
		}

		if modelName == "" {
			fmt.Println("Model name cannot be empty. Please try again.")
			continue
		}

		return modelName, nil
	}
}

// getModelUrl prompts the user to provide the model server URL.
//
// If left blank, a default of "http://localhost:11434" is used.
// The input is validated to ensure it is a properly formatted HTTP or HTTPS URL.
func getModelUrl(reader *bufio.Reader) (string, error) {
	for {
		modelUrl := "http://localhost:11434"
		fmt.Print("Provide the model server URL. If left blank nibl will use http://localhost:11434: ")

		input, err := readLine(reader)
		if err != nil {
			return "", err
		}

		if input != "" {
			parsedUrl, err := url.ParseRequestURI(input)
			if err != nil {
				fmt.Println("Invalid URL format. Please try again.")
				continue
			}

			if parsedUrl.Scheme != "http" && parsedUrl.Scheme != "https" {
				fmt.Println("URL must start with http:// or https://")
				continue
			}

			modelUrl = parsedUrl.String()
		}

		return modelUrl, nil
	}
}

// readLine reads a single line from the provided reader and trims
// leading and trailing whitespace.
//
// It returns an error if reading from the reader fails.
func readLine(reader *bufio.Reader) (string, error) {
	res, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(res), nil
}

// validateLocale parses and normalises a locale string using BCP 47 rules.
//
// It returns the canonical locale representation or an error if the input is invalid.
func validateLocale(input string) (string, error) {
	tag, err := language.Parse(input)
	if err != nil {
		return "", err
	}

	return tag.String(), nil
}
