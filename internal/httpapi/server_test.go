package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dawang20250107/ziweidoushu/data"
	"github.com/dawang20250107/ziweidoushu/internal/ai"
	"github.com/dawang20250107/ziweidoushu/internal/config"
	"github.com/dawang20250107/ziweidoushu/internal/corpus"
	"github.com/dawang20250107/ziweidoushu/internal/knowledge"
)

func newTestServer(t *testing.T, mutate func(*config.Config)) *httptest.Server {
	t.Helper()
	store, err := corpus.NewStore(data.FS, "classics")
	if err != nil {
		t.Fatal(err)
	}
	kb, err := knowledge.Load(data.FS)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Addr:            ":0",
		RateLimitRPS:    0, // 测试默认关闭限流
		RequestTimeout:  10 * time.Second,
		ShutdownTimeout: time.Second,
		ChartCacheSize:  256,
		AI:              ai.Config{MaxConcurrency: 2, MaxTokens: 512, Timeout: 5 * time.Second},
	}
	if mutate != nil {
		mutate(&cfg)
	}
	interp := ai.NewInterpreter(nil, kb, store, cfg.AI) // 无供应商 → 规则化降级
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s := New(cfg, logger, store, kb, interp, Deps{})
	ts := httptest.NewServer(s.http.Handler)
	t.Cleanup(ts.Close)
	return ts
}

func postJSON(t *testing.T, url string, body any) (*http.Response, []byte) {
	t.Helper()
	raw, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp, data
}

func getJSON(t *testing.T, url string) (*http.Response, []byte) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp, data
}

var chartBody = map[string]any{
	"year": 1990, "month": 6, "day": 15, "hour": 5, "gender": "male",
}

