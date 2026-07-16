// Package ai AI 解读层:多供应商 LLM 接入(Anthropic 原生 + OpenAI 兼容协议)、
// 流式输出、上游并发限流,以及无 API Key 时的规则化降级解读。
package ai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Message 对话消息。
type Message struct {
	Role    string `json:"role"` // user | assistant
	Content string `json:"content"`
}

// Request 一次补全请求。
type Request struct {
	System      string
	Messages    []Message
	MaxTokens   int
	Temperature float64
}

// Provider LLM 供应商抽象。
type Provider interface {
	Name() string
	// Stream 流式补全:增量文本经 onDelta 逐段回调(可为 nil),返回完整文本。
	Stream(ctx context.Context, req Request, onDelta func(string) error) (string, error)
}

// Config AI 层配置(全部来自环境变量)。
type Config struct {
	Provider string // anthropic | openai | off | ""(自动探测)

	AnthropicAPIKey  string
	AnthropicBaseURL string
	AnthropicModel   string

	OpenAIAPIKey  string
	OpenAIBaseURL string
	OpenAIModel   string

	MaxConcurrency int           // 上游并发上限(信号量)
	Timeout        time.Duration // 单次上游调用超时
	MaxTokens      int
}

// ConfigFromEnv 读取环境变量。
func ConfigFromEnv() Config {
	c := Config{
		Provider:         os.Getenv("AI_PROVIDER"),
		AnthropicAPIKey:  os.Getenv("ANTHROPIC_API_KEY"),
		AnthropicBaseURL: envOr("ANTHROPIC_BASE_URL", "https://api.anthropic.com"),
		AnthropicModel:   envOr("ANTHROPIC_MODEL", "claude-sonnet-5"),
		OpenAIAPIKey:     os.Getenv("OPENAI_API_KEY"),
		OpenAIBaseURL:    envOr("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIModel:      envOr("OPENAI_MODEL", "gpt-4o-mini"),
		MaxConcurrency:   envInt("AI_MAX_CONCURRENCY", 8),
		Timeout:          time.Duration(envInt("AI_TIMEOUT_SECONDS", 120)) * time.Second,
		MaxTokens:        envInt("AI_MAX_TOKENS", 2048),
	}
	return c
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

// ErrNoProvider 未配置任何可用供应商。
var ErrNoProvider = errors.New("未配置 AI 供应商(设置 ANTHROPIC_API_KEY 或 OPENAI_API_KEY)")

// NewProvider 按配置构建供应商;返回 nil 表示走规则化降级(不报错)。
func NewProvider(c Config) (Provider, error) {
	client := &http.Client{Timeout: c.Timeout}
	switch c.Provider {
	case "off":
		return nil, nil
	case "anthropic":
		if c.AnthropicAPIKey == "" {
			return nil, fmt.Errorf("AI_PROVIDER=anthropic 但未设置 ANTHROPIC_API_KEY")
		}
		return newAnthropic(c, client), nil
	case "openai":
		if c.OpenAIAPIKey == "" {
			return nil, fmt.Errorf("AI_PROVIDER=openai 但未设置 OPENAI_API_KEY")
		}
		return newOpenAI(c, client), nil
	case "":
		// 自动探测:优先 Anthropic
		if c.AnthropicAPIKey != "" {
			return newAnthropic(c, client), nil
		}
		if c.OpenAIAPIKey != "" {
			return newOpenAI(c, client), nil
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("未知 AI_PROVIDER: %q(支持 anthropic/openai/off)", c.Provider)
	}
}
