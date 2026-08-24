package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"net/mail"
	"net/smtp"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/dawang20250107/ziweidoushu/internal/store"
)

func b64s(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }

// DeviceHash 设备指纹:UA 归一后哈希(环境检测底账键;粗粒度足矣——
// 只求「同一浏览器不再重复升级验证」,不做跨端追踪)。
func DeviceHash(deviceInfo string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(deviceInfo)))
	return hex.EncodeToString(sum[:16])
}

// ── 邮件通道 ─────────────────────────────────────────────────

// EmailProvider 验证码邮件通道抽象(SMTP / 事务邮件服务商按此接入)。
type EmailProvider interface {
	Name() string
	Send(ctx context.Context, to, code, purpose string) error
}

// DevEmail 开发通道:不真发邮件,验证码写日志(便于本地与测试)。
type DevEmail struct{ Logger *slog.Logger }

func (d *DevEmail) Name() string { return "dev" }
func (d *DevEmail) Send(_ context.Context, to, code, purpose string) error {
	d.Logger.Warn("DEV 邮件通道(不真实发送)", "to", to, "code", code, "purpose", purpose)
	return nil
}

// SMTPConfig SMTP 通道配置(465 隐式 TLS / 587 STARTTLS 皆可)。
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string // 发件人,如 观星台 <noreply@example.com>
}

// SMTPEmail 标准 SMTP 通道(企业邮箱/阿里云邮件推送/腾讯云 SES 通用)。
type SMTPEmail struct{ Cfg SMTPConfig }

func (s *SMTPEmail) Name() string { return "smtp" }

var purposeLabel = map[string]string{"register": "注册", "reset": "重置密码", "login": "登录验证"}

func (s *SMTPEmail) Send(ctx context.Context, to, code, purpose string) error {
	subject := fmt.Sprintf("【观星台】%s验证码:%s", purposeLabel[purpose], code)
	body := fmt.Sprintf("您的%s验证码为 %s,10 分钟内有效。若非本人操作请忽略此邮件。",
		purposeLabel[purpose], code)
	msg := strings.Join([]string{
		"From: " + s.Cfg.From,
		"To: " + to,
		"Subject: =?UTF-8?B?" + b64s(subject) + "?=",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: base64",
		"",
		b64s(body),
	}, "\r\n")

	addr := net.JoinHostPort(s.Cfg.Host, strconv.Itoa(s.Cfg.Port))
	auth := smtp.PlainAuth("", s.Cfg.Username, s.Cfg.Password, s.Cfg.Host)
	from := s.Cfg.Username

	// 465:隐式 TLS 手工拨号;587/25:smtp.SendMail 自带 STARTTLS 协商
	if s.Cfg.Port == 465 {
		d := tls.Dialer{Config: &tls.Config{ServerName: s.Cfg.Host}}
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err != nil {
			return err
		}
		c, err := smtp.NewClient(conn, s.Cfg.Host)
		if err != nil {
			return err
		}
		defer c.Close()
		if err := c.Auth(auth); err != nil {
			return err
		}
		if err := c.Mail(from); err != nil {
			return err
		}
		if err := c.Rcpt(to); err != nil {
			return err
		}
		w, err := c.Data()
		if err != nil {
			return err
		}
		if _, err := w.Write([]byte(msg)); err != nil {
			return err
		}
		if err := w.Close(); err != nil {
			return err
		}
		return c.Quit()
	}
	// 587/25:SendMail 内部完成 STARTTLS 协商 + AUTH
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}

// ── 校验与密码 ───────────────────────────────────────────────

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// ValidateEmail 归一(小写去空白)并校验邮箱格式。
func ValidateEmail(email string) (string, error) {
	e := strings.ToLower(strings.TrimSpace(email))
	if len(e) > 254 || !emailRe.MatchString(e) {
		return "", errors.New("邮箱格式不正确")
	}
	if _, err := mail.ParseAddress(e); err != nil {
		return "", errors.New("邮箱格式不正确")
	}
	return e, nil
}

