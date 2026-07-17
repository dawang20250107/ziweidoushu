// 四柱视角(八字附加层):在紫微盘四柱干支之上推导五行、十神、
// 藏干与纳音——八字不作独立产品,作为紫微盘的同源视角呈现。
package ziwei

// HiddenStem 地支藏干(本气在前)。
type HiddenStem struct {
	Stem    string `json:"stem"`
	Element string `json:"element"`
	ShiShen string `json:"shiShen"` // 相对日主的十神
}

// SiZhuPillar 一柱。
type SiZhuPillar struct {
	Name          string       `json:"name"` // 年柱/月柱/日柱/时柱
	Stem          string       `json:"stem"`
	Branch        string       `json:"branch"`
	StemElement   string       `json:"stemElement"`
	BranchElement string       `json:"branchElement"` // 地支本气五行
	StemShiShen   string       `json:"stemShiShen"`   // 日柱天干为「日主」
	Hidden        []HiddenStem `json:"hidden"`
	NaYin         string       `json:"naYin"`
}

// SiZhuView 四柱视角。
type SiZhuView struct {
	DayMaster        string         `json:"dayMaster"` // 日干
	DayMasterElement string         `json:"dayMasterElement"`
	Pillars          [4]SiZhuPillar `json:"pillars"`
	// ElementCount 八字五行分布(四天干 + 四地支本气,共 8 字)。
	ElementCount map[string]int `json:"elementCount"`
}

var (
	szStems    = []rune("甲乙丙丁戊己庚辛壬癸")
	szBranches = []rune("子丑寅卯辰巳午未申酉戌亥")
	// 五行序:木0 火1 土2 金3 水4(顺生)
	elementNames = []string{"木", "火", "土", "金", "水"}
	// 天干五行:甲乙木 丙丁火 戊己土 庚辛金 壬癸水
	stemElement = []int{0, 0, 1, 1, 2, 2, 3, 3, 4, 4}
	// 地支本气五行
	branchElement = []int{4, 2, 0, 0, 2, 1, 1, 2, 3, 3, 2, 4}
	// 地支藏干(本气在前)
	hiddenStems = [][]rune{
		[]rune("癸"), []rune("己癸辛"), []rune("甲丙戊"), []rune("乙"),
		[]rune("戊乙癸"), []rune("丙庚戊"), []rune("丁己"), []rune("己丁乙"),
		[]rune("庚壬戊"), []rune("辛"), []rune("戊辛丁"), []rune("壬甲"),
	}
	// 六十甲子纳音(两柱一音,索引 = 甲子序/2)
	naYinNames = []string{
		"海中金", "炉中火", "大林木", "路旁土", "剑锋金", "山头火",
		"涧下水", "城头土", "白蜡金", "杨柳木", "泉中水", "屋上土",
		"霹雳火", "松柏木", "长流水", "沙中金", "山下火", "平地木",
		"壁上土", "金箔金", "覆灯火", "天河水", "大驿土", "钗钏金",
		"桑柘木", "大溪水", "沙中土", "天上火", "石榴木", "大海水",
	}
)

func stemIndex(r rune) int {
	for i, s := range szStems {
		if s == r {
			return i
		}
	}
	return -1
}

func branchIndex(r rune) int {
	for i, b := range szBranches {
		if b == r {
			return i
		}
	}
	return -1
}

// shiShen 日主天干 → 另一天干的十神。
// 同五行:同阴阳比肩/异劫财;我生:同食神/异伤官;我克:同偏财/异正财;
// 克我:同七杀/异正官;生我:同偏印/异正印。
func shiShen(dayStem, other int) string {
	de, oe := stemElement[dayStem], stemElement[other]
	same := dayStem%2 == other%2
	pick := func(a, b string) string {
		if same {
			return a
		}
		return b
	}
	switch {
	case de == oe:
		return pick("比肩", "劫财")
	case (de+1)%5 == oe: // 我生
		return pick("食神", "伤官")
	case (de+2)%5 == oe: // 我克
		return pick("偏财", "正财")
	case (oe+2)%5 == de: // 克我
		return pick("七杀", "正官")
	default: // 生我
		return pick("偏印", "正印")
	}
}

// naYin 干支 → 纳音(六十甲子序)。
func naYin(stem, branch int) string {
	for j := 0; j < 60; j++ {
		if j%10 == stem && j%12 == branch {
			return naYinNames[j/2]
		}
	}
	return ""
}

// buildSiZhu 由四柱干支字符串构建四柱视角。
func buildSiZhu(fp FourPillars) *SiZhuView {
	pillarStrs := []string{fp.Year, fp.Month, fp.Day, fp.Hour}
	names := []string{"年柱", "月柱", "日柱", "时柱"}

	dayRunes := []rune(fp.Day)
	if len(dayRunes) != 2 {
		return nil
	}
	dayStem := stemIndex(dayRunes[0])
	if dayStem < 0 {
		return nil
	}

	view := &SiZhuView{
		DayMaster:        string(szStems[dayStem]),
		DayMasterElement: elementNames[stemElement[dayStem]],
		ElementCount:     map[string]int{"木": 0, "火": 0, "土": 0, "金": 0, "水": 0},
	}

	for i, ps := range pillarStrs {
		rs := []rune(ps)
		if len(rs) != 2 {
			return nil
		}
		s, b := stemIndex(rs[0]), branchIndex(rs[1])
		if s < 0 || b < 0 {
			return nil
		}
		p := SiZhuPillar{
			Name:          names[i],
			Stem:          string(rs[0]),
			Branch:        string(rs[1]),
			StemElement:   elementNames[stemElement[s]],
			BranchElement: elementNames[branchElement[b]],
			NaYin:         naYin(s, b),
		}
		if i == 2 {
			p.StemShiShen = "日主"
		} else {
			p.StemShiShen = shiShen(dayStem, s)
		}
		for _, hr := range hiddenStems[b] {
			hs := stemIndex(hr)
			p.Hidden = append(p.Hidden, HiddenStem{
				Stem:    string(hr),
				Element: elementNames[stemElement[hs]],
				ShiShen: shiShen(dayStem, hs),
			})
		}
		view.ElementCount[p.StemElement]++
		view.ElementCount[p.BranchElement]++
		view.Pillars[i] = p
	}
	return view
}
