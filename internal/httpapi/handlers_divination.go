package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dawang20250107/ziweidoushu/internal/ai"
	"github.com/dawang20250107/ziweidoushu/internal/auth"
	"github.com/dawang20250107/ziweidoushu/internal/daliuren"
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
	RecordID string `json:"recordId,omitempty"` // 卦档记录:AI 解卦回填目标
}

// optionalClaims 起卦免费接口的可选鉴权:带合法 Bearer 则识别用户(用于卦档),否则匿名。
func (s *Server) optionalClaims(r *http.Request) *auth.Claims {
	if s.auth == nil {
		return nil
	}
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return nil
	}
	claims, err := s.auth.Authenticate(r.Context(), strings.TrimPrefix(header, "Bearer "))
	if err != nil {
		return nil
	}
	return claims
}

// liuyaoSummary 卦档摘要:「地天泰 → 山风蛊」。
func liuyaoSummary(r *liuyao.Result) string {
	if r.BianName != "" {
		return r.BenName + " → " + r.BianName
	}
	return r.BenName
}

// meihuaSummary 卦档摘要:「泽火革 · 用克体」。
func meihuaSummary(r *meihua.Result) string {
	return fmt.Sprintf("%s · %s", r.Ben.Name, r.Relation)
}

// saveDivinationRecord 登录起卦自动存档;失败仅记日志,不影响起卦响应。
func (s *Server) saveDivinationRecord(r *http.Request, kind, question, summary string, payload any, castAt time.Time) string {
	claims := s.optionalClaims(r)
	if claims == nil || s.store == nil {
		return ""
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	id, err := s.store.SaveDivination(r.Context(), claims.Sub, kind, question, summary, raw, castAt)
	if err != nil {
		s.logger.Error("卦档存档失败", "user", claims.Sub, "err", err)
		return ""
	}
	return id
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
	castAt := time.Now()
	recordID := s.saveDivinationRecord(r, "liuyao", req.Question, liuyaoSummary(result), result, castAt)
	resp := map[string]any{"result": result, "castAt": castAt.Unix()}
	if recordID != "" {
		resp["recordId"] = recordID
	}
	writeJSON(w, http.StatusOK, resp)
}

// daliurenSummary 卦档摘要:「涉害课 · 三传巳丑酉」。
func daliurenSummary(r *daliuren.Result) string {
	return fmt.Sprintf("%s课 · 三传%s%s%s", r.KeType, r.Chuan[0], r.Chuan[1], r.Chuan[2])
}

// handleDaLiuRen 大六壬起课(免费):天地盘/四课/三传/课体 + 确定性断语。
func (s *Server) handleDaLiuRen(w http.ResponseWriter, r *http.Request) {
	var req divinationRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if len(req.Question) > 200 {
		writeError(w, http.StatusBadRequest, "question_too_long", "所问之事请精简至 200 字内")
		return
	}
	at, err := castTime(req.CastAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_cast_time", err.Error())
		return
	}
	result, err := daliuren.CastByTime(at)
	if err != nil {
		writeError(w, http.StatusBadRequest, "cast_failed", err.Error())
		return
	}
	recordID := s.saveDivinationRecord(r, "daliuren", req.Question, daliurenSummary(result), result, at)
	resp := map[string]any{"result": result, "castAt": at.Unix()}
	if recordID != "" {
		resp["recordId"] = recordID
	}
	writeJSON(w, http.StatusOK, resp)
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
	castAt := time.Now()
	recordID := s.saveDivinationRecord(r, "meihua", req.Question, meihuaSummary(result), result, castAt)
	resp := map[string]any{"result": result, "castAt": castAt.Unix()}
	if recordID != "" {
		resp["recordId"] = recordID
	}
	writeJSON(w, http.StatusOK, resp)
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
	recordID := s.saveDivinationRecord(r, "xiaoliuren", req.Question,
		result.Result.Name+" · "+result.Result.Luck, result, time.Now())
	resp := map[string]any{"result": result}
	if recordID != "" {
		resp["recordId"] = recordID
	}
	writeJSON(w, http.StatusOK, resp)
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
	// 卦档:已有记录回填解卦;无记录(如匿名起卦后才登录)补建一条带解卦的档
	var kind, summary string
	var payloadAny any
	if liuyaoResult != nil {
		kind, summary, payloadAny = "liuyao", liuyaoSummary(liuyaoResult), liuyaoResult
	} else {
		kind, summary, payloadAny = "meihua", meihuaSummary(meihuaResult), meihuaResult
	}
	if req.RecordID != "" {
		if aerr := s.store.AttachDivinationReading(r.Context(), claims.Sub, req.RecordID, reading.Text, reading.Provider); aerr != nil {
			s.logger.Error("卦档解卦回填失败", "user", claims.Sub, "record", req.RecordID, "err", aerr)
		}
	} else if raw, merr := json.Marshal(payloadAny); merr == nil {
		if id, serr := s.store.SaveDivination(r.Context(), claims.Sub, kind, req.Question, summary, raw, time.Now()); serr == nil {
			if aerr := s.store.AttachDivinationReading(r.Context(), claims.Sub, id, reading.Text, reading.Provider); aerr != nil {
				s.logger.Error("卦档解卦写入失败", "user", claims.Sub, "err", aerr)
			}
		}
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

// ── 卦档:列表 / 详情 / 删除(需登录)────────────────────────

func (s *Server) handleListDivinations(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	limit, offset := 50, 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= 100 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	kind := r.URL.Query().Get("kind")
	if kind != "meihua" && kind != "liuyao" && kind != "xiaoliuren" && kind != "daliuren" {
		kind = "" // 非法/缺省一律全部
	}
	records, total, err := s.store.ListDivinations(r.Context(), claims.Sub, kind, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"records": records, "total": total})
}

// validRecordID uuid 形状校验(避免非法值下探到 pg 报 500)。
func validRecordID(id string) bool {
	return len(id) == 36
}

func (s *Server) handleGetDivination(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	id := r.PathValue("id")
	if !validRecordID(id) {
		writeError(w, http.StatusNotFound, "not_found", "卦档不存在")
		return
	}
	rec, err := s.store.GetDivination(r.Context(), claims.Sub, id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "卦档不存在")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"record": rec})
}

func (s *Server) handleDeleteDivination(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	id := r.PathValue("id")
	if !validRecordID(id) {
		writeError(w, http.StatusNotFound, "not_found", "卦档不存在")
		return
	}
	err := s.store.DeleteDivination(r.Context(), claims.Sub, id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "卦档不存在")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "delete_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}
