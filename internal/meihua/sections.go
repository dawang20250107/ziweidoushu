// 梅花易数分节深断层:卦象总论/体用之辨/过程与结局/类象取应/应期(确定性)。
// 义理依邵子《梅花易数》体用总诀归纳、原创行文;只做呈现纵深,
// 不改 Judge 的推演与吉凶评分。
package meihua

import (
	"fmt"
	"strings"
)

// JudgeSection 断语分节(免费确定性层的呈现单元)。
type JudgeSection struct {
	Key   string `json:"key"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

// wuxingShengKe 五行生克白话链(a 对 b)。
func wuxingShengKe(a, b string) string {
	idx := map[string]int{"木": 0, "火": 1, "土": 2, "金": 3, "水": 4}
	ai, aok := idx[a]
	bi, bok := idx[b]
	if !aok || !bok || a == b {
		return a + b + "同气比和"
	}
	chain := map[[2]int]string{
		{0, 1}: "木生火", {1, 2}: "火生土", {2, 3}: "土生金", {3, 4}: "金生水", {4, 0}: "水生木",
		{0, 2}: "木克土", {2, 4}: "土克水", {4, 1}: "水克火", {1, 3}: "火克金", {3, 0}: "金克木",
	}
	if s, ok := chain[[2]int{ai, bi}]; ok {
		return s
	}
	if s, ok := chain[[2]int{bi, ai}]; ok {
		return s
	}
	return ""
}

// buildSections 组装分节深断(Judge 推演完毕后调用)。
func (r *Result) buildSections(j *Judgment) {
	// ── 卦象总论:上下卦象组合 + 卦辞 + 总断 ──
	var zl strings.Builder
	zl.WriteString(fmt.Sprintf("本卦【%s】,上%s(%s)下%s(%s)——%s之上有%s,此其象也。",
		r.Ben.Name, r.Ben.Upper.Name, r.Ben.Upper.Nature, r.Ben.Lower.Name, r.Ben.Lower.Nature,
		r.Ben.Lower.Nature, r.Ben.Upper.Nature))
	if r.Ben.GuaCi != "" {
		zl.WriteString(fmt.Sprintf("《周易》%s卦辞:「%s」——经文为骨,断辞扣此而发。", r.Ben.Name, r.Ben.GuaCi))
	}
	zl.WriteString(fmt.Sprintf("第 %d 爻动,变出【%s】;动之所在,事之枢机所在。", r.Moving, r.Bian.Name))
	j.Sections = append(j.Sections, JudgeSection{Key: "zonglun", Title: "卦象总论", Text: zl.String()})

	// ── 体用之辨:体用定位 + 五行生克链条 + 卦气旺衰 ──
	var ty strings.Builder
	side := "下卦"
	if r.TiIsUpper {
		side = "上卦"
	}
	ty.WriteString(fmt.Sprintf("动爻在%s之侧,故以不动之%s【%s·%s】为体(我方、事主),动侧【%s·%s】为用(彼方、事应)。",
		map[bool]string{true: "下", false: "上"}[!r.TiIsUpper]+"卦", side,
		r.TiTrigram.Name, r.TiTrigram.Element, r.YongTrigram.Name, r.YongTrigram.Element))
	if sk := wuxingShengKe(r.TiTrigram.Element, r.YongTrigram.Element); sk != "" {
		ty.WriteString(fmt.Sprintf("五行之链:%s——故为「%s」,%s。", sk, r.Relation, r.Verdict))
	}
	ty.WriteString(fmt.Sprintf("体卦月令之气【%s】:气旺则纵受克而有抵,气衰则虽得生而受之有限——旺衰是体用生克的放大器。", j.TiQi))
	j.Sections = append(j.Sections, JudgeSection{Key: "tiyong", Title: "体用之辨", Text: ty.String()})

	// ── 过程与结局:互卦中段、变卦归宿(引 Judge 已推演的党势细节) ──
	var gj strings.Builder
	gj.WriteString(fmt.Sprintf("互卦【%s】藏于本卦之中,主事之中段经过;", r.Hu.Name))
	gj.WriteString(fmt.Sprintf("变卦【%s】为动极之归,主事之结局。", r.Bian.Name))
	if r.Bian.GuaCi != "" {
		gj.WriteString(fmt.Sprintf("变卦卦辞:「%s」——结局之象扣此参详。", r.Bian.GuaCi))
	}
	for _, p := range j.Points {
		if strings.HasPrefix(p, "过程(互卦") || strings.HasPrefix(p, "结局(变卦") {
			gj.WriteString(p)
			if !strings.HasSuffix(p, "。") {
				gj.WriteString("。")
			}
		}
	}
	j.Sections = append(j.Sections, JudgeSection{Key: "guocheng", Title: "过程与结局", Text: gj.String()})

	// ── 类象取应:体用之卦落到人事物(邵子八卦类占义) ──
	var lx strings.Builder
	if lore, ok := LoreOf(r.TiTrigram.Name); ok {
		lx.WriteString(fmt.Sprintf("体卦%s之类象:人物应%s;身体应%s;器物应%s;方位在%s。",
			r.TiTrigram.Name, lore.Renlun, lore.Shenti, lore.Jingwu, lore.Fangwei))
	}
	if lore, ok := LoreOf(r.YongTrigram.Name); ok {
		lx.WriteString(fmt.Sprintf("用卦%s之类象:人物应%s;器物应%s;方位在%s。",
			r.YongTrigram.Name, lore.Renlun, lore.Jingwu, lore.Fangwei))
	}
	lx.WriteString("类象之用:所问之事中若见此类人、此类物、此方位,即是卦象落地之处——外应所指,断辞所归。")
	j.Sections = append(j.Sections, JudgeSection{Key: "leixiang", Title: "类象取应", Text: lx.String()})

	// ── 应期 ──
	if j.YingQi != "" {
		j.Sections = append(j.Sections, JudgeSection{Key: "yingqi", Title: "应期", Text: j.YingQi})
	}
}
