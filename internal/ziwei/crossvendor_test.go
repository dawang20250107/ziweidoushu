package ziwei

import "testing"

// 跨软件名盘对照:以市面专业软件「文墨天机专业版 pro 2.5.9」的实盘输出为
// 外部基准(用户提供截图,2026-07 核对),锚定安星硬核层——历法转换、五行局、
// 命身宫、十四主星、六吉六煞禄马、生年四化、大限起限、命主身主、双套四柱。
// 该层《紫微斗数全书·安星诀》口径唯一,任何软件在此层分歧即有一方为错;
// 杂曜集合/命名(截空 vs 截路空亡)与起运天数舍入属产品/派别差异,不在此锚定。
//
// 命例:1996-07-14 00:20(早子时)男,农历丙子年五月廿九子时,阳男金四局。
func TestCrossVendorWenMo19960714(t *testing.T) {
	c, err := Generate(BirthInfo{Year: 1996, Month: 7, Day: 14, Hour: 0, Gender: Male}, Options{ReferenceYear: 2026})
	if err != nil {
		t.Fatal(err)
	}

	// 历法与总纲
	if c.LunarDateText != "一九九六年五月廿九" {
		t.Errorf("农历: got %s", c.LunarDateText)
	}
	if c.WuxingJuName != "金四局" {
		t.Errorf("五行局: got %s", c.WuxingJuName)
	}
	if c.MingZhu != "破军" || c.ShenZhu != "火星" {
		t.Errorf("命主/身主: got %s/%s want 破军/火星", c.MingZhu, c.ShenZhu)
	}
	// 双套四柱:紫微(非节气)与四柱视角(节气)——文墨同样双列
	if fp := c.FourPillars; fp.Year != "丙子" || fp.Month != "甲午" || fp.Day != "壬子" || fp.Hour != "庚子" {
		t.Errorf("紫微四柱: got %v want 丙子 甲午 壬子 庚子", fp)
	}
	if c.SiZhu == nil {
		t.Fatal("SiZhu 为空")
	}
	if got := c.SiZhu.Pillars[1].Stem + c.SiZhu.Pillars[1].Branch; got != "乙未" {
		t.Errorf("节气月柱: got %s want 乙未(小暑后未月)", got)
	}

	// 命身同宫甲午(子时生人命身同宫)
	ming := c.MingGong()
	if Branches[c.MingGongBranch] != "午" || Branches[c.ShenGongBranch] != "午" || Stems[ming.Stem] != "甲" {
		t.Errorf("命身宫: 命%s%s 身%s want 甲午同宫", Stems[ming.Stem], Branches[c.MingGongBranch], Branches[c.ShenGongBranch])
	}

	// 十四主星落宫 + 六吉六煞禄马(文墨实盘逐颗核对)
	wantStars := map[string]string{
		"紫微": "午", "破军": "申", "廉贞": "戌", "天府": "戌", "太阴": "亥",
		"贪狼": "子", "天同": "丑", "巨门": "丑", "武曲": "寅", "天相": "寅",
		"太阳": "卯", "天梁": "卯", "七杀": "辰", "天机": "巳",
		"文昌": "戌", "文曲": "辰", "左辅": "申", "右弼": "午",
		"天魁": "亥", "天钺": "酉", "禄存": "巳", "擎羊": "午", "陀罗": "辰",
		"火星": "寅", "铃星": "戌", "地空": "亥", "地劫": "亥", "天马": "寅",
	}
	for star, wantBranch := range wantStars {
		p := findStar(c, star)
		if p == nil {
			t.Errorf("%s 未安", star)
			continue
		}
		if Branches[p.Branch] != wantBranch {
			t.Errorf("%s: got %s宫 want %s宫", star, Branches[p.Branch], wantBranch)
		}
	}

	// 丙年生年四化:同禄 机权 昌科 廉忌
	wantHua := map[string]SiHua{"天同": "禄", "天机": "权", "文昌": "科", "廉贞": "忌"}
	for star, want := range wantHua {
		p := findStar(c, star)
		found := false
		for _, s := range p.Stars {
			if s.Name == star && s.SiHua == want {
				found = true
			}
		}
		if !found {
			t.Errorf("%s 应化%s", star, want)
		}
	}

	// 大限:金四局命宫 4-13 起,阳男顺行,福德申 24-33
	if ming.DaXianStart != 4 || ming.DaXianEnd != 13 {
		t.Errorf("命宫大限: got %d-%d want 4-13", ming.DaXianStart, ming.DaXianEnd)
	}
	fude := c.PalaceByName("福德")
	if fude == nil || fude.DaXianStart != 24 || Branches[fude.Branch] != "申" {
		t.Errorf("福德大限: want 申宫 24 起")
	}
}

// findStar 全盘找星,返回所在宫。
func findStar(c *Chart, name string) *Palace {
	for i := range c.Palaces {
		for _, s := range c.Palaces[i].Stars {
			if s.Name == name {
				return &c.Palaces[i]
			}
		}
	}
	return nil
}
