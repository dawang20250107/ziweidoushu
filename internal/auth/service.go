package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/dawang20250107/ziweidoushu/internal/store"
)

// SMSProvider 短信通道抽象(阿里云/腾讯云后续按此接入)。
type SMSProvider interface {
	Name() string
	// Send 发送验证码短信。
	Send(ctx context.Context, phone, code string) error
}

// DevSMS 开发通道:不真发短信,验证码写日志(便于本地与测试)。
type DevSMS struct{ Logger *slog.Logger }

func (d *DevSMS) Name() string { return "dev" }
func (d *DevSMS) Send(_ context.Context, phone, code string) error {
	d.Logger.Warn("DEV 短信通道(不真实发送)", "phone", phone, "code", code)
	return nil
}

// Config 鉴权服务配置。
type Config struct {
	JWTSecret     string
	JWTPrevSecret string
	AccessTTL     time.Duration // 默认 30 分钟
	RefreshTTL    time.Duration // 默认 30 天
	CodeTTL       time.Duration // 验证码有效期,默认 5 分钟(邮箱码 10 分钟另有下限)
	DevEchoCode   bool          // dev 通道下把验证码回显到接口(仅本地/E2E)
	// InviteRequired 内测闸门:邮箱注册须携带有效邀请码。
	InviteRequired bool
	// DeviceStrict 环境检测:陌生设备密码登录须邮箱验证码升级。
	DeviceStrict bool
}

// Service 鉴权服务。
type Service struct {
	st    *store.Store
	jwt   *JWT
	sms   SMSProvider
	email EmailProvider
	cfg   Config
	log   *slog.Logger

	// 会话版本进程内缓存(10s TTL):吊销延迟上限 10s,DB 压力可忽略
	svMu    sync.Mutex
	svCache map[string]svEntry
}

type svEntry struct {
	ver     int
	expires time.Time
}

// NewService 构建鉴权服务(email 为 nil 时用 Dev 通道,验证码写日志)。
func NewService(st *store.Store, sms SMSProvider, email EmailProvider, cfg Config, log *slog.Logger) (*Service, error) {
	if cfg.AccessTTL == 0 {
		cfg.AccessTTL = 30 * time.Minute
	}
	if cfg.RefreshTTL == 0 {
		cfg.RefreshTTL = 30 * 24 * time.Hour
	}
	if cfg.CodeTTL == 0 {
		cfg.CodeTTL = 5 * time.Minute
	}
	j, err := NewJWT(cfg.JWTSecret, cfg.JWTPrevSecret, cfg.AccessTTL)
	if err != nil {
		return nil, err
	}
	if email == nil {
		email = &DevEmail{Logger: log}
	}
	return &Service{st: st, jwt: j, sms: sms, email: email, cfg: cfg, log: log, svCache: map[string]svEntry{}}, nil
}

// 频控参数(docs/architecture/auth.md)
const (
	smsMinInterval = 60 * time.Second
	smsMaxPerHour  = 5
	smsMaxPerDayIP = 20
	smsMaxPerDay   = 10
)

// ErrRateLimited 触发频控。
var ErrRateLimited = errors.New("发送过于频繁,请稍后再试")

// SendCode 发送登录验证码。返回 devEcho(仅 DevEchoCode 开启时非空)。
func (s *Service) SendCode(ctx context.Context, phone, clientIP string) (string, error) {
	if err := ValidatePhone(phone); err != nil {
		return "", err
	}
	// 60s 间隔
	last, err := s.st.LastSMSAt(ctx, phone)
	if err != nil {
		return "", err
	}
	if time.Since(last) < smsMinInterval {
		return "", ErrRateLimited
	}
	// 小时/日配额
	byPhoneHour, _, err := s.st.CountRecentSMS(ctx, phone, clientIP, time.Hour)
	if err != nil {
		return "", err
	}
	byPhoneDay, byIPDay, err := s.st.CountRecentSMS(ctx, phone, clientIP, 24*time.Hour)
	if err != nil {
		return "", err
	}
	if byPhoneHour >= smsMaxPerHour || byPhoneDay >= smsMaxPerDay || byIPDay >= smsMaxPerDayIP {
		return "", ErrRateLimited
	}

	code, err := RandomCode(6)
	if err != nil {
		return "", err
	}
	if err := s.st.InsertSMSCode(ctx, phone, store.HashCode(code), "login", clientIP, s.cfg.CodeTTL); err != nil {
		return "", err
	}
	if err := s.sms.Send(ctx, phone, code); err != nil {
		return "", fmt.Errorf("短信发送失败: %w", err)
	}
	if s.cfg.DevEchoCode && s.sms.Name() == "dev" {
		return code, nil
	}
	return "", nil
}

// TokenPair 登录/刷新产物。
type TokenPair struct {
	Access    string    `json:"access"`
	Refresh   string    `json:"refresh"`
	ExpiresAt time.Time `json:"expiresAt"` // access 过期时间
}

// ErrCodeInvalid 验证码错误或过期(统一话术防遍历)。
var ErrCodeInvalid = errors.New("验证码错误或已过期")

