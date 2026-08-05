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

// TestHemingTiming 婚嫁流年:共振年须双方指数均达标,且 timing 维度随响应给出。
func TestHemingTiming(t *testing.T) {
	a, _ := ziwei.Generate(ziwei.BirthInfo{Year: 1990, Month: 6, Day: 15, Hour: 6, Gender: ziwei.Male}, ziwei.Options{ReferenceYear: 2026})
	b, _ := ziwei.Generate(ziwei.BirthInfo{Year: 1992, Month: 3, Day: 8, Hour: 4, Gender: ziwei.Female}, ziwei.Options{ReferenceYear: 2026})
	hr := buildHemingReading(a, b)
	// timing 维度必然存在(有共振年或说明无共振)。
	var hasTiming bool
	for _, s := range hr.Sections {
		if s.Key == "timing" {
			hasTiming = true
			if s.Text == "" {
				t.Error("婚嫁流年断语为空")
			}
		}
	}
	if !hasTiming {
		t.Error("缺婚嫁流年维度")
	}
	// 若给出共振年,须落在未来十年窗口内。
	for _, ty := range hr.Timing {
		if ty.Year < 2026 || ty.Year >= 2036 {
			t.Errorf("共振年 %d 超出十年窗口", ty.Year)
		}
		if ty.Note == "" || ty.GanZhi == "" {
			t.Errorf("共振年 %d 缺干支或说明", ty.Year)
		}
	}
	// 单人婚嫁指数:流年行至夫妻宫应加分。
	fuqi := a.PalaceByName("夫妻")
	found := false
	for y := 2026; y < 2040 && !found; y++ {
		h, err := ziwei.GenerateHoroscope(a, y, 6, 15, 6)
		if err != nil {
			continue
		}
		if h.Yearly.PalaceBranch == fuqi.Branch {
			if sc, _ := personMarriageYear(a, y); sc < 2 {
				t.Errorf("流年行至夫妻宫之年 %d 婚嫁指数应≥2,得 %d", y, sc)
			}
			found = true
		}
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
