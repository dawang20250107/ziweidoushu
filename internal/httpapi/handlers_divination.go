package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/dawang20250107/ziweidoushu/internal/ai"
	"github.com/dawang20250107/ziweidoushu/internal/liuyao"
	"github.com/dawang20250107/ziweidoushu/internal/meihua"
	"github.com/dawang20250107/ziweidoushu/internal/store"
)

// ── 占卜:梅花易数 + 小六壬 ──────────────────────────────────
// 起卦与卦象展示免费(引流);AI 深度解卦按次付费(credit_type=divination)。

type divinationRequest struct {
	Kind     string `json:"kind,omitempty"`     // meihua(默认)| liuyao
	Method   string `json:"method"`             // 梅花:time | number;六爻:shake | tosses
	Numbers  []int  `json:"numbers,omitempty"`  // 梅花数字起卦
	Tosses   []int  `json:"tosses,omitempty"`   // 六爻:六爻背面数(自下而上,每爻 0-3)
	CastAt   int64  `json:"castAt,omitempty"`   // 起卦时刻(unix 秒,缺省=当下)
	Question string `json:"question,omitempty"` // 求测之事
}

// castTime 解析起卦时刻(近 24h 防伪造)。
func castTime(castAt int64) (time.Time, error) {
	at := time.Now()
	if castAt > 0 {
		at = time.Unix(castAt, 0)
		if at.After(time.Now().Add(time.Minute)) || time.Since(at) > 24*time.Hour {
			return at, errors.New("起卦时刻须在近 24 小时内")
		}
	}
	return at, nil
}

// castMeihua 依请求重推卦象(服务端起卦,客户端不可伪造)。
func castMeihua(req divinationRequest) (*meihua.Result, error) {
	switch req.Method {
	case "number":
		r, err := meihua.ByNumbers(req.Numbers, req.Question)
		if err != nil {
			return nil, err
		}
		return &r, nil
	default: // time
		at, err := castTime(req.CastAt)
		if err != nil {
			return nil, err
		}
		r, err := meihua.ByTime(at, req.Question)
		if err != nil {
			return nil, err
		}
		return &r, nil
	}
}

// castLiuYao 六爻起卦(服务端摇卦或按用户报爻重现)。
func castLiuYao(req divinationRequest) (*liuyao.Result, error) {
	at, err := castTime(req.CastAt)
	if err != nil {
		return nil, err
	}
	if req.Method == "tosses" || len(req.Tosses) > 0 {
		return liuyao.ByTosses(req.Tosses, at, req.Question)
	}
	return liuyao.Shake(at, req.Question)
}

// handleLiuYao 六爻摇卦(免费)。
func (s *Server) handleLiuYao(w http.ResponseWriter, r *http.Request) {
	var req divinationRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if len(req.Question) > 200 {
		writeError(w, http.StatusBadRequest, "question_too_long", "所问之事请精简至 200 字内")
		return
	}
	result, err := castLiuYao(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "cast_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"result": result, "castAt": time.Now().Unix()})
}

// handleMeihua 梅花易数起卦(免费)。
func (s *Server) handleMeihua(w http.ResponseWriter, r *http.Request) {
	var req divinationRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if len(req.Question) > 200 {
		writeError(w, http.StatusBadRequest, "question_too_long", "所问之事请精简至 200 字内")
		return
	}
	result, err := castMeihua(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "cast_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"result": result, "castAt": time.Now().Unix()})
}

// handleXiaoLiuRen 小六壬快占(免费)。
func (s *Server) handleXiaoLiuRen(w http.ResponseWriter, r *http.Request) {
	var req divinationRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if len(req.Question) > 200 {
		writeError(w, http.StatusBadRequest, "question_too_long", "所问之事请精简至 200 字内")
		return
	}
	at := time.Now()
	if req.CastAt > 0 {
		at = time.Unix(req.CastAt, 0)
		if at.After(time.Now().Add(time.Minute)) || time.Since(at) > 24*time.Hour {
			writeError(w, http.StatusBadRequest, "cast_failed", "起算时刻须在近 24 小时内")
			return
		}
	}
	result, err := meihua.XiaoLiuRen(at, req.Question)
	if err != nil {
		writeError(w, http.StatusBadRequest, "cast_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"result": result})
}

// handleDivineAI AI 深度解卦:消耗 1 次 divination;AI 失败自动退还。
func (s *Server) handleDivineAI(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	var req divinationRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Question == "" {
		writeError(w, http.StatusBadRequest, "question_required", "请写明求测之事,解卦才有落点")
		return
	}
	if len(req.Question) > 200 {
		writeError(w, http.StatusBadRequest, "question_too_long", "所问之事请精简至 200 字内")
		return
	}
	// 付费解卦必须走真实 LLM
	if !s.interp.HasProvider() {
		writeError(w, http.StatusServiceUnavailable, "ai_unavailable", "AI 服务暂不可用,未扣次数")
		return
	}
	// 按占法起卦(服务端重推,客户端不可伪造)
	var meihuaResult *meihua.Result
	var liuyaoResult *liuyao.Result
	var castErr error
	if req.Kind == "liuyao" {
		liuyaoResult, castErr = castLiuYao(req)
	} else {
		meihuaResult, castErr = castMeihua(req)
	}
	if castErr != nil {
		writeError(w, http.StatusBadRequest, "cast_failed", castErr.Error())
		return
	}

	if err := s.store.ConsumeCredit(r.Context(), claims.Sub, "divination", req.Question); err != nil {
		if errors.Is(err, store.ErrNoCredits) {
			writeError(w, http.StatusPaymentRequired, "no_credits", "解卦次数不足,请先购买")
			return
		}
		writeError(w, http.StatusInternalServerError, "consume_failed", err.Error())
		return
	}

	s.metrics.aiRequests.Add(1)
	var reading ai.Result
	var err error
	if liuyaoResult != nil {
		reading, err = s.interp.DivineLiuYao(r.Context(), liuyaoResult, nil)
	} else {
		reading, err = s.interp.Divine(r.Context(), meihuaResult, nil)
	}
	if err != nil {
		s.metrics.aiErrors.Add(1)
		if rerr := s.store.RefundCredit(context.WithoutCancel(r.Context()), claims.Sub, "divination", "refund:divine"); rerr != nil {
			s.logger.Error("解卦次数退还失败,需人工补偿", "user", claims.Sub, "err", rerr)
		}
		if errors.Is(err, ai.ErrNoProvider) {
			writeError(w, http.StatusServiceUnavailable, "ai_unavailable", "AI 服务暂不可用,未扣次数")
			return
		}
		writeAIError(w, err)
		return
	}
	credits, _ := s.store.CreditBalances(r.Context(), claims.Sub)
	var resultAny any = meihuaResult
	if liuyaoResult != nil {
		resultAny = liuyaoResult
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"result":           resultAny,
		"reading":          reading,
		"remainingCredits": credits["divination"],
	})
}
