package cache

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type mockRW struct {
	readCalled  bool
	writeCalled bool
}

func (m *mockRW) Read(r io.Reader) (map[string]string, error) {
	m.readCalled = true
	return map[string]string{}, nil
}

func (m *mockRW) Write(w io.Writer, obj map[string]string, _ bool) error {
	m.writeCalled = true
	return nil
}

func Test_Read_Integration(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(path string) error
		expectErr   bool
		expectedMap map[string]string
	}{
		{
			name: "Cache exists",
			setup: func(path string) error {
				data, _ := json.Marshal(map[string]string{"key1": "value1"})
				return os.WriteFile(path, data, 0o664)
			},
			expectedMap: map[string]string{"key1": "value1"},
		},
		{
			name:      "Cache does not exist",
			setup:     func(string) error { return nil },
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			path := filepath.Join(tmp, "cache.json")

			if err := tt.setup(path); err != nil {
				t.Fatal(err)
			}

			fc := NewFileCache(path)

			got, err := fc.Read()

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.expectedMap, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_Write_Integration(t *testing.T) {
	t.Run("Cache exists", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "cache.json")
		if err := os.WriteFile(path, []byte(""), 0o664); err != nil {

			t.Fatal(err)
		}

		fc := NewFileCache(path)

		expected := map[string]string{"key1": "value1"}

		if err := fc.Write(expected); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		var cache map[string]string
		if err := json.Unmarshal(data, &cache); err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(expected, cache); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("Cache does not exist", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "cache.json")

		fc := NewFileCache(path)

		expected := map[string]string{"key1": "value1"}

		if err := fc.Write(expected); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("Cache dir does not exist", func(t *testing.T) {
		tmp := t.TempDir()

		fc := NewFileCache(tmp)

		expected := map[string]string{"key1": "value1"}

		if err := fc.Write(expected); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func Test_Read_Unit(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "cache.json")

	if err := os.WriteFile(path, []byte(""), 0o664); err != nil {
		t.Fatal(err)
	}

	mock := &mockRW{}

	fc := fileCache{
		path:       path,
		readWriter: mock,
	}

	_, _ = fc.Read()

	if !mock.readCalled {
		t.Fatal("expected Read to be called")
	}
}

func Test_Write_Unit(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "cache.json")

	if err := os.WriteFile(path, []byte(""), 0o664); err != nil {
		t.Fatal(err)
	}

	mock := &mockRW{}

	fc := fileCache{
		path:       path,
		readWriter: mock,
	}

	_ = fc.Write(map[string]string{})

	if !mock.writeCalled {
		t.Fatal("expected Write to be called")
	}
}
