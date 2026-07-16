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

// anthropicProvider Anthropic Messages API(原生 SSE 流式)。
type anthropicProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func newAnthropic(c Config, client *http.Client) *anthropicProvider {
	return &anthropicProvider{
		apiKey:  c.AnthropicAPIKey,
		baseURL: strings.TrimRight(c.AnthropicBaseURL, "/"),
		model:   c.AnthropicModel,
		client:  client,
	}
}

func (p *anthropicProvider) Name() string { return "anthropic/" + p.model }

func (p *anthropicProvider) Stream(ctx context.Context, req Request, onDelta func(string) error) (string, error) {
	body := map[string]any{
		"model":      p.model,
		"max_tokens": req.MaxTokens,
		"stream":     true,
		"messages":   req.Messages,
	}
	if req.System != "" {
		body["system"] = req.System
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/v1/messages", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("anthropic 请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("anthropic 返回 %d: %s", resp.StatusCode, string(msg))
	}

	var full strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		var ev struct {
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
		}
		if err := json.Unmarshal([]byte(payload), &ev); err != nil {
			continue
		}
		if ev.Type == "content_block_delta" && ev.Delta.Type == "text_delta" && ev.Delta.Text != "" {
			full.WriteString(ev.Delta.Text)
			if onDelta != nil {
				if err := onDelta(ev.Delta.Text); err != nil {
					return full.String(), err
				}
			}
		}
		if ev.Type == "message_stop" {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return full.String(), fmt.Errorf("anthropic 流读取失败: %w", err)
	}
	return full.String(), nil
}
