package httpapi

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"time"

	"github.com/dawang20250107/ziweidoushu/internal/auth"
	"github.com/dawang20250107/ziweidoushu/internal/store"
)

// ── 内测鉴权:邮箱验证码 + 自设密码 + 邀请码 ─────────────────
// 流程口径:注册(邮箱码+密码+邀请码)→ 平时密码登录 → 忘记走邮箱码重置;
// strict 模式下陌生设备登录须邮箱码升级(need_verify)。

// handleEmailSendCode 发送邮箱验证码(purpose: register / reset / login)。
func (s *Server) handleEmailSendCode(w http.ResponseWriter, r *http.Request) {
	if !s.guardAuth(w) {
		return
	}
	var req struct {
		Email   string `json:"email"`
		Purpose string `json:"purpose"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	devCode, err := s.auth.SendEmailCode(r.Context(), req.Email, req.Purpose, clientIP(r))
	if err != nil {
		status, code := http.StatusBadRequest, "send_failed"
		switch {
		case errors.Is(err, auth.ErrRateLimited):
			status, code = http.StatusTooManyRequests, "rate_limited"
		case errors.Is(err, store.ErrEmailTaken):
			code = "email_taken"
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

// handleEmailRegister 邮箱注册:验证码 + 自设密码 + 邀请码(内测闸门)。
func (s *Server) handleEmailRegister(w http.ResponseWriter, r *http.Request) {
	if !s.guardAuth(w) {
		return
	}
	var req struct {
		Email    string `json:"email"`
		Code     string `json:"code"`
		Password string `json:"password"`
		Invite   string `json:"invite,omitempty"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	pair, user, err := s.auth.RegisterEmail(r.Context(), req.Email, req.Code, req.Password, req.Invite, r.UserAgent())
	if err != nil {
		status, code := http.StatusBadRequest, "register_failed"
		switch {
		case errors.Is(err, store.ErrInviteInvalid):
			code = "invite_invalid"
		case errors.Is(err, store.ErrEmailTaken):
			code = "email_taken"
		case errors.Is(err, auth.ErrCodeInvalid):
			code = "code_invalid"
		}
		writeError(w, status, code, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tokens": pair, "user": userView(user), "created": true})
}

// handleEmailLogin 邮箱密码登录;strict 模式陌生设备返回 need_verify,
// 客户端引导发 login 用途验证码后带 code 重试同一端点。
func (s *Server) handleEmailLogin(w http.ResponseWriter, r *http.Request) {
	if !s.guardAuth(w) {
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Code     string `json:"code,omitempty"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	pair, user, err := s.auth.LoginEmail(r.Context(), req.Email, req.Password, req.Code, r.UserAgent())
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrNeedVerify):
			// 密码已对、仅差环境验证:200 + needVerify,前端走验证码分支
			writeJSON(w, http.StatusOK, map[string]any{"needVerify": true})
		case errors.Is(err, auth.ErrCodeInvalid):
			writeError(w, http.StatusUnauthorized, "code_invalid", err.Error())
		default:
			writeError(w, http.StatusUnauthorized, "login_failed", err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tokens": pair, "user": userView(user)})
}

// handlePasswordReset 忘记密码:邮箱验证码 + 新密码;成功后全端下线。
func (s *Server) handlePasswordReset(w http.ResponseWriter, r *http.Request) {
	if !s.guardAuth(w) {
		return
	}
	var req struct {
		Email       string `json:"email"`
		Code        string `json:"code"`
		NewPassword string `json:"newPassword"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := s.auth.ResetPassword(r.Context(), req.Email, req.Code, req.NewPassword); err != nil {
		code := "reset_failed"
		if errors.Is(err, auth.ErrCodeInvalid) {
			code = "code_invalid"
		}
		writeError(w, http.StatusBadRequest, code, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"reset": true})
}

// handlePasswordChange 登录态改密:旧密码 + 新密码;其余端下线。
func (s *Server) handlePasswordChange(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := s.auth.ChangePassword(r.Context(), claims.Sub, req.OldPassword, req.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, "change_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"changed": true})
}

// handleAdminInvites 铸造邀请码(内测运营):Bearer ADMIN_TOKEN。
func (s *Server) handleAdminInvites(w http.ResponseWriter, r *http.Request) {
	token := s.cfg.AdminToken
	if token == "" {
		writeError(w, http.StatusForbidden, "disabled", "未配置 ADMIN_TOKEN,管理接口已禁用")
		return
	}
	authz := r.Header.Get("Authorization")
	if len(authz) < 8 || subtle.ConstantTimeCompare([]byte(authz), []byte("Bearer "+token)) != 1 {
		writeError(w, http.StatusUnauthorized, "unauthorized", "凭证无效")
		return
	}
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store_disabled", "未配置 DATABASE_URL")
		return
	}
	var req struct {
		Count   int    `json:"count"`
		MaxUses int    `json:"maxUses"`
		Note    string `json:"note,omitempty"`
		// ExpiresInDays 有效天数,0 = 不过期
		ExpiresInDays int `json:"expiresInDays,omitempty"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Count < 1 || req.Count > 200 {
		writeError(w, http.StatusBadRequest, "bad_count", "count 须为 1-200")
		return
	}
	if req.MaxUses < 1 || req.MaxUses > 10000 {
		writeError(w, http.StatusBadRequest, "bad_max_uses", "maxUses 须为 1-10000")
		return
	}
	var expires *time.Time
	if req.ExpiresInDays > 0 {
		t := time.Now().AddDate(0, 0, req.ExpiresInDays)
		expires = &t
	}
	codes, err := s.store.CreateInvites(r.Context(), req.Count, req.MaxUses, req.Note, expires)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"codes": codes})
}
