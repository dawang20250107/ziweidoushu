package store

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

// TestDivinationRecordLifecycle 卦档全生命周期:存档 → 列表 → 回填解卦 → 详情 → 删除 → 越权防护。
func TestDivinationRecordLifecycle(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	user := newTestUser(t, st)
	other := newTestUser(t, st)

	payload := json.RawMessage(`{"benName":"地天泰","bianName":"山风蛊"}`)
	id, err := st.SaveDivination(ctx, user.ID, "liuyao", "老屋翻修可动工否", "地天泰 → 山风蛊", payload, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.SaveDivination(ctx, user.ID, "meihua", "此事可成否", "泽火革 · 用克体",
		json.RawMessage(`{"ben":{"name":"泽火革"}}`), time.Now()); err != nil {
		t.Fatal(err)
	}

	// 列表:两条,未解卦
	records, total, err := st.ListDivinations(ctx, user.ID, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(records) != 2 {
		t.Fatalf("total=%d len=%d", total, len(records))
	}
	for _, r := range records {
		if r.HasReading {
			t.Fatalf("未解卦记录 hasReading 应为 false: %+v", r)
		}
		if r.Payload != nil {
			t.Fatal("列表不应含 payload")
		}
	}

	// 回填解卦 → 详情可见
	if err := st.AttachDivinationReading(ctx, user.ID, id, "解卦文本……", "anthropic"); err != nil {
		t.Fatal(err)
	}
	rec, err := st.GetDivination(ctx, user.ID, id)
	if err != nil {
		t.Fatal(err)
	}
	if !rec.HasReading || rec.Reading == "" || rec.ReadingProvider != "anthropic" {
		t.Fatalf("解卦回填后详情异常: %+v", rec)
	}
	if string(rec.Payload) == "" || rec.Summary != "地天泰 → 山风蛊" {
		t.Fatalf("详情 payload/summary 异常: %+v", rec)
	}

	// 越权:他人不可见/不可回填/不可删
	if _, err := st.GetDivination(ctx, other.ID, id); !errors.Is(err, ErrNotFound) {
		t.Fatalf("越权读取应 ErrNotFound, got %v", err)
	}
	if err := st.AttachDivinationReading(ctx, other.ID, id, "x", "y"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("越权回填应 ErrNotFound, got %v", err)
	}
	if err := st.DeleteDivination(ctx, other.ID, id); !errors.Is(err, ErrNotFound) {
		t.Fatalf("越权删除应 ErrNotFound, got %v", err)
	}

	// 本人删除
	if err := st.DeleteDivination(ctx, user.ID, id); err != nil {
		t.Fatal(err)
	}
	if _, err := st.GetDivination(ctx, user.ID, id); !errors.Is(err, ErrNotFound) {
		t.Fatal("删除后应不可见")
	}
	if _, total, _ := st.ListDivinations(ctx, user.ID, 50, 0); total != 1 {
		t.Fatalf("删除后 total=%d", total)
	}
}

// TestDivinationRecordCap 超出 200 条截断最旧。
func TestDivinationRecordCap(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	user := newTestUser(t, st)

	payload := json.RawMessage(`{}`)
	base := time.Now().Add(-time.Hour)
	for i := 0; i < divinationRecordCap+5; i++ {
		if _, err := st.SaveDivination(ctx, user.ID, "meihua", "q", "s", payload, base.Add(time.Duration(i)*time.Second)); err != nil {
			t.Fatal(err)
		}
	}
	_, total, err := st.ListDivinations(ctx, user.ID, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total > divinationRecordCap {
		t.Fatalf("应截断至 %d 条, got %d", divinationRecordCap, total)
	}
}