func TestChartEndpoint(t *testing.T) {
	ts := newTestServer(t, nil)

	resp, body := postJSON(t, ts.URL+"/api/v1/chart", chartBody)
	if resp.StatusCode != 200 {
		t.Fatalf("排盘接口状态码 %d: %s", resp.StatusCode, body)
	}
	var out struct {
		OK   bool `json:"ok"`
		Data struct {
			Chart struct {
				WuxingJuName string `json:"wuxingJuName"`
				Palaces      []struct {
					Name  string `json:"name"`
					Stars []struct {
						Name string `json:"name"`
					} `json:"stars"`
				} `json:"palaces"`
			} `json:"chart"`
			Patterns []struct {
				Name string `json:"name"`
			} `json:"patterns"`
			Reading struct {
				Overview string `json:"overview"`
				Sections []struct {
					Key string `json:"key"`
				} `json:"sections"`
			} `json:"reading"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &out); err != nil || !out.OK {
		t.Fatalf("响应解析失败: %v %s", err, body)
	}
	if len(out.Data.Chart.Palaces) != 12 {
		t.Fatalf("宫位数 %d != 12", len(out.Data.Chart.Palaces))
	}
	if out.Data.Chart.WuxingJuName == "" {
		t.Fatal("五行局缺失")
	}
	// 多维断语须随盘序列化到 /api/v1/chart:含四化落宫、仆役维度与流年维度。
	keys := map[string]bool{}
	for _, s := range out.Data.Reading.Sections {
		keys[s.Key] = true
	}
	for _, want := range []string{"sihua", "jiaoyou", "liunian", "ming"} {
		if !keys[want] {
			t.Errorf("排盘响应的断语缺失维度 %s(全键:%v)", want, keys)
		}
	}

	// 非法输入
	resp, _ = postJSON(t, ts.URL+"/api/v1/chart", map[string]any{"year": 1700, "month": 1, "day": 1, "hour": 0, "gender": "male"})
	if resp.StatusCode != 400 {
		t.Fatalf("非法年份应 400, got %d", resp.StatusCode)
	}
	resp, _ = postJSON(t, ts.URL+"/api/v1/chart", map[string]any{"year": 1990, "month": 2, "day": 30, "hour": 0, "gender": "male"})
	if resp.StatusCode != 400 {
		t.Fatalf("不存在的日期应 400, got %d", resp.StatusCode)
	}
}

func TestHoroscopeEndpoint(t *testing.T) {
	ts := newTestServer(t, nil)
	body := map[string]any{
		"year": 1990, "month": 6, "day": 15, "hour": 5, "gender": "male",
		"target": map[string]any{"year": 2026, "month": 7, "day": 16, "hour": 6},
	}
	resp, raw := postJSON(t, ts.URL+"/api/v1/horoscope", body)
	if resp.StatusCode != 200 {
		t.Fatalf("运限接口 %d: %s", resp.StatusCode, truncate(raw))
	}
	var out struct {
		Data struct {
			Horoscope struct {
				NominalAge int `json:"nominalAge"`
				Decadal    struct {
					Name        string   `json:"name"`
					PalaceNames []string `json:"palaceNames"`
				} `json:"decadal"`
				Yearly struct {
					Stars [][]struct {
						Name string `json:"name"`
					} `json:"stars"`
				} `json:"yearly"`
			} `json:"horoscope"`
			Reading struct {
				Sections []struct {
					Key string `json:"key"`
				} `json:"sections"`
			} `json:"reading"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	h := out.Data.Horoscope
	if h.NominalAge != 37 {
		t.Errorf("虚岁: got %d want 37", h.NominalAge)
	}
	if h.Decadal.Name != "大限" || len(h.Decadal.PalaceNames) != 12 {
		t.Errorf("大限层异常: %+v", h.Decadal)
	}
	count := 0
	for _, cell := range h.Yearly.Stars {
		count += len(cell)
	}
	if count != 11 { // 流曜十颗 + 年解
		t.Errorf("流年流曜数: got %d want 11", count)
	}
	// 运限逐层断语须随响应一并返回(大限/流年/流月/流日/流时 五层)。
	rkeys := map[string]bool{}
	for _, sct := range out.Data.Reading.Sections {
		rkeys[sct.Key] = true
	}
	for _, want := range []string{"decadal", "yearly", "monthly", "daily", "hourly"} {
		if !rkeys[want] {
			t.Errorf("运限响应缺断语层 %s(全键:%v)", want, rkeys)
		}
	}

	// 越界目标
	body["target"] = map[string]any{"year": 1900, "month": 1, "day": 1, "hour": 0}
	resp, _ = postJSON(t, ts.URL+"/api/v1/horoscope", body)
	if resp.StatusCode != 400 {
		t.Errorf("早于出生的目标应 400: %d", resp.StatusCode)
	}
}

func TestCorpusEndpoints(t *testing.T) {
	ts := newTestServer(t, nil)

	resp, body := getJSON(t, ts.URL+"/api/v1/books")
	if resp.StatusCode != 200 || !strings.Contains(string(body), "骨髓赋") {
		t.Fatalf("书目接口异常: %d %s", resp.StatusCode, body)
	}
	resp, _ = getJSON(t, ts.URL+"/api/v1/books/gusuifu")
	if resp.StatusCode != 200 {
		t.Fatalf("取书失败: %d", resp.StatusCode)
	}
	resp, _ = getJSON(t, ts.URL+"/api/v1/books/gusuifu/chapters/0")
	if resp.StatusCode != 200 {
		t.Fatalf("取章节失败: %d", resp.StatusCode)
	}
	resp, _ = getJSON(t, ts.URL+"/api/v1/books/nonexistent")
	if resp.StatusCode != 404 {
		t.Fatalf("不存在的书应 404: %d", resp.StatusCode)
	}
	resp, body = getJSON(t, ts.URL+"/api/v1/search?q="+urlEncode("紫微"))
	if resp.StatusCode != 200 {
		t.Fatalf("搜索接口异常: %d %s", resp.StatusCode, truncate(body))
	}
	var searchOut struct {
		Data struct {
			Count int `json:"count"`
			Hits  []struct {
				Snippet string `json:"snippet"`
			} `json:"hits"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &searchOut); err != nil {
		t.Fatal(err)
	}
	if searchOut.Data.Count == 0 || !strings.Contains(searchOut.Data.Hits[0].Snippet, "<mark>") {
		t.Fatalf("搜索结果异常: %+v", searchOut.Data)
	}
}

func TestKnowledgeEndpoints(t *testing.T) {
	ts := newTestServer(t, nil)
	for _, path := range []string{
		"/api/v1/nihai/tianji", "/api/v1/nihai/renji", "/api/v1/nihai/diji", "/api/v1/nihai/bio",
		"/api/v1/knowledge/stars", "/api/v1/knowledge/topics", "/api/v1/knowledge/heming",
		"/api/v1/cities", "/api/v1/famous", "/api/v1/meta",
		"/api/v1/sihua/liunian?year=2026", "/api/v1/sihua/liuyue?year=2026&month=3",
	} {
		resp, body := getJSON(t, ts.URL+path)
		if resp.StatusCode != 200 {
			t.Errorf("%s: %d %s", path, resp.StatusCode, truncate(body))
		}
	}
	resp, _ := getJSON(t, ts.URL+"/api/v1/nihai/unknown")
	if resp.StatusCode != 404 {
		t.Errorf("未知分区应 404: %d", resp.StatusCode)
	}
	resp, _ = getJSON(t, ts.URL+"/api/v1/famous/ma-yun/chart")
	if resp.StatusCode != 200 {
		t.Errorf("名人排盘失败: %d", resp.StatusCode)
	}
}

func TestInterpretFallbackJSON(t *testing.T) {
	ts := newTestServer(t, nil)
	body := map[string]any{}
	for k, v := range chartBody {
		body[k] = v
	}
	body["topic"] = "love"
	resp, raw := postJSON(t, ts.URL+"/api/v1/ai/interpret", body)
	if resp.StatusCode != 200 {
		t.Fatalf("解读接口 %d: %s", resp.StatusCode, truncate(raw))
	}
	var out struct {
		Data struct {
			Text     string `json:"text"`
			Provider string `json:"provider"`
			Degraded bool   `json:"degraded"`
			Reading  struct {
				Overview string `json:"overview"`
				Sections []struct {
					Key  string `json:"key"`
					Text string `json:"text"`
				} `json:"sections"`
			} `json:"reading"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if !out.Data.Degraded || out.Data.Provider != "fallback/rule-based" {
		t.Fatalf("无 Key 时应降级: %+v", out.Data)
	}
	if !strings.Contains(out.Data.Text, "命格总论") {
		t.Fatalf("降级文本异常: %s", truncate([]byte(out.Data.Text)))
	}
	// 结构化多维断语须随盘生成(≥12 维度)。
	if len(out.Data.Reading.Sections) < 12 || out.Data.Reading.Overview == "" {
		t.Fatalf("多维断语骨架异常:维度数=%d", len(out.Data.Reading.Sections))
	}
}

func TestInterpretSSE(t *testing.T) {
	ts := newTestServer(t, nil)
	body := map[string]any{}
	for k, v := range chartBody {
		body[k] = v
	}
	body["stream"] = true
	raw, _ := json.Marshal(body)
	resp, err := http.Post(ts.URL+"/api/v1/ai/interpret", "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		t.Fatalf("SSE Content-Type 错误: %s", ct)
	}
	streamed, _ := io.ReadAll(resp.Body)
	text := string(streamed)
	if !strings.Contains(text, "event: delta") || !strings.Contains(text, "event: done") {
		t.Fatalf("SSE 事件缺失: %s", truncate(streamed))
	}
}

func TestTimingEndpoints(t *testing.T) {
	ts := newTestServer(t, nil)

	// 事项目录。
	resp, raw := getJSON(t, ts.URL+"/api/v1/timing/events")
	if resp.StatusCode != 200 {
		t.Fatalf("择吉目录 %d: %s", resp.StatusCode, truncate(raw))
	}
	var cat struct {
		Data struct {
			Events []struct {
				Key    string `json:"key"`
				Palace string `json:"palace"`
			} `json:"events"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &cat); err != nil {
		t.Fatal(err)
	}
	if len(cat.Data.Events) != 11 { // 7 择吉 + 4 避忌
		t.Fatalf("事项应 11 项,得 %d", len(cat.Data.Events))
	}

	// 事项择吉。
	resp, raw = postJSON(t, ts.URL+"/api/v1/timing/event", map[string]any{
		"year": 1990, "month": 6, "day": 15, "hour": 5, "gender": "male", "event": "wealth",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("事项择吉 %d: %s", resp.StatusCode, truncate(raw))
	}
	var out struct {
		Data struct {
			Timing struct {
				Palace  string `json:"palace"`
				Summary string `json:"summary"`
				Advice  string `json:"advice"`
			} `json:"timing"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out.Data.Timing.Palace != "财帛" || out.Data.Timing.Summary == "" || out.Data.Timing.Advice == "" {
		t.Errorf("求财择吉结果异常: %+v", out.Data.Timing)
	}

	// 未知事项 400。
	resp, _ = postJSON(t, ts.URL+"/api/v1/timing/event", map[string]any{
		"year": 1990, "month": 6, "day": 15, "hour": 5, "gender": "male", "event": "nope",
	})
	if resp.StatusCode != 400 {
		t.Errorf("未知事项应 400,得 %d", resp.StatusCode)
	}
}

func TestHemingEndpoint(t *testing.T) {
	ts := newTestServer(t, nil)
	resp, raw := postJSON(t, ts.URL+"/api/v1/heming", map[string]any{
		"a": map[string]any{"year": 1990, "month": 6, "day": 15, "hour": 5, "gender": "male"},
		"b": map[string]any{"year": 1992, "month": 3, "day": 8, "hour": 8, "gender": "female"},
	})
	if resp.StatusCode != 200 {
		t.Fatalf("合盘接口 %d: %s", resp.StatusCode, truncate(raw))
	}
	var out struct {
		Data struct {
			A struct {
				FuqiStars []string `json:"fuqiStars"`
				Readings  []any    `json:"readings"`
			} `json:"a"`
			Methodology string `json:"methodology"`
			Reading     struct {
				Score    int    `json:"score"`
				Level    string `json:"level"`
				Sections []struct {
					Key string `json:"key"`
				} `json:"sections"`
			} `json:"reading"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Data.A.FuqiStars) == 0 || out.Data.Methodology == "" {
		t.Fatalf("合盘数据不完整: %+v", out.Data)
	}
	// 合盘确定性契合断语须随响应返回。
	if out.Data.Reading.Score < 20 || out.Data.Reading.Level == "" {
		t.Errorf("合盘契合断语缺失或分数越界: %+v", out.Data.Reading)
	}
	rkeys := map[string]bool{}
	for _, sct := range out.Data.Reading.Sections {
		rkeys[sct.Key] = true
	}
	for _, want := range []string{"nianming", "sihuafly", "echo", "advice"} {
		if !rkeys[want] {
			t.Errorf("合盘断语缺维度 %s(全键:%v)", want, rkeys)
		}
	}
}

func TestRateLimit(t *testing.T) {
	ts := newTestServer(t, func(c *config.Config) {
		c.RateLimitRPS = 1
		c.RateLimitBurst = 3
	})
	limited := false
	for i := 0; i < 10; i++ {
		resp, _ := getJSON(t, ts.URL+"/api/v1/books")
		if resp.StatusCode == http.StatusTooManyRequests {
			limited = true
			break
		}
	}
	if !limited {
		t.Fatal("突发请求应触发限流")
	}
	// 探针豁免
	resp, _ := getJSON(t, ts.URL+"/healthz")
	if resp.StatusCode != 200 {
		t.Fatalf("healthz 不应被限流: %d", resp.StatusCode)
	}
}

func TestConcurrentChartRequests(t *testing.T) {
	ts := newTestServer(t, nil)
	var wg sync.WaitGroup
	errCh := make(chan error, 64)
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			body := map[string]any{
				"year": 1950 + i, "month": (i % 12) + 1, "day": (i % 28) + 1,
				"hour": i % 12, "gender": []string{"male", "female"}[i%2],
			}
			raw, _ := json.Marshal(body)
			resp, err := http.Post(ts.URL+"/api/v1/chart", "application/json", bytes.NewReader(raw))
			if err != nil {
				errCh <- err
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				b, _ := io.ReadAll(resp.Body)
				errCh <- fmt.Errorf("并发排盘 %d: %s", resp.StatusCode, truncate(b))
			}
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Error(err)
	}
}

func urlEncode(s string) string {
	var sb strings.Builder
	for _, b := range []byte(s) {
		sb.WriteString(fmt.Sprintf("%%%02X", b))
	}
	return sb.String()
}

func truncate(b []byte) string {
	if len(b) > 300 {
		return string(b[:300]) + "…"
	}
	return string(b)
}
