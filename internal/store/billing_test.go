package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

// 集成测试:需要 TEST_DATABASE_URL 指向可用的 PostgreSQL(会自动应用迁移)。

func newTestStore(t *testing.T) *Store {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("未设置 TEST_DATABASE_URL,跳过数据库集成测试")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	st, err := Open(ctx, url)
	if err != nil {
		t.Fatalf("连接测试库失败: %v", err)
	}
	t.Cleanup(st.Close)
	return st
}

// newTestUser 造一个独立测试用户(手机号按纳秒唯一)。
func newTestUser(t *testing.T, st *Store) *User {
	t.Helper()
	phone := fmt.Sprintf("19%09d", time.Now().UnixNano()%1_000_000_000)
	user, _, err := st.FindOrCreateUserByPhone(context.Background(), phone)
	if err != nil {
		t.Fatalf("建测试用户失败: %v", err)
	}
	return user
}

// buyAndFulfill 下单 → dev 支付 → 履约。
func buyAndFulfill(t *testing.T, st *Store, userID, productID string) *Order {
	t.Helper()
	ctx := context.Background()
	p, err := st.GetProduct(ctx, productID)
	if err != nil {
		t.Fatalf("取商品 %s 失败: %v", productID, err)
	}
	order, err := st.CreateOrder(ctx, userID, p, time.Hour)
	if err != nil {
		t.Fatalf("下单失败: %v", err)
	}
	if order.AmountCents != p.PriceCents {
		t.Fatalf("订单金额应取自商品: got %d want %d", order.AmountCents, p.PriceCents)
	}
	if err := st.MarkOrderPaid(ctx, order.ID, "dev", "txn-"+order.OrderNo); err != nil {
		t.Fatalf("标记支付失败: %v", err)
	}
	if err := st.FulfillOrder(ctx, order.ID); err != nil {
		t.Fatalf("履约失败: %v", err)
	}
	return order
}

// TestCreditsFulfillIdempotent 次卡履约:重复回调/重复履约只发一次。
func TestCreditsFulfillIdempotent(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	user := newTestUser(t, st)

	order := buyAndFulfill(t, st, user.ID, "report_5")

	// 渠道重复回调 + 重复履约:都应幂等无害
	if err := st.MarkOrderPaid(ctx, order.ID, "dev", "txn-dup"); err != nil {
		t.Fatalf("重复支付回调应幂等: %v", err)
	}
	if err := st.FulfillOrder(ctx, order.ID); err != nil {
		t.Fatalf("重复履约应幂等: %v", err)
	}

	credits, err := st.CreditBalances(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if credits["deep_report"] != 5 {
		t.Fatalf("5 次卡重复履约后余额应为 5, got %d", credits["deep_report"])
	}

	// 未支付订单不可履约
	p, _ := st.GetProduct(ctx, "report_1")
	fresh, err := st.CreateOrder(ctx, user.ID, p, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.FulfillOrder(ctx, fresh.ID); !errors.Is(err, ErrOrderState) {
		t.Fatalf("未支付订单履约应报 ErrOrderState, got %v", err)
	}
}

// TestConsumeCreditConcurrent 并发消费:余额恰好扣完、永不为负、账本对账平。
func TestConsumeCreditConcurrent(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	user := newTestUser(t, st)

	buyAndFulfill(t, st, user.ID, "report_5") // 5 次
	buyAndFulfill(t, st, user.ID, "report_5") // 再 5 次 → 共 10

	const workers = 25 // 超发 15,应恰好 10 成功
	var ok, noCredits int64
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			err := st.ConsumeCredit(ctx, user.ID, "deep_report", fmt.Sprintf("race-%d", n))
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				ok++
			case errors.Is(err, ErrNoCredits):
				noCredits++
			default:
				t.Errorf("意外错误: %v", err)
			}
		}(i)
	}
	wg.Wait()

	if ok != 10 || noCredits != 15 {
		t.Fatalf("并发扣减应恰好 10 成功 15 拒绝, got ok=%d noCredits=%d", ok, noCredits)
	}
	credits, err := st.CreditBalances(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if credits["deep_report"] != 0 {
		t.Fatalf("扣完后余额应为 0, got %d", credits["deep_report"])
	}

	// 退还补偿后余额回升
	if err := st.RefundCredit(ctx, user.ID, "deep_report", "refund:race-0"); err != nil {
		t.Fatal(err)
	}
	credits, _ = st.CreditBalances(ctx, user.ID)
	if credits["deep_report"] != 1 {
		t.Fatalf("退还后余额应为 1, got %d", credits["deep_report"])
	}

	// 账本对账:sum(delta) == 余额
	var sum int
	if err := st.pool.QueryRow(ctx, `SELECT coalesce(sum(delta), 0) FROM credit_ledger
		WHERE user_id = $1 AND credit_type = 'deep_report'`, user.ID).Scan(&sum); err != nil {
		t.Fatal(err)
	}
	if sum != 1 {
		t.Fatalf("账本流水和应等于余额 1, got %d", sum)
	}
}

// TestSubscriptionExtend 订阅续费顺延 + users.tier 缓存刷新。
func TestSubscriptionExtend(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	user := newTestUser(t, st)

	buyAndFulfill(t, st, user.ID, "pro_1m")
	u, err := st.GetUser(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if u.Tier != "pro" {
		t.Fatalf("履约后层级应为 pro, got %s", u.Tier)
	}

	ents, err := st.ActiveEntitlements(ctx, user.ID)
	if err != nil || len(ents) != 1 {
		t.Fatalf("应有 1 条有效权益: %v %d", err, len(ents))
	}
	firstEnd := ents[0].EndsAt

	// 续费:第二段应从第一段到期顺延,而不是从现在起算(用户不吃亏)
	buyAndFulfill(t, st, user.ID, "pro_1m")
	ents, err = st.ActiveEntitlements(ctx, user.ID)
	if err != nil || len(ents) != 2 {
		t.Fatalf("应有 2 条有效权益: %v %d", err, len(ents))
	}
	var second *Entitlement
	for i := range ents {
		if !ents[i].EndsAt.Equal(firstEnd) {
			second = &ents[i]
		}
	}
	if second == nil {
		t.Fatal("未找到第二段权益")
	}
	if delta := second.StartsAt.Sub(firstEnd); delta < -time.Second || delta > time.Second {
		t.Fatalf("续费应从上段到期顺延: 第二段起点 %v, 上段终点 %v", second.StartsAt, firstEnd)
	}

	u, _ = st.GetUser(ctx, user.ID)
	if u.Tier != "pro" || u.TierExpiresAt == nil {
		t.Fatalf("层级缓存异常: tier=%s expires=%v", u.Tier, u.TierExpiresAt)
	}
	if delta := u.TierExpiresAt.Sub(second.EndsAt); delta < -time.Second || delta > time.Second {
		t.Fatalf("tier_expires_at 应为最晚到期: got %v want %v", u.TierExpiresAt, second.EndsAt)
	}
}
