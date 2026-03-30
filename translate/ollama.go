package translate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Ollama represents a client for interacting with an Ollama API instance
// to perform language model operations.
type Ollama struct {
	Model string
	Url   string
}

// Translate takes a source string and converts it from the fromLocale
// to the toLocale using the configured Ollama model.
//
// It returns the translated text or an error if the request fails.
func (o Ollama) Translate(source, fromLocale, toLocale string) (string, error) {
	// TODO: use Go ollama module

	// TODO: model heartbeat

	prompt := fmt.Sprintf(`You are a professional %[1]s to %[2]s translator.
		Your goal is to accurately convey the meaning and nuances of the original %[1]s
		text while adhering to %[2]s grammar, vocabulary, and cultural sensetivities.
		Produce only the %[2]s translation, without any additional explanations or commentary.
		Please translate the following %[1]s text into %[2]s:\n\n %[3]s`, fromLocale, toLocale, source)

	payload := map[string]any{
		"model":      o.Model,
		"prompt":     prompt,
		"stream":     false,
		"think":      false,
		"keep_alive": "10s",
		"opts": map[string]any{
			"num_ctx": 4096,
		},
	}
	payloadB, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	res, err := http.Post(o.Url+"/api/generate", "application/json", bytes.NewReader(payloadB))
	if err != nil {
		return "", err
	}

	// TODO: properly type `o`
	var out map[string]string
	if res.StatusCode == 200 {
		defer res.Body.Close()

		json.NewDecoder(res.Body).Decode(&out)
	} else {
		return "", fmt.Errorf("translation of key %s to %s failed", source, toLocale)
	}

	return strings.TrimSuffix(out["response"], "\n"), nil
}
