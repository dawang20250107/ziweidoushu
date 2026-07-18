package auth

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/dawang20250107/ziweidoushu/internal/store"
)

// 集成测试:需要 TEST_DATABASE_URL 指向可用的 PostgreSQL(会自动应用迁移)。
// 本地运行示例见 Makefile 的 test-db 目标。

func newTestService(t *testing.T) (*Service, *store.Store) {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("未设置 TEST_DATABASE_URL,跳过数据库集成测试")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	st, err := store.Open(ctx, url)
	if err != nil {
		t.Fatalf("连接测试库失败: %v", err)
	}
	t.Cleanup(st.Close)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc, err := NewService(st, &DevSMS{Logger: logger}, Config{
		JWTSecret:   testSecret,
		DevEchoCode: true,
		AccessTTL:   time.Minute,
	}, logger)
	if err != nil {
		t.Fatal(err)
	}
	return svc, st
}

func TestLoginFlow(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	phone := "138" + time.Now().Format("040515")[0:6] + "00" // 避免多次运行撞频控

	// 1) 发码(dev 回显)
	code, err := svc.SendCode(ctx, phone, "203.0.113.9")
	if err != nil {
		t.Fatalf("发码失败: %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("dev 回显码异常: %q", code)
	}

	// 60s 间隔频控
	if _, err := svc.SendCode(ctx, phone, "203.0.113.9"); err != ErrRateLimited {
		t.Fatalf("60s 内重发应触发频控, got %v", err)
	}

	// 2) 错误码拒绝
	if _, _, _, err := svc.VerifyCode(ctx, phone, "000000", "test-ua"); err != ErrCodeInvalid {
		t.Fatalf("错误码应拒绝, got %v", err)
	}

	// 3) 正确码登录(新用户)
	pair, user, created, err := svc.VerifyCode(ctx, phone, code, "test-ua")
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	if !created || user.Tier != "free" {
		t.Fatalf("新用户创建异常: created=%v tier=%s", created, user.Tier)
	}

	// 4) 验证码一次性:同码重放拒绝
	if _, _, _, err := svc.VerifyCode(ctx, phone, code, "test-ua"); err != ErrCodeInvalid {
		t.Fatalf("验证码重放应拒绝, got %v", err)
	}

	// 5) access 校验
	claims, err := svc.Authenticate(ctx, pair.Access)
	if err != nil || claims.Sub != user.ID {
		t.Fatalf("access 校验失败: %v", err)
	}

	// 6) 刷新旋转
	pair2, _, err := svc.Refresh(ctx, pair.Refresh, "test-ua")
	if err != nil {
		t.Fatalf("刷新失败: %v", err)
	}
	if pair2.Refresh == pair.Refresh {
		t.Fatal("刷新令牌未旋转")
	}

	// 7) 复用检测:旧 refresh 再用 → 拒绝并全端下线(新 refresh 也随之失效)
	if _, _, err := svc.Refresh(ctx, pair.Refresh, "test-ua"); err == nil {
		t.Fatal("旧刷新令牌复用应被拒绝")
	}
	if _, _, err := svc.Refresh(ctx, pair2.Refresh, "test-ua"); err == nil {
		t.Fatal("复用检测应级联撤销新令牌")
	}

	// 8) 会话版本已 bump:旧 access 失效(缓存失效路径)
	if _, err := svc.Authenticate(ctx, pair2.Access); err == nil {
		t.Fatal("全端下线后旧 access 应失效")
	}
}

func TestLogoutAll(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	phone := "139" + time.Now().Format("040515")[0:6] + "01"

	code, err := svc.SendCode(ctx, phone, "203.0.113.10")
	if err != nil {
		t.Fatal(err)
	}
	pair, user, _, err := svc.VerifyCode(ctx, phone, code, "ua")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Logout(ctx, user.ID, pair.Refresh, true); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Authenticate(ctx, pair.Access); err == nil {
		t.Fatal("全端登出后 access 应失效")
	}
	if _, _, err := svc.Refresh(ctx, pair.Refresh, "ua"); err == nil {
		t.Fatal("全端登出后 refresh 应失效")
	}
}
