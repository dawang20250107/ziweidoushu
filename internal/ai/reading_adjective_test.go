package ai

import (
	"strings"
	"testing"

	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// 杂曜点缀层:有宫位语义才开口、Stars 与 ExtraStars 同扫、至多三条。
func TestAdjectiveNotes(t *testing.T) {
	p := &ziwei.Palace{
		Name: "财帛",
		Stars: []ziwei.Star{
			{Name: "武曲", Type: ziwei.StarMajor}, // 主星不入杂曜层
			{Name: "破碎", Type: ziwei.StarMinor},
		},
		ExtraStars: []ziwei.Star{
			{Name: "大耗", Type: ziwei.StarMinor},
			{Name: "劫煞", Type: ziwei.StarMinor},
		},
	}
	notes := adjectiveNotes(p, "财帛")
	if len(notes) != 3 {
		t.Fatalf("财帛应得 3 条杂曜断语(破碎/大耗/劫煞),得 %d: %v", len(notes), notes)
	}
	joined := strings.Join(notes, ";")
	for _, kw := range []string{"破碎", "大耗", "劫煞"} {
		if !strings.Contains(joined, kw) {
			t.Errorf("断语应含 %s: %s", kw, joined)
		}
	}

	// 无语义宫位保持沉默(兄弟宫未收录 → 不写通用套话)
	if got := adjectiveNotes(p, "兄弟"); len(got) != 0 {
		t.Errorf("兄弟宫应沉默,得 %v", got)
	}

	// 上限三条:再多的杂曜也不喧宾夺主
	p.Stars = append(p.Stars, ziwei.Star{Name: "天虚", Type: ziwei.StarMinor})
	p.Name = "命宫"
	many := adjectiveNotes(&ziwei.Palace{
		Name: "命宫",
		Stars: []ziwei.Star{
			{Name: "天刑", Type: ziwei.StarMinor},
			{Name: "天姚", Type: ziwei.StarMinor},
			{Name: "华盖", Type: ziwei.StarMinor},
			{Name: "天哭", Type: ziwei.StarMinor},
			{Name: "阴煞", Type: ziwei.StarMinor},
		},
	}, "命宫")
	if len(many) != 3 {
		t.Errorf("杂曜断语应封顶 3 条,得 %d", len(many))
	}
}

// 断语骨架整链:补充杂曜经 buildReading 注入宫位断语(大耗入财帛盘例:
// 1990-6-15 午年生人大耗在丑,丑宫为其疾厄……选一张大耗恰落六亲外主域的盘)。
func TestAdjectiveInReading(t *testing.T) {
	// 1987-2-1 卯年:大耗申、劫煞申、龙德戌。排盘后看申/戌落何宫,若为
	// 已收录宫位则断语应出现对应短句;至少验证 ExtraStars 已入盘。
	c, err := ziwei.Generate(ziwei.BirthInfo{Year: 1987, Month: 2, Day: 1, Hour: 6, Gender: ziwei.Male}, ziwei.Options{ReferenceYear: 2026})
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for i := range c.Palaces {
		found += len(c.Palaces[i].ExtraStars)
	}
	if found != 3 {
		t.Fatalf("补充杂曜应共 3 颗入盘,得 %d", found)
	}
	rd := buildReading(c, nil)
	if rd == nil {
		t.Fatal("buildReading 为空")
	}
}
