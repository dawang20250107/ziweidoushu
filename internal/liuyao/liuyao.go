// Package liuyao 六爻(火珠林法)装卦引擎:以钱代蓍摇卦、京房纳甲、
// 八宫世应、六亲、六神、动爻变卦。断辞由 AI 层引语料
// (火珠林/增删卜易/卜筮正宗均在研究语料)补充,本包只负责可验证的装卦计算。
//
// 正确性:八宫六十四卦按变爻规律程序生成,与《卜筮正宗》定表逐卦对照(单测)。
package liuyao

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/6tail/lunar-go/calendar"
	"github.com/dawang20250107/ziweidoushu/internal/meihua"
)

// ── 基础表 ───────────────────────────────────────────────────

var (
	stems    = []rune("甲乙丙丁戊己庚辛壬癸")
	branches = []rune("子丑寅卯辰巳午未申酉戌亥")
	// 地支五行:木0 火1 土2 金3 水4(与 sizhu 同序)
	// 基础表:与 liuyao/daliuren/ziwei(sizhu) 三处刻意各自持有(引擎解耦);口径须一致(子水始,木0火1土2金3水4),改此须同步余处,各引擎黄金基准会捕获不一致。
	branchElement = []int{4, 2, 0, 0, 2, 1, 1, 2, 3, 3, 2, 4}
	elementNames  = []string{"木", "火", "土", "金", "水"}
	// 八卦宫五行(先天数序 乾兑离震巽坎艮坤)
	trigramElement = []int{3, 3, 1, 0, 0, 4, 2, 2}
)

// 京房纳甲:每纯卦内/外三爻的天干与三支(自下而上)。索引=先天数-1。
var najia = [8]struct {
	innerStem rune
	outerStem rune
	inner     [3]rune
	outer     [3]rune
}{
	{'甲', '壬', [3]rune{'子', '寅', '辰'}, [3]rune{'午', '申', '戌'}}, // 乾
	{'丁', '丁', [3]rune{'巳', '卯', '丑'}, [3]rune{'亥', '酉', '未'}}, // 兑
	{'己', '己', [3]rune{'卯', '丑', '亥'}, [3]rune{'酉', '未', '巳'}}, // 离
	{'庚', '庚', [3]rune{'子', '寅', '辰'}, [3]rune{'午', '申', '戌'}}, // 震
	{'辛', '辛', [3]rune{'丑', '亥', '酉'}, [3]rune{'未', '巳', '卯'}}, // 巽
	{'戊', '戊', [3]rune{'寅', '辰', '午'}, [3]rune{'申', '戌', '子'}}, // 坎
	{'丙', '丙', [3]rune{'辰', '午', '申'}, [3]rune{'戌', '子', '寅'}}, // 艮
	{'乙', '癸', [3]rune{'未', '巳', '卯'}, [3]rune{'丑', '亥', '酉'}}, // 坤
}

// 六神(按日干起,自初爻循环向上)。
var liuShen = []string{"青龙", "朱雀", "勾陈", "腾蛇", "白虎", "玄武"}

// liuShenStart 日干 → 初爻六神索引:甲乙青龙 丙丁朱雀 戊勾陈 己腾蛇 庚辛白虎 壬癸玄武。
func liuShenStart(dayStem int) int {
	switch {
	case dayStem <= 1:
		return 0
	case dayStem <= 3:
		return 1
	case dayStem == 4:
		return 2
	case dayStem == 5:
		return 3
	case dayStem <= 7:
		return 4
	default:
		return 5
	}
}

// ── 八宫与世应(规律生成,单测与定表对照)────────────────────

// palaceHexagrams 生成某宫八卦的六爻爻象序列(自下而上)。
// 宫序:0 纯卦,1-5 一至五世(依次再变初~五爻),6 游魂(变四爻),7 归魂(内卦复原)。
func palaceHexagrams(pure [6]bool) [8][6]bool {
	var out [8][6]bool
	cur := pure
	out[0] = cur
	for i := 1; i <= 5; i++ {
		cur[i-1] = !cur[i-1]
		out[i] = cur
	}
	// 游魂:在五世卦基础上再变四爻
	cur[3] = !cur[3]
	out[6] = cur
	// 归魂:游魂内卦复原为本宫内卦
	copy(cur[:3], pure[:3])
	out[7] = cur
	return out
}

