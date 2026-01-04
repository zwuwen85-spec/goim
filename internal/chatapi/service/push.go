package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PushClient handles pushing messages through goim Logic
type PushClient struct {
	endpoint string
	client   *http.Client
}

// NewPushClient creates a new push client
func NewPushClient(endpoint string) *PushClient {
	return &PushClient{
		endpoint: endpoint,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// PushKeys pushes a message to specific users by their keys
func (p *PushClient) PushKeys(ctx context.Context, operation int32, keys []string, msg []byte) error {
	url := fmt.Sprintf("%s/goim/push/keys?operation=%d", p.endpoint, operation)

	// Create request body
	reqBody := map[string]interface{}{
		"keys": keys,
		"msg":  string(msg),
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("push failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}

// PushRoom pushes a message to a room
func (p *PushClient) PushRoom(ctx context.Context, operation int32, typ, room string, msg []byte) error {
	url := fmt.Sprintf("%s/goim/push/room?operation=%d&type=%s&room=%s", p.endpoint, operation, typ, room)

	// Create request body
	reqBody := map[string]interface{}{
		"msg": string(msg),
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("push failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}

// GetOnlineUsers gets the online user count from goim
func (p *PushClient) GetOnlineTotal(ctx context.Context) (int64, error) {
	url := fmt.Sprintf("%s/goim/online/total", p.endpoint)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("parse response: %w", err)
	}

	if code, ok := result["code"].(float64); ok && code != 0 {
		return 0, fmt.Errorf("api error: %v", result)
	}

	if total, ok := result["data"].(float64); ok {
		return int64(total), nil
	}

	return 0, fmt.Errorf("invalid response format")
}
