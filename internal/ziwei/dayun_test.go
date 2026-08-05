package ziwei

import "testing"

// TestDaYun 大运/流年:起运顺逆、首运接月柱、运干十神/纳音、当前运与流年标注。
// 大运起运精算委托 lunar-go(与四柱同库);本测试锁定我方接线与十神/纳音注入。
func TestDaYun(t *testing.T) {
	// 甲子年 丙寅月 甲戌日 庚午时,男 —— 阳年男顺行,首运接月柱丙寅之后=丁卯。
	c, err := Generate(BirthInfo{Year: 1984, Month: 2, Day: 10, Hour: 6, Gender: Male},
		Options{ReferenceYear: 2024})
	if err != nil {
		t.Fatal(err)
	}
	dy := c.SiZhu.DaYun
	if dy == nil {
		t.Fatal("大运为空")
	}
	if !dy.Forward {
		t.Errorf("甲子年(阳)男应顺行,得逆行")
	}
	if len(dy.List) < 8 {
		t.Fatalf("大运步数应≥8,得 %d", len(dy.List))
	}
	if dy.List[0].GanZhi != "丁卯" {
		t.Errorf("顺行首运应接月柱丙寅之后=丁卯,得 %s", dy.List[0].GanZhi)
	}
	if dy.StartAge != dy.List[0].StartAge {
		t.Errorf("视图起运虚岁 %d 应等于首运 %d", dy.StartAge, dy.List[0].StartAge)
	}

	dayStem := stemIndex([]rune(c.SiZhu.DayMaster)[0])
	curCount := 0
	for _, e := range dy.List {
		// 运干十神与纳音须与我方 shiShen/naYin 一致
		rs := []rune(e.GanZhi)
		s, b := stemIndex(rs[0]), branchIndex(rs[1])
		if e.StemShiShen != shiShen(dayStem, s) {
			t.Errorf("运%d %s 十神:得「%s」应「%s」", e.Index, e.GanZhi, e.StemShiShen, shiShen(dayStem, s))
		}
		if e.NaYin != naYin(s, b) {
			t.Errorf("运%d %s 纳音:得「%s」应「%s」", e.Index, e.GanZhi, e.NaYin, naYin(s, b))
		}
		if e.IsCurrent {
			curCount++
		}
	}
	if curCount != 1 {
		t.Errorf("当前大运应恰 1 步,得 %d", curCount)
	}

	// 当前大运流年:10 条,恰一条为参照年(2024)。
	if len(dy.CurrentLiuNian) != 10 {
		t.Fatalf("当前大运流年应 10 条,得 %d", len(dy.CurrentLiuNian))
	}
	nCur := 0
	for _, l := range dy.CurrentLiuNian {
		if l.IsCurrent {
			nCur++
			if l.Year != 2024 {
				t.Errorf("当前流年应为 2024,得 %d", l.Year)
			}
		}
		rs := []rune(l.GanZhi)
		if s := stemIndex(rs[0]); l.StemShiShen != shiShen(dayStem, s) {
			t.Errorf("流年%d %s 十神不一致", l.Year, l.GanZhi)
		}
	}
	if nCur != 1 {
		t.Errorf("流年当前标注应恰 1 条,得 %d", nCur)
	}
}
