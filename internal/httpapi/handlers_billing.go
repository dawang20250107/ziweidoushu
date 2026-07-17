package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/dawang20250107/ziweidoushu/internal/store"
)

// ── 商品与订单 ────────────────────────────────────────────────

func (s *Server) handleProducts(w http.ResponseWriter, r *http.Request) {
	if !s.guardAuth(w) {
		return
	}
	products, err := s.store.ListProducts(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "products_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"products":      products,
		"devPayEnabled": s.cfg.PayDevEnabled,
	})
}

func (s *Server) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	var req struct {
		ProductID string `json:"productId"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	product, err := s.store.GetProduct(r.Context(), req.ProductID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "商品不存在或已下架")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "order_failed", err.Error())
		return
	}
	order, err := s.store.CreateOrder(r.Context(), claims.Sub, product, 2*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "order_failed", err.Error())
		return
	}
	// 支付参数:微信/支付宝渠道接入后在此返回 prepay 参数;当前仅 dev 渠道
	writeJSON(w, http.StatusOK, map[string]any{
		"order":         order,
		"devPayEnabled": s.cfg.PayDevEnabled,
	})
}

func (s *Server) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	order, err := s.store.GetOrder(r.Context(), r.PathValue("id"), claims.Sub)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "订单不存在")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "order_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"order": order})
}

func (s *Server) handleListOrders(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	orders, err := s.store.ListOrders(r.Context(), claims.Sub)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "orders_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"orders": orders})
}

// handleDevPay dev 支付渠道:模拟渠道回调,标记支付成功并立即履约。
// 仅 PAY_DEV_ENABLED=1 时可用(本地/E2E;接入微信/支付宝后生产禁开)。
func (s *Server) handleDevPay(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.PayDevEnabled {
		writeError(w, http.StatusForbidden, "dev_pay_disabled", "dev 支付渠道未启用")
		return
	}
	claims := currentClaims(r.Context())
	orderID := r.PathValue("id")
	order, err := s.store.GetOrder(r.Context(), orderID, claims.Sub)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "订单不存在")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "pay_failed", err.Error())
		return
	}
	if time.Now().After(order.ExpiresAt) && order.Status == "created" {
		writeError(w, http.StatusBadRequest, "order_expired", "订单已超时,请重新下单")
		return
	}
	if err := s.store.MarkOrderPaid(r.Context(), orderID, "dev", "dev-"+order.OrderNo); err != nil {
		writeError(w, http.StatusBadRequest, "pay_failed", err.Error())
		return
	}
	if err := s.store.FulfillOrder(r.Context(), orderID); err != nil {
		s.logger.Error("履约失败(订单已支付,待补偿)", "order", orderID, "err", err)
		writeError(w, http.StatusInternalServerError, "fulfill_failed", "支付成功但权益发放异常,请联系客服")
		return
	}
	order, _ = s.store.GetOrder(r.Context(), orderID, claims.Sub)
	writeJSON(w, http.StatusOK, map[string]any{"order": order})
}

// ── 我的权益与次数 ────────────────────────────────────────────

func (s *Server) handleMyEntitlements(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	ents, err := s.store.ActiveEntitlements(r.Context(), claims.Sub)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "entitlements_failed", err.Error())
		return
	}
	credits, err := s.store.CreditBalances(r.Context(), claims.Sub)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "credits_failed", err.Error())
		return
	}
	user, err := s.store.GetUser(r.Context(), claims.Sub)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "user_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"tier":         user.Tier,
		"entitlements": ents,
		"credits":      credits,
	})
}

// ── 深度报告(按次付费消费点)────────────────────────────────

// handleDeepReport 生成深度主题报告:消耗 1 次 deep_report;AI 失败自动退还。
func (s *Server) handleDeepReport(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	var req struct {
		chartRequest
		Topic string `json:"topic"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Topic == "" {
		req.Topic = "overview"
	}
	// 付费报告必须走真实 LLM:降级模式不收钱、不扣次数
	if !s.interp.HasProvider() {
		writeError(w, http.StatusServiceUnavailable, "ai_unavailable", "AI 服务暂不可用,深度报告未扣次数")
		return
	}
	resp, err := s.computeChart(req.chartRequest)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_birth", err.Error())
		return
	}

	// 先扣次数(原子),AI 失败再补偿退还
	if err := s.store.ConsumeCredit(r.Context(), claims.Sub, "deep_report", req.Topic); err != nil {
		if errors.Is(err, store.ErrNoCredits) {
			writeError(w, http.StatusPaymentRequired, "no_credits", "深度报告次数不足,请先购买")
			return
		}
		writeError(w, http.StatusInternalServerError, "consume_failed", err.Error())
		return
	}

	s.metrics.aiRequests.Add(1)
	question := "请生成一份完整的深度报告:结构化分层论述,先总论后分项,每一项给出命盘依据(宫位/星曜/四化/格局)与可操作建议,篇幅充分。"
	result, err := s.interp.Interpret(r.Context(), resp.Chart, resp.Patterns, req.Topic, question, nil)
	if err != nil {
		s.metrics.aiErrors.Add(1)
		// 生成失败退还次数;请求可能已断开,退款用不可取消上下文
		if rerr := s.store.RefundCredit(context.WithoutCancel(r.Context()), claims.Sub, "deep_report", "refund:"+req.Topic); rerr != nil {
			s.logger.Error("次数退还失败,需人工补偿", "user", claims.Sub, "err", rerr)
		}
		writeAIError(w, err)
		return
	}
	credits, _ := s.store.CreditBalances(r.Context(), claims.Sub)
	writeJSON(w, http.StatusOK, map[string]any{
		"report":           result,
		"topic":            req.Topic,
		"remainingCredits": credits["deep_report"],
	})
}
