package ai

// 事项择吉引擎:按所问事项定对应宫,逐层择时——利年 → 利月 → 利日。
//
// 事项 → 宫映射(倪师三合派口径):考试签约看官禄、求财开业看财帛、婚嫁看夫妻、
// 搬迁置产看田宅、出行远行看迁移。以流年/流月/流日命宫行至该宫为主信号,
// 相关四化(年层认可)、流曜(禄马昌曲魁钺鸾喜)会照为加强,逢流羊陀/化忌则宜避让。
// 流月/流日的宫干四化属飞星派,故月日层只据命宫行至与流曜择时,不用四化。

import (
	"fmt"
	"strings"

	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// eventSpec 一类事项的择吉规格。
type eventSpec struct {
	Label   string        // 中文事项名
	Palace  string        // 对应本命宫
	Mutagen []ziwei.SiHua // 利此事的流年四化(飞入对应宫)
	Stars   []rune        // 利此事的流曜末字
}

// eventOrder 择吉事项顺序(前端下拉展示用)。
var eventOrder = []string{"exam", "contract", "wealth", "career", "marriage", "relocate", "travel"}

var eventSpecs = map[string]eventSpec{
	"exam":     {"考试·文书", "官禄", []ziwei.SiHua{ziwei.HuaKe, ziwei.HuaQuan}, []rune{'昌', '曲', '魁', '钺'}},
	"contract": {"签约·合作", "官禄", []ziwei.SiHua{ziwei.HuaLu, ziwei.HuaKe}, []rune{'昌', '曲', '魁', '钺'}},
	"wealth":   {"求财·开业", "财帛", []ziwei.SiHua{ziwei.HuaLu}, []rune{'禄', '马'}},
	"career":   {"求职·升迁", "官禄", []ziwei.SiHua{ziwei.HuaQuan, ziwei.HuaLu}, []rune{'魁', '钺'}},
	"marriage": {"婚嫁·求偶", "夫妻", []ziwei.SiHua{ziwei.HuaLu}, []rune{'鸾', '喜', '禄'}},
	"relocate": {"搬迁·置产", "田宅", []ziwei.SiHua{ziwei.HuaLu}, []rune{'禄'}},
	"travel":   {"出行·远行", "迁移", []ziwei.SiHua{ziwei.HuaLu}, []rune{'马', '禄'}},
}

// avoidSpec 忌事(避险)择时规格:找该宫逢煞忌之年月，宜避。
type avoidSpec struct {
	Label  string
	Palace string
}

var avoidOrder = []string{"surgery", "litigation", "travelrisk", "investrisk"}

var avoidSpecs = map[string]avoidSpec{
	"surgery":    {"手术·开刀", "疾厄"},
	"litigation": {"诉讼·官非", "官禄"},
	"travelrisk": {"远行·避险", "迁移"},
	"investrisk": {"投资·防破", "财帛"},
}

// EventCatalogItem 事项目录项(供前端渲染下拉)。
type EventCatalogItem struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Palace string `json:"palace"`
	Kind   string `json:"kind"` // auspicious=择吉 / avoid=避忌
}

// EventCatalog 全部事项(先择吉后避忌)。
func EventCatalog() []EventCatalogItem {
	out := make([]EventCatalogItem, 0, len(eventOrder)+len(avoidOrder))
	for _, k := range eventOrder {
		s := eventSpecs[k]
		out = append(out, EventCatalogItem{Key: k, Label: s.Label, Palace: s.Palace, Kind: "auspicious"})
	}
	for _, k := range avoidOrder {
		s := avoidSpecs[k]
		out = append(out, EventCatalogItem{Key: k, Label: s.Label, Palace: s.Palace, Kind: "avoid"})
	}
	return out
}

// natalPalaceQuality 本命某宫的吉凶底分(与逐宫综论同源:庙陷+四化+煞吉+应验档)。
func natalPalaceQuality(chart *ziwei.Chart, p *ziwei.Palace) int {
	majors, borrowed := palaceMajors(p)
	bright, dim, sihua, sha, lucky := scanPalace(p)
	score := 0
	for _, name := range majors {
		if bright[name] {
			score += 2
		} else if dim[name] {
			score -= 2
		} else {
			score++
		}
		if h, ok := sihua[name]; ok {
			if h == ziwei.HuaJi {
				score -= 3
			} else {
				score += 2
			}
		}
	}
	score += len(lucky) - len(sha)
	if !borrowed {
		_, d := escalations(chart, p, majors, dim, sihua)
		score += d
	}
	return score
}

