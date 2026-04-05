package project

import (
	"encoding/json"
	"errors"
	"ni81/config"
	"ni81/serialization"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type mockCache struct {
	readData map[string]string
	written  map[string]string
}

func (m *mockCache) Read() (map[string]string, error) {
	return m.readData, nil
}

func (m *mockCache) Write(obj map[string]string) error {
	m.written = obj
	return nil
}

type mockTranslator struct{}

func (m mockTranslator) Translate(s, from, to string) (string, error) {
	return s + "_translated_" + to, nil
}

// translator that FAILS
type failingTranslator struct{}

func (f failingTranslator) Translate(string, string, string) (string, error) {
	return "", errors.New("translation failed")
}

func Test_CreateCache(t *testing.T) {
	tmp := t.TempDir()

	// Setup default locale file
	defaultLocale := "en"
	path := filepath.Join(tmp, defaultLocale+".json")

	input := map[string]string{"key": "value"}
	data, _ := json.Marshal(input)

	if err := os.WriteFile(path, data, 0664); err != nil {
		t.Fatal(err)
	}

	mockCache := &mockCache{}

	p := project{
		defaultLocale:  defaultLocale,
		localeDir:      tmp,
		cache:          mockCache,
		jsonReadWriter: serialization.JSONReadWriter{},
	}

	if err := p.CreateCache(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if diff := cmp.Diff(input, mockCache.written); diff != "" {
		t.Error(diff)
	}
}

func Test_readTranslations(t *testing.T) {
	tmp := t.TempDir()

	input := `{"a": {"b": "c"}}`
	if err := os.WriteFile(filepath.Join(tmp, "en.json"), []byte(input), 0664); err != nil {
		t.Fatal(err)
	}

	p := project{
		localeDir:      tmp,
		jsonReadWriter: serialization.JSONReadWriter{},
	}

	out, err := p.readTranslations("en")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := map[string]string{"a.b": "c"}
	if diff := cmp.Diff(expected, out); diff != "" {
		t.Error(diff)
	}
}

func Test_writeTranslations(t *testing.T) {
	tmp := t.TempDir()

	path := filepath.Join(tmp, "en.json")
	if err := os.WriteFile(path, []byte("{}"), 0664); err != nil {
		t.Fatal(err)
	}

	p := project{
		localeDir:      tmp,
		jsonReadWriter: serialization.JSONReadWriter{},
	}

	input := map[string]string{"a.b": "c"}

	if err := p.writeTranslations("en", input); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)

	var out map[string]any
	_ = json.Unmarshal(data, &out)

	expected := map[string]any{
		"a": map[string]any{"b": "c"},
	}

	if diff := cmp.Diff(expected, out); diff != "" {
		t.Error(diff)
	}
}

func Test_Translate(t *testing.T) {
	tmp := t.TempDir()

	// default locale (new)
	newDefault := map[string]string{
		"k1": "v1",
	}

	data, _ := json.Marshal(newDefault)
	if err := os.WriteFile(filepath.Join(tmp, "en.json"), data, 0664); err != nil {
		t.Fatal(err)
	}

	// target locale
	if err := os.WriteFile(filepath.Join(tmp, "es.json"), []byte("{}"), 0664); err != nil {
		t.Fatal(err)
	}

	mockCache := &mockCache{
		readData: map[string]string{}, // oldDefault empty
	}

	p := project{
		defaultLocale:  "en",
		targetLocales:  []string{"es"},
		localeDir:      tmp,
		cache:          mockCache,
		jsonReadWriter: serialization.JSONReadWriter{},
		translator:     mockTranslator{},
	}

	if err := p.Translate(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// verify translated file
	data, _ = os.ReadFile(filepath.Join(tmp, "es.json"))

	var out map[string]any
	_ = json.Unmarshal(data, &out)

	expected := map[string]any{
		"k1": "v1_translated_es",
	}

	if diff := cmp.Diff(expected, out); diff != "" {
		t.Error(diff)
	}

	// verify cache updated
	if diff := cmp.Diff(newDefault, mockCache.written); diff != "" {
		t.Error(diff)
	}
}

func Test_Translate_Failure(t *testing.T) {
	tmp := t.TempDir()

	// default locale
	defaultData := map[string]string{
		"k1": "v1",
	}
	data, _ := json.Marshal(defaultData)

	if err := os.WriteFile(filepath.Join(tmp, "en.json"), data, 0664); err != nil {
		t.Fatal(err)
	}

	// target locale
	if err := os.WriteFile(filepath.Join(tmp, "es.json"), []byte("{}"), 0664); err != nil {
		t.Fatal(err)
	}

	mockCache := &mockCache{
		readData: map[string]string{}, // force "new key"
	}

	p := project{
		defaultLocale:  "en",
		targetLocales:  []string{"es"},
		localeDir:      tmp,
		cache:          mockCache,
		jsonReadWriter: serialization.JSONReadWriter{},
		translator:     failingTranslator{},
	}

	err := p.Translate()
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}

	// cache should NOT be written
	if mockCache.written != nil {
		t.Errorf("Expected cache not to be updated, got %v", mockCache.written)
	}

	// target file should remain unchanged (empty JSON)
	data, err = os.ReadFile(filepath.Join(tmp, "es.json"))
	if err != nil {
		t.Fatal(err)
	}

	var out map[string]any
	_ = json.Unmarshal(data, &out)

	if len(out) != 0 {
		t.Errorf("Expected empty output, got %v", out)
	}
}

func Test_NewProject(t *testing.T) {
	tmp := t.TempDir()

	cfg := []byte(`
[i18n]
default_locale = "en"
locales = ["en", "es"]
locale_dir = "i18n"

[model]
name = "test"
url = "http://localhost"
`)

	if err := os.WriteFile(filepath.Join(tmp, config.ConfigName), cfg, 0664); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(filepath.Join(tmp, "i18n"), 0755); err != nil {
		t.Fatal(err)
	}

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	p, err := NewProject(config.ConfigName)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if p.defaultLocale != "en" {
		t.Errorf("Expected en, got %s", p.defaultLocale)
	}

	if diff := cmp.Diff([]string{"es"}, p.targetLocales); diff != "" {
		t.Error(diff)
	}

	expectedDir := filepath.Join(tmp, "i18n")
	if p.localeDir != expectedDir {
		t.Errorf("Expected %s, got %s", expectedDir, p.localeDir)
	}
}
