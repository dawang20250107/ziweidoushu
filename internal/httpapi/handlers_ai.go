package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dawang20250107/ziweidoushu/internal/ai"
	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// interpretRequest AI 解读请求。
type interpretRequest struct {
	chartRequest
	Topic    string `json:"topic,omitempty"`    // overview/love/career/wealth/health/...
	Question string `json:"question,omitempty"` // 自由提问
	Stream   bool   `json:"stream,omitempty"`   // true = SSE 流式
}

func (s *Server) handleInterpret(w http.ResponseWriter, r *http.Request) {
	var req interpretRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if len([]rune(req.Question)) > 500 {
		writeError(w, http.StatusBadRequest, "question_too_long", "问题长度请控制在 500 字内")
		return
	}
	resp, err := s.computeChart(req.chartRequest)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_birth", err.Error())
		return
	}
	s.metrics.aiRequests.Add(1)

	stream := req.Stream || strings.Contains(r.Header.Get("Accept"), "text/event-stream")
	if !stream {
		result, err := s.interp.Interpret(r.Context(), resp.Chart, resp.Patterns, req.Topic, req.Question, nil)
		if err != nil {
			s.metrics.aiErrors.Add(1)
			writeAIError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
		return
	}

	// ── SSE 流式 ──
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "no_stream", "当前连接不支持流式输出")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // 关闭 Nginx 缓冲
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	sendEvent := func(event string, payload any) error {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, raw); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	result, err := s.interp.Interpret(r.Context(), resp.Chart, resp.Patterns, req.Topic, req.Question,
		func(delta string) error {
			return sendEvent("delta", map[string]string{"text": delta})
		})
	if err != nil {
		s.metrics.aiErrors.Add(1)
		_ = sendEvent("error", map[string]string{"message": aiErrorMessage(err)})
		return
	}
	_ = sendEvent("done", map[string]any{
		"provider": result.Provider,
		"degraded": result.Degraded,
		"length":   len([]rune(result.Text)),
	})
}

// hemingRequest 合盘请求。
type hemingRequest struct {
	A      chartRequest `json:"a"`
	B      chartRequest `json:"b"`
	WithAI bool         `json:"withAI,omitempty"`
}

// hemingPersonView 单人合盘视图。
type hemingPersonView struct {
	Chart      *ziwei.Chart    `json:"chart"`
	Patterns   []ziwei.Pattern `json:"patterns,omitempty"`
	FuqiStars  []string        `json:"fuqiStars"`    // 夫妻宫主星(空宫借对宫)
	FuqiBorrow bool            `json:"fuqiBorrowed"` // 是否借宫
	Readings   []hemingReading `json:"readings"`     // 夫妻宫断语
}

type hemingReading struct {
	Star         string `json:"star"`
	Summary      string `json:"summary"`
	Good         string `json:"good"`
	Bad          string `json:"bad"`
	SpouseTraits string `json:"spouseTraits"`
	Timing       string `json:"timing"`
	NiQuote      string `json:"niQuote,omitempty"`
}

func (s *Server) handleHeming(w http.ResponseWriter, r *http.Request) {
	var req hemingRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	a, err := s.computeChart(req.A)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_birth_a", "甲方生辰:"+err.Error())
		return
	}
	b, err := s.computeChart(req.B)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_birth_b", "乙方生辰:"+err.Error())
		return
	}

	out := map[string]any{
		"a":             s.hemingView(a),
		"b":             s.hemingView(b),
		"methodology":   s.kb.Heming.Methodology,
		"scoreCriteria": s.kb.Heming.ScoreCriteria,
	}

	if req.WithAI {
		s.metrics.aiRequests.Add(1)
		question := "请依据倪海厦体系的合盘方法论,对甲乙双方命盘的婚配契合度做整体分析:双方夫妻宫主星互参、四化互飞影响、性格互补与冲突点、婚期与相处建议。"
		combined := "甲方命盘:\n" + ai.ChartSummary(a.Chart) + "\n乙方命盘:\n" + ai.ChartSummary(b.Chart)
		result, err := s.interp.Interpret(r.Context(), a.Chart, a.Patterns, "love", question+"\n\n"+combined, nil)
		if err != nil {
			s.metrics.aiErrors.Add(1)
			out["ai"] = map[string]string{"error": aiErrorMessage(err)}
		} else {
			out["ai"] = result
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) hemingView(resp *chartResponse) hemingPersonView {
	view := hemingPersonView{Chart: resp.Chart, Patterns: resp.Patterns}
	fuqi := resp.Chart.PalaceByName("夫妻")
	if fuqi == nil {
		return view
	}
	stars := fuqi.MajorStarNames()
	if len(stars) == 0 {
		stars = fuqi.BorrowedStars
		view.FuqiBorrow = true
	}
	view.FuqiStars = stars
	for _, name := range stars {
		if entry, ok := s.kb.Heming.StarInFuqi[name]; ok {
			view.Readings = append(view.Readings, hemingReading{
				Star: name, Summary: entry.Summary, Good: entry.Good, Bad: entry.Bad,
				SpouseTraits: entry.SpouseTraits, Timing: entry.Timing, NiQuote: entry.NiQuote,
			})
		}
	}
	return view
}

func writeAIError(w http.ResponseWriter, err error) {
	if errors.Is(err, ai.ErrBusy) {
		w.Header().Set("Retry-After", "3")
		writeError(w, http.StatusTooManyRequests, "ai_busy", err.Error())
		return
	}
	writeError(w, http.StatusBadGateway, "ai_failed", aiErrorMessage(err))
}

func aiErrorMessage(err error) string {
	if errors.Is(err, ai.ErrBusy) {
		return err.Error()
	}
	return "AI 解读暂不可用: " + err.Error()
}