// ValidatePassword 密码策略:8-72 位,须含字母与数字(bcrypt 上限 72 字节)。
func ValidatePassword(pw string) error {
	if len(pw) < 8 || len(pw) > 72 {
		return errors.New("密码须为 8-72 位")
	}
	var hasLetter, hasDigit bool
	for _, r := range pw {
		switch {
		case r >= '0' && r <= '9':
			hasDigit = true
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
			hasLetter = true
		}
	}
	if !hasLetter || !hasDigit {
		return errors.New("密码须同时包含字母与数字")
	}
	return nil
}

// dummyHash 用户不存在时的假比对目标:恒定耗时防邮箱枚举。
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-anti-enum"), bcrypt.DefaultCost)

// ── 服务流程 ─────────────────────────────────────────────────

// ErrNeedVerify 陌生设备登录须验证码升级(strict 模式)。
var ErrNeedVerify = errors.New("新设备登录,需邮箱验证码确认")

// ErrBadCredential 邮箱或密码错误(统一话术防枚举)。
var ErrBadCredential = errors.New("邮箱或密码错误")

// SendEmailCode 发送邮箱验证码(注册/重置/登录升级),频控与短信同刻度。
// register 用途且邮箱已注册时直接报错,省一次无效邮件。
func (s *Service) SendEmailCode(ctx context.Context, email, purpose, clientIP string) (string, error) {
	e, err := ValidateEmail(email)
	if err != nil {
		return "", err
	}
	if purpose != "register" && purpose != "reset" && purpose != "login" {
		return "", errors.New("不支持的验证码用途")
	}
	if purpose == "register" {
		if _, _, err := s.st.FindUserByEmail(ctx, e); err == nil {
			return "", store.ErrEmailTaken
		}
	}
	last, err := s.st.LastEmailAt(ctx, e)
	if err != nil {
		return "", err
	}
	if time.Since(last) < smsMinInterval {
		return "", ErrRateLimited
	}
	byEmailHour, _, err := s.st.CountRecentEmail(ctx, e, clientIP, time.Hour)
	if err != nil {
		return "", err
	}
	byEmailDay, byIPDay, err := s.st.CountRecentEmail(ctx, e, clientIP, 24*time.Hour)
	if err != nil {
		return "", err
	}
	if byEmailHour >= smsMaxPerHour || byEmailDay >= smsMaxPerDay || byIPDay >= smsMaxPerDayIP {
		return "", ErrRateLimited
	}
	code, err := RandomCode(6)
	if err != nil {
		return "", err
	}
	if err := s.st.InsertEmailCode(ctx, e, store.HashCode(code), purpose, clientIP, s.cfg.CodeTTL); err != nil {
		return "", err
	}
	if err := s.email.Send(ctx, e, code, purpose); err != nil {
		return "", fmt.Errorf("邮件发送失败: %w", err)
	}
	if s.cfg.DevEchoCode && s.email.Name() == "dev" {
		return code, nil
	}
	return "", nil
}

// RegisterEmail 邮箱注册:验证码 + 自设密码 + 邀请码(内测闸门)→ 建户签发。
func (s *Service) RegisterEmail(ctx context.Context, email, code, password, invite, deviceInfo string) (*TokenPair, *store.User, error) {
	e, err := ValidateEmail(email)
	if err != nil {
		return nil, nil, err
	}
	if err := ValidatePassword(password); err != nil {
		return nil, nil, err
	}
	if s.cfg.InviteRequired {
		if invite == "" {
			return nil, nil, store.ErrInviteInvalid
		}
		if err := s.st.CheckInvite(ctx, strings.ToUpper(strings.TrimSpace(invite))); err != nil {
			return nil, nil, err
		}
	}
	if len(code) != 6 {
		return nil, nil, ErrCodeInvalid
	}
	ok, err := s.st.ConsumeEmailCode(ctx, e, store.HashCode(code), "register")
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		_ = s.st.FailEmailAttempt(ctx, e)
		return nil, nil, ErrCodeInvalid
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}
	at := strings.IndexByte(e, '@')
	user, err := s.st.CreateUserByEmail(ctx, e, string(hash), "星友"+e[:min(at, 6)])
	if err != nil {
		return nil, nil, err
	}
	if s.cfg.InviteRequired {
		if err := s.st.ConsumeInvite(ctx, strings.ToUpper(strings.TrimSpace(invite)), user.ID); err != nil {
			return nil, nil, err
		}
	}
	_, _ = s.st.TouchDevice(ctx, user.ID, DeviceHash(deviceInfo))
	pair, err := s.issuePair(ctx, user, deviceInfo, nil)
	if err != nil {
		return nil, nil, err
	}
	return pair, user, nil
}

