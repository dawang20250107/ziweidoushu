// ziweidoushu API 服务入口。
//
// 基于倪海厦《天纪》体系的紫微斗数排盘 + 古籍查阅 + AI 解读平台(纯后端,API-first,
// 供 Web/小程序等多端接入)。
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/dawang20250107/ziweidoushu/data"
	"github.com/dawang20250107/ziweidoushu/internal/ai"
	"github.com/dawang20250107/ziweidoushu/internal/auth"
	"github.com/dawang20250107/ziweidoushu/internal/config"
	"github.com/dawang20250107/ziweidoushu/internal/corpus"
	"github.com/dawang20250107/ziweidoushu/internal/httpapi"
	"github.com/dawang20250107/ziweidoushu/internal/knowledge"
	dbstore "github.com/dawang20250107/ziweidoushu/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := config.FromEnv()

	// ── 数据加载(内置古籍 + 知识库)──
	store, err := corpus.NewStore(data.FS, "classics")
	if err != nil {
		logger.Error("古籍加载失败", "err", err)
		os.Exit(1)
	}
	if cfg.CorpusExternalDir != "" {
		n, err := store.LoadExternalDir(cfg.CorpusExternalDir)
		if err != nil {
			logger.Error("外部古籍目录加载失败", "dir", cfg.CorpusExternalDir, "err", err)
			os.Exit(1)
		}
		logger.Info("外部古籍加载完成", "count", n, "dir", cfg.CorpusExternalDir)
	}
	kb, err := knowledge.Load(data.FS)
	if err != nil {
		logger.Error("知识库加载失败", "err", err)
		os.Exit(1)
	}

	// ── AI 供应商 ──
	provider, err := ai.NewProvider(cfg.AI)
	if err != nil {
		logger.Error("AI 供应商配置错误", "err", err)
		os.Exit(1)
	}
	interp := ai.NewInterpreter(provider, kb, store, cfg.AI)

	// ── 启动服务(SIGINT/SIGTERM 优雅停机)──
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// ── 用户体系(可选:需 DATABASE_URL + JWT_SECRET)──
	var deps httpapi.Deps
	switch {
	case cfg.DatabaseURL == "":
		logger.Info("无库模式:用户体系未启用(未配置 DATABASE_URL)")
	case cfg.JWTSecret == "":
		logger.Error("已配置 DATABASE_URL 但缺少 JWT_SECRET(≥32 字节)")
		os.Exit(1)
	default:
		st, err := dbstore.Open(ctx, cfg.DatabaseURL)
		if err != nil {
			logger.Error("数据库初始化失败", "err", err)
			os.Exit(1)
		}
		defer st.Close()
		authSvc, err := auth.NewService(st, &auth.DevSMS{Logger: logger}, auth.Config{
			JWTSecret:     cfg.JWTSecret,
			JWTPrevSecret: cfg.JWTPrevSecret,
			DevEchoCode:   cfg.SMSDevEchoCode,
		}, logger)
		if err != nil {
			logger.Error("鉴权服务初始化失败", "err", err)
			os.Exit(1)
		}
		deps = httpapi.Deps{Auth: authSvc, Store: st}
		logger.Info("用户体系已启用", "sms", "dev(接入云厂商前不真实发送)")
	}

	srv := httpapi.New(cfg, logger, store, kb, interp, deps)
	if err := srv.Run(ctx); err != nil {
		logger.Error("服务异常退出", "err", err)
		os.Exit(1)
	}
	logger.Info("服务已停止")
}
