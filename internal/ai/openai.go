package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Terry-Mao/goim/internal/ai/conf"
	log "github.com/golang/glog"
)

// OpenAI is the OpenAI implementation of the AI service
type OpenAI struct {
	config     *Config
	httpClient *http.Client
}

// NewOpenAI creates a new OpenAI service
func NewOpenAI(cfg *conf.AI) *OpenAI {
	config := &Config{
		Provider:    cfg.Provider,
		APIKey:      cfg.APIKey,
		Model:       cfg.Model,
		Temperature: cfg.Temperature,
		MaxTokens:   cfg.MaxTokens,
		BaseURL:     cfg.BaseURL,
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}

	return &OpenAI{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ChatRequest represents the OpenAI chat request
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// ChatResponse represents the OpenAI chat response
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a choice in the response
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Chat sends a message to the AI and returns the response
func (o *OpenAI) Chat(ctx context.Context, botID int64, personality *Personality, history []Message, userMessage string) (string, error) {
	// Build messages with system prompt
	messages := make([]Message, 0, len(history)+2)

	// Add system prompt
	systemPrompt := BuildSystemPrompt(personality)
	messages = append(messages, Message{
		Role:    "system",
		Content: systemPrompt,
	})

	// Add history
	messages = append(messages, history...)

	// Add user message
	messages = append(messages, Message{
		Role:    "user",
		Content: userMessage,
	})

	// Build request
	reqBody := ChatRequest{
		Model:    o.config.Model,
		Messages: messages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	// Create HTTP request
	url := o.config.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.config.APIKey)

	// Send request
	resp, err := o.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("OpenAI API error: status=%d, body=%s", resp.StatusCode, string(body))
		return "", fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	// Parse response
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// StreamChat streams the AI response (not implemented for now)
func (o *OpenAI) StreamChat(ctx context.Context, botID int64, personality *Personality, history []Message, userMessage string, callback func(chunk string)) error {
	// TODO: Implement SSE streaming
	return fmt.Errorf("streaming not implemented")
}

// SetModel sets the AI model to use
func (o *OpenAI) SetModel(model string) {
	o.config.Model = model
}

// SetTemperature sets the temperature for responses
func (o *OpenAI) SetTemperature(temp float64) {
	o.config.Temperature = temp
}
