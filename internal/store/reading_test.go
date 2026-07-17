package store

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestReadingProgressUpsert 进度 upsert:同书二次写入取后值,列表按最近读排序。
func TestReadingProgressUpsert(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	user := newTestUser(t, st)

	// 首次读「甲书」第 0 章
	if err := st.UpsertReadingProgress(ctx, user.ID, "jiashu", 0, "p1"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Millisecond)
	// 读「乙书」第 2 章(更晚,应排最前)
	if err := st.UpsertReadingProgress(ctx, user.ID, "yishu", 2, "p9"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Millisecond)
	// 回到「甲书」推进到第 3 章 —— 同书二次写入应覆盖为后值,且刷新为最近
	if err := st.UpsertReadingProgress(ctx, user.ID, "jiashu", 3, "p42"); err != nil {
		t.Fatal(err)
	}

	list, err := st.ListReadingProgress(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("应有 2 本书进度(同书合并为一行), got %d", len(list))
	}
	// 最近读的「甲书」应在最前,且取后值(第 3 章 p42)
	if list[0].BookSlug != "jiashu" {
		t.Fatalf("最近读的书应排最前, got %s", list[0].BookSlug)
	}
	if list[0].ChapterIdx != 3 || list[0].ParagraphID != "p42" {
		t.Fatalf("同书二次写入应取后值, got 第%d章 %s", list[0].ChapterIdx, list[0].ParagraphID)
	}
	if list[1].BookSlug != "yishu" || list[1].ChapterIdx != 2 {
		t.Fatalf("乙书进度异常: %+v", list[1])
	}
}

// TestBookmarksCRUD 书签增删查:新增、按书过滤、同段幂等、删除。
func TestBookmarksCRUD(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	user := newTestUser(t, st)

	b1, err := st.AddBookmark(ctx, user.ID, "jiashu", 0, "p1", "紫微星者,帝座也")
	if err != nil {
		t.Fatal(err)
	}
	if b1.ID == "" || b1.Excerpt != "紫微星者,帝座也" {
		t.Fatalf("新增书签异常: %+v", b1)
	}
	if _, err := st.AddBookmark(ctx, user.ID, "jiashu", 1, "p7", "天府为库"); err != nil {
		t.Fatal(err)
	}
	if _, err := st.AddBookmark(ctx, user.ID, "yishu", 0, "p1", "别书一段"); err != nil {
		t.Fatal(err)
	}

	// 同段再加:幂等,返回既有 id,摘录不被覆盖
	dup, err := st.AddBookmark(ctx, user.ID, "jiashu", 0, "p1", "覆盖尝试")
	if err != nil {
		t.Fatal(err)
	}
	if dup.ID != b1.ID {
		t.Fatalf("同段重复加书签应幂等返回既有, got %s want %s", dup.ID, b1.ID)
	}
	if dup.Excerpt != "紫微星者,帝座也" {
		t.Fatalf("幂等加书签不应覆盖既有摘录, got %s", dup.Excerpt)
	}

	// 全部列表:3 条
	all, err := st.ListBookmarks(ctx, user.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("应有 3 条书签, got %d", len(all))
	}

	// 按书过滤:甲书 2 条
	jiashu, err := st.ListBookmarks(ctx, user.ID, "jiashu")
	if err != nil {
		t.Fatal(err)
	}
	if len(jiashu) != 2 {
		t.Fatalf("甲书应有 2 条书签, got %d", len(jiashu))
	}
	for _, b := range jiashu {
		if b.BookSlug != "jiashu" {
			t.Fatalf("过滤应只含甲书, got %s", b.BookSlug)
		}
	}

	// 删除一条并幂等复删
	if err := st.DeleteBookmark(ctx, user.ID, b1.ID); err != nil {
		t.Fatal(err)
	}
	if err := st.DeleteBookmark(ctx, user.ID, b1.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("复删已删书签应 ErrNotFound, got %v", err)
	}
	after, _ := st.ListBookmarks(ctx, user.ID, "")
	if len(after) != 2 {
		t.Fatalf("删除后应剩 2 条, got %d", len(after))
	}
}

// TestDeleteOthersBookmarkForbidden 越权删除他人书签:返回 ErrNotFound(→404),原书签不受影响。
func TestDeleteOthersBookmarkForbidden(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	owner := newTestUser(t, st)
	attacker := newTestUser(t, st)

	bm, err := st.AddBookmark(ctx, owner.ID, "jiashu", 0, "p1", "他人书签")
	if err != nil {
		t.Fatal(err)
	}

	// 攻击者以自己的身份删他人书签:校验归属失败,ErrNotFound
	if err := st.DeleteBookmark(ctx, attacker.ID, bm.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("越权删除应 ErrNotFound, got %v", err)
	}
	// 非法 UUID 同样 ErrNotFound(不泄露存在性)
	if err := st.DeleteBookmark(ctx, attacker.ID, "not-a-uuid"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("非法 id 应 ErrNotFound, got %v", err)
	}

	// 物主的书签仍在
	still, _ := st.ListBookmarks(ctx, owner.ID, "")
	if len(still) != 1 {
		t.Fatalf("越权删除不应影响物主书签, got %d", len(still))
	}
}
