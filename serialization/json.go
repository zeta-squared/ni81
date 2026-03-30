package serialization

import (
	"encoding/json"
	"maps"
	"os"
	"strings"
)

// ParseJSON decodes a JSON byte slice into a flattened JSON object.
//
// If the input data is empty or contains invalid JSON, it returns an empty map and an error.
func ParseJSON(data []byte) (map[string]string, error) {
	if len(data) == 0 {
		return map[string]string{}, nil
	}

	obj := make(map[string]any)
	err := json.Unmarshal(data, &obj)

	return Flatten("", obj), err // NOTE:inefficient as we may flatten already flattened JSON
}

// WriteJSON writes a map[string]T into a JSON file with indentation.
//
// If an error occurs during marshaling or writing to the file, it returns an error.
func WriteJSON[T any](path string, obj map[string]T) error {
	o, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, o, 0664)
}

// Flatten converts a nested map[string]any into a flat map[string]string.
//
// It prefixes keys with the given prefix and handles nested maps recursively.
func Flatten(prefix string, m map[string]any) map[string]string {
	flat := make(map[string]string)
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch val := v.(type) {
		case map[string]any:
			maps.Copy(flat, Flatten(key, val))
		case string:
			flat[key] = val
		}
	}

	return flat
}

// Unflatten converts a flat map[string]string back into a nested map[string]any.
//
// It splits keys on the '.' delimiter to reconstruct the original nested structure.
func Unflatten(flat map[string]string) map[string]any {
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
