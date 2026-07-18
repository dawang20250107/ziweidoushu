package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dawang20250107/ziweidoushu/internal/store"
)

// ── 命盘档案库 ────────────────────────────────────────────────

// freeProfileLimit 免费层档案上限;pro/master 不限。
const freeProfileLimit = 3

// profileCreateRequest 新建档案请求:生辰输入 + 档案元信息。
type profileCreateRequest struct {
	chartRequest
	Label     string `json:"label"`
	Relation  string `json:"relation,omitempty"` // self/family/friend/client/other
	IsDefault bool   `json:"isDefault,omitempty"`
}

func (s *Server) handleCreateProfile(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	var req profileCreateRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Label == "" {
		req.Label = req.Name
	}
	if req.Label == "" {
		writeError(w, http.StatusBadRequest, "invalid_label", "请为档案填写名称")
		return
	}

	// 生辰必须能排出盘才允许入库;快照与档案同生同版本(数据锁定)
	resp, err := s.computeChart(req.chartRequest)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_birth", err.Error())
		return
	}

	// 免费层数量闸门(计数+插入非原子,极端并发下可能略超,业务上可接受)
	user, err := s.store.GetUser(r.Context(), claims.Sub)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "profile_failed", err.Error())
		return
	}
	if user.Tier == "free" {
		n, err := s.store.CountProfiles(r.Context(), claims.Sub)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "profile_failed", err.Error())
			return
		}
		if n >= freeProfileLimit {
			writeError(w, http.StatusForbidden, "profile_limit",
				"免费版最多保存 3 份档案,升级会员解锁无限档案")
			return
		}
	}

	birthInput, err := json.Marshal(req.chartRequest)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "profile_failed", err.Error())
		return
	}
	snapshot, err := json.Marshal(resp)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "profile_failed", err.Error())
		return
	}
	p, err := s.store.CreateProfile(r.Context(), claims.Sub, req.Label, req.Relation,
		birthInput, snapshot, Version, req.IsDefault)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "profile_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"profile": p})
}

func (s *Server) handleListProfiles(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	profiles, err := s.store.ListProfiles(r.Context(), claims.Sub)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "profiles_failed", err.Error())
		return
	}
	if profiles == nil {
		profiles = []store.Profile{}
	}
	limit := 0 // 0 = 不限
	if user, err := s.store.GetUser(r.Context(), claims.Sub); err == nil && user.Tier == "free" {
		limit = freeProfileLimit
	}
	writeJSON(w, http.StatusOK, map[string]any{"profiles": profiles, "limit": limit})
}

func (s *Server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	p, err := s.store.GetProfile(r.Context(), claims.Sub, r.PathValue("id"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "档案不存在")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "profile_failed", err.Error())
		return
	}
	// 引擎升级后快照口径可能变化:按档案原始生辰实时重排,保证口径永远最新
	if p.EngineVersion != Version {
		var birth chartRequest
		if json.Unmarshal(p.BirthInput, &birth) == nil {
			if resp, err := s.computeChart(birth); err == nil {
				if snap, err := json.Marshal(resp); err == nil {
					p.ChartSnapshot = snap
					p.EngineVersion = Version
				}
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"profile": p})
}

func (s *Server) handleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	err := s.store.DeleteProfile(r.Context(), claims.Sub, r.PathValue("id"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "档案不存在")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "profile_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *Server) handleSetDefaultProfile(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	err := s.store.SetDefaultProfile(r.Context(), claims.Sub, r.PathValue("id"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "档案不存在")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "profile_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
