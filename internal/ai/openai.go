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
)

// openaiProvider OpenAI Chat Completions 兼容协议。
// 通过 OPENAI_BASE_URL 可对接 DeepSeek、通义千问、智谱、Moonshot 等国产模型服务。
type openaiProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func newOpenAI(c Config, client *http.Client) *openaiProvider {
	return &openaiProvider{
		apiKey:  c.OpenAIAPIKey,
		baseURL: strings.TrimRight(c.OpenAIBaseURL, "/"),
		model:   c.OpenAIModel,
		client:  client,
	}
}

func (p *openaiProvider) Name() string { return "openai-compatible/" + p.model }

func (p *openaiProvider) Stream(ctx context.Context, req Request, onDelta func(string) error) (string, error) {
	msgs := make([]Message, 0, len(req.Messages)+1)
	if req.System != "" {
		msgs = append(msgs, Message{Role: "system", Content: req.System})
	}
	msgs = append(msgs, req.Messages...)

	body := map[string]any{
		"model":      p.model,
		"stream":     true,
		"messages":   msgs,
		"max_tokens": req.MaxTokens,
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("openai 兼容接口请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("openai 兼容接口返回 %d: %s", resp.StatusCode, string(msg))
	}

	var full strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			break
		}
		var ev struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(payload), &ev); err != nil {
			continue
		}
		if len(ev.Choices) > 0 && ev.Choices[0].Delta.Content != "" {
			full.WriteString(ev.Choices[0].Delta.Content)
			if onDelta != nil {
				if err := onDelta(ev.Choices[0].Delta.Content); err != nil {
					return full.String(), err
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return full.String(), fmt.Errorf("openai 兼容流读取失败: %w", err)
	}
	return full.String(), nil
}
