package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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
			Timeout: 0, // No timeout for streaming
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

// StreamChoice represents a choice in a streaming response
type StreamChoice struct {
	Index        int           `json:"index"`
	Delta        MessageDelta  `json:"delta"`
	FinishReason string        `json:"finish_reason"`
}

// MessageDelta represents the delta content in streaming
type MessageDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// StreamResponse represents a single chunk in the streaming response
type StreamResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
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

// StreamChat streams the AI response
func (o *OpenAI) StreamChat(ctx context.Context, botID int64, personality *Personality, history []Message, userMessage string, callback func(chunk string)) error {
	log.Infof("[StreamChat] Starting stream request for botID=%d", botID)

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

	// Build request with stream=true
	reqBody := ChatRequest{
		Model:    o.config.Model,
		Messages: messages,
		Stream:   true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	// Create HTTP request
	url := o.config.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.config.APIKey)
	req.Header.Set("Accept", "text/event-stream")

	log.Infof("[StreamChat] Sending request to %s with model %s", url, o.config.Model)

	// Send request
	resp, err := o.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	log.Infof("[StreamChat] Got response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("OpenAI API error: status=%d, body=%s", resp.StatusCode, string(body))
		return fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	// Read streaming response
	scanner := bufio.NewScanner(resp.Body)

	// Increase buffer size for long lines
	const maxCapacity = 256 * 1024 // 256KB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	chunkCount := 0
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for SSE format
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Remove "data: " prefix
		data := strings.TrimPrefix(line, "data: ")

		// Check for [DONE] marker
		if data == "[DONE]" {
			log.Infof("[StreamChat] Received [DONE] marker after %d chunks", chunkCount)
			break
		}

		// Parse JSON chunk
		var streamResp StreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			log.Warningf("Failed to parse stream chunk: %v, data: %s", err, data)
			continue
		}

		// Extract content from delta
		if len(streamResp.Choices) > 0 {
			delta := streamResp.Choices[0].Delta.Content
			if delta != "" {
				chunkCount++
				if chunkCount%10 == 0 {
					log.Infof("[StreamChat] Received %d chunks so far", chunkCount)
				}
				callback(delta)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Errorf("[StreamChat] Scanner error: %v", err)
		return fmt.Errorf("read stream: %w", err)
	}

	log.Infof("[StreamChat] Stream completed with %d total chunks", chunkCount)
	return nil
}

// SetModel sets the AI model to use
func (o *OpenAI) SetModel(model string) {
	o.config.Model = model
}

// SetTemperature sets the temperature for responses
func (o *OpenAI) SetTemperature(temp float64) {
	o.config.Temperature = temp
}
