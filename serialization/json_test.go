package serialization

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_Flatten(t *testing.T) {
	data := []struct {
		file     string
		expected map[string]string
	}{
		{"unflattened-1.json", map[string]string{"key1": "value1", "key2.keyA": "subValue1", "key2.keyB": "subValue2", "key3.keyA.keya": "subSubValue1"}},
		{"unflattened-2.json", map[string]string{"key1": "value1"}},
		{"empty.json", map[string]string{}},
	}

	for _, d := range data {
		t.Run(d.file, func(t *testing.T) {
			data, err := os.ReadFile(fmt.Sprintf("testdata/%s", d.file))
			if err != nil {
				log.Fatal(err)
			}

			obj := map[string]any{}
			err = json.Unmarshal(data, &obj)
			if err != nil {
				log.Fatal(err)
			}

			flat := Flatten("", obj)
			fmt.Println(obj)
			fmt.Println(flat)

			if diff := cmp.Diff(d.expected, flat); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_Unflatten(t *testing.T) {
	data := []struct {
		file     string
		expected map[string]any
	}{
		{"flattened-1.json", map[string]any{"key1": "value1", "key2": map[string]any{"keyA": "subValue1", "keyB": "subValue2"}, "key3": map[string]any{"keyA": map[string]any{"keya": "subSubValue1"}}}},
		{"flattened-2.json", map[string]any{"key1": "value1"}},
		{"empty.json", map[string]any{}},
	}

	for _, d := range data {
		t.Run(d.file, func(t *testing.T) {
			data, err := os.ReadFile(fmt.Sprintf("testdata/%s", d.file))
			if err != nil {
				log.Fatal(err)
			}

			obj := map[string]string{}
			err = json.Unmarshal(data, &obj)
			if err != nil {
				log.Fatal(err)
			}

			out := Unflatten(obj)

			if diff := cmp.Diff(d.expected, out); diff != "" {
				t.Error(diff)
			}
		})
	}
}
