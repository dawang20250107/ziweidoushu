package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/dawang20250107/ziweidoushu/internal/auth"
	"github.com/dawang20250107/ziweidoushu/internal/store"
)

type authCtxKey int

const claimsKey authCtxKey = iota

// currentClaims 取当前请求的鉴权载荷,未登录返回 nil。
func currentClaims(ctx context.Context) *auth.Claims {
	c, _ := ctx.Value(claimsKey).(*auth.Claims)
	return c
}

// requireAuth 鉴权中间件:Authorization: Bearer <access>。
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.auth == nil {
			writeError(w, http.StatusServiceUnavailable, "auth_disabled", "用户体系未启用(未配置 DATABASE_URL/JWT_SECRET)")
			return
		}
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "unauthorized", "请先登录")
			return
		}
		claims, err := s.auth.Authenticate(r.Context(), strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			code := "unauthorized"
			if errors.Is(err, auth.ErrTokenExpired) {
				code = "token_expired"
			}
			writeError(w, http.StatusUnauthorized, code, "登录状态失效,请重新登录")
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), claimsKey, claims)))
	}
}

// guardAuth 未启用用户体系时统一 503。
func (s *Server) guardAuth(w http.ResponseWriter) bool {
	if s.auth == nil {
		writeError(w, http.StatusServiceUnavailable, "auth_disabled", "用户体系未启用(未配置 DATABASE_URL/JWT_SECRET)")
		return false
	}
	return true
}

func (s *Server) handleSMSSend(w http.ResponseWriter, r *http.Request) {
	if !s.guardAuth(w) {
		return
	}
	var req struct {
		Phone string `json:"phone"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	devCode, err := s.auth.SendCode(r.Context(), req.Phone, clientIP(r))
	if err != nil {
		status, code := http.StatusBadRequest, "sms_failed"
		if errors.Is(err, auth.ErrRateLimited) {
			status, code = http.StatusTooManyRequests, "rate_limited"
		}
		writeError(w, status, code, err.Error())
		return
	}
	resp := map[string]any{"sent": true}
	if devCode != "" {
		resp["devCode"] = devCode // 仅 dev 通道 + 显式开关时回显
	}
	writeJSON(w, http.StatusOK, resp)
}

// userView 用户对外视图。
func userView(u *store.User) map[string]any {
	return map[string]any{
		"id":       u.ID,
		"nickname": u.Nickname,
		"avatar":   u.AvatarURL,
		"tier":     u.Tier,
	}
}

func (s *Server) handleSMSVerify(w http.ResponseWriter, r *http.Request) {
	if !s.guardAuth(w) {
		return
	}
	var req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	pair, user, created, err := s.auth.VerifyCode(r.Context(), req.Phone, req.Code, r.UserAgent())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "verify_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"tokens":  pair,
		"user":    userView(user),
		"created": created,
	})
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if !s.guardAuth(w) {
		return
	}
	var req struct {
		Refresh string `json:"refresh"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	pair, user, err := s.auth.Refresh(r.Context(), req.Refresh, r.UserAgent())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "refresh_failed", "登录状态失效,请重新登录")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tokens": pair, "user": userView(user)})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	var req struct {
		Refresh string `json:"refresh,omitempty"`
		All     bool   `json:"all,omitempty"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := s.auth.Logout(r.Context(), claims.Sub, req.Refresh, req.All); err != nil {
		writeError(w, http.StatusInternalServerError, "logout_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"loggedOut": true})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	user, err := s.store.GetUser(r.Context(), claims.Sub)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "用户不存在")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": userView(user)})
}
