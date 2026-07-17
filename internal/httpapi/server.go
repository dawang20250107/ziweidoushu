// Package httpapi HTTP API 层:路由、并发中间件、SSE 流式输出。
//
// 高并发设计:
//   - 服务完全无状态(缓存/限流均为进程内加速,不影响正确性),可水平扩容;
//   - 排盘为纯 CPU 计算(微秒级)+ 分片 LRU 缓存;
//   - 语料检索基于不可变快照,读路径零锁竞争;
//   - AI 上游调用有信号量限流与超时,SSE 流式回传;
//   - 单 IP 令牌桶限流、请求体大小限制、panic 兜底、优雅停机。
package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/dawang20250107/ziweidoushu/internal/ai"
	"github.com/dawang20250107/ziweidoushu/internal/auth"
	"github.com/dawang20250107/ziweidoushu/internal/config"
	"github.com/dawang20250107/ziweidoushu/internal/corpus"
	"github.com/dawang20250107/ziweidoushu/internal/knowledge"
	"github.com/dawang20250107/ziweidoushu/internal/store"
)

// Version 服务版本(构建时可用 -ldflags 覆盖)。
var Version = "2.0.0"

// Server API 服务。
type Server struct {
	cfg     config.Config
	logger  *slog.Logger
	corpus  *corpus.Store
	kb      *knowledge.Base
	interp  *ai.Interpreter
	auth    *auth.Service // nil = 用户体系未启用(无库模式)
	store   *store.Store  // nil = 无库模式
	cache   *lruCache
	limiter *ipLimiter
	metrics *metrics
	http    *http.Server
}

// Deps 可选依赖(无库模式下 Auth/Store 为 nil)。
type Deps struct {
	Auth  *auth.Service
	Store *store.Store
}

// New 组装服务。
func New(cfg config.Config, logger *slog.Logger, corpusStore *corpus.Store, kb *knowledge.Base, interp *ai.Interpreter, deps Deps) *Server {
	s := &Server{
		cfg:     cfg,
		logger:  logger,
		corpus:  corpusStore,
		kb:      kb,
		interp:  interp,
		auth:    deps.Auth,
		store:   deps.Store,
		cache:   newLRUCache(cfg.ChartCacheSize),
		limiter: newIPLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst),
		metrics: newMetrics(),
	}

	mux := http.NewServeMux()

	// 探针与指标
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /readyz", s.handleReady)
	mux.HandleFunc("GET /metrics", s.metrics.handler())

	// 排盘
	mux.HandleFunc("POST /api/v1/chart", s.handleChart)
	mux.HandleFunc("POST /api/v1/horoscope", s.handleHoroscope)
	mux.HandleFunc("GET /api/v1/famous", s.handleFamousList)
	mux.HandleFunc("GET /api/v1/famous/{id}/chart", s.handleFamousChart)

	// 四化
	mux.HandleFunc("GET /api/v1/sihua/liunian", s.handleLiuNianSiHua)
	mux.HandleFunc("GET /api/v1/sihua/liuyue", s.handleLiuYueSiHua)

	// 古籍
	mux.HandleFunc("GET /api/v1/books", s.handleBooks)
	mux.HandleFunc("GET /api/v1/books/{slug}", s.handleBook)
	mux.HandleFunc("GET /api/v1/books/{slug}/chapters/{idx}", s.handleChapter)
	mux.HandleFunc("GET /api/v1/search", s.handleSearch)
	mux.HandleFunc("POST /api/v1/admin/corpus/reload", s.handleCorpusReload)

	// 倪海厦知识库
	mux.HandleFunc("GET /api/v1/nihai/{section}", s.handleNihai)
	mux.HandleFunc("GET /api/v1/knowledge/stars", s.handleKnowledgeStars)
	mux.HandleFunc("GET /api/v1/knowledge/topics", s.handleKnowledgeTopics)
	mux.HandleFunc("GET /api/v1/knowledge/heming", s.handleKnowledgeHeming)
	mux.HandleFunc("GET /api/v1/cities", s.handleCities)

	// AI 解读与合盘
	mux.HandleFunc("POST /api/v1/ai/interpret", s.handleInterpret)
	mux.HandleFunc("POST /api/v1/heming", s.handleHeming)

	// 用户体系(未配置 DATABASE_URL/JWT_SECRET 时统一 503)
	mux.HandleFunc("POST /api/v1/auth/sms/send", s.handleSMSSend)
	mux.HandleFunc("POST /api/v1/auth/sms/verify", s.handleSMSVerify)
	mux.HandleFunc("POST /api/v1/auth/refresh", s.handleRefresh)
	mux.HandleFunc("POST /api/v1/auth/logout", s.requireAuth(s.handleLogout))
	mux.HandleFunc("GET /api/v1/me", s.requireAuth(s.handleMe))
	mux.HandleFunc("GET /api/v1/me/entitlements", s.requireAuth(s.handleMyEntitlements))

	// 变现:商品(订阅 + 次卡)/ 订单 / dev 支付渠道
	mux.HandleFunc("GET /api/v1/products", s.handleProducts)
	mux.HandleFunc("POST /api/v1/orders", s.requireAuth(s.handleCreateOrder))
	mux.HandleFunc("GET /api/v1/orders", s.requireAuth(s.handleListOrders))
	mux.HandleFunc("GET /api/v1/orders/{id}", s.requireAuth(s.handleGetOrder))
	mux.HandleFunc("POST /api/v1/orders/{id}/dev-pay", s.requireAuth(s.handleDevPay))

	// 深度报告(按次付费消费点)
	mux.HandleFunc("POST /api/v1/ai/report", s.requireAuth(s.handleDeepReport))

	// 命盘档案库
	mux.HandleFunc("POST /api/v1/profiles", s.requireAuth(s.handleCreateProfile))
	mux.HandleFunc("GET /api/v1/profiles", s.requireAuth(s.handleListProfiles))
	mux.HandleFunc("GET /api/v1/profiles/{id}", s.requireAuth(s.handleGetProfile))
	mux.HandleFunc("DELETE /api/v1/profiles/{id}", s.requireAuth(s.handleDeleteProfile))
	mux.HandleFunc("POST /api/v1/profiles/{id}/default", s.requireAuth(s.handleSetDefaultProfile))

	// 元信息
	mux.HandleFunc("GET /api/v1/meta", s.handleMeta)

	handler := chain(mux,
		s.withRecover,
		withRequestID,
		s.metrics.withInFlight,
		s.withLogging,
		s.withCORS,
		s.withRateLimit,
		withMaxBody(1<<20), // 1 MiB
	)

	s.http = &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		// WriteTimeout 需容纳 AI SSE 长流,取 AI 超时 + 余量。
		WriteTimeout: cfg.AI.Timeout + 30*time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return s
}

// Run 启动并阻塞直到 ctx 取消,随后优雅停机。
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("服务启动", "addr", s.cfg.Addr, "version", Version, "ai", s.interp.ProviderName())
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	s.logger.Info("收到停机信号,优雅退出中")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()
	s.limiter.close()
	return s.http.Shutdown(shutdownCtx)
}

func (s *Server) handleReady(w http.ResponseWriter, _ *http.Request) {
	if s.corpus == nil || s.kb == nil {
		writeError(w, http.StatusServiceUnavailable, "not_ready", "数据未加载完成")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ready": true})
}
