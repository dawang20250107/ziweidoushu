package ai

// 单盘流年择时:扫描未来十年,推出健康预警年与财官催旺年。
//
// 与合盘婚嫁流年同理,以流年引动本命关键宫(疾厄/财帛/官禄)为信号:
//   - 健康预警——流年命宫行至疾厄、流羊流陀会照疾厄、流年化忌入疾厄;
//   - 财运催旺——流年命宫行至财帛、流禄流曜会照财帛、流年化禄入财帛;
//   - 事业升迁——流年命宫行至官禄、流魁钺会照官禄、流年化权/科入官禄。
// 倪师体系:流年四化(当年年干)为认可的动态层,据以择时。

import (
	"fmt"
	"strings"

	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// TimingYear 一个被引动的流年。
type TimingYear struct {
	Year   int    `json:"year"`
	GanZhi string `json:"ganZhi"`
	Note   string `json:"note"`
}

// liuYaoAround 某流限层在 branch 及其三方四正会照的流曜末字集合。
func liuYaoAround(scope ziwei.HoroscopeScope, branch int) map[rune]bool {
	set := map[rune]bool{}
	for _, off := range []int{0, 4, 6, 8} {
		bi := ((branch+off)%12 + 12) % 12
		if bi < len(scope.Stars) {
			for _, st := range scope.Stars[bi] {
				set[lastRune(st.Name)] = true
			}
		}
	}
	return set
}

// mutagenInPalace 流年某化星是否落于本命指定宫。
func mutagenInPalace(chart *ziwei.Chart, star, palaceName string) bool {
	if star == "" {
		return false
	}
	fp := findStarPalace(chart, star)
	return fp != nil && fp.Name == palaceName
}

// annualTiming 扫描 ReferenceYear 起未来十年,分出健康预警/财运/事业催旺三类年份。
func annualTiming(chart *ziwei.Chart) (health, wealth, career []TimingYear) {
	base := chart.ReferenceYear
	if base <= 0 {
		return nil, nil, nil
	}
	jie := chart.PalaceByName("疾厄")
	cai := chart.PalaceByName("财帛")
	guan := chart.PalaceByName("官禄")
	for y := base; y < base+10; y++ {
		h, err := ziwei.GenerateHoroscope(chart, y, 6, 15, 6)
		if err != nil {
			continue
		}
		yr := h.Yearly
		gz := ziwei.Stems[ziwei.YearStemIndex(y)] + ziwei.Branches[ziwei.YearBranchIndex(y)]

		// 健康预警。
		if jie != nil {
			score := 0
			var rs []string
			if yr.PalaceBranch == jie.Branch {
				score++
				rs = append(rs, "流年运程行至疾厄宫、以调养为重")
			}
			la := liuYaoAround(yr, jie.Branch)
			if la['羊'] || la['陀'] {
				score++
				rs = append(rs, "流羊陀会照疾厄、防外伤暗疾")
			}
			if mutagenInPalace(chart, yr.Mutagen[3], "疾厄") {
				score += 2
				rs = append(rs, "流年化忌入疾厄、防病厄是非")
			}
			if score >= 2 {
				health = append(health, TimingYear{Year: y, GanZhi: gz, Note: strings.Join(rs, ";")})
			}
		}

		// 财运催旺。
		if cai != nil {
			score := 0
			var rs []string
			if yr.PalaceBranch == cai.Branch {
				score++
				rs = append(rs, "流年运程行至财帛宫、财事为重心")
			}
			if la := liuYaoAround(yr, cai.Branch); la['禄'] {
				score++
				rs = append(rs, "流禄会照财帛、进财顺")
			}
			if mutagenInPalace(chart, yr.Mutagen[0], "财帛") || mutagenInPalace(chart, yr.Mutagen[0], "命宫") {
				score += 2
				rs = append(rs, "流年化禄入财帛/命、财源增益")
			}
			if score >= 2 {
				wealth = append(wealth, TimingYear{Year: y, GanZhi: gz, Note: strings.Join(rs, ";")})
			}
		}

		// 事业升迁。
		if guan != nil {
			score := 0
			var rs []string
			if yr.PalaceBranch == guan.Branch {
				score++
				rs = append(rs, "流年运程行至官禄宫、事业为重心")
			}
			if la := liuYaoAround(yr, guan.Branch); la['魁'] || la['钺'] {
				score++
				rs = append(rs, "流魁钺会照官禄、贵人提携")
			}
			if mutagenInPalace(chart, yr.Mutagen[1], "官禄") {
				score += 2
				rs = append(rs, "流年化权入官禄、掌权升迁")
			}
			if mutagenInPalace(chart, yr.Mutagen[2], "官禄") || mutagenInPalace(chart, yr.Mutagen[2], "命宫") {
				score++
				rs = append(rs, "流年化科入官禄/命、名声考试之喜")
			}
			if score >= 2 {
				career = append(career, TimingYear{Year: y, GanZhi: gz, Note: strings.Join(rs, ";")})
			}
		}
	}
	return health, wealth, career
}

// sectionForHealthTiming 健康预警年维度。
func sectionForHealthTiming(chart *ziwei.Chart, health []TimingYear) *ReadingSection {
	if chart.ReferenceYear <= 0 {
		return nil
	}
	level := "neutral"
	text := "未来十年疾厄宫无显著流年引动,健康大致平顺(仍宜定期体检、规律作息)。"
	if len(health) > 0 {
		level = "caution"
		text = "未来十年健康须留意的年份——" + joinTimingYears(health) + "。此等年份疾厄宫受流年引动,宜提前体检、注意作息与情绪,防患未然(命理提示不代医疗诊断)。"
	}
	return &ReadingSection{Key: "healthtiming", Title: "健康预警·流年", Level: level, Text: text}
}

// sectionForFortuneTiming 财官催旺年维度。
func sectionForFortuneTiming(chart *ziwei.Chart, wealth, career []TimingYear) *ReadingSection {
	if chart.ReferenceYear <= 0 {
		return nil
	}
	var b strings.Builder
	if len(wealth) > 0 {
		b.WriteString("财运催旺之年——" + joinTimingYears(wealth) + ";宜把握求财、投资、开源之机。")
	} else {
		b.WriteString("未来十年财帛宫无显著流年催旺,财运平稳,宜稳中积累。")
	}
	if len(career) > 0 {
		b.WriteString("事业升迁之年——" + joinTimingYears(career) + ";宜争取晋升、考试、开创之事。")
	} else {
		b.WriteString("事业无显著流年催旺,宜厚积待时、稳步经营。")
	}
	level := "neutral"
	if len(wealth) > 0 || len(career) > 0 {
		level = "good"
	}
	return &ReadingSection{Key: "fortunetiming", Title: "财官择时·流年", Level: level, Text: b.String()}
}

func joinTimingYears(ts []TimingYear) string {
	var parts []string
	for _, t := range ts {
		parts = append(parts, fmt.Sprintf("%d年(%s):%s", t.Year, t.GanZhi, t.Note))
	}
	return strings.Join(parts, ";")
}
