package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ── 商品 ─────────────────────────────────────────────────────

// Product 商品(订阅 或 次卡)。
type Product struct {
	ID                 string  `json:"id"`
	Title              string  `json:"title"`
	Kind               string  `json:"kind"` // subscription | credits
	Tier               *string `json:"tier,omitempty"`
	DurationDays       *int    `json:"durationDays,omitempty"`
	CreditType         *string `json:"creditType,omitempty"`
	CreditAmount       *int    `json:"creditAmount,omitempty"`
	PriceCents         int     `json:"priceCents"`
	OriginalPriceCents *int    `json:"originalPriceCents,omitempty"`
	SortOrder          int     `json:"-"`
}

// ListProducts 上架商品(按 sort_order)。
func (s *Store) ListProducts(ctx context.Context) ([]Product, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, title, kind, tier, duration_days,
		credit_type, credit_amount, price_cents, original_price_cents, sort_order
		FROM products WHERE active ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Title, &p.Kind, &p.Tier, &p.DurationDays,
			&p.CreditType, &p.CreditAmount, &p.PriceCents, &p.OriginalPriceCents, &p.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// GetProduct 单个上架商品。
func (s *Store) GetProduct(ctx context.Context, id string) (*Product, error) {
	var p Product
	err := s.pool.QueryRow(ctx, `SELECT id, title, kind, tier, duration_days,
		credit_type, credit_amount, price_cents, original_price_cents, sort_order
		FROM products WHERE id = $1 AND active`, id).
		Scan(&p.ID, &p.Title, &p.Kind, &p.Tier, &p.DurationDays,
			&p.CreditType, &p.CreditAmount, &p.PriceCents, &p.OriginalPriceCents, &p.SortOrder)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ── 订单 ─────────────────────────────────────────────────────

// Order 订单。
type Order struct {
	ID          string     `json:"id"`
	OrderNo     string     `json:"orderNo"`
	UserID      string     `json:"-"`
	ProductID   string     `json:"productId"`
	AmountCents int        `json:"amountCents"`
	Status      string     `json:"status"`
	Channel     *string    `json:"channel,omitempty"`
	PaidAt      *time.Time `json:"paidAt,omitempty"`
	FulfilledAt *time.Time `json:"fulfilledAt,omitempty"`
	ExpiresAt   time.Time  `json:"expiresAt"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// CreateOrder 创建订单(金额取自商品,不信任客户端)。
func (s *Store) CreateOrder(ctx context.Context, userID string, p *Product, ttl time.Duration) (*Order, error) {
	suffix := make([]byte, 4)
	if _, err := rand.Read(suffix); err != nil {
		return nil, err
	}
	orderNo := fmt.Sprintf("Z%s%s", time.Now().Format("20060102150405"), hex.EncodeToString(suffix))
	var o Order
	err := s.pool.QueryRow(ctx, `INSERT INTO orders (order_no, user_id, product_id, amount_cents, expires_at)
		VALUES ($1, $2, $3, $4, now() + $5::interval)
		RETURNING id, order_no, user_id, product_id, amount_cents, status, channel, paid_at, fulfilled_at, expires_at, created_at`,
		orderNo, userID, p.ID, p.PriceCents, ttl).
		Scan(&o.ID, &o.OrderNo, &o.UserID, &o.ProductID, &o.AmountCents, &o.Status,
			&o.Channel, &o.PaidAt, &o.FulfilledAt, &o.ExpiresAt, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// GetOrder 按 id 取订单(校验归属)。
func (s *Store) GetOrder(ctx context.Context, orderID, userID string) (*Order, error) {
	var o Order
	err := s.pool.QueryRow(ctx, `SELECT id, order_no, user_id, product_id, amount_cents, status,
		channel, paid_at, fulfilled_at, expires_at, created_at
		FROM orders WHERE id = $1 AND user_id = $2`, orderID, userID).
		Scan(&o.ID, &o.OrderNo, &o.UserID, &o.ProductID, &o.AmountCents, &o.Status,
			&o.Channel, &o.PaidAt, &o.FulfilledAt, &o.ExpiresAt, &o.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// ListOrders 用户订单列表(近 50 条)。
func (s *Store) ListOrders(ctx context.Context, userID string) ([]Order, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, order_no, user_id, product_id, amount_cents, status,
		channel, paid_at, fulfilled_at, expires_at, created_at
		FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.OrderNo, &o.UserID, &o.ProductID, &o.AmountCents, &o.Status,
			&o.Channel, &o.PaidAt, &o.FulfilledAt, &o.ExpiresAt, &o.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// ErrOrderState 订单状态不满足操作前置条件。
var ErrOrderState = errors.New("订单状态不允许该操作")

// MarkOrderPaid 渠道回调:created/paying → paid(带渠道交易号,重复回调幂等)。
func (s *Store) MarkOrderPaid(ctx context.Context, orderID, channel, channelTxnID string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE orders
		SET status = 'paid', channel = $2, channel_txn_id = $3, paid_at = now()
		WHERE id = $1 AND status IN ('created', 'paying')`, orderID, channel, channelTxnID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		// 已是 paid/fulfilled(重复回调)不算错;其他状态报错
		var status string
		if err := s.pool.QueryRow(ctx, "SELECT status FROM orders WHERE id = $1", orderID).Scan(&status); err != nil {
			return err
		}
		if status == "paid" || status == "fulfilled" {
			return nil
		}
		return fmt.Errorf("%w: %s", ErrOrderState, status)
	}
	return nil
}

// FulfillOrder 履约(单事务):paid → fulfilled + 发放权益/次数 + 刷新用户层级缓存。
// 幂等:重复调用因状态前置条件与唯一约束而安全。
func (s *Store) FulfillOrder(ctx context.Context, orderID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.WithoutCancel(ctx)) //nolint:errcheck

	// 行锁订单,避免并发双履约
	var userID, productID, status string
	err = tx.QueryRow(ctx, `SELECT user_id, product_id, status FROM orders
		WHERE id = $1 FOR UPDATE`, orderID).Scan(&userID, &productID, &status)
	if err != nil {
		return err
	}
	if status == "fulfilled" {
		return nil // 幂等
	}
	if status != "paid" {
		return fmt.Errorf("%w: %s", ErrOrderState, status)
	}

	var kind string
	var tier *string
	var durationDays, creditAmount *int
	var creditType *string
	err = tx.QueryRow(ctx, `SELECT kind, tier, duration_days, credit_type, credit_amount
		FROM products WHERE id = $1`, productID).
		Scan(&kind, &tier, &durationDays, &creditType, &creditAmount)
	if err != nil {
		return err
	}

	switch kind {
	case "subscription":
		// 顺延:起点 = max(now, 当前层级到期);跨层级购买按新层级即时生效
		var startsAt time.Time
		err = tx.QueryRow(ctx, `SELECT greatest(now(),
			coalesce((SELECT max(ends_at) FROM entitlements WHERE user_id = $1 AND tier = $2 AND ends_at > now()), now()))`,
			userID, *tier).Scan(&startsAt)
		if err != nil {
			return err
		}
		endsAt := startsAt.AddDate(0, 0, *durationDays)
		if _, err := tx.Exec(ctx, `INSERT INTO entitlements (user_id, order_id, tier, starts_at, ends_at)
			VALUES ($1, $2, $3, $4, $5)`, userID, orderID, *tier, startsAt, endsAt); err != nil {
			return err
		}
		// 刷新 users.tier 缓存(取当前最高有效层级:master > pro)
		if _, err := tx.Exec(ctx, `UPDATE users SET
			tier = coalesce((SELECT tier FROM entitlements WHERE user_id = $1 AND ends_at > now()
				ORDER BY CASE tier WHEN 'master' THEN 0 ELSE 1 END, ends_at DESC LIMIT 1), 'free'),
			tier_expires_at = (SELECT max(ends_at) FROM entitlements WHERE user_id = $1 AND ends_at > now())
			WHERE id = $1`, userID); err != nil {
			return err
		}
	case "credits":
		if _, err := tx.Exec(ctx, `INSERT INTO credit_ledger (user_id, credit_type, delta, reason, order_id)
			VALUES ($1, $2, $3, 'purchase', $4)`, userID, *creditType, *creditAmount, orderID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO credit_balances (user_id, credit_type, balance)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id, credit_type)
			DO UPDATE SET balance = credit_balances.balance + $3, updated_at = now()`,
			userID, *creditType, *creditAmount); err != nil {
			return err
		}
	default:
		return fmt.Errorf("未知商品类型: %s", kind)
	}

	if _, err := tx.Exec(ctx, `UPDATE orders SET status = 'fulfilled', fulfilled_at = now()
		WHERE id = $1`, orderID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ── 次数余额与消费 ───────────────────────────────────────────

// ErrNoCredits 余额不足。
var ErrNoCredits = errors.New("次数不足,请先购买")

// CreditBalances 用户全部次数余额。
func (s *Store) CreditBalances(ctx context.Context, userID string) (map[string]int, error) {
	rows, err := s.pool.Query(ctx, `SELECT credit_type, balance FROM credit_balances WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var t string
		var b int
		if err := rows.Scan(&t, &b); err != nil {
			return nil, err
		}
		out[t] = b
	}
	return out, rows.Err()
}