// VerifyCode 验证码登录:校验→找/建用户→签发双 token。
func (s *Service) VerifyCode(ctx context.Context, phone, code, deviceInfo string) (*TokenPair, *store.User, bool, error) {
	if err := ValidatePhone(phone); err != nil {
		return nil, nil, false, err
	}
	if len(code) != 6 {
		return nil, nil, false, ErrCodeInvalid
	}
	ok, err := s.st.ConsumeSMSCode(ctx, phone, store.HashCode(code), "login")
	if err != nil {
		return nil, nil, false, err
	}
	if !ok {
		_ = s.st.FailSMSAttempt(ctx, phone)
		return nil, nil, false, ErrCodeInvalid
	}
	user, created, err := s.st.FindOrCreateUserByPhone(ctx, phone)
	if err != nil {
		return nil, nil, false, err
	}
	if user.Status != "active" {
		return nil, nil, false, errors.New("账号不可用")
	}
	pair, err := s.issuePair(ctx, user, deviceInfo, nil)
	if err != nil {
		return nil, nil, false, err
	}
	return pair, user, created, nil
}

// Refresh 旋转刷新:旧 token 失效、签发新对;复用检测触发全端下线。
func (s *Service) Refresh(ctx context.Context, refreshToken, deviceInfo string) (*TokenPair, *store.User, error) {
	rt, err := s.st.FindRefreshToken(ctx, HashToken(refreshToken))
	if errors.Is(err, store.ErrNotFound) {
		return nil, nil, ErrTokenInvalid
	}
	if err != nil {
		return nil, nil, err
	}
	if rt.RevokedAt != nil {
		// 已撤销令牌被再次使用 → 疑似令牌被盗:撤销该用户全部刷新令牌并强制全端重登
		s.log.Warn("刷新令牌复用检测触发,全端下线", "user", rt.UserID)
		_ = s.st.RevokeAllRefreshTokens(ctx, rt.UserID)
		_ = s.st.BumpSessionVer(ctx, rt.UserID)
		s.invalidateSV(rt.UserID)
		return nil, nil, ErrTokenInvalid
	}
	if time.Now().After(rt.ExpiresAt) {
		return nil, nil, ErrTokenExpired
	}
	user, err := s.st.GetUser(ctx, rt.UserID)
	if err != nil {
		return nil, nil, err
	}
	if user.Status != "active" {
		return nil, nil, errors.New("账号不可用")
	}
	if err := s.st.RevokeRefreshToken(ctx, rt.ID); err != nil {
		return nil, nil, err
	}
	pair, err := s.issuePair(ctx, user, deviceInfo, &rt.ID)
	if err != nil {
		return nil, nil, err
	}
	return pair, user, nil
}

// Logout 撤销当前设备刷新令牌;all=true 时全端下线(bump 会话版本)。
func (s *Service) Logout(ctx context.Context, userID, refreshToken string, all bool) error {
	if refreshToken != "" {
		if rt, err := s.st.FindRefreshToken(ctx, HashToken(refreshToken)); err == nil && rt.UserID == userID {
			_ = s.st.RevokeRefreshToken(ctx, rt.ID)
		}
	}
	if all {
		if err := s.st.RevokeAllRefreshTokens(ctx, userID); err != nil {
			return err
		}
		if err := s.st.BumpSessionVer(ctx, userID); err != nil {
			return err
		}
		s.invalidateSV(userID)
	}
	return nil
}

// Authenticate 校验访问令牌 + 会话版本,返回 Claims。
func (s *Service) Authenticate(ctx context.Context, token string) (*Claims, error) {
	claims, err := s.jwt.Verify(token, time.Now())
	if err != nil {
		return nil, err
	}
	ver, err := s.sessionVer(ctx, claims.Sub)
	if err != nil {
		return nil, err
	}
	if ver != claims.SV {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

func (s *Service) issuePair(ctx context.Context, user *store.User, deviceInfo string, rotatedFrom *string) (*TokenPair, error) {
	now := time.Now()
	access, err := s.jwt.Sign(user.ID, user.Tier, user.SessionVer, now)
	if err != nil {
		return nil, err
	}
	refresh, hash, err := NewRefreshToken()
	if err != nil {
		return nil, err
	}
	if _, err := s.st.InsertRefreshToken(ctx, user.ID, hash, deviceInfo, s.cfg.RefreshTTL, rotatedFrom); err != nil {
		return nil, err
	}
	return &TokenPair{Access: access, Refresh: refresh, ExpiresAt: now.Add(s.cfg.AccessTTL)}, nil
}

// sessionVer 带 10s 进程内缓存的会话版本读取。
func (s *Service) sessionVer(ctx context.Context, userID string) (int, error) {
	s.svMu.Lock()
	if e, ok := s.svCache[userID]; ok && time.Now().Before(e.expires) {
		s.svMu.Unlock()
		return e.ver, nil
	}
	s.svMu.Unlock()

	ver, err := s.st.SessionVer(ctx, userID)
	if err != nil {
		return 0, err
	}
	s.svMu.Lock()
	if len(s.svCache) > 100_000 { // 兜底防无界增长
		s.svCache = map[string]svEntry{}
	}
	s.svCache[userID] = svEntry{ver: ver, expires: time.Now().Add(10 * time.Second)}
	s.svMu.Unlock()
	return ver, nil
}

func (s *Service) invalidateSV(userID string) {
	s.svMu.Lock()
	delete(s.svCache, userID)
	s.svMu.Unlock()
}
