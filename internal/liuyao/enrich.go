// 六爻断层补强:六冲六合卦性、伏神之法、六神事象、爻位象义、三合局。
// 由《六爻玄奇》对照查漏定题,规则均按《增删卜易》《卜筮正宗》同源口径义引落地
// (该书独家「扶抑格」体系不采,取舍记录见 research/liuyao-xuanqi-notes.md)。
package liuyao

import "fmt"

// ── 六冲/六合卦性 ────────────────────────────────────────────

// guaXingZhi 六爻三对(初四/二五/三上)支支相冲为六冲卦、相合为六合卦。
// 六冲主散、动而不久;六合主成、绊住而久长(增删卜易·冲合章义)。
func guaXingZhi(bs [6]int) string {
	chong, he := true, true
	for i := 0; i < 3; i++ {
		a, b := bs[i], bs[i+3]
		if (a+6)%12 != b {
			chong = false
		}
		if liuHe[a] != b {
			he = false
		}
	}
	switch {
	case chong:
		return "六冲"
	case he:
		return "六合"
	}
	return ""
}

// ── 伏神之法(用神不上卦)────────────────────────────────────

// FuShen 伏神:用神六亲不上卦时,从本宫首卦纳甲取之,伏于本卦同位爻(飞神)下。
// 出伏条件(卜筮正宗·飞伏章同源口径):飞神旬空或被日冲则伏神自出;
// 伏神得日月生扶、或飞来生伏,皆能出;伏神休囚又遭飞神克者难出。
type FuShen struct {
	LiuQin  string `json:"liuQin"`
	Branch  string `json:"branch"`
	Element string `json:"element"`
	Pos     int    `json:"pos"`     // 所伏爻位(本宫首卦位)
	Fei     string `json:"fei"`     // 飞神(本卦同位爻)支
	CanOut  bool   `json:"canOut"`  // 能否出伏
	Note    string `json:"note"`    // 机理一句
	ChuFuRi string `json:"chuFuRi"` // 出伏应期支(值伏神/冲飞神之日)
}

// applyFuShen 用神不上卦时取伏神(须在 applyYongShen 之后调用)。
func (r *Result) applyFuShen() {
	if len(r.YongShenPos) > 0 || r.YongShen == "" || r.YongShen == "世爻" {
		return
	}
	palaceEl := trigramElement[r.palaceIdx]
	na := najia[r.palaceIdx]
	dayB := branchIndexOf(r.DayBranch)
	dayS := 0
	for i, s := range stems {
		if string(s) == r.DayStem {
			dayS = i
		}
	}
	monthB := branchIndexOf(r.MonthJian)
	kongA, kongB := xunKongBranches(dayS, dayB)

	for i := 0; i < 6; i++ {
		var br rune
		if i < 3 {
			br = na.inner[i]
		} else {
			br = na.outer[i-3]
		}
		fuB := branchIndexOf(string(br))
		fuEl := branchElement[fuB]
		if liuQin(palaceEl, fuEl) != r.YongShen {
			continue
		}
		fei := r.Yaos[i]
		feiB := branchIndexOf(fei.Branch)
		feiEl := branchElement[feiB]

		fs := &FuShen{
			LiuQin:  r.YongShen,
			Branch:  string(br),
			Element: elementNames[fuEl],
			Pos:     i + 1,
			Fei:     fei.Branch,
		}
		// 出伏机理逐条判定
		fuWang := monthState(branchElement[monthB], fuEl)
		fuDayRel := dayRelation(dayB, fuB)
		fuKong := fuB == kongA || fuB == kongB
		feiChong := (feiB+6)%12 == dayB
		feiShengFu := (feiEl+1)%5 == fuEl
		feiKeFu := (feiEl+2)%5 == fuEl
		switch {
		case fei.XunKong:
			fs.CanOut, fs.Note = true, "飞神旬空,伏神乘虚自出"
		case feiChong:
			fs.CanOut, fs.Note = true, "飞神被日辰冲开,伏神得出"
		case feiShengFu:
			fs.CanOut, fs.Note = true, "飞来生伏,得长生之义,能出"
		case fuWang == "旺" || fuWang == "相":
			fs.CanOut, fs.Note = true, "伏神得月令旺相,有气能出"
		case fuDayRel == "临" || fuDayRel == "生" || fuDayRel == "扶":
			fs.CanOut, fs.Note = true, "伏神得日辰生扶,有气能出"
		case feiKeFu:
			fs.CanOut, fs.Note = false, "飞来克伏又休囚,受制难出"
		case fuKong:
			fs.CanOut, fs.Note = false, "伏神休囚又旬空,难出"
		default:
			fs.CanOut, fs.Note = false, "伏神休囚无气,待生扶之期"
		}
		fs.ChuFuRi = fmt.Sprintf("%s(值伏神)或%s(冲飞神)",
			string(br), string(branches[(feiB+6)%12]))
		r.FuShen = fs
		return // 取自下而上首位所伏
	}
}

// ── 六神事象(只作事象修饰,不入吉凶分——「六神不主吉凶」)──

var liuShenNote = map[string]string{
	"青龙": "主喜庆婚酒、财喜临门,事带吉彩",
	"朱雀": "主口舌文书、信息言辞,涉讼防口角",
	"勾陈": "主田土屋宅、迟滞牵连,旧事纠缠",
	"腾蛇": "主虚惊怪异、梦扰缠绕,事多虚诈不实",
	"白虎": "主伤病孝服、刀兵血光,其性急烈",
	"玄武": "主暗昧盗失、私情欺瞒,事有不明",
}

// ── 爻位象义(黄金策爻位说同源)──────────────────────────────

var yaoPosNote = [6]string{
	"初爻(足下宅基,事之始)",
	"二爻(宅舍内室,近身之地)",
	"三爻(门户胸腹)",
	"四爻(门户胸腹,内外之交)",
	"五爻(道路尊位)",
	"上爻(头顶远末,事之终)",
}

// ── 三合局(申子辰水/寅午戌火/巳酉丑金/亥卯未木)────────────

var sanHeGroups = [4][3]int{{8, 0, 4}, {2, 6, 10}, {5, 9, 1}, {11, 3, 7}}
var sanHeEl = [4]int{4, 1, 3, 0}

// sanHeJu 动爻(明动/暗动)是否会成三合局:三支中至少二支为动爻、
// 缺一字可由日建或月建补(增删卜易·三合成局章义:虚一待用以日月填实)。
// 返回 (局五行索引, 局名, 是否成局)。
func (r *Result) sanHeJu() (int, string, bool) {
	dong := map[int]bool{}
	for _, y := range r.Yaos {
		if y.Moving || y.AnDong {
			dong[branchIndexOf(y.Branch)] = true
		}
	}
	if len(dong) < 2 {
		return 0, "", false
	}
	dayB := branchIndexOf(r.DayBranch)
	monthB := branchIndexOf(r.MonthJian)
	for gi, g := range sanHeGroups {
		nDong, covered := 0, true
		for _, b := range g {
			switch {
			case dong[b]:
				nDong++
			case b == dayB || b == monthB:
				// 日月补一字
			default:
				covered = false
			}
		}
		if covered && nDong >= 2 {
			name := fmt.Sprintf("%s%s%s%s局",
				string(branches[g[0]]), string(branches[g[1]]), string(branches[g[2]]),
				elementNames[sanHeEl[gi]])
			return sanHeEl[gi], name, true
		}
	}
	return 0, "", false
}
