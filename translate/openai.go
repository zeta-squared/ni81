package translate

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type chatClient interface {
	CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// openaicCleint is a Translator implementation that uses an OpenAI model
// to perform language translations.
type openaiClient struct {
	name   string
	client chatClient
}

// Translate converts the source text from fromLocale to toLocale using
// the configured OpenAI model.
//
// It verifies server availability, constructs a translation prompt,
// and invokes the Ollama API. An error is returned if the server is
// unreachable or the translation fails.
func (o openaiClient) Translate(source, fromLocale, toLocale string) (string, error) {
	systemMessage := openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf(`You are a professional %[1]s to %[2]s translator.
			Your goal is to accurately convey the meaning and nuances of the original %[1]s
			text while adhering to %[2]s grammar, vocabulary, and cultural sensetivities.
			Produce only the %[2]s translation, without any additional explanations or commentary.`, fromLocale, toLocale),
	}
	userMessage := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: fmt.Sprintf(`Please translate the following %[1]s text into %[2]s:\n\n %[3]s`, fromLocale, toLocale, source),
	}
	messages := []openai.ChatCompletionMessage{systemMessage, userMessage}
	chatCompletionRequest := openai.ChatCompletionRequest{
		Model:    o.name,
		Messages: messages,
		Stream:   false,
	}

	resp, err := o.client.CreateChatCompletion(context.Background(), chatCompletionRequest)
	if err != nil {
		return "", fmt.Errorf("translation of key %s to %s failed | %w", source, toLocale, err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("translation of key %s to %s failed: response contained no choices", source, toLocale)
	}

	return resp.Choices[0].Message.Content, nil
}

// NewOpenAiClient creates a new OpenAI Translator using the provided
// model name, base URL, and api key.
func NewOpenAiClient(name, baseUrl, apiKey string) (openaiClient, error) {
	myConfig := openai.DefaultConfig(apiKey)
	myConfig.BaseURL = baseUrl
	client := openai.NewClientWithConfig(myConfig)

	return openaiClient{
		name:   name,
		client: client,
	}, nil
}