// baseQualityNote 本命底色对择吉/避忌的加权提示。
func baseQualityNote(palace string, q int, avoid bool) (word, note string) {
	switch {
	case q >= 3:
		word = "佳"
	case q <= -2:
		word = "弱"
	default:
		word = "中"
	}
	if avoid {
		switch word {
		case "弱":
			note = fmt.Sprintf("本命%s宫本偏弱(落陷/逢煞忌),忌年忌月尤须谨慎、能避则避。", palace)
		case "佳":
			note = fmt.Sprintf("本命%s宫根基尚佳,纵逢忌时亦有缓冲,仍以稳妥为上。", palace)
		default:
			note = fmt.Sprintf("本命%s宫中平,忌时避其锋、平时无需过虑。", palace)
		}
		return
	}
	switch word {
	case "佳":
		note = fmt.Sprintf("本命%s宫根基佳(庙旺/得吉),此事本有底气,逢利时更易水到渠成。", palace)
	case "弱":
		note = fmt.Sprintf("本命%s宫偏弱(落陷/逢煞忌),纵逢利时亦须加倍用心、量力而为,择吉为辅。", palace)
	default:
		note = fmt.Sprintf("本命%s宫中平,成事在人为,利时把握、平时经营。", palace)
	}
	return
}

// EventTiming 某事项的择时结果(择吉或避忌)。
type EventTiming struct {
	Event       string       `json:"event"`
	Palace      string       `json:"palace"`
	Kind        string       `json:"kind"` // auspicious / avoid
	Summary     string       `json:"summary"`
	Years       []TimingYear `json:"years"`               // 择吉=利年 / 避忌=忌年
	BestMonth   string       `json:"bestMonth,omitempty"` // 择吉=利月 / 避忌=忌月
	BestDays    []string     `json:"bestDays,omitempty"`  // 择吉=利日
	BaseQuality string       `json:"baseQuality"`         // 本命该宫底色:佳/中/弱
	BaseNote    string       `json:"baseNote"`
	Advice      string       `json:"advice"`
}

// BuildEventTiming 事项择时:择吉(利年→利月→利日)或避忌(忌年→忌月)。
func (it *Interpreter) BuildEventTiming(chart *ziwei.Chart, eventKey string) *EventTiming {
	if _, ok := avoidSpecs[eventKey]; ok {
		return buildAvoidTiming(chart, eventKey)
	}
	return buildEventTiming(chart, eventKey)
}

func mutagenStar(mut [4]string, h ziwei.SiHua) string {
	switch h {
	case ziwei.HuaLu:
		return mut[0]
	case ziwei.HuaQuan:
		return mut[1]
	case ziwei.HuaKe:
		return mut[2]
	case ziwei.HuaJi:
		return mut[3]
	}
	return ""
}