// LoginEmail 邮箱密码登录。环境检测:strict 模式下陌生设备须验证码升级
// (返回 ErrNeedVerify,客户端引导发 login 用途验证码后带 code 重试)。
func (s *Service) LoginEmail(ctx context.Context, email, password, code, deviceInfo string) (*TokenPair, *store.User, error) {
	e, err := ValidateEmail(email)
	if err != nil {
		return nil, nil, ErrBadCredential
	}
	user, hash, err := s.st.FindUserByEmail(ctx, e)
	if errors.Is(err, store.ErrNotFound) {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password)) // 恒时防枚举
		return nil, nil, ErrBadCredential
	}
	if err != nil {
		return nil, nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return nil, nil, ErrBadCredential
	}
	if user.Status != "active" {
		return nil, nil, errors.New("账号不可用")
	}
	dh := DeviceHash(deviceInfo)
	if s.cfg.DeviceStrict {
		known, err := s.st.KnownDevice(ctx, user.ID, dh)
		if err != nil {
			return nil, nil, err
		}
		if !known {
			if code == "" {
				return nil, nil, ErrNeedVerify
			}
			ok, err := s.st.ConsumeEmailCode(ctx, e, store.HashCode(code), "login")
			if err != nil {
				return nil, nil, err
			}
			if !ok {
				_ = s.st.FailEmailAttempt(ctx, e)
				return nil, nil, ErrCodeInvalid
			}
		}
	}
	_, _ = s.st.TouchDevice(ctx, user.ID, dh)
	pair, err := s.issuePair(ctx, user, deviceInfo, nil)
	if err != nil {
		return nil, nil, err
	}
	return pair, user, nil
}

// ResetPassword 忘记密码:邮箱验证码 + 新密码;成功后全端下线防盗号残留。
func (s *Service) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	e, err := ValidateEmail(email)
	if err != nil {
		return err
	}
	if err := ValidatePassword(newPassword); err != nil {
		return err
	}
	user, _, err := s.st.FindUserByEmail(ctx, e)
	if errors.Is(err, store.ErrNotFound) {
		return ErrCodeInvalid // 不暴露邮箱是否注册
	}
	if err != nil {
		return err
	}
	if len(code) != 6 {
		return ErrCodeInvalid
	}
	ok, err := s.st.ConsumeEmailCode(ctx, e, store.HashCode(code), "reset")
	if err != nil {
		return err
	}
	if !ok {
		_ = s.st.FailEmailAttempt(ctx, e)
		return ErrCodeInvalid
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.st.SetPassword(ctx, user.ID, string(hash)); err != nil {
		return err
	}
	_ = s.st.RevokeAllRefreshTokens(ctx, user.ID)
	if err := s.st.BumpSessionVer(ctx, user.ID); err != nil {
		return err
	}
	s.invalidateSV(user.ID)
	return nil
}

// ChangePassword 登录态改密:旧密码校验 + 新密码;其余端下线。
func (s *Service) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	if err := ValidatePassword(newPassword); err != nil {
		return err
	}
	hash, err := s.st.PasswordHash(ctx, userID)
	if err != nil {
		return err
	}
	if hash == "" || bcrypt.CompareHashAndPassword([]byte(hash), []byte(oldPassword)) != nil {
		return errors.New("原密码不正确")
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.st.SetPassword(ctx, userID, string(newHash)); err != nil {
		return err
	}
	_ = s.st.RevokeAllRefreshTokens(ctx, userID)
	if err := s.st.BumpSessionVer(ctx, userID); err != nil {
		return err
	}
	s.invalidateSV(userID)
	return nil
}

// ConstantTimeEqual 管理令牌恒时比较。
func ConstantTimeEqual(a, b string) bool {
	return len(a) == len(b) && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