// ConsumeCredit 原子扣减 1 次并记账;余额不足返回 ErrNoCredits。
func (s *Store) ConsumeCredit(ctx context.Context, userID, creditType, ref string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.WithoutCancel(ctx)) //nolint:errcheck

	tag, err := tx.Exec(ctx, `UPDATE credit_balances SET balance = balance - 1, updated_at = now()
		WHERE user_id = $1 AND credit_type = $2 AND balance >= 1`, userID, creditType)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNoCredits
	}
	if _, err := tx.Exec(ctx, `INSERT INTO credit_ledger (user_id, credit_type, delta, reason, ref)
		VALUES ($1, $2, -1, 'consume:' || $2, $3)`, userID, creditType, ref); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// RefundCredit 补偿退还 1 次(如 AI 生成失败)。
func (s *Store) RefundCredit(ctx context.Context, userID, creditType, ref string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.WithoutCancel(ctx)) //nolint:errcheck
	if _, err := tx.Exec(ctx, `UPDATE credit_balances SET balance = balance + 1, updated_at = now()
		WHERE user_id = $1 AND credit_type = $2`, userID, creditType); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO credit_ledger (user_id, credit_type, delta, reason, ref)
		VALUES ($1, $2, 1, 'refund:' || $2, $3)`, userID, creditType, ref); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ── 权益视图 ─────────────────────────────────────────────────

// Entitlement 权益记录。
type Entitlement struct {
	Tier     string    `json:"tier"`
	StartsAt time.Time `json:"startsAt"`
	EndsAt   time.Time `json:"endsAt"`
}

// ActiveEntitlements 用户当前有效权益。
func (s *Store) ActiveEntitlements(ctx context.Context, userID string) ([]Entitlement, error) {
	rows, err := s.pool.Query(ctx, `SELECT tier, starts_at, ends_at FROM entitlements
		WHERE user_id = $1 AND ends_at > now() ORDER BY ends_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Entitlement
	for rows.Next() {
		var e Entitlement
		if err := rows.Scan(&e.Tier, &e.StartsAt, &e.EndsAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// CloseExpiredOrders 关闭超时未支付订单(由后台任务周期调用)。
func (s *Store) CloseExpiredOrders(ctx context.Context) (int64, error) {
	tag, err := s.pool.Exec(ctx, `UPDATE orders SET status = 'closed'
		WHERE status IN ('created', 'paying') AND expires_at < now()`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
