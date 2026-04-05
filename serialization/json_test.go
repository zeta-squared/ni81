package serialization

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Test_Read
func Test_Read(t *testing.T) {
	noErr := func(err error) bool { return err == nil }

	data := []struct {
		name        string
		input       []byte
		expectedMap map[string]string
		checkErr    func(error) bool
	}{
		{
			name:  "Nested JSON Test",
			input: []byte(`{ "key1": "value1", "key2": { "keyA": "subValue1", "keyB": "subValue2" }, "key3": { "keyA": { "keya": "subSubValue1" } } }`),
			expectedMap: map[string]string{
				"key1":           "value1",
				"key2.keyA":      "subValue1",
				"key2.keyB":      "subValue2",
				"key3.keyA.keya": "subSubValue1",
			},
			checkErr: noErr,
		},
		{
			name:  "Flat JSON Test",
			input: []byte(`{ "key1": "value1", "key2": "value2" }`),
			expectedMap: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			checkErr: noErr,
		},
		{
			name:        "Empty JSON Test",
			input:       []byte(`{}`),
			expectedMap: map[string]string{},
			checkErr:    noErr,
		},
		{
			name:        "Malformed JSON Test",
			input:       []byte(`{ "key1": "value1 }`),
			expectedMap: nil,
			checkErr: func(err error) bool {
				var syntaxError *json.SyntaxError
				return errors.As(err, &syntaxError)
			},
		},
	}

	j := JSONReadWriter{}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			obj, err := j.Read(bytes.NewReader(d.input))

			if !d.checkErr(err) {
				t.Fatalf("Unexpected error: %v", err)
			}

			if diff := cmp.Diff(d.expectedMap, obj); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_WriteJSON(t *testing.T) {
	noErr := func(err error) bool { return err == nil }

	data := []struct {
		name         string
		input        map[string]string
		unflattenObj bool
		output       map[string]any
		checkErr     func(error) bool
	}{
		{
			name: "Flat write",
			input: map[string]string{
				"key1":      "value1",
				"key2.keyA": "subValue1",
			},
			unflattenObj: false,
			output: map[string]any{
				"key1":      "value1",
				"key2.keyA": "subValue1",
			},
			checkErr: noErr,
		},
		{
			name: "Unflatten on write",
			input: map[string]string{
				"key1":      "value1",
				"key2.keyA": "subValue1",
			},
			unflattenObj: true,
			output: map[string]any{
				"key1": "value1",
				"key2": map[string]any{
					"keyA": "subValue1",
				},
			},
			checkErr: noErr,
		},
		{
			name:         "Empty JSON Test",
			input:        map[string]string{},
			unflattenObj: false,
			output:       map[string]any{},
			checkErr:     noErr,
		},
	}

	j := JSONReadWriter{}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			var b bytes.Buffer

			err := j.Write(&b, d.input, d.unflattenObj)
			if !d.checkErr(err) {
				t.Fatalf("Unexpected error: %v", err)
			}

			var out map[string]any
			if err := json.Unmarshal(b.Bytes(), &out); err != nil {
				t.Fatalf("Invalid JSON output: %v", err)
			}

			if diff := cmp.Diff(d.output, out); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_Flatten(t *testing.T) {
	data := []struct {
		name     string
		input    map[string]any
		expected map[string]string
	}{
		{
			name: "Nested JSON Test",
			input: map[string]any{
				"key1": "value1",
				"key2": map[string]any{
					"keyA": "subValue1",
					"keyB": "subValue2",
				},
				"key3": map[string]any{
					"keyA": map[string]any{
						"keya": "subSubValue1",
					},
				},
			},
			expected: map[string]string{
				"key1":           "value1",
				"key2.keyA":      "subValue1",
				"key2.keyB":      "subValue2",
				"key3.keyA.keya": "subSubValue1",
			},
		},
		{
			name: "Flat JSON Test",
			input: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name:     "Empty JSON Test",
			input:    map[string]any{},
			expected: map[string]string{},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			if diff := cmp.Diff(d.expected, flatten("", d.input)); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_Unflatten(t *testing.T) {
	data := []struct {
		name     string
		input    map[string]string
		expected map[string]any
	}{
		{
			name: "Nested JSON Test",
			input: map[string]string{
				"key1":           "value1",
				"key2.keyA":      "subValue1",
				"key2.keyB":      "subValue2",
				"key3.keyA.keya": "subSubValue1",
			},
			expected: map[string]any{
				"key1": "value1",
				"key2": map[string]any{
					"keyA": "subValue1",
					"keyB": "subValue2",
				},
				"key3": map[string]any{
					"keyA": map[string]any{
						"keya": "subSubValue1",
					},
				},
			},
		},
		{
			name: "Flat JSON Test",
			input: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			expected: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name:     "Empty JSON Test",
			input:    map[string]string{},
			expected: map[string]any{},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			if diff := cmp.Diff(d.expected, unflatten(d.input)); diff != "" {
				t.Error(diff)
			}
		})
	}
}
