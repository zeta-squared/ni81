package serialization

import (
	"encoding/json"
	"io"
	"maps"
	"strings"
)

// ReadWriter defines methods for reading and writing translation data.
// Implementations control how translation files are serialised and deserialised.
type ReadWriter interface {
	Read(r io.Reader) (map[string]string, error)
	Write(w io.Writer, obj map[string]string, unflatten bool) error
}

// JSONReadWriter implements ReadWriter for JSON-based translation files.
// It supports flattening nested JSON structures on read and optionally
// unflattening them on write.
type JSONReadWriter struct {
}

// Read parses JSON data from the provided reader and returns a flattened
// map of translation keys to values. Empty input results in an empty map.
func (JSONReadWriter) Read(r io.Reader) (map[string]string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return map[string]string{}, nil
	}

	obj := make(map[string]any)
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return nil, err
	}

	return flatten("", obj), err
}

// Write serialises the provided translation map to JSON and writes it to w.
//
// If unflattenObj is true, the map is first converted into a nested structure
// before encoding. Otherwise, the flat structure is written directly.
func (JSONReadWriter) Write(w io.Writer, obj map[string]string, unflattenObj bool) error {
	var o []byte
	var err error

	if unflattenObj {
		o, err = json.MarshalIndent(unflatten(obj), "", "  ")
	} else {
		o, err = json.MarshalIndent(obj, "", "  ")
	}

	if err != nil {
		return err
	}

	_, err = w.Write(o)

	return err
}

// flatten converts a nested map into a flat map using dot-separated keys.
//
// For example, {"a": {"b": "c"}} becomes {"a.b": "c"}.
func flatten(prefix string, m map[string]any) map[string]string {
	flat := make(map[string]string)
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch val := v.(type) {
		case map[string]any:
			maps.Copy(flat, flatten(key, val))
		case string:
			flat[key] = val
		}
	}

	return flat
}

// unflatten converts a flat map with dot-separated keys into a nested map.
//
// For example, {"a.b": "c"} becomes {"a": {"b": "c"}}.
func unflatten(flat map[string]string) map[string]any {
	root := make(map[string]any)

	for k, v := range flat {
		parts := strings.Split(k, ".")
		current := root

		for i, p := range parts {
			if i == len(parts)-1 {
				current[p] = v
				break
			}

			if _, ok := current[p]; !ok {
				current[p] = make(map[string]any)
			}

			current = current[p].(map[string]any)
		}
	}

	return root
}
