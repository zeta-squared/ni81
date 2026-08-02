package translate

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
)

type mockOpenaiClient struct {
	resp    openai.ChatCompletionResponse
	err     error
	lastReq openai.ChatCompletionRequest
}

func newMockOpenaiClient(content string, err error) *mockOpenaiClient {
	return &mockOpenaiClient{
		resp: openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{Content: content},
				},
			},
		},
		err: err,
	}
}

func (m *mockOpenaiClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	m.lastReq = req
	return m.resp, m.err
}

func TestTranslate_Success(t *testing.T) {
	mock := newMockOpenaiClient("Hola mundo", nil)
	client := openaiClient{name: "gpt-4o", client: mock}

	got, err := client.Translate("Hello world", "English", "Spanish")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "Hola mundo"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTranslate_Error(t *testing.T) {
	sentinel := errors.New("server unreachable")
	mock := newMockOpenaiClient("", sentinel)
	client := openaiClient{name: "gpt-4o", client: mock}

	_, err := client.Translate("Hello world", "English", "Spanish")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, sentinel) {
		t.Errorf("expected error to wrap %v, got %v", sentinel, err)
	}

	if !strings.Contains(err.Error(), "translation of key Hello world to Spanish failed") {
		t.Errorf("expected wrapped message in %q", err.Error())
	}
}

func TestTranslate_EmptyChoices(t *testing.T) {
	mock := &mockOpenaiClient{resp: openai.ChatCompletionResponse{}}
	client := openaiClient{name: "gpt-4o", client: mock}

	_, err := client.Translate("Hello world", "English", "Spanish")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "response contained no choices") {
		t.Errorf("expected no-choices message in %q", err.Error())
	}
}

func TestTranslate_RequestFormatting(t *testing.T) {
	mock := newMockOpenaiClient("Hola mundo", nil)
	client := openaiClient{name: "gpt-4o", client: mock}

	if _, err := client.Translate("Hello world", "English", "Spanish"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.lastReq.Model != "gpt-4o" {
		t.Errorf("got model %q, want %q", mock.lastReq.Model, "gpt-4o")
	}

	if mock.lastReq.Stream {
		t.Error("expected Stream to be false")
	}

	if len(mock.lastReq.Messages) != 2 {
		t.Fatalf("got %d messages, want 2", len(mock.lastReq.Messages))
	}

	if mock.lastReq.Messages[0].Role != openai.ChatMessageRoleSystem {
		t.Errorf("got first message role %q, want %q", mock.lastReq.Messages[0].Role, openai.ChatMessageRoleSystem)
	}
	if !strings.Contains(mock.lastReq.Messages[0].Content, "English to Spanish") {
		t.Errorf("system message missing locales: %q", mock.lastReq.Messages[0].Content)
	}

	if mock.lastReq.Messages[1].Role != openai.ChatMessageRoleUser {
		t.Errorf("got second message role %q, want %q", mock.lastReq.Messages[1].Role, openai.ChatMessageRoleUser)
	}
	if !strings.Contains(mock.lastReq.Messages[1].Content, "Hello world") {
		t.Errorf("user message missing source text: %q", mock.lastReq.Messages[1].Content)
	}
}

func TestNewOpenAiClient(t *testing.T) {
	client, err := NewOpenAiClient("gpt-4o", "https://api.example.com/v1", "secret-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.name != "gpt-4o" {
		t.Errorf("got name %q, want %q", client.name, "gpt-4o")
	}
}
