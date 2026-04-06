package translate

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/ollama/ollama/api"
)

type mockOllamaClient struct {
	heartbeatErr error
	generateErr  error
	response     string
}

func (m *mockOllamaClient) Heartbeat(ctx context.Context) error {
	return m.heartbeatErr
}

func (m *mockOllamaClient) Generate(ctx context.Context, _ *api.GenerateRequest, fn api.GenerateResponseFunc) error {
	if m.generateErr != nil {
		return m.generateErr
	}
	return fn(api.GenerateResponse{Response: m.response})
}

func Test_Translate_HeartbeatFailure(t *testing.T) {
	o := ollama{
		client: &mockOllamaClient{
			heartbeatErr: errors.New("no server"),
		},
		name: "test",
	}

	_, err := o.Translate("hello", "en", "es")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_Translate_GenerateFailure(t *testing.T) {
	o := ollama{
		client: &mockOllamaClient{
			generateErr: errors.New("fail"),
		},
		name: "test",
	}

	_, err := o.Translate("hello", "en", "es")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_Translate_Success(t *testing.T) {
	o := ollama{
		client: &mockOllamaClient{
			response: " hola ",
		},
		name: "test",
	}

	out, err := o.Translate("hello", "en", "es")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "hola"
	if out != expected {
		t.Fatalf("expected %s, got %s", expected, out)
	}
}

func Test_Translate_EmptyResponse(t *testing.T) {
	o := ollama{
		client: &mockOllamaClient{
			response: "",
		},
		name: "test",
	}

	_, err := o.Translate("hello", "en", "es")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_NewOllamaFromConfig(t *testing.T) {
	base, err := url.Parse("http://localhost:11434")
	if err != nil {
		t.Fatal(err)
	}

	httpClient := &http.Client{}
	name := "test-model"

	o, err := NewOllamaFromConfig(name, base, httpClient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if o.name != name {
		t.Fatalf("expected name %s, got %s", name, o.name)
	}

	if o.client == nil {
		t.Fatal("expected client to be non-nil")
	}
}
