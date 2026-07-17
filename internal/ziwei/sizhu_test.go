package ziwei

import "testing"

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
