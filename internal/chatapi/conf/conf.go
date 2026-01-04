package conf

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// ChatAPI configuration
type Config struct {
	HTTP    *HTTPServer    `toml:"httpServer"`
	MySQL   *MySQL         `toml:"mysql"`
	Redis   *Redis         `toml:"redis"`
	Logic   *LogicRPC      `toml:"logic"`
	JWT     *JWTConfig     `toml:"jwt"`
	AI      *AIConfig      `toml:"ai"`
}

type HTTPServer struct {
	Addr         string `toml:"addr"`
	ReadTimeout  int    `toml:"readTimeout"`
	WriteTimeout int    `toml:"writeTimeout"`
}

type MySQL struct {
	DSN         string `toml:"dsn"`
	MaxIdleConn int    `toml:"maxIdle"`
	MaxOpenConn int    `toml:"maxOpen"`
	MaxLifetime int    `toml:"maxLifetime"`
}

type Redis struct {
	Network  string `toml:"network"`
	Addr     string `toml:"addr"`
	Auth     string `toml:"auth"`
	Database int    `toml:"database"`
}

type LogicRPC struct {
	AppID    string `toml:"appid"`
	Timeout  int    `toml:"timeout"`
	Endpoint string `toml:"endpoint"`
}

type JWTConfig struct {
	Secret     string `toml:"secret"`
	ExpireTime int    `toml:"expireTime"` // hours
}

type AIConfig struct {
	Provider    string  `toml:"provider"`    // openai, anthropic, local
	APIKey      string  `toml:"apiKey"`
	BaseURL     string  `toml:"baseUrl"`
	Model       string  `toml:"model"`       // gpt-3.5-turbo, gpt-4, etc.
	Temperature float64 `toml:"temperature"` // 0.0 - 1.0
	MaxTokens   int     `toml:"maxTokens"`
}

var (
	Conf *Config
)

// Init load config from file
func Init(cfgPath string) error {
	f, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("read config file error: %w", err)
	}

	Conf = &Config{}
	if err := toml.Unmarshal(f, Conf); err != nil {
		return fmt.Errorf("unmarshal config error: %w", err)
	}

	return nil
}
