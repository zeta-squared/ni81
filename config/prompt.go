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

	defaultLocale, err := getDefaultLocale(reader, locales)
	if err != nil {
		return Config{}, err
	}

	modelName, err := getModelName(reader)
	if err != nil {
		return Config{}, err
	}

	modelUrl, err := getModelUrl(reader)
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
			Url:  modelUrl,
		},
	}, nil
}

// getLocaleDir prompts the user for the directory containing translation files
// and returns the absolute path.
func getLocaleDir(reader *bufio.Reader) (string, error) {
	for {
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

		localeDir = filepath.Join(cwd, localeDir)
		if fileutil.NotExists(localeDir) != nil {
			fmt.Printf("Directory %s does not exist. Please try again.\n", localeDir)
			continue
		}

		return localeDir, nil
	}
}

// getProjectLocales identifies existing translation files in the localeDir and
// allows the user to manually add or confirm supported language codes.
func getProjectLocales(reader *bufio.Reader, localeDir string) ([]string, error) {
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

localesLoop:
	for {
		if strings.ToLower(strings.TrimSpace(ans)) == "y" {
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

			dedupedLocales[validatedLocale] = true
		}

		if len(invalids) == 1 {
			fmt.Printf("Invalid locale: %s\n", strings.Join(invalids, ", "))
			continue
		} else if len(invalids) > 0 {
			fmt.Printf("Invalid locales: %s\n", strings.Join(invalids, ", "))
			continue
		}

		for i := range existingLocales {
			dedupedLocales[existingLocales[i]] = true
		}

		tomlLocales := make([]string, 0, len(dedupedLocales))
		for k := range dedupedLocales {
			tomlLocales = append(tomlLocales, k)
		}

		return tomlLocales, nil
	}
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

// getDefaultLocale prompts the user to specify the primary fallback language for the project.
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
			fmt.Printf("Default locale must be one of %s. Please try again.", strings.Join(options, ", "))
			continue
		}

		return validatedLocale, nil
	}
}

// getModelName prompts the user for the name of the Ollama model to be used for translations.
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

// getModelUrl prompts the user for the server URL of the Ollama model to be used for translations.
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

// readLine captures a single line of input from the provided reader and trims leading/trailing whitespace.
func readLine(reader *bufio.Reader) (string, error) {
	res, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(res), nil
}

func validateLocale(input string) (string, error) {
	tag, err := language.Parse(input)
	if err != nil {
		return "", err
	}

	return tag.String(), nil
}
