package ai

// 运限逐层断语:大限 → 流年 → 流月 → 流日 → 流时。
//
// 与本命多维断语互补——本命看一生格局,本层看某一时点的运势。读法遵倪师体系:
//   - 每层「命宫」落于本命何宫、坐何星曜,定该时段主轴(本命星曜为骨,恒定);
//   - 流曜(流禄/流羊/流马/流鸾…)会照本层命宫三方四正,添吉助或增波折;
//   - 流年四化(当年年干)为倪师认可的动态层,飞入本命宫位定当年动向;
//   - 流月/流日/流时的宫干四化属飞星派,仅附注参考、不作定断(与引擎口径一致)。

import (
	"fmt"
	"strings"

	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// HoroscopeReading 运限逐层断语。
type HoroscopeReading struct {
	Target   string           `json:"target"` // 目标公历日期
	Sections []ReadingSection `json:"sections"`
}

// liuYaoEffect 流曜(按名称末字)对本时段的定性。
var liuYaoEffect = map[rune]string{
	'禄': "流禄照,本期财禄进益、机遇顺遂",
	'魁': "流魁临,本期得贵人提携(昼贵)",
	'钺': "流钺临,本期得贵人暗助(夜贵)",
	'昌': "流昌照,利文书、考试、签约",
	'曲': "流曲照,利才艺、文书、口才之事",
	'马': "流马动,本期奔波变动、宜主动出行求变",
	'鸾': "流鸾照,本期主姻缘、婚庆、桃花",
	'喜': "流喜照,本期有喜庆、添丁进口之兆",
	'羊': "流羊冲,本期防刑伤、是非、血光,忌急躁",
	'陀': "流陀缠,本期防拖延、暗损、旧事纠缠",
}

// scopeUnit 层名 → 时段量词(用于「本X以此为主轴」)。
var scopeUnit = map[string]string{
	"大限": "十年", "童限": "童限", "流年": "年", "流月": "月", "流日": "日", "流时": "时辰",
}

// BuildHoroscopeReading 由本命盘 + 运限叠加生成逐层断语(确定性,不走 LLM)。
func (it *Interpreter) BuildHoroscopeReading(chart *ziwei.Chart, h *ziwei.Horoscope) *HoroscopeReading {
	return buildHoroscopeReading(chart, h)
}

func buildHoroscopeReading(chart *ziwei.Chart, h *ziwei.Horoscope) *HoroscopeReading {
	if chart == nil || h == nil {
		return nil
	}
	hr := &HoroscopeReading{Target: h.TargetSolarDate}
	add := func(key string, sc ziwei.HoroscopeScope, isYear bool) {
		if s := hscopeSection(chart, sc, key, isYear); s != nil {
			hr.Sections = append(hr.Sections, *s)
		}
	}
	add("decadal", h.Decadal, false)
	add("yearly", h.Yearly, true)
	add("monthly", h.Monthly, false)
	add("daily", h.Daily, false)
	add("hourly", h.Hourly, false)
	return hr
}

// hscopeSection 组装某一运限层的断语。
func hscopeSection(chart *ziwei.Chart, sc ziwei.HoroscopeScope, key string, isYear bool) *ReadingSection {
	p := chart.PalaceByBranch(sc.PalaceBranch)
	if p == nil {
		return nil
	}
	var b strings.Builder
	var tags []string
	score := 0

	unit := scopeUnit[sc.Name]
	if unit == "" {
		unit = "期"
	}
	b.WriteString(fmt.Sprintf("%s命宫落于本命【%s】", sc.Name, palaceLabel(p.Name)))
	majors, borrowed := palaceMajors(p)
	if len(majors) > 0 {
		var cl []string
		for _, n := range majors {
			cl = append(cl, n+"("+starTrait[n][0]+")")
		}
		lead := "坐"
		if borrowed {
			lead = "空借"
		}
		b.WriteString(fmt.Sprintf("(%s %s),本%s以此为主轴:%s。", lead, strings.Join(majors, "、"), unit, strings.Join(cl, ";")))
		if pn, pt := pairTraitOf(majors); pt != "" {
			b.WriteString(fmt.Sprintf("【%s】:%s", pn, firstSentence(pt)))
		}
		tags = append(tags, majors...)
	} else {
		b.WriteString(",本宫无正曜,借对宫参看,本" + unit + "宜守常、随三方之势。")
	}

	// 流曜会照(本层命宫 + 三方四正)。
	seen := map[string]bool{}
	for _, off := range []int{0, 4, 6, 8} {
		bi := ((sc.PalaceBranch+off)%12 + 12) % 12
		if bi >= len(sc.Stars) {
			continue
		}
		for _, st := range sc.Stars[bi] {
			r := lastRune(st.Name)
			eff, ok := liuYaoEffect[r]
			if !ok || seen[st.Name] {
				continue
			}
			seen[st.Name] = true
			b.WriteString(eff + ";")
			switch r {
			case '羊', '陀':
				score--
			case '禄', '魁', '钺', '昌', '曲', '喜':
				score++
			}
		}
	}

	if isYear {
		// 流年四化(倪师认可):飞入本命宫位定当年动向。
		b.WriteString("流年四化:")
		for _, hv := range []struct {
			h    ziwei.SiHua
			star string
		}{{ziwei.HuaLu, sc.Mutagen[0]}, {ziwei.HuaQuan, sc.Mutagen[1]}, {ziwei.HuaKe, sc.Mutagen[2]}, {ziwei.HuaJi, sc.Mutagen[3]}} {
			if hv.star == "" {
				continue
			}
			where := "本命盘外"
			if fp := findStarPalace(chart, hv.star); fp != nil {
				where = "本命" + palaceLabel(fp.Name)
			}
			b.WriteString(fmt.Sprintf("%s化%s飞入%s,%s;", hv.star, string(hv.h), where, liuNianHuaWord[hv.h]))
			if hv.h == ziwei.HuaJi {
				score--
			}
		}
	} else {
		b.WriteString(fmt.Sprintf("(%s以宫气与流曜为主;其宫干四化属飞星派,仅供参考不作定断)", sc.Name))
	}

	return &ReadingSection{
		Key: key, Title: sc.Name + "·" + sc.Stem + sc.Branch, Palace: p.Name,
		Stars: tags, Level: levelOf(score), Text: b.String(),
	}
}

// lastRune 取字符串末字(流曜按末字判吉凶)。
func lastRune(s string) rune {
	rs := []rune(s)
	if len(rs) == 0 {
		return 0
	}
	return rs[len(rs)-1]
}
