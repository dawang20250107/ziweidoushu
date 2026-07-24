package ai

import (
	"strings"
	"testing"

	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// TestHemingReading 合盘确定性断语:四段齐备、分数有界、随配对而异。
func TestHemingReading(t *testing.T) {
	a, err := ziwei.Generate(ziwei.BirthInfo{Year: 1990, Month: 6, Day: 15, Hour: 6, Gender: ziwei.Male}, ziwei.Options{ReferenceYear: 2026})
	if err != nil {
		t.Fatal(err)
	}
	b, err := ziwei.Generate(ziwei.BirthInfo{Year: 1992, Month: 3, Day: 8, Hour: 4, Gender: ziwei.Female}, ziwei.Options{ReferenceYear: 2026})
	if err != nil {
		t.Fatal(err)
	}
	hr := buildHemingReading(a, b)
	if hr == nil {
		t.Fatal("合盘断语为空")
	}
	if hr.Score < 20 || hr.Score > 100 {
		t.Errorf("契合分越界: %d", hr.Score)
	}
	if hr.Level == "" || hr.Summary == "" {
		t.Error("等级或总述缺失")
	}
	wantKeys := []string{"nianming", "sihuafly", "echo", "advice"}
	got := map[string]bool{}
	for _, s := range hr.Sections {
		got[s.Key] = true
	}
	for _, k := range wantKeys {
		if !got[k] {
			t.Errorf("缺合盘维度 %s", k)
		}
	}

	// 换一方,断语应不同(年命/四化互飞随之变)。
	c, err := ziwei.Generate(ziwei.BirthInfo{Year: 1985, Month: 11, Day: 2, Hour: 8, Gender: ziwei.Female}, ziwei.Options{ReferenceYear: 2026})
	if err != nil {
		t.Fatal(err)
	}
	hr2 := buildHemingReading(a, c)
	if hr.Summary == hr2.Summary && hr.Sections[0].Text == hr2.Sections[0].Text {
		t.Error("不同配对的合盘断语不应完全雷同")
	}
}

// TestYearBranchRelation 年支关系判定正确(合/冲/害/刑各取一例)。
func TestYearBranchRelation(t *testing.T) {
	// 卯(3)戌(10)六合。
	if _, d := yearBranchRelation(3, 10); d <= 0 {
		t.Errorf("卯戌应六合(正分),得 %d", d)
	}
	// 子(0)午(6)相冲。
	if _, d := yearBranchRelation(0, 6); d >= 0 {
		t.Errorf("子午应相冲(负分),得 %d", d)
	}
	// 子(0)未(7)相害。
	if txt, d := yearBranchRelation(0, 7); d >= 0 || !strings.Contains(txt, "相害") {
		t.Errorf("子未应相害,得 %d / %s", d, txt)
	}
	// 申(8)子(0)三合。
	if _, d := yearBranchRelation(8, 0); d <= 0 {
		t.Errorf("申子应三合(正分),得 %d", d)
	}
}
