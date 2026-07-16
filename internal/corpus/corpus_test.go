package corpus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dawang20250107/ziweidoushu/data"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(data.FS, "classics")
	if err != nil {
		t.Fatalf("加载内置古籍失败: %v", err)
	}
	return s
}

func TestEmbeddedBooks(t *testing.T) {
	s := newTestStore(t)
	books := s.Books()
	if len(books) != 3 {
		t.Fatalf("内置古籍数量: got %d want 3", len(books))
	}
	// 与迁移时的统计对齐:共 75 段
	stats := s.Stats()
	if stats["paragraphs"] != 75 {
		t.Errorf("总段落数: got %d want 75", stats["paragraphs"])
	}
	if b := s.Book("gusuifu"); b == nil || b.Title != "骨髓赋" {
		t.Errorf("按 slug 取书失败: %+v", b)
	}
	if b, c := s.Chapter("quanji", 0); b == nil || c == nil || len(c.Paragraphs) == 0 {
		t.Errorf("取章节失败")
	}
	if _, c := s.Chapter("quanji", 999); c != nil {
		t.Errorf("越界章节应返回 nil")
	}
}

func TestSearch(t *testing.T) {
	s := newTestStore(t)

	hits := s.Search("紫微", 30)
	if len(hits) == 0 {
		t.Fatal("搜索「紫微」应有命中")
	}
	for _, h := range hits {
		if !strings.Contains(h.Text, "紫微") {
			t.Errorf("命中原文不含关键词: %s", h.Text)
		}
		if !strings.Contains(h.Snippet, "<mark>紫微</mark>") {
			t.Errorf("片段未高亮: %s", h.Snippet)
		}
	}

	// 单字查询走一元索引
	if len(s.Search("命", 10)) == 0 {
		t.Error("单字搜索应有命中")
	}
	// 无命中
	if n := len(s.Search("量子力学", 10)); n != 0 {
		t.Errorf("不存在的词命中 %d 处", n)
	}
	// 空查询
	if n := len(s.Search("  ", 10)); n != 0 {
		t.Errorf("空查询应返回空")
	}
	// limit 生效
	if n := len(s.Search("命", 3)); n > 3 {
		t.Errorf("limit 未生效: %d", n)
	}
}

func TestExternalIngest(t *testing.T) {
	s := newTestStore(t)
	dir := t.TempDir()

	// 合法外部书目(模拟后期投放的倪师著作)
	ok := `{
		"title": "测试天机道", "slug": "tianjidao-test", "dynasty": "当代", "author": "倪海厦",
		"intro": "测试导入", "wordCount": 10,
		"chapters": [{"title": "第一章", "paragraphs": [{"id": "tj-1-1", "idx": 1, "text": "紫微星君临测试段落"}]}]
	}`
	if err := os.WriteFile(filepath.Join(dir, "tianjidao.json"), []byte(ok), 0o644); err != nil {
		t.Fatal(err)
	}
	n, err := s.LoadExternalDir(dir)
	if err != nil {
		t.Fatalf("外部导入失败: %v", err)
	}
	if n != 1 {
		t.Fatalf("导入数量: got %d want 1", n)
	}
	if len(s.Books()) != 4 {
		t.Fatalf("导入后书目: got %d want 4", len(s.Books()))
	}
	if b := s.Book("tianjidao-test"); b == nil || b.Source != "external" {
		t.Fatalf("外部书目缺失或来源标记错误: %+v", b)
	}
	// 新内容立即可检索
	hits := s.Search("星君临", 10)
	if len(hits) != 1 || hits[0].BookSlug != "tianjidao-test" {
		t.Fatalf("外部内容未进索引: %+v", hits)
	}

	// 重复导入幂等(热加载)
	if _, err := s.LoadExternalDir(dir); err != nil {
		t.Fatalf("重复导入失败: %v", err)
	}
	if len(s.Books()) != 4 {
		t.Fatalf("重复导入后书目应仍为 4, got %d", len(s.Books()))
	}

	// 非法 JSON 拒绝且不破坏现有数据
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte("{broken"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.LoadExternalDir(dir); err == nil {
		t.Fatal("非法 JSON 应报错")
	}
	if len(s.Books()) != 4 {
		t.Fatalf("失败导入不应破坏现有数据: got %d", len(s.Books()))
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	s := newTestStore(t)
	dir := t.TempDir()
	book := `{"title":"并发书","slug":"conc","dynasty":"当代","author":"测试","intro":"x","wordCount":1,
		"chapters":[{"title":"一","paragraphs":[{"id":"c-1","idx":1,"text":"并发安全测试文本"}]}]}`
	if err := os.WriteFile(filepath.Join(dir, "conc.json"), []byte(book), 0o644); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 50; i++ {
			if _, err := s.LoadExternalDir(dir); err != nil {
				t.Errorf("并发导入失败: %v", err)
				return
			}
		}
	}()
	for i := 0; i < 2000; i++ {
		s.Search("紫微", 10)
		s.Books()
		s.Book("gusuifu")
	}
	<-done
}
