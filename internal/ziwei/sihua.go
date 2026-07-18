package ziwei

// 四化工具:年干 / 大限宫干 / 流年干 / 流月干 四化映射 + 宫干自化检测 + 来因宫追溯。
//
// 倪海厦《天纪》体系口径:
//   本命四化 = 出生年天干四化(静态基础,已在排盘时标注到星曜上);
//   流年四化 = 当年年干四化(一年动态)。
// 大限宫干四化、宫干自化、来因宫属飞星派工具,倪师不采用
// (《天纪 03》:四化星永远固定不动),此处保留仅供研究比对。

// SiHuaSet 某一天干引发的四化四星。
type SiHuaSet struct {
	StemIndex int    `json:"stemIndex"`
	StemName  string `json:"stemName"`
	Lu        string `json:"lu"`   // 化禄
	Quan      string `json:"quan"` // 化权
	Ke        string `json:"ke"`   // 化科
	Ji        string `json:"ji"`   // 化忌
}

// SiHuaByStem 天干索引(0-9)→ 四化四星。
func SiHuaByStem(stemIndex int) SiHuaSet {
	if stemIndex < 0 || stemIndex > 9 {
		return SiHuaSet{StemIndex: -1}
	}
	t := mutagenTable[stemIndex]
	return SiHuaSet{
		StemIndex: stemIndex,
		StemName:  Stems[stemIndex],
		Lu:        t[0], Quan: t[1], Ke: t[2], Ji: t[3],
	}
}

// YearStemIndex 公历年份 → 年柱天干索引(0=甲)。
func YearStemIndex(year int) int { return fix10(year - 4) }

// YearBranchIndex 公历年份 → 年柱地支索引(0=子)。
func YearBranchIndex(year int) int { return fix12(year - 4) }

// LiuNianSiHua 流年四化(按公历年份的年干)。
func LiuNianSiHua(year int) SiHuaSet { return SiHuaByStem(YearStemIndex(year)) }

// LiuYueStemIndex 流月天干(五虎遁:由年干与农历月序推正月起干)。
func LiuYueStemIndex(yearStem, month int) int {
	if yearStem < 0 || yearStem > 9 {
		return -1
	}
	return fix10(tigerRule[yearStem] + (month-1)%12)
}

// LiuYueSiHua 流月四化。month 为农历月 1-12。
func LiuYueSiHua(yearStem, month int) SiHuaSet {
	return SiHuaByStem(LiuYueStemIndex(yearStem, month))
}

// DaXianSiHuaOf 大限宫干四化(飞星派工具,仅供研究)。
// dxIndex 为 Chart.DaXians 序号,越界返回 StemIndex=-1 的空集。
func DaXianSiHuaOf(c *Chart, dxIndex int) SiHuaSet {
	if c == nil || dxIndex < 0 || dxIndex >= len(c.DaXians) {
		return SiHuaSet{StemIndex: -1}
	}
	p := c.PalaceByBranch(c.DaXians[dxIndex].PalaceBranch)
	if p == nil {
		return SiHuaSet{StemIndex: -1}
	}
	return SiHuaByStem(p.Stem)
}

// SelfSiHuaMark 宫干自化标记。
type SelfSiHuaMark struct {
	SiHua    SiHua  `json:"siHua"`
	StarName string `json:"starName"`
}

// DetectSelfSiHua 检测某宫宫干自化:宫干引发的四化星恰在本宫。
func DetectSelfSiHua(p *Palace) []SelfSiHuaMark {
	if p == nil {
		return nil
	}
	set := SiHuaByStem(p.Stem)
	var out []SelfSiHuaMark
	for _, pair := range []struct {
		h SiHua
		s string
	}{{HuaLu, set.Lu}, {HuaQuan, set.Quan}, {HuaKe, set.Ke}, {HuaJi, set.Ji}} {
		if pair.s != "" && p.HasStar(pair.s) {
			out = append(out, SelfSiHuaMark{SiHua: pair.h, StarName: pair.s})
		}
	}
	return out
}

// FindIncomingPalaces 来因宫追溯:哪些宫位的宫干把 starName 化为 sihua。
func FindIncomingPalaces(c *Chart, starName string, sihua SiHua) []*Palace {
	if c == nil {
		return nil
	}
	var out []*Palace
	for i := range c.Palaces {
		set := SiHuaByStem(c.Palaces[i].Stem)
		var target string
		switch sihua {
		case HuaLu:
			target = set.Lu
		case HuaQuan:
			target = set.Quan
		case HuaKe:
			target = set.Ke
		case HuaJi:
			target = set.Ji
		}
		if target == starName {
			out = append(out, &c.Palaces[i])
		}
	}
	return out
}
