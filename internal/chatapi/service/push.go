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
	// Build URL with operation and keys as query parameters
	url := fmt.Sprintf("%s/goim/push/keys?operation=%d", p.endpoint, operation)
	for _, key := range keys {
		url += fmt.Sprintf("&keys=%s", key)
	}

	// Create request with raw message bytes as body
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(msg))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")

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
	fmt.Printf("[PushRoom] Calling URL: %s, msg_len=%d\n", url, len(msg))

	// Create request with raw message bytes as body
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(msg))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		fmt.Printf("[PushRoom] Request failed: %v\n", err)
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("[PushRoom] Non-200 status: %d, body: %s\n", resp.StatusCode, string(body))
		return fmt.Errorf("push failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	fmt.Printf("[PushRoom] Success: status=%d\n", resp.StatusCode)
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
