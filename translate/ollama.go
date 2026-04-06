package translate

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

// ollamaClient defines the subset of the Ollama API client used for translation.
// It allows mocking of the client for testing.
type ollamaClient interface {
	Heartbeat(ctx context.Context) error
	Generate(ctx context.Context, req *api.GenerateRequest, fn api.GenerateResponseFunc) error
}

// ollama is a Translator implementation that uses an Ollama model
// to perform language translations.
type ollama struct {
	name   string
	client ollamaClient
}

// Translate converts the source text from fromLocale to toLocale using
// the configured Ollama model.
//
// It verifies server availability, constructs a translation prompt,
// and invokes the Ollama API. An error is returned if the server is
// unreachable or the translation fails.
func (o ollama) Translate(source, fromLocale, toLocale string) (string, error) {
	if err := o.client.Heartbeat(context.Background()); err != nil {
		return "", fmt.Errorf("Could not find active server")
	}

	prompt := fmt.Sprintf(`You are a professional %[1]s to %[2]s translator.
		Your goal is to accurately convey the meaning and nuances of the original %[1]s
		text while adhering to %[2]s grammar, vocabulary, and cultural sensetivities.
		Produce only the %[2]s translation, without any additional explanations or commentary.
		Please translate the following %[1]s text into %[2]s:\n\n %[3]s`, fromLocale, toLocale, source)

	stream := false
	generateRequest := api.GenerateRequest{
		Model:     o.name,
		Prompt:    prompt,
		Stream:    &stream,
		Think:     &api.ThinkValue{Value: false},
		KeepAlive: &api.Duration{Duration: time.Second * 10},
		Options: map[string]any{
			"num_ctx": 4096,
		},
	}

	var translation string
	generateCallback := func(res api.GenerateResponse) error {
		if res.Response == "" {
			return fmt.Errorf("unexpected translation of %s to empty string", source)
		}
		translation = strings.TrimSpace(res.Response)
		return nil
	}
	err := o.client.Generate(context.Background(), &generateRequest, generateCallback)
	if err != nil {
		return "", fmt.Errorf("translation of key %s to %s failed", source, toLocale)
	}

	return translation, nil
}

// NewOllamaFromConfig creates a new ollama Translator using the provided
// model name, base URL, and HTTP client.
func NewOllamaFromConfig(name string, base *url.URL, client *http.Client) (ollama, error) {
	return ollama{
		client: api.NewClient(base, client),
		name:   name,
	}, nil
}