func buildEventTiming(chart *ziwei.Chart, eventKey string) *EventTiming {
	spec, ok := eventSpecs[eventKey]
	if !ok || chart == nil {
		return nil
	}
	target := chart.PalaceByName(spec.Palace)
	base := chart.ReferenceYear
	if target == nil || base <= 0 {
		return nil
	}

	et := &EventTiming{Event: spec.Label, Palace: spec.Palace, Kind: "auspicious"}
	et.BaseQuality, et.BaseNote = baseQualityNote(spec.Palace, natalPalaceQuality(chart, target), false)

	// 利年:扫未来十年。
	for y := base; y < base+10; y++ {
		h, err := ziwei.GenerateHoroscope(chart, y, 6, 15, 6)
		if err != nil {
			continue
		}
		score, rs := yearEventScore(chart, h.Yearly, target.Branch, spec)
		if dp := chart.PalaceByBranch(h.Decadal.PalaceBranch); dp != nil && dp.Name == spec.Palace {
			score++
			rs = append(rs, "大限亦行"+spec.Palace+"乡、尤验")
		}
		if score >= 2 {
			gz := ziwei.Stems[ziwei.YearStemIndex(y)] + ziwei.Branches[ziwei.YearBranchIndex(y)]
			et.Years = append(et.Years, TimingYear{Year: y, GanZhi: gz, Note: strings.Join(rs, ";")})
		}
	}

	// 利月/利日:就近利年(无则本年)内下钻。
	dy := base
	if len(et.Years) > 0 {
		dy = et.Years[0].Year
	}
	bestMonth := 0
	bestScore := 0
	for _, md := range monthDaySamples() {
		h, err := ziwei.GenerateHoroscope(chart, dy, md[0], md[1], 6)
		if err != nil {
			continue
		}
		s := scopeEventScore(h.Monthly, target.Branch, spec)
		if s > bestScore {
			bestScore, bestMonth = s, md[0]
		}
	}
	if bestMonth > 0 {
		et.BestMonth = fmt.Sprintf("%d 年阳历 %d 月前后", dy, bestMonth)
		for d := 1; d <= 28 && len(et.BestDays) < 3; d++ {
			h, err := ziwei.GenerateHoroscope(chart, dy, bestMonth, d, 6)
			if err != nil {
				continue
			}
			if h.Daily.PalaceBranch == target.Branch {
				extra := ""
				if la := liuYaoAround(h.Daily, target.Branch); anyRune(la, spec.Stars) {
					extra = "、流曜助"
				}
				et.BestDays = append(et.BestDays, fmt.Sprintf("阳历 %d 月 %d 日前后(流日命宫行%s%s)", bestMonth, d, spec.Palace, extra))
			}
		}
	}

	// 概述与建议。
	if len(et.Years) > 0 {
		et.Summary = fmt.Sprintf("为「%s」择吉(应事之宫:%s)。未来十年有利年 %d 个,就近可循利月利日行事。", spec.Label, spec.Palace, len(et.Years))
	} else {
		et.Summary = fmt.Sprintf("为「%s」择吉(应事之宫:%s)。未来十年%s宫无强流年引动,宜择本年利月利日、或俟大限行至此乡之期。", spec.Label, spec.Palace, spec.Palace)
	}
	et.Advice = eventAdvice(spec, len(et.Years) > 0)
	return et
}

// buildAvoidTiming 忌事择时:找该宫逢流年化忌/流羊陀之年月,宜避。
func buildAvoidTiming(chart *ziwei.Chart, eventKey string) *EventTiming {
	spec, ok := avoidSpecs[eventKey]
	if !ok || chart == nil {
		return nil
	}
	target := chart.PalaceByName(spec.Palace)
	base := chart.ReferenceYear
	if target == nil || base <= 0 {
		return nil
	}
	et := &EventTiming{Event: spec.Label, Palace: spec.Palace, Kind: "avoid"}
	et.BaseQuality, et.BaseNote = baseQualityNote(spec.Palace, natalPalaceQuality(chart, target), true)

	for y := base; y < base+10; y++ {
		h, err := ziwei.GenerateHoroscope(chart, y, 6, 15, 6)
		if err != nil {
			continue
		}
		yr := h.Yearly
		score := 0
		var rs []string
		if star := mutagenStar(yr.Mutagen, ziwei.HuaJi); star != "" && mutagenInPalace(chart, star, spec.Palace) {
			score += 2
			rs = append(rs, "流年化忌入"+spec.Palace)
		}
		if la := liuYaoAround(yr, target.Branch); la['羊'] || la['陀'] {
			score++
			rs = append(rs, "流羊陀会照"+spec.Palace)
		}
		if yr.PalaceBranch == target.Branch && score > 0 {
			rs = append(rs, "流年运程亦行至此宫、感受更切")
		}
		if score >= 2 {
			gz := ziwei.Stems[ziwei.YearStemIndex(y)] + ziwei.Branches[ziwei.YearBranchIndex(y)]
			et.Years = append(et.Years, TimingYear{Year: y, GanZhi: gz, Note: strings.Join(rs, ";")})
		}
	}

	// 忌月:就近忌年(无则本年)内该宫逢煞忌之月。
	dy := base
	if len(et.Years) > 0 {
		dy = et.Years[0].Year
	}
	worstMonth, worst := 0, 0
	for _, md := range monthDaySamples() {
		h, err := ziwei.GenerateHoroscope(chart, dy, md[0], md[1], 6)
		if err != nil {
			continue
		}
		s := 0
		if h.Monthly.PalaceBranch == target.Branch {
			s++
		}
		if la := liuYaoAround(h.Monthly, target.Branch); la['羊'] || la['陀'] {
			s += 2
		}
		if s > worst {
			worst, worstMonth = s, md[0]
		}
	}
	if worstMonth > 0 && worst >= 2 {
		et.BestMonth = fmt.Sprintf("%d 年阳历 %d 月前后(该宫逢煞、宜避)", dy, worstMonth)
	}

	if len(et.Years) > 0 {
		et.Summary = fmt.Sprintf("为「%s」避忌(应事之宫:%s)。未来十年宜避之年 %d 个,此事若可择期,避开为上、不得已则加倍谨慎。", spec.Label, spec.Palace, len(et.Years))
	} else {
		et.Summary = fmt.Sprintf("为「%s」避忌(应事之宫:%s)。未来十年%s宫无强煞忌引动,风险平平,循常谨慎即可。", spec.Label, spec.Palace, spec.Palace)
	}
	et.Advice = avoidAdvice(spec.Palace) + et.BaseNote
	return et
}

