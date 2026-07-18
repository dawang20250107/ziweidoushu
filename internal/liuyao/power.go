// 元忌力量评估器:《增删卜易》元神有力五条/无力六条、忌神有力五条/
// 无力七条,在旺衰/四态/动变标注层上机械化对照(考据见
// research/liuyao-yongshen.md)。输出条目命中与倾向,判定权仍在解卦层
// ——「总贵用神有气,用神无根则元神有力亦难生,忌神无力亦休喜」。
package liuyao

import "fmt"

// PowerNote 一爻的力量评估(命中条目为经文条款的直述)。
type PowerNote struct {
	Pos     int      `json:"pos"`
	Verdict string   `json:"verdict"` // 有力/无力/中平
	Reasons []string `json:"reasons"`
}

func elIdxOf(name string) int {
	for i, n := range elementNames {
		if n == name {
			return i
		}
	}
	return 0
}

// movingAt 位置集合中是否有动爻(明动或暗动)。
func (r *Result) movingAt(pos []int) bool {
	for _, p := range pos {
		y := r.Yaos[p-1]
		if y.Moving || y.AnDong {
			return true
		}
	}
	return false
}

// assessPower 按增删条目评估一爻;otherDongPos 为「同动即改性」的另一方
// (元神看忌神、忌神看元神/仇神)。
func (r *Result) assessPower(y Yao, isYuan bool) PowerNote {
	el := elIdxOf(y.Element)
	moving := y.Moving || y.AnDong
	wangXiang := y.MonthState == "旺" || y.MonthState == "相"
	xiuQiu := !wangXiang
	linRiYue := y.DayRelation == "临" || y.Branch == r.MonthJian
	riYueSheng := y.DayRelation == "生" || y.DayRelation == "扶" || y.MonthState == "相"
	dongSheng, dongKe, dongMu := false, false, false
	for _, z := range r.Yaos {
		if z.Pos == y.Pos || !(z.Moving || z.AnDong) {
			continue
		}
		ze := elIdxOf(z.Element)
		if (ze+1)%5 == el {
			dongSheng = true
		}
		if (ze+2)%5 == el {
			dongKe = true
		}
		// 动墓仅取明动(暗动力半,墓库收藏须明动;增删原例未土日冲丑暗动不作酉金入墓论)
		if z.Moving && branchIndexOf(z.Branch) == muBranch[el] {
			dongMu = true
		}
	}

	var strong, weak []string
	// ── 有力条 ──
	if wangXiang || linRiYue || riYueSheng || dongSheng {
		strong = append(strong, "旺相/临日月/得日月动爻生扶(有力一)")
	}
	if moving && (y.BianRelation == "回头生" || y.BianRelation == "化进神") {
		strong = append(strong, "动化回头生/化进神(有力二)")
	}
	if y.DayStage == "长生" || y.DayStage == "帝旺" {
		strong = append(strong, "长生帝旺于日辰(有力三/四)")
	}
	if isYuan && moving && r.movingAt(r.JiShenPos) {
		strong = append(strong, "元神与忌神同动,忌贪生元反成连生(有力四)")
	}
	if !isYuan && moving {
		var chouPos []int
		for _, z := range r.Yaos {
			if z.LiuQin == r.ChouShen {
				chouPos = append(chouPos, z.Pos)
			}
		}
		if r.movingAt(chouPos) {
			strong = append(strong, "忌神与仇神同动(有力五)")
		}
	}
	if wangXiang && moving && y.XunKong {
		strong = append(strong, "旺动临空,动不为空,待冲空实空之日(有力五)")
	}
	// ── 无力条 ──
	if xiuQiu && (!moving || y.DayRelation == "克" || dongKe) {
		weak = append(weak, "休囚不动/动而被日月动爻克(无力一)")
	}
	if xiuQiu && (y.XunKong || y.YuePo || y.RiPo) {
		weak = append(weak, "休囚又逢旬空月破(无力二)")
	}
	if !isYuan && !moving && (y.XunKong || y.YuePo) {
		weak = append(weak, "忌神静临空破(忌无力二)")
	}
	if xiuQiu && moving && y.BianRelation == "化退神" {
		weak = append(weak, "衰而动化退神(无力三/四)")
	}
	if xiuQiu && (y.DayStage == "绝" || y.BianRelation == "化绝") {
		weak = append(weak, "衰而又绝(无力四/五)")
	}
	if y.DayStage == "墓" || y.BianRelation == "化墓" || dongMu {
		weak = append(weak, "入墓(日墓/化墓/动墓,无力五/三)")
	}
	if moving && (y.BianRelation == "化绝" || y.BianRelation == "回头克") {
		weak = append(weak, "动而化绝化克(无力六)")
	}
	if !isYuan && moving && r.movingAt(r.YuanShenPos) {
		weak = append(weak, "忌神与元神同动,贪生元神忘克用神(忌无力七)")
	}

	verdict := "中平"
	reasons := []string{}
	switch {
	case len(weak) > 0: // 冲突时保守:根蒂有伤则不作有力论
		verdict = "无力"
		reasons = append(reasons, weak...)
		if len(strong) > 0 {
			reasons = append(reasons, "另有:"+strong[0])
		}
	case len(strong) > 0:
		verdict = "有力"
		reasons = strong
	}
	return PowerNote{Pos: y.Pos, Verdict: verdict, Reasons: reasons}
}

func branchIndexOf(b string) int {
	rs := []rune(b)
	if len(rs) == 0 {
		return -1
	}
	for j, br := range branches {
		if br == rs[0] {
			return j
		}
	}
	return -1
}

// applyPower 对元神/忌神各爻评估力量。
func (r *Result) applyPower() {
	r.YuanShenPower, r.JiShenPower = []PowerNote{}, []PowerNote{}
	for _, p := range r.YuanShenPos {
		r.YuanShenPower = append(r.YuanShenPower, r.assessPower(r.Yaos[p-1], true))
	}
	for _, p := range r.JiShenPos {
		r.JiShenPower = append(r.JiShenPower, r.assessPower(r.Yaos[p-1], false))
	}
}

// PowerText 供提示词用的一行摘要:「第6爻有力(旺相…;动化进神…)」。
func PowerText(notes []PowerNote) string {
	if len(notes) == 0 {
		return "不上卦"
	}
	s := ""
	for i, n := range notes {
		if i > 0 {
			s += ";"
		}
		s += fmt.Sprintf("第%d爻%s", n.Pos, n.Verdict)
		if len(n.Reasons) > 0 {
			s += "(" + n.Reasons[0]
			for _, rr := range n.Reasons[1:] {
				s += ";" + rr
			}
			s += ")"
		}
	}
	return s
}
