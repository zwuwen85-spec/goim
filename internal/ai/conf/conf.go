package conf

// AI holds the AI service configuration
type AI struct {
	Provider    string  `toml:"provider"`
	APIKey      string  `toml:"api_key"`
	Model       string  `toml:"model"`
	Temperature float64 `toml:"temperature"`
	MaxTokens   int     `toml:"max_tokens"`
	BaseURL     string  `toml:"base_url"`
}

// FromChatAPIConfig creates an AI config from chatapi.AIConfig
func FromChatAPIConfig(provider, apiKey, baseURL, model string, temperature float64, maxTokens int) *AI {
	return &AI{
		Provider:    provider,
		APIKey:      apiKey,
		Model:       model,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		BaseURL:     baseURL,
	}
}
