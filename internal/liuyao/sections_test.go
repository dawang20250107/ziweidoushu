package liuyao

import (
	"strings"
	"testing"
	"time"
)

// TestJudgeSections 分节深断:核心节齐备、逐爻细览六行、静卦亦有动变节。
func TestJudgeSections(t *testing.T) {
	at := time.Date(2026, 8, 5, 10, 30, 0, 0, time.FixedZone("CST", 8*3600))
	r, err := ByTosses([]int{1, 2, 0, 3, 1, 2}, at, "此番转职可成否?")
	if err != nil {
		t.Fatal(err)
	}
	j := r.Judgment
	if j == nil {
		t.Fatal("应有断语")
	}
	keys := map[string]string{}
	for _, s := range j.Sections {
		keys[s.Key] = s.Text
	}
	for _, want := range []string{"yongshen", "dongbian", "zhuyao"} {
		if keys[want] == "" {
			t.Errorf("缺分节 %s", want)
		}
	}
	// 逐爻细览应六行(自上而下)
	if got := len(splitLines(keys["zhuyao"])); got != 6 {
		t.Errorf("逐爻细览应 6 行,得 %d: %q", got, keys["zhuyao"])
	}
	// 静卦(无动爻)动变节应为「六爻安静」文本
	r2, err := ByTosses([]int{1, 2, 1, 2, 1, 2}, at, "测静卦")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range r2.Judgment.Sections {
		if s.Key == "dongbian" && len(s.Text) > 20 {
			found = true
		}
	}
	if !found {
		t.Error("静卦亦应有动变作用节(六爻安静论)")
	}
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i, r := range s {
		if r == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

// TestYongShenOverride 显式取用:问者指明优先于问辞推断;非法值回退推断。
func TestYongShenOverride(t *testing.T) {
	at := time.Date(2026, 8, 5, 10, 30, 0, 0, time.FixedZone("CST", 8*3600))
	// 问辞含「财」本会推妻财;显式指定官鬼须以官鬼为用
	r, err := ByTossesYong([]int{1, 2, 0, 3, 1, 2}, at, "求财之事", "官鬼")
	if err != nil {
		t.Fatal(err)
	}
	if r.YongShen != "官鬼" {
		t.Errorf("显式取用官鬼应生效,得 %s", r.YongShen)
	}
	if !strings.Contains(r.YongShenBasis, "问者指明") {
		t.Errorf("依据应标注问者指明: %s", r.YongShenBasis)
	}
	// 非法值回退问辞推断(妻财)
	r2, _ := ByTossesYong([]int{1, 2, 0, 3, 1, 2}, at, "求财之事", "乱写")
	if r2.YongShen != "妻财" {
		t.Errorf("非法取用应回退推断妻财,得 %s", r2.YongShen)
	}
}