func avoidAdvice(palace string) string {
	switch palace {
	case "疾厄":
		return "忌年忌月宜避择期手术、慎防意外与旧疾复发,平日强健体魄、定期检查(不代医疗诊断);"
	case "官禄":
		return "忌年忌月慎涉诉讼、纠纷、担保背书,合约条款严审、留证据、少与人争;"
	case "迁移":
		return "忌年忌月远行宜谨慎,注意交通与人身安全、避险地险事、备应急;"
	case "财帛":
		return "忌年忌月忌大额投资、投机、借贷担保,守成为上、防被骗破财;"
	default:
		return "忌年忌月宜避其锋、稳妥行事;"
	}
}

// yearEventScore 流年层评分(四化为倪师认可,参与计分)。
func yearEventScore(chart *ziwei.Chart, y ziwei.HoroscopeScope, targetBranch int, spec eventSpec) (int, []string) {
	score := 0
	var rs []string
	if y.PalaceBranch == targetBranch {
		score++
		rs = append(rs, "流年运程行至"+spec.Palace+"宫")
	}
	la := liuYaoAround(y, targetBranch)
	for _, r := range spec.Stars {
		if la[r] {
			score++
			rs = append(rs, "流"+string(r)+"会照"+spec.Palace)
			break // 流曜加分至多一次,免过度堆叠
		}
	}
	for _, h := range spec.Mutagen {
		if star := mutagenStar(y.Mutagen, h); star != "" && mutagenInPalace(chart, star, spec.Palace) {
			score += 2
			rs = append(rs, "流年化"+string(h)+"入"+spec.Palace)
		}
	}
	if la['羊'] || la['陀'] {
		score--
		rs = append(rs, "(逢流羊陀,宜择清净时或另择年)")
	}
	if star := mutagenStar(y.Mutagen, ziwei.HuaJi); star != "" && mutagenInPalace(chart, star, spec.Palace) {
		score--
		rs = append(rs, "(流年化忌扰"+spec.Palace+",此年宜缓)")
	}
	return score, rs
}

// scopeEventScore 流月/流日层评分(不用宫干四化,只据命宫行至 + 流曜)。
func scopeEventScore(sc ziwei.HoroscopeScope, targetBranch int, spec eventSpec) int {
	score := 0
	if sc.PalaceBranch == targetBranch {
		score += 2
	}
	la := liuYaoAround(sc, targetBranch)
	if anyRune(la, spec.Stars) {
		score++
	}
	if la['羊'] || la['陀'] {
		score--
	}
	return score
}

func anyRune(set map[rune]bool, rs []rune) bool {
	for _, r := range rs {
		if set[r] {
			return true
		}
	}
	return false
}

func eventAdvice(spec eventSpec, hasYears bool) string {
	base := map[string]string{
		"官禄": "利年利月宜争取考试、面谈、签署、晋升;文昌文曲魁钺之日尤佳。",
		"财帛": "利年利月宜开业、投资、谈价、进货;禄星之日进财顺。",
		"夫妻": "利年利月宜订婚、成婚、告白、见家长;红鸾天喜之日尤吉。",
		"田宅": "利年利月宜置产、搬迁、动土、装修;宜避冲煞之日。",
		"迁移": "利年利月宜出行、远行、移居、外派;天马之日利动。",
	}[spec.Palace]
	if !hasYears {
		return "近十年虽无强引动,仍可于本年利月利日行事,或俟大限行至此乡再图大举。" + base
	}
	return base + "凡事仍以本人诚意与准备为本,择吉为辅、不可尽恃。"
}