// shiPositions 宫序 → 世爻位置(1-6)。
var shiPositions = [8]int{6, 1, 2, 3, 4, 5, 4, 3}

type palaceEntry struct {
	palace int // 宫(先天数-1)
	seq    int // 宫内序 0-7
	shi    int // 世爻 1-6
}

// hexIndex 六爻爻象 → 64 序(下卦*8+上卦 各 0-7,用先天数-1)。
func hexKey(lines [6]bool) int {
	lo := meihua.TrigramByLines([3]bool{lines[0], lines[1], lines[2]}).Num - 1
	hi := meihua.TrigramByLines([3]bool{lines[3], lines[4], lines[5]}).Num - 1
	return hi*8 + lo
}

// palaceIndex 全 64 卦 → 八宫归属(初始化时按规律生成)。
var palaceIndex map[int]palaceEntry

func init() {
	palaceIndex = make(map[int]palaceEntry, 64)
	for p := 0; p < 8; p++ {
		t := meihua.TrigramByNum(p + 1)
		pure := [6]bool{t.Lines[0], t.Lines[1], t.Lines[2], t.Lines[0], t.Lines[1], t.Lines[2]}
		for seq, lines := range palaceHexagrams(pure) {
			palaceIndex[hexKey(lines)] = palaceEntry{palace: p, seq: seq, shi: shiPositions[seq]}
		}
	}
}

// ── 装卦 ─────────────────────────────────────────────────────

// Yao 一爻(自下而上第 Pos 爻)。
// 旺衰/旬空/日辰作用为装卦即算的客观标注层(依《增删卜易》《卜筮正宗》,
// 规则考据见 research/liuyao-wangshuai.md)。
type Yao struct {
	Pos     int    `json:"pos"`  // 1-6
	Yang    bool   `json:"yang"` // 阳爻
	Moving  bool   `json:"moving"`
	Stem    string `json:"stem"`   // 纳甲天干
	Branch  string `json:"branch"` // 纳甲地支
	Element string `json:"element"`
	LiuQin  string `json:"liuQin"`            // 六亲
	LiuShen string `json:"liuShen"`           // 六神(按日干)
	IsShi   bool   `json:"isShi"`             // 世
	IsYing  bool   `json:"isYing"`            // 应
	BianYao *Yao   `json:"bianYao,omitempty"` // 动爻之变(仅动爻有,变卦对应爻)

	MonthState  string `json:"monthState"`            // 对月建旺衰:旺/相/休/囚/死
	YuePo       bool   `json:"yuePo,omitempty"`       // 月破(爻支冲月建)
	XunKong     bool   `json:"xunKong,omitempty"`     // 旬空(按日干支所在旬)
	DayRelation string `json:"dayRelation,omitempty"` // 日辰对爻:临/冲/合/扶/生/克/泄/耗
	AnDong      bool   `json:"anDong,omitempty"`      // 暗动(静爻旺相逢日冲)
	RiPo        bool   `json:"riPo,omitempty"`        // 日破(静爻休囚死逢日冲)

	DayStage     string `json:"dayStage,omitempty"`     // 对日辰四态:长生/帝旺/墓/绝(野鹤口径)
	BianRelation string `json:"bianRelation,omitempty"` // 动爻之变:化进神/化退神/伏吟/反吟/化长生/化墓/化绝/化合/回头生/回头克
}

