package corpus

import (
	"os"
	"path/filepath"
	"testing"
)

// TestBookCategoryHeuristic 板块启发式:内置四书 + 典型研究书名逐一钉住,
// 显式 category 字段优先,主题类目先于倪师通论。
func TestBookCategoryHeuristic(t *testing.T) {
	cases := []struct {
		title, slug, explicit, want string
	}{
		{"骨髓赋", "gusuifu", "", CatZiwei},
		{"梅花易数", "meihuayishu", "", CatMeihua},
		{"紫微斗数全书", "quanshu", "", CatZiwei},
		{"紫微斗数全集", "quanji", "", CatZiwei},
		{"王維德《卜筮正宗》", "r-1", "", CatLiuYao},
		{"增删卜易", "r-2", "", CatLiuYao},
		{"壬归", "r-3", "", CatLiuRen},
		{"六壬断案01", "r-4", "", CatLiuRen},
		{"渊海子平", "r-5", "", CatBazi},
		{"公笃相法", "r-6", "", CatXiang},
		{"伊川易传", "r-7", "", CatZhouYi},
		{"倪海厦-天纪-听课笔记", "r-8", "", CatNi},
		{"倪海厦紫微斗数讲义", "r-9", "", CatZiwei},   // 主题先于倪师通论
		{"正统铁板神数", "r-10", "", CatMisc},      // 未归类 → misc,不被任何解读线引用
		{"某书", "r-11", CatLiuRen, CatLiuRen}, // 显式字段优先
	}
	for _, c := range cases {
		b := &Book{Title: c.title, Slug: c.slug, Category: c.explicit}
		if got := bookCategory(b); got != c.want {
			t.Errorf("bookCategory(%q)=%q, want %q", c.title, got, c.want)
		}
	}
}

// TestSearchAllInFilters 板块过滤检索:同一命中词跨板块的书,
// 过滤后只返回允许板块;空板块退化为全域。
func TestSearchAllInFilters(t *testing.T) {
	s := newTestStore(t)
	dir := t.TempDir()
	// 两本研究书同含「天罡秘要诀」:一本六壬、一本杂项(铁板神数)
	for _, spec := range []struct{ file, title, slug string }{
		{"r-lr.json", "六壬心镜", "r-lr"},
		{"r-misc.json", "铁板神数密钥", "r-misc"},
	} {
		j := `{"title":"` + spec.title + `","slug":"` + spec.slug + `","dynasty":"研究","author":"内部",
			"intro":"t","wordCount":9,"research":true,
			"chapters":[{"title":"卷一","paragraphs":[{"id":"` + spec.slug + `-1","idx":1,"text":"天罡秘要诀所在段落"}]}]}`
		if err := os.WriteFile(filepath.Join(dir, spec.file), []byte(j), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.LoadExternalDir(dir); err != nil {
		t.Fatal(err)
	}

	all := s.SearchAllIn("天罡秘要诀", 10)
	if len(all) != 2 {
		t.Fatalf("空板块应退化全域命中 2,得 %d", len(all))
	}
	lr := s.SearchAllIn("天罡秘要诀", 10, CatLiuRen)
	if len(lr) != 1 || lr[0].BookSlug != "r-lr" {
		t.Fatalf("六壬板块过滤应只命中 r-lr,得 %+v", lr)
	}
	if hits := s.SearchAllIn("天罡秘要诀", 10, CatZiwei); len(hits) != 0 {
		t.Fatalf("紫微板块不应命中六壬/杂项语料,得 %+v", hits)
	}
	// 紫微线板块组不应命中梅花书(实测教训:梅花脉诀混入紫微报告)
	if hits := s.SearchAllIn("梅花", 20, CatZiwei, CatXiang, CatBazi, CatNi); len(hits) != 0 {
		for _, h := range hits {
			if h.BookSlug == "meihuayishu" {
				t.Fatalf("紫微线板块组命中了《梅花易数》: %+v", h)
			}
		}
	}
}
