// Package config 服务配置(全部环境变量驱动,便于容器化与横向扩容)。
package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dawang20250107/ziweidoushu/internal/ai"
)

// Config 服务配置。
type Config struct {
	// Addr 监听地址,默认 ":8080"。
	Addr string
	// CorpusExternalDir 外部古籍目录(后期投放古籍资料/倪师著作),空则只用内置数据。
	CorpusExternalDir string

	// RateLimitRPS 单 IP 每秒请求配额,<=0 关闭限流。
	RateLimitRPS float64
	// RateLimitBurst 单 IP 突发配额。
	RateLimitBurst int

	// CORSAllowOrigins 允许的跨域来源,"*" 表示全部(小程序侧通常走服务端转发,默认放开)。
	CORSAllowOrigins []string

	// RequestTimeout 常规接口超时(AI 流式接口单独用 AI_TIMEOUT_SECONDS)。
	RequestTimeout time.Duration
	// ShutdownTimeout 优雅停机等待时长。
	ShutdownTimeout time.Duration

	// ChartCacheSize 命盘结果 LRU 缓存条数,<=0 关闭缓存。
	ChartCacheSize int

	// AdminToken 管理接口令牌(语料热加载);空则禁用管理接口。
	AdminToken string

	// DatabaseURL PostgreSQL 连接串;空则以无库模式运行(用户体系不可用)。
	DatabaseURL string
	// JWTSecret 访问令牌签名密钥(≥32 字节);JWTPrevSecret 轮换期兼容旧密钥。
	JWTSecret     string
	JWTPrevSecret string
	// SMSDevEchoCode dev 短信通道下把验证码回显到接口(仅本地/E2E,生产禁开)。
	SMSDevEchoCode bool
	// SMTP 邮件通道(邮箱验证码):Host 为空则走 dev 通道(验证码写日志)。
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
	// InviteRequired 内测闸门:注册须携带有效邀请码(默认开;公测置 0 关闭)。
	InviteRequired bool
	// DeviceStrict 环境检测:陌生设备密码登录须邮箱验证码升级(默认关)。
	DeviceStrict bool
	// PayDevEnabled dev 支付渠道开关(模拟支付回调即时履约;仅本地/E2E,生产禁开)。
	PayDevEnabled bool

	AI ai.Config
}

// FromEnv 读取环境变量构建配置。
func FromEnv() Config {
	return Config{
		Addr:              envOr("LISTEN_ADDR", ":8080"),
		CorpusExternalDir: os.Getenv("CORPUS_EXTERNAL_DIR"),
		RateLimitRPS:      envFloat("RATE_LIMIT_RPS", 20),
		RateLimitBurst:    envInt("RATE_LIMIT_BURST", 40),
		CORSAllowOrigins:  splitCSV(envOr("CORS_ALLOW_ORIGINS", "*")),
		RequestTimeout:    time.Duration(envInt("REQUEST_TIMEOUT_SECONDS", 15)) * time.Second,
		ShutdownTimeout:   time.Duration(envInt("SHUTDOWN_TIMEOUT_SECONDS", 20)) * time.Second,
		ChartCacheSize:    envInt("CHART_CACHE_SIZE", 4096),
		AdminToken:        os.Getenv("ADMIN_TOKEN"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		JWTPrevSecret:     os.Getenv("JWT_SECRET_PREV"),
		SMSDevEchoCode:    os.Getenv("SMS_DEV_ECHO_CODE") == "1",
		SMTPHost:          os.Getenv("SMTP_HOST"),
		SMTPPort:          envInt("SMTP_PORT", 465),
		SMTPUsername:      os.Getenv("SMTP_USERNAME"),
		SMTPPassword:      os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:          envOr("SMTP_FROM", os.Getenv("SMTP_USERNAME")),
		InviteRequired:    envOr("AUTH_INVITE_REQUIRED", "1") == "1",
		DeviceStrict:      os.Getenv("AUTH_DEVICE_STRICT") == "1",
		PayDevEnabled:     os.Getenv("PAY_DEV_ENABLED") == "1",
		AI:                ai.ConfigFromEnv(),
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