// Result 六爻装卦结果。
type Result struct {
	Question  string `json:"question,omitempty"`
	LunarText string `json:"lunarText"` // 日干支定六神:「甲子日」
	DayStem   string `json:"dayStem"`
	DayBranch string `json:"dayBranch"`
	MonthJian string `json:"monthJian"` // 月建地支
	RiJian    string `json:"riJian"`    // 日辰地支(同 DayBranch,断卦口径字段)

	BenName    string `json:"benName"`
	BianName   string `json:"bianName,omitempty"` // 有动爻才有变卦
	Palace     string `json:"palace"`             // 本卦所属宫,如「乾宫」
	PalaceSeq  string `json:"palaceSeq"`          // 纯卦/一世…游魂/归魂
	Yaos       [6]Yao `json:"yaos"`
	MovingNums []int  `json:"movingNums"` // 动爻位置列表(可为空=静卦)

	// Tosses 摇卦原始记录(每爻背面数 0-3;报数起卦为空)
	Tosses []int `json:"tosses,omitempty"`

	// 用神建议(事类→六亲映射,见 yongshen.go;仅建议,解卦层可按事理改取)
	YongShen      string `json:"yongShen,omitempty"`
	YongShenBasis string `json:"yongShenBasis,omitempty"`
	YongShenPos   []int  `json:"yongShenPos"` // 用神所在爻位;空=用神不上卦(伏神之法另论)

	// 元忌仇链(生用者元、克用者忌、生忌克元者仇——增删卜易·元神章)
	YuanShen    string `json:"yuanShen,omitempty"`
	YuanShenPos []int  `json:"yuanShenPos"`
	JiShen      string `json:"jiShen,omitempty"`
	JiShenPos   []int  `json:"jiShenPos"`
	ChouShen    string `json:"chouShen,omitempty"`

	// 元忌力量评估(增删有力/无力条目对照,见 power.go;倾向供参,判定在解卦层)
	YuanShenPower []PowerNote `json:"yuanShenPower"`
	JiShenPower   []PowerNote `json:"jiShenPower"`

	// Judgment 确定性断语(用神旺衰/元忌力量/动变/世应/应期,见 judge.go)。
	Judgment *Judgment `json:"judgment,omitempty"`
}

var seqNames = []string{"八纯卦", "一世卦", "二世卦", "三世卦", "四世卦", "五世卦", "游魂卦", "归魂卦"}

// ── 旺衰/旬空/日辰作用(research/liuyao-wangshuai.md)────────

// liuHe 地支六合对家:子丑 寅亥 卯戌 辰酉 巳申 午未。
var liuHe = [12]int{1, 0, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2}

// monthState 五态公式:当令旺、令生相、生令休、克令囚、令克死。
func monthState(lingEl, yaoEl int) string {
	switch {
	case yaoEl == lingEl:
		return "旺"
	case (lingEl+1)%5 == yaoEl:
		return "相"
	case (yaoEl+1)%5 == lingEl:
		return "休"
	case (yaoEl+2)%5 == lingEl:
		return "囚"
	default:
		return "死"
	}
}

// xunKongBranches 由日干支六十甲子序推旬空两支(六甲旬空诀)。
func xunKongBranches(dayStem, dayBranch int) (int, int) {
	j := 0
	for ; j < 60; j++ {
		if j%10 == dayStem && j%12 == dayBranch {
			break
		}
	}
	a := (10 - 2*(j/10) + 24) % 12
	return a, (a + 1) % 12
}

// ── 生旺墓绝与动变作用(research/liuyao-dongbian.md)──────────
// 十二长生野鹤只验四态(「余得验者,止验生旺墓绝,其余不验」);
// 土长生在申为野鹤以天时占验裁定(水土同宫)。索引=五行(木火土金水)。

var (
	csBranch  = []int{11, 2, 8, 5, 8} // 长生:木亥 火寅 土申 金巳 水申
	dwBranch  = []int{3, 6, 0, 9, 0}  // 帝旺:木卯 火午 土子 金酉 水子
	muBranch  = []int{7, 10, 4, 1, 4} // 墓:木未 火戌 土辰 金丑 水辰
	jueBranch = []int{8, 11, 5, 2, 5} // 绝:木申 火亥 土巳 金寅 水巳
	// 进神对(增删卜易):寅→卯 巳→午 申→酉 亥→子 丑→辰 辰→未 未→戌
	jinShen = map[int]int{2: 3, 5: 6, 8: 9, 11: 0, 1: 4, 4: 7, 7: 10}
	// 退神对:子→亥 卯→寅 午→巳 酉→申 辰→丑 未→辰 戌→未
	tuiShen = map[int]int{0: 11, 3: 2, 6: 5, 9: 8, 4: 1, 7: 4, 10: 7}
)

