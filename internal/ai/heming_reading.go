package ai

// 合盘(双人)确定性断语骨架:不走 LLM,由两张命盘的可计算特征推出婚配契合度。
//
// 三条可机械化的合盘核心信号(倪海厦三合派口径):
//   1. 年命相合——双方年支的六合/三合/相冲/相刑/相害(生肖气类);
//   2. 四化互飞——一方生年四化飞入对方命宫/夫妻宫(禄科情深有助,忌则牵绊挑剔);
//   3. 夫妻宫呼应——一方夫妻宫星曜正应对方命宫(对方近其命定的另一半)。
// 三项加权成契合分,并给出分向断语。内容为公版通则原创提炼(非第三方文本)。

import (
	"fmt"
	"strings"

	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// HemingReading 合盘确定性断语。
type HemingReading struct {
	Score    int              `json:"score"` // 20-100 契合度
	Level    string           `json:"level"` // 上上缘/上等姻缘/中上可成/中平宜经营/宜慎重
	Summary  string           `json:"summary"`
	Sections []ReadingSection `json:"sections"`
	Timing   []HemingYear     `json:"timing,omitempty"` // 婚嫁共振流年
}

// HemingYear 双方婚庆信号共振的流年。
type HemingYear struct {
	Year   int    `json:"year"`
	GanZhi string `json:"ganZhi"`
	Note   string `json:"note"`
}

// BuildHemingReading 由甲乙双方命盘生成合盘契合断语。
func (it *Interpreter) BuildHemingReading(a, b *ziwei.Chart) *HemingReading {
	return buildHemingReading(a, b)
}

func buildHemingReading(a, b *ziwei.Chart) *HemingReading {
	if a == nil || b == nil {
		return nil
	}
	score := 60
	var secs []ReadingSection

	relText, relDelta := yearBranchRelation(a.LunarInfo.YearBranch, b.LunarInfo.YearBranch)
	score += relDelta
	secs = append(secs, ReadingSection{Key: "nianming", Title: "年命相合", Level: deltaLevel(relDelta), Text: relText})

	flyText, flyDelta, flyHasJi := sihuaCross(a, b)
	score += flyDelta
	flyLevel := deltaLevel(flyDelta)
	if flyHasJi && flyLevel == "neutral" { // 禄忌相抵仍有忌,提示留意
		flyLevel = "caution"
	}
	secs = append(secs, ReadingSection{Key: "sihuafly", Title: "四化互飞", Level: flyLevel, Text: flyText})

	echoText, echoDelta := fuqiEcho(a, b)
	score += echoDelta
	secs = append(secs, ReadingSection{Key: "echo", Title: "夫妻宫呼应", Level: deltaLevel(echoDelta), Text: echoText})

	if score > 100 {
		score = 100
	}
	if score < 20 {
		score = 20
	}
	// 婚嫁流年:未来十年双方婚庆信号共振之年。
	timing := hemingTiming(a, b)
	secs = append(secs, ReadingSection{Key: "timing", Title: "婚嫁流年", Level: timingLevel(timing), Text: timingText(timing)})

	level := hemingLevel(score)
	summary := fmt.Sprintf("合盘契合度 %d 分(%s)。%s三项合参:年命之气、四化互飞之情、夫妻宫之应,吉则顺缘、不足处以后天经营补之。",
		score, level, strongestSignal(relDelta, flyDelta, echoDelta))
	secs = append(secs, ReadingSection{Key: "advice", Title: "相处建议", Level: "neutral", Text: hemingAdvice(score, relDelta, flyDelta, echoDelta, flyHasJi)})

	return &HemingReading{Score: score, Level: level, Summary: summary, Sections: secs, Timing: timing}
}

// personMarriageYear 某人某年的婚嫁指数 + 触发说明。
// 信号:流年命宫入夫妻宫、流鸾流喜照命/夫妻、流年化禄入夫妻(化忌扰夫妻则减)。
func personMarriageYear(chart *ziwei.Chart, year int) (score int, reasons []string) {
	h, err := ziwei.GenerateHoroscope(chart, year, 6, 15, 6)
	if err != nil {
		return 0, nil
	}
	fuqi := chart.PalaceByName("夫妻")
	ming := chart.MingGong()
	if fuqi == nil || ming == nil {
		return 0, nil
	}
	y := h.Yearly
	if y.PalaceBranch == fuqi.Branch {
		score += 2
		reasons = append(reasons, "流年行至夫妻宫")
	}
	luanXi := false
	for bi := 0; bi < len(y.Stars); bi++ {
		if bi != fuqi.Branch && bi != ming.Branch {
			continue
		}
		for _, st := range y.Stars[bi] {
			if r := lastRune(st.Name); r == '鸾' || r == '喜' {
				luanXi = true
			}
		}
	}
	if luanXi {
		score += 2
		reasons = append(reasons, "流年红鸾天喜照命/夫妻")
	}
	if s := y.Mutagen[0]; s != "" { // 流年化禄
		if fp := findStarPalace(chart, s); fp != nil && fp.Name == "夫妻" {
			score++
			reasons = append(reasons, "流年化禄入夫妻宫")
		}
	}
	if s := y.Mutagen[3]; s != "" { // 流年化忌
		if fp := findStarPalace(chart, s); fp != nil && fp.Name == "夫妻" {
			score--
			reasons = append(reasons, "流年化忌扰夫妻(婚事宜缓)")
		}
	}
	return score, reasons
}

// hemingTiming 未来十年双方婚庆信号共振之年(双方指数均达标)。
func hemingTiming(a, b *ziwei.Chart) []HemingYear {
	base := a.ReferenceYear
	if base <= 0 {
		return nil
	}
	var out []HemingYear
	for y := base; y < base+10; y++ {
		sa, ra := personMarriageYear(a, y)
		sb, rb := personMarriageYear(b, y)
		if sa >= 2 && sb >= 2 {
			gz := ziwei.Stems[ziwei.YearStemIndex(y)] + ziwei.Branches[ziwei.YearBranchIndex(y)]
			out = append(out, HemingYear{
				Year:   y,
				GanZhi: gz,
				Note:   fmt.Sprintf("甲方%s;乙方%s", strings.Join(ra, "、"), strings.Join(rb, "、")),
			})
		}
	}
	return out
}

func timingLevel(ts []HemingYear) string {
	if len(ts) > 0 {
		return "good"
	}
	return "neutral"
}

func timingText(ts []HemingYear) string {
	if len(ts) == 0 {
		return "未来十年双方婚庆信号未见强烈共振,姻缘更需主动经营;亦可各自参看流年催旺(红鸾天喜、流年行至夫妻宫之年)。"
	}
	var ys []string
	for _, t := range ts {
		ys = append(ys, fmt.Sprintf("%d年(%s):%s", t.Year, t.GanZhi, t.Note))
	}
	return "未来十年双方婚庆信号共振、宜把握的年份:" + strings.Join(ys, ";") + "。此为双方红鸾天喜、流年行至夫妻宫等信号叠合之年,利订婚、成婚、感情升温。"
}

// yearBranchRelation 年支关系:六合/三合/相冲/相刑/相害/同支/中性。
func yearBranchRelation(b1, b2 int) (text string, delta int) {
	b1 = ((b1 % 12) + 12) % 12
	b2 = ((b2 % 12) + 12) % 12
	z := ziwei.Branches
	switch {
	case liuHe(b1, b2):
		return fmt.Sprintf("双方年支%s%s六合,天然投契、气类相求,相处易生默契、逢难能互补,是姻缘的良好底色。", z[b1], z[b2]), 12
	case b1 != b2 && b1%4 == b2%4:
		return fmt.Sprintf("双方年支%s%s同属三合局,志趣相近、易成同盟,相辅相成、聚合有力。", z[b1], z[b2]), 10
	case (b1+6)%12 == b2:
		return fmt.Sprintf("双方年支%s%s相冲,个性、节奏差异大,易生摩擦口角,须多包容退让、忌硬碰硬（冲亦主动、能彼此激发,善处则为助力）。", z[b1], z[b2]), -12
	case xiangXing(b1, b2):
		return fmt.Sprintf("双方年支%s%s相刑,相处中易有磨擦与心结,宜坦诚沟通、勿积怨。", z[b1], z[b2]), -8
	case xiangHai(b1, b2):
		return fmt.Sprintf("双方年支%s%s相害,易因小事暗生嫌隙,宜多体谅、勿钻牛角尖。", z[b1], z[b2]), -6
	case b1 == b2:
		return fmt.Sprintf("双方同属%s,气类相同、易相知,然亦易同其所短,须互相提醒补足。", z[b1]), 2
	default:
		return fmt.Sprintf("双方年支%s%s无显著刑冲会合,年命之气中性,契合更看四化互飞与夫妻宫呼应。", z[b1], z[b2]), 0
	}
}

func liuHe(b1, b2 int) bool {
	if b1 == b2 {
		return false
	}
	s := b1 + b2
	return s == 1 || s == 13
}

// xiangXing 相刑:子卯无礼、寅巳申无恩、丑戌未恃势、辰午酉亥自刑。
func xiangXing(b1, b2 int) bool {
	groups := [][]int{{0, 3}, {2, 5, 8}, {1, 10, 7}}
	for _, g := range groups {
		if inInts(g, b1) && inInts(g, b2) && b1 != b2 {
			return true
		}
	}
	self := map[int]bool{4: true, 6: true, 9: true, 11: true}
	return b1 == b2 && self[b1]
}

// xiangHai 相害:子未、丑午、寅巳、卯辰、申亥、酉戌。
func xiangHai(b1, b2 int) bool {
	pairs := [][2]int{{0, 7}, {1, 6}, {2, 5}, {3, 4}, {8, 11}, {9, 10}}
	for _, p := range pairs {
		if (b1 == p[0] && b2 == p[1]) || (b1 == p[1] && b2 == p[0]) {
			return true
		}
	}
	return false
}

func inInts(list []int, v int) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// sihuaCross 四化互飞:一方生年四化飞入对方命宫/夫妻宫。
func sihuaCross(a, b *ziwei.Chart) (text string, delta int, hasJi bool) {
	var parts []string
	oneWay := func(src, dst *ziwei.Chart, srcName, dstName string) int {
		set := ziwei.SiHuaByStem(src.LunarInfo.YearStem)
		d := 0
		for _, hv := range []struct {
			h    ziwei.SiHua
			star string
		}{{ziwei.HuaLu, set.Lu}, {ziwei.HuaQuan, set.Quan}, {ziwei.HuaKe, set.Ke}, {ziwei.HuaJi, set.Ji}} {
			if hv.star == "" {
				continue
			}
			fp := findStarPalace(dst, hv.star)
			if fp == nil || (fp.Name != "命宫" && fp.Name != "夫妻") {
				continue
			}
			switch hv.h {
			case ziwei.HuaLu:
				parts = append(parts, fmt.Sprintf("%s生年%s化禄飞入%s的%s,%s对%s情深有助、乐于付出", srcName, hv.star, dstName, palaceLabel(fp.Name), srcName, dstName))
				d += 8
			case ziwei.HuaKe:
				parts = append(parts, fmt.Sprintf("%s%s化科飞入%s的%s,相敬有礼、名分和顺", srcName, hv.star, dstName, palaceLabel(fp.Name)))
				d += 5
			case ziwei.HuaQuan:
				parts = append(parts, fmt.Sprintf("%s%s化权飞入%s的%s,%s于%s较主导,双刃宜善用", srcName, hv.star, dstName, palaceLabel(fp.Name), srcName, dstName))
				d += 3
			case ziwei.HuaJi:
				parts = append(parts, fmt.Sprintf("%s生年%s化忌飞入%s的%s,%s对%s易挑剔牵绊、放不下,须自觉察", srcName, hv.star, dstName, palaceLabel(fp.Name), srcName, dstName))
				d -= 8
				hasJi = true
			}
		}
		return d
	}
	delta += oneWay(a, b, "甲方", "乙方")
	delta += oneWay(b, a, "乙方", "甲方")
	if len(parts) == 0 {
		return "双方生年四化未直接飞入对方命宫或夫妻宫,情感牵引不强不弱,缘分更靠日常经营。", 0, false
	}
	return strings.Join(parts, ";") + "。", delta, hasJi
}

// fuqiEcho 夫妻宫呼应:一方夫妻宫主星是否正应对方命宫主星。
func fuqiEcho(a, b *ziwei.Chart) (text string, delta int) {
	aFuqi := majorsOfPalace(a, "夫妻")
	bFuqi := majorsOfPalace(b, "夫妻")
	aMing := majorsOfPalace(a, "命宫")
	bMing := majorsOfPalace(b, "命宫")

	var parts []string
	if hit := intersect(aFuqi, bMing); len(hit) > 0 {
		parts = append(parts, fmt.Sprintf("甲方夫妻宫见 %s,正应乙方命宫——乙方近甲方命定的另一半样貌", strings.Join(hit, "、")))
		delta += 8
	}
	if hit := intersect(bFuqi, aMing); len(hit) > 0 {
		parts = append(parts, fmt.Sprintf("乙方夫妻宫见 %s,正应甲方命宫——甲方近乙方命定的另一半样貌", strings.Join(hit, "、")))
		delta += 8
	}
	if len(parts) == 0 {
		return "双方夫妻宫星曜与对方命宫星系不同,并非彼此第一眼的理想型,重在后天了解、以互补相守。", 0
	}
	if len(parts) == 2 {
		return strings.Join(parts, ";") + "——双向呼应,尤为难得,主一见倾心、相守相知。", delta
	}
	return parts[0] + "。", delta
}

// majorsOfPalace 取某宫主星(空宫借对宫)。
func majorsOfPalace(c *ziwei.Chart, name string) []string {
	p := c.PalaceByName(name)
	if p == nil {
		return nil
	}
	m, _ := palaceMajors(p)
	return m
}

func intersect(a, b []string) []string {
	set := map[string]bool{}
	for _, x := range b {
		set[x] = true
	}
	var out []string
	for _, x := range a {
		if set[x] {
			out = append(out, x)
		}
	}
	return out
}

func deltaLevel(d int) string {
	switch {
	case d > 0:
		return "good"
	case d < 0:
		return "caution"
	default:
		return "neutral"
	}
}

func hemingLevel(score int) string {
	switch {
	case score >= 85:
		return "上上缘"
	case score >= 72:
		return "上等姻缘"
	case score >= 58:
		return "中上可成"
	case score >= 45:
		return "中平宜经营"
	default:
		return "宜慎重"
	}
}

func strongestSignal(rel, fly, echo int) string {
	if fly <= -8 || rel <= -12 {
		return "有须留意的冲忌之处，"
	}
	if fly >= 8 || echo >= 8 || rel >= 10 {
		return "有天成的相契之处，"
	}
	return ""
}

func hemingAdvice(score, rel, fly, echo int, flyHasJi bool) string {
	var b strings.Builder
	if rel < 0 {
		b.WriteString("年命有刑冲害,遇分歧宜先退一步、就事论事,勿翻旧账;")
	}
	if flyHasJi {
		b.WriteString("四化互飞见忌,一方对另一方易过度在意或挑剔,当觉察此为『在乎』之变形,化苛求为体谅;")
	}
	if echo > 0 {
		b.WriteString("夫妻宫有呼应,珍惜这份天然的吸引,多创造相处与共同目标;")
	}
	if rel >= 0 && !flyHasJi && echo >= 0 {
		b.WriteString("三项皆无破,底子和顺,平实经营即可长久;")
	}
	if score >= 72 {
		b.WriteString("整体契合佳,宜顺缘把握。")
	} else if score >= 58 {
		b.WriteString("整体中上可成,用心经营能得美满。")
	} else {
		b.WriteString("整体宜慎重,须双方都有磨合的诚意与耐心方能长久。")
	}
	return b.String()
}
