package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client calls Yandex Cloud AI /v1/responses API (prompt + input -> output_text).
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	folderID   string
	promptID   string
}

// Config for the Yandex responses API.
type Config struct {
	APIKey   string // required
	BaseURL  string // e.g. https://ai.api.cloud.yandex.net/v1
	Project  string // folder ID (OpenAI-Project header)
	PromptID string // prompt id (e.g. fvtch4g1pmhosuh3r2cu)
}

type prompt struct {
	ID        string            `json:"id"`
	Variables map[string]string `json:"variables,omitempty"`
}

type responseRequest struct {
	Prompt prompt `json:"prompt"`
	Input  string `json:"input"`
}

// responseData matches Yandex API: output[].content[].text (type output_text)
type responseData struct {
	Output []outputMessage `json:"output"`
}

type outputMessage struct {
	Content []outputContent `json:"content"`
}

type outputContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// NewClient builds a client for the Yandex responses API.
func NewClient(cfg Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("llm: APIKey is required")
	}
	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://ai.api.cloud.yandex.net/v1"
	}
	if cfg.PromptID == "" {
		return nil, fmt.Errorf("llm: PromptID is required")
	}
	return &Client{
		httpClient: &http.Client{},
		baseURL:    baseURL,
		apiKey:     cfg.APIKey,
		folderID:   cfg.Project,
		promptID:   cfg.PromptID,
	}, nil
}

// ChatMessage is a single message in a conversation (role + content).
type ChatMessage struct {
	Role    string
	Content string
}

// Chat sends the conversation to the model. The full message history is
// formatted into a single input string (User: ... / Assistant: ...) so the
// prompt receives context; the last message should be from the user.
func (c *Client) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("llm: at least one message required")
	}
	input := formatMessagesAsInput(messages)

	reqBody := responseRequest{
		Prompt: prompt{ID: c.promptID},
		Input:  input,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("llm: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/responses", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("llm: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Key "+c.apiKey)
	req.Header.Set("OpenAI-Project", c.folderID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm: request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("llm: read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llm: %s: %s", resp.Status, string(body))
	}

	var data responseData
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("llm: parse response: %w", err)
	}
	text := extractOutputText(data)
	if text == "" {
		return "", fmt.Errorf("llm: no output_text in response")
	}
	return text, nil
}

func extractOutputText(data responseData) string {
	var b strings.Builder
	for _, msg := range data.Output {
		for _, c := range msg.Content {
			if c.Type == "output_text" && c.Text != "" {
				if b.Len() > 0 {
					b.WriteString("\n")
				}
				b.WriteString(c.Text)
			}
		}
	}
	return b.String()
}

func formatMessagesAsInput(messages []ChatMessage) string {
	var b strings.Builder
	for i, m := range messages {
		if i > 0 {
			b.WriteString("\n")
		}
		role := "User"
		if m.Role == "assistant" || m.Role == "system" {
			role = "Assistant"
		}
		b.WriteString(role)
		b.WriteString(": ")
		b.WriteString(m.Content)
	}
	return b.String()
}