// dayStage 爻对日辰之四态(长生/帝旺/墓/绝);土绝于巳论生不论绝,不标。
func dayStage(el, dayBranch int) string {
	switch dayBranch {
	case csBranch[el]:
		return "长生"
	case dwBranch[el]:
		return "帝旺"
	case muBranch[el]:
		return "墓"
	case jueBranch[el]:
		if el == 2 { // 土绝于巳:巳火反能生土,论生不论绝
			return ""
		}
		return "绝"
	}
	return ""
}

// bianRelation 动爻与变爻作用。优先级:伏吟(同支)> 进/退神 > 反吟(冲)>
// 化长生(金化巳论长生不论克)> 化墓 > 化绝(土化巳论生不论绝)> 化合 >
// 回头生 > 回头克;化泄/化耗不标。
func bianRelation(benB, bianB int) string {
	be, ve := branchElement[benB], branchElement[bianB]
	switch {
	case bianB == benB:
		return "伏吟"
	case jinShen[benB] == bianB && be == ve:
		return "化进神"
	case tuiShen[benB] == bianB && be == ve:
		return "化退神"
	case (benB+6)%12 == bianB:
		return "反吟"
	case bianB == csBranch[be]:
		return "化长生"
	case bianB == muBranch[be]:
		return "化墓"
	case bianB == jueBranch[be] && be != 2:
		return "化绝"
	case liuHe[benB] == bianB:
		return "化合"
	case (ve+1)%5 == be:
		return "回头生"
	case (ve+2)%5 == be:
		return "回头克"
	}
	return ""
}

// dayRelation 日辰对爻:临(同支)/冲/合优先,余按五行生克。
func dayRelation(dayBranch, yaoBranch int) string {
	switch {
	case yaoBranch == dayBranch:
		return "临"
	case (yaoBranch+6)%12 == dayBranch:
		return "冲"
	case liuHe[yaoBranch] == dayBranch:
		return "合"
	}
	de, ye := branchElement[dayBranch], branchElement[yaoBranch]
	switch {
	case de == ye:
		return "扶"
	case (de+1)%5 == ye: // 日生爻
		return "生"
	case (de+2)%5 == ye: // 日克爻
		return "克"
	case (ye+1)%5 == de: // 爻生日
		return "泄"
	default: // 爻克日
		return "耗"
	}
}

// liuQin 宫五行 → 爻五行 的六亲。
func liuQin(palaceEl, yaoEl int) string {
	switch {
	case palaceEl == yaoEl:
		return "兄弟"
	case (palaceEl+1)%5 == yaoEl: // 我生
		return "子孙"
	case (palaceEl+2)%5 == yaoEl: // 我克
		return "妻财"
	case (yaoEl+2)%5 == palaceEl: // 克我
		return "官鬼"
	default: // 生我
		return "父母"
	}
}

