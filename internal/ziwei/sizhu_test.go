package ziwei

import (
	"strings"
	"testing"
)

// TestSiZhuView 四柱视角 spot 基准:1990-06-15 午时男
// 四柱 庚午 壬午 辛亥 甲午,日主辛金。
// 十神:庚=劫财 壬=伤官 甲=正财;午藏丁己=七杀/偏印;亥藏壬甲=伤官/正财。
// 纳音:路旁土/杨柳木/钗钏金/沙中金。五行:金2 水2 木1 火3 土0。
func TestSiZhuView(t *testing.T) {
	chart, err := Generate(BirthInfo{Year: 1990, Month: 6, Day: 15, Hour: 6, Gender: Male}, Options{ReferenceYear: 2026})
	if err != nil {
		t.Fatal(err)
	}
	sz := chart.SiZhu
	if sz == nil {
		t.Fatal("SiZhu 为空")
	}
	if sz.DayMaster != "辛" || sz.DayMasterElement != "金" {
		t.Fatalf("日主: got %s%s want 辛金", sz.DayMaster, sz.DayMasterElement)
	}

	wantPillars := []struct {
		stem, branch, shiShen, naYin string
	}{
		{"庚", "午", "劫财", "路旁土"},
		{"壬", "午", "伤官", "杨柳木"},
		{"辛", "亥", "日主", "钗钏金"},
		{"甲", "午", "正财", "沙中金"},
	}
	for i, w := range wantPillars {
		p := sz.Pillars[i]
		if p.Stem != w.stem || p.Branch != w.branch {
			t.Fatalf("柱%d 干支: got %s%s want %s%s", i, p.Stem, p.Branch, w.stem, w.branch)
		}
		if p.StemShiShen != w.shiShen {
			t.Fatalf("柱%d 天干十神: got %s want %s", i, p.StemShiShen, w.shiShen)
		}
		if p.NaYin != w.naYin {
			t.Fatalf("柱%d 纳音: got %s want %s", i, p.NaYin, w.naYin)
		}
	}

	// 午藏干:丁(七杀)己(偏印)
	yearHidden := sz.Pillars[0].Hidden
	if len(yearHidden) != 2 || yearHidden[0].Stem != "丁" || yearHidden[0].ShiShen != "七杀" ||
		yearHidden[1].Stem != "己" || yearHidden[1].ShiShen != "偏印" {
		t.Fatalf("午藏干十神: %+v", yearHidden)
	}
	// 亥藏干:壬(伤官)甲(正财)
	dayHidden := sz.Pillars[2].Hidden
	if len(dayHidden) != 2 || dayHidden[0].ShiShen != "伤官" || dayHidden[1].ShiShen != "正财" {
		t.Fatalf("亥藏干十神: %+v", dayHidden)
	}

	want := map[string]int{"金": 2, "水": 2, "木": 1, "火": 3, "土": 0}
	for k, v := range want {
		if sz.ElementCount[k] != v {
			t.Fatalf("五行分布 %s: got %d want %d(全部:%v)", k, sz.ElementCount[k], v, sz.ElementCount)
		}
	}
}

// TestNaYinTable 六十甲子纳音表抽查(首/中/尾)。
func TestNaYinTable(t *testing.T) {
	cases := []struct {
		stem, branch int
		want         string
	}{
		{0, 0, "海中金"},  // 甲子
		{1, 1, "海中金"},  // 乙丑
		{2, 2, "炉中火"},  // 丙寅
		{6, 6, "路旁土"},  // 庚午
		{8, 10, "大海水"}, // 壬戌
		{9, 11, "大海水"}, // 癸亥
	}
	for _, c := range cases {
		if got := naYin(c.stem, c.branch); got != c.want {
			t.Fatalf("纳音 %d/%d: got %s want %s", c.stem, c.branch, got, c.want)
		}
	}
}

// TestGeJu 月令取格分支覆盖:八格/建禄/阳刃/月劫/杂气透干/皆不透。
func TestGeJu(t *testing.T) {
	cases := []struct {
		fp        FourPillars
		wantName  string
		basisPart string
	}{
		// 辛日午月:本气丁(七杀)不透,己藏中气亦不透 → 仍以本气论七杀
		{FourPillars{Year: "庚午", Month: "壬午", Day: "辛亥", Hour: "甲午"}, "七杀格", "本气"},
		// 甲日寅月:禄位 → 建禄格
		{FourPillars{Year: "甲寅", Month: "丙寅", Day: "甲子", Hour: "乙丑"}, "建禄格", "禄位"},
		// 丙日午月:刃位 → 阳刃格
		{FourPillars{Year: "壬子", Month: "丙午", Day: "丙申", Hour: "戊子"}, "阳刃格", "刃位"},
		// 乙日寅月:本气甲为劫财,非禄非刃 → 月劫格
		{FourPillars{Year: "壬寅", Month: "壬寅", Day: "乙亥", Hour: "丁丑"}, "月劫格", "劫财"},
		// 戊日丑月:本气己劫财,四库杂气,癸透月干 → 正财格
		{FourPillars{Year: "壬子", Month: "癸丑", Day: "戊午", Hour: "丁巳"}, "正财格", "杂气"},
		// 甲日申月:本气庚(七杀)透年干 → 七杀格(本气透干)
		{FourPillars{Year: "庚申", Month: "甲申", Day: "甲子", Hour: "乙亥"}, "七杀格", "透干"},
		// 甲日申月:庚不透而壬透 → 取壬偏印格
		{FourPillars{Year: "壬申", Month: "戊申", Day: "甲子", Hour: "乙亥"}, "偏印格", "不透"},
		// 丙日酉月:酉唯一藏辛(正财)不透 → 本气秉令正财格
		{FourPillars{Year: "甲午", Month: "癸酉", Day: "丙子", Hour: "戊戌"}, "正财格", "秉令"},
	}
	for _, c := range cases {
		sz := buildSiZhu(c.fp)
		if sz == nil || sz.GeJu == nil {
			t.Fatalf("%v: 未取格", c.fp)
		}
		if sz.GeJu.Name != c.wantName {
			t.Fatalf("%v: got %s(%s) want %s", c.fp, sz.GeJu.Name, sz.GeJu.Basis, c.wantName)
		}
		if c.basisPart != "" && !strings.Contains(sz.GeJu.Basis, c.basisPart) {
			t.Fatalf("%v: basis %q 缺 %q", c.fp, sz.GeJu.Basis, c.basisPart)
		}
		if sz.GeJu.Note == "" || sz.GeJu.Source == "" {
			t.Fatalf("%v: 断语/出处缺失 %+v", c.fp, sz.GeJu)
		}
	}
}