// assemble 由六爻阴阳+动爻+日干支装出完整卦。
func assemble(lines [6]bool, moving []int, dayStem, dayBranch int, monthJian rune) (*Result, error) {
	key := hexKey(lines)
	entry, ok := palaceIndex[key]
	if !ok {
		return nil, fmt.Errorf("装卦失败:卦象无宫位归属")
	}
	palaceEl := trigramElement[entry.palace]
	lower := meihua.TrigramByLines([3]bool{lines[0], lines[1], lines[2]})
	upper := meihua.TrigramByLines([3]bool{lines[3], lines[4], lines[5]})

	movingSet := map[int]bool{}
	for _, m := range moving {
		if m < 1 || m > 6 {
			return nil, fmt.Errorf("动爻位置非法: %d", m)
		}
		movingSet[m] = true
	}

	// 变卦
	bianLines := lines
	for m := range movingSet {
		bianLines[m-1] = !bianLines[m-1]
	}

	r := &Result{
		BenName:   meihua.HexagramNameByNums(upper.Num, lower.Num),
		Palace:    meihua.TrigramByNum(entry.palace+1).Name + "宫",
		PalaceSeq: seqNames[entry.seq],
		DayStem:   string(stems[dayStem]),
		DayBranch: string(branches[dayBranch]),
		RiJian:    string(branches[dayBranch]),
		MonthJian: string(monthJian),
		// 静卦时也须输出 [] 而非 null(JSON 列表契约)
		MovingNums: []int{},
	}
	if len(movingSet) > 0 {
		bl := meihua.TrigramByLines([3]bool{bianLines[0], bianLines[1], bianLines[2]})
		bu := meihua.TrigramByLines([3]bool{bianLines[3], bianLines[4], bianLines[5]})
		r.BianName = meihua.HexagramNameByNums(bu.Num, bl.Num)
	}

	// 变卦纳甲按变卦自身的内外卦纳
	bianLower := meihua.TrigramByLines([3]bool{bianLines[0], bianLines[1], bianLines[2]})
	bianUpper := meihua.TrigramByLines([3]bool{bianLines[3], bianLines[4], bianLines[5]})

	shenStart := liuShenStart(dayStem)
	for i := 0; i < 6; i++ {
		pos := i + 1
		var stem rune
		var branch rune
		if i < 3 {
			na := najia[lower.Num-1]
			stem, branch = na.innerStem, na.inner[i]
		} else {
			na := najia[upper.Num-1]
			stem, branch = na.outerStem, na.outer[i-3]
		}
		bIdx := 0
		for j, br := range branches {
			if br == branch {
				bIdx = j
			}
		}
		// 应爻 = 世 ±3(规范化到 1-6)
		ying := entry.shi + 3
		if ying > 6 {
			ying -= 6
		}
		yao := Yao{
			Pos:     pos,
			Yang:    lines[i],
			Moving:  movingSet[pos],
			Stem:    string(stem),
			Branch:  string(branch),
			Element: elementNames[branchElement[bIdx]],
			LiuQin:  liuQin(palaceEl, branchElement[bIdx]),
			LiuShen: liuShen[(shenStart+i)%6],
			IsShi:   pos == entry.shi,
			IsYing:  pos == ying,
		}

		// 旺衰/旬空/日辰作用标注(装卦即算的客观事实层)
		monthIdx := 0
		for j, br := range branches {
			if br == monthJian {
				monthIdx = j
			}
		}
		yao.MonthState = monthState(branchElement[monthIdx], branchElement[bIdx])
		yao.YuePo = (bIdx+6)%12 == monthIdx
		kongA, kongB := xunKongBranches(dayStem, dayBranch)
		yao.XunKong = bIdx == kongA || bIdx == kongB
		yao.DayRelation = dayRelation(dayBranch, bIdx)
		yao.DayStage = dayStage(branchElement[bIdx], dayBranch)
		if !yao.Moving && yao.DayRelation == "冲" {
			if yao.MonthState == "旺" || yao.MonthState == "相" {
				yao.AnDong = true
			} else {
				yao.RiPo = true
			}
		}

		// 动爻:装变卦对应爻(六亲仍以本宫五行论)
		if yao.Moving {
			var bs rune
			var bb rune
			if i < 3 {
				na := najia[bianLower.Num-1]
				bs, bb = na.innerStem, na.inner[i]
			} else {
				na := najia[bianUpper.Num-1]
				bs, bb = na.outerStem, na.outer[i-3]
			}
			bbIdx := 0
			for j, br := range branches {
				if br == bb {
					bbIdx = j
				}
			}
			yao.BianYao = &Yao{
				Pos: pos, Yang: bianLines[i],
				Stem: string(bs), Branch: string(bb),
				Element: elementNames[branchElement[bbIdx]],
				LiuQin:  liuQin(palaceEl, branchElement[bbIdx]),
			}
			yao.BianRelation = bianRelation(bIdx, bbIdx)
		}
		r.Yaos[i] = yao
		if yao.Moving {
			r.MovingNums = append(r.MovingNums, pos)
		}
	}
	return r, nil
}

// ── 研究校验导出(tools/liuyaoverify 以书校机)──────────────

// AssembleForResearch 直连装卦:月支+日干支即可,无需历日(书中占例只记月日)。
func AssembleForResearch(lines [6]bool, moving []int, dayStem, dayBranch int, monthJian rune) (*Result, error) {
	return assemble(lines, moving, dayStem, dayBranch, monthJian)
}

// XunKongForResearch 六甲旬空两支索引。
func XunKongForResearch(dayStem, dayBranch int) (int, int) {
	return xunKongBranches(dayStem, dayBranch)
}

// ── 起卦入口 ─────────────────────────────────────────────────

// tossToYao 背面数 → (阳爻, 动)。1背=少阳、2背=少阴、3背=老阳动、0背=老阴动。
func tossToYao(backs int) (yang, moving bool, err error) {
	switch backs {
	case 1:
		return true, false, nil
	case 2:
		return false, false, nil
	case 3:
		return true, true, nil
	case 0:
		return false, true, nil
	default:
		return false, false, fmt.Errorf("每爻背面数须为 0-3,得到 %d", backs)
	}
}

// dayGanZhi 当前时刻 → 农历日干支索引与月建。
func dayGanZhi(t time.Time) (dayStem, dayBranch int, monthJian rune, lunarText string, err error) {
	if t.Year() < 1902 || t.Year() > 2098 {
		return 0, 0, ' ', "", fmt.Errorf("时间超出支持范围(1902-2098)")
	}
	lunar := calendar.NewSolarFromDate(t).GetLunar()
	dgz := []rune(lunar.GetDayInGanZhi())
	if len(dgz) != 2 {
		return 0, 0, ' ', "", fmt.Errorf("日干支解析失败")
	}
	for i, s := range stems {
		if s == dgz[0] {
			dayStem = i
		}
	}
	for i, b := range branches {
		if b == dgz[1] {
			dayBranch = i
		}
	}
	mgz := []rune(lunar.GetMonthInGanZhi())
	monthJian = mgz[len(mgz)-1]
	lunarText = fmt.Sprintf("%s月%s日(%s日)", lunar.GetMonthInChinese(), lunar.GetDayInChinese(), lunar.GetDayInGanZhi())
	return dayStem, dayBranch, monthJian, lunarText, nil
}

// ByTosses 报爻起卦:六爻背面数(自下而上)。
func ByTosses(tosses []int, at time.Time, question string) (*Result, error) {
	if len(tosses) != 6 {
		return nil, fmt.Errorf("六爻需六次摇卦记录")
	}
	var lines [6]bool
	var moving []int
	for i, backs := range tosses {
		yang, mv, err := tossToYao(backs)
		if err != nil {
			return nil, err
		}
		lines[i] = yang
		if mv {
			moving = append(moving, i+1)
		}
	}
	ds, db, mj, lt, err := dayGanZhi(at)
	if err != nil {
		return nil, err
	}
	r, err := assemble(lines, moving, ds, db, mj)
	if err != nil {
		return nil, err
	}
	r.Question = question
	r.LunarText = lt
	r.Tosses = append([]int(nil), tosses...)
	r.applyYongShen()
	r.applyPower()
	r.Judgment = r.Judge() // 确定性断语,随起卦即出(免费层)
	return r, nil
}

// Shake 服务端摇卦:crypto/rand 模拟三枚铜钱六掷。
func Shake(at time.Time, question string) (*Result, error) {
	tosses := make([]int, 6)
	buf := make([]byte, 18) // 6 爻 × 3 枚
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	for i := 0; i < 6; i++ {
		backs := 0
		for c := 0; c < 3; c++ {
			if buf[i*3+c]%2 == 0 {
				backs++
			}
		}
		tosses[i] = backs
	}
	return ByTosses(tosses, at, question)
}
