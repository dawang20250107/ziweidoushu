// Package daliuren 大六壬起课引擎:天地盘(月将加时)、四课(干寄宫)、
// 三传(九宗门)。规则依《大六壬理法易简》(北海闲人)为纲,考据与走例
// 见 research/daliuren-survey.md;体系口径与正确性基准以理法易简走例钉死。
//
// 本包只做「起课」这一确定性骨架;天将(贵人昼夜)待贵人全表跨书互证后
// 另加,AI 断课层引语料处理。
package daliuren

import "fmt"

// 地支索引:子0 丑1 寅2 卯3 辰4 巳5 午6 未7 申8 酉9 戌10 亥11
var branches = []rune("子丑寅卯辰巳午未申酉戌亥")

// 天干索引:甲0…癸9
var stems = []rune("甲乙丙丁戊己庚辛壬癸")

// 五行:木0 火1 土2 金3 水4。地支本气五行。
// 基础表:与 liuyao/daliuren/ziwei(sizhu) 三处刻意各自持有(引擎解耦);口径须一致(子水始,木0火1土2金3水4),改此须同步余处,各引擎黄金基准会捕获不一致。
var branchElement = []int{4, 2, 0, 0, 2, 1, 1, 2, 3, 3, 2, 4}

// 天干寄宫(理法易简起课定式):甲寅 乙辰 丙巳 丁未 戊巳 己未 庚申 辛戌 壬亥 癸丑。
// (丙戊同寄巳、丁己同寄未——阳干寄本气方、阴干寄墓库方)
var stemJiGong = []int{2, 4, 5, 7, 5, 7, 8, 10, 11, 1}

// 中气 → 月将(理法易简:换将按中气,自雨水亥将起,逐中气退一支)。
// 索引为「中气序」:0=雨水…11=大寒;月将 = (11 - idx) 号地支?见 monthGeneral。
// 中气次序:雨水 春分 谷雨 小满 夏至 大暑 处暑 秋分 霜降 小雪 冬至 大寒
// 对应月将:亥   戌   酉   申   未   午   巳   辰   卯   寅   丑   子
var midQiGeneral = []int{11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0}

// MonthGeneralByMidQi 按中气序(0=雨水…11=大寒)取月将地支索引。
func MonthGeneralByMidQi(midQiIndex int) int {
	return midQiGeneral[midQiIndex%12]
}

// Ke 一课:下神(地盘/本位)加上神(天盘)。
type Ke struct {
	Lower string `json:"lower"` // 下神(第一课为日干寄宫、第二课为干上神、第三课为日支、第四课为支上神)
	Upper string `json:"upper"` // 上神(下神落在天盘上所乘之神)
}

// Result 一课的起课结果。
type Result struct {
	DayStem    string `json:"dayStem"`
	DayBranch  string `json:"dayBranch"`
	HourBranch string `json:"hourBranch"` // 占时地支
	MonthGen   string `json:"monthGen"`   // 月将
	// TianPan[i] = 地盘 i 位上所乘的天盘之神(月将加时排布)
	TianPan [12]string `json:"tianPan"`
	Ke      [4]Ke      `json:"ke"`     // 四课(自第一至第四)
	Chuan   [3]string  `json:"chuan"`  // 三传(初/中/末)
	KeType  string     `json:"keType"` // 课体:贼克/比用/涉害/遥克/昴星/别责/八专/伏吟/返吟
	// TianJiang[i] = 地盘 i 位上所乘十二天将(贵人歌两书互证,亥~辰顺布)。
	TianJiang [12]string `json:"tianJiang"`
	// ChuanJiang 三传所乘天将;GuiIsDay 本课用昼贵与否。
	ChuanJiang [3]string `json:"chuanJiang"`
	GuiIsDay   bool      `json:"guiIsDay"`
	// XunKong 日干支所在旬的空亡两支;ChuanDunGan 三传旬遁干(传落空亡则空串)。
	XunKong     [2]string `json:"xunKong"`
	ChuanDunGan [3]string `json:"chuanDunGan"`
	// 活时报数起课溯源(正时为零值):BaoShu 所报之数,HourNote 如「活时·报数7」。
	BaoShu   int    `json:"baoShu,omitempty"`
	HourNote string `json:"hourNote,omitempty"`

	// Judgment 确定性断语(课体 + 三传对日干生克,见 judge.go)。
	Judgment *Judgment `json:"judgment,omitempty"`
	// NianMingInfo 年命上神(问者提供出生年时计算,正时课的个人化分断)。
	NianMingInfo *NianMing `json:"nianMing,omitempty"`
}

// tianPanOf 天盘布局:月将加于占时之上,顺行十二地支。
// 地盘 i 位所乘天盘 = 月将 + (i - 占时),即将加时后地盘各位对应的天盘神。
func tianPanOf(monthGen, hour int) [12]int {
	var tp [12]int
	// 月将落在地盘「占时」位;地盘 hour 位天盘=monthGen,地盘 (hour+k) 位天盘=monthGen+k
	for i := 0; i < 12; i++ {
		tp[i] = ((monthGen+(i-hour))%12 + 12) % 12
	}
	return tp
}

// upperOf 取地盘某支位上所乘的天盘之神(下神→上神)。
func upperOf(tp [12]int, lower int) int {
	return tp[lower]
}

// Cast 起课:日干支、占时、月将(地支索引)。
func Cast(dayStem, dayBranch, hour, monthGen int) (*Result, error) {
	if dayStem < 0 || dayStem > 9 || dayBranch < 0 || dayBranch > 11 || hour < 0 || hour > 11 || monthGen < 0 || monthGen > 11 {
		return nil, fmt.Errorf("起课参数越界")
	}
	tp := tianPanOf(monthGen, hour)

	// 四课:1=干寄宫及其上神;2=(1之上神)及其上神;3=日支及其上神;4=(3之上神)及其上神
	ji := stemJiGong[dayStem]
	k1u := upperOf(tp, ji)
	k2u := upperOf(tp, k1u)
	k3u := upperOf(tp, dayBranch)
	k4u := upperOf(tp, k3u)
	ke := [4]Ke{
		{string(branches[ji]), string(branches[k1u])},
		{string(branches[k1u]), string(branches[k2u])},
		{string(branches[dayBranch]), string(branches[k3u])},
		{string(branches[k3u]), string(branches[k4u])},
	}

	r := &Result{
		DayStem:    string(stems[dayStem]),
		DayBranch:  string(branches[dayBranch]),
		HourBranch: string(branches[hour]),
		MonthGen:   string(branches[monthGen]),
		Ke:         ke,
	}
	for i := 0; i < 12; i++ {
		r.TianPan[i] = string(branches[tp[i]])
	}

	chuan, ktype := deriveChuan(dayStem, dayBranch, ji, tp, [4][2]int{
		{ji, k1u}, {k1u, k2u}, {dayBranch, k3u}, {k3u, k4u},
	})
	r.Chuan = [3]string{string(branches[chuan[0]]), string(branches[chuan[1]]), string(branches[chuan[2]])}
	r.KeType = ktype
	// 旬空与旬遁:旬首(甲)所临支 = 日支 - 日干;旬内十支得遁干,余二支即空亡
	xunFirst := ((dayBranch-dayStem)%12 + 12) % 12
	r.XunKong = [2]string{string(branches[(xunFirst+10)%12]), string(branches[(xunFirst+11)%12])}
	for c := 0; c < 3; c++ {
		if off := ((chuan[c]-xunFirst)%12 + 12) % 12; off < 10 {
			r.ChuanDunGan[c] = string(stems[off])
		}
	}
	// 十二天将:随天盘布于地盘十二位;三传乘将取该传之神所在地盘位
	jiang, _, isDay := placeTianJiang(tp, dayStem, hour)
	r.GuiIsDay = isDay
	for i := 0; i < 12; i++ {
		r.TianJiang[i] = TianJiangNames[jiang[i]]
	}
	for c := 0; c < 3; c++ {
		for i := 0; i < 12; i++ {
			if tp[i] == chuan[c] {
				r.ChuanJiang[c] = TianJiangNames[jiang[i]]
				break
			}
		}
	}
	r.Judgment = r.Judge() // 确定性断语,随起课即出
	return r, nil
}

// ── 九宗门三传推导(理法易简简化口径)──────────────────────

// keSheng a 生 b / a 克 b(五行索引)。
func ke(a, b int) bool { return (branchElement[a]+2)%5 == branchElement[b] } // a 克 b

// keEl 五行层面 a 克 b(供第一课以日干本气论克)。
func keEl(aEl, bEl int) bool { return (aEl+2)%5 == bEl }

var (
	mengSet = map[int]bool{2: true, 5: true, 8: true, 11: true} // 孟:寅巳申亥
	zhongS  = map[int]bool{0: true, 6: true, 3: true, 9: true}  // 仲:子午卯酉
	// 季:辰戌丑未(余下)
)

// xing 三刑:子卯互刑;寅巳申环;丑戌未环;辰午酉亥自刑(伏吟中末自刑取冲)。
var xing = map[int]int{0: 3, 3: 0, 2: 5, 5: 8, 8: 2, 1: 10, 10: 7, 7: 1, 4: 4, 6: 6, 9: 9, 11: 11}

// yiMa 日支三合驿马:申子辰马寅、寅午戌马申、巳酉丑马亥、亥卯未马巳。
func yiMa(b int) int {
	switch b % 4 {
	case 0: // 子辰申 之一(0,4,8)
		return 2
	case 2: // 寅午戌(2,6,10)
		return 8
	case 1: // 丑巳酉(1,5,9)
		return 11
	default: // 卯未亥(3,7,11)
		return 5
	}
}

// classRank 孟0 仲1 季2(涉害取用序:先孟后仲后季)。
func classRank(b int) int {
	if mengSet[b] {
		return 0
	}
	if zhongS[b] {
		return 1
	}
	return 2
}

// chuanFrom 由初传(天盘神)顺天盘推中末:中=初之上神,末=中之上神。
func chuanFrom(chu int, tp [12]int) [3]int {
	mid := tp[chu]
	end := tp[mid]
	return [3]int{chu, mid, end}
}

// deriveChuan 起三传,返回三传地支索引与课体名。
// courses[i] = {下神, 上神}(第 i+1 课)。
func deriveChuan(dayStem, dayBranch, jiGong int, tp [12]int, courses [4][2]int) ([3]int, string) {
	ganUpper := courses[0][1] // 干上神(第一课上神)
	zhiUpper := courses[2][1] // 支上神(第三课上神)
	yangDay := dayStem%2 == 0

	// 伏吟:天盘==地盘
	fuyin := true
	for i := 0; i < 12; i++ {
		if tp[i] != i {
			fuyin = false
			break
		}
	}
	// 返吟:天盘==冲地盘
	fanyin := true
	for i := 0; i < 12; i++ {
		if tp[i] != (i+6)%12 {
			fanyin = false
			break
		}
	}

	// 收集克:下贼上(下克上)优先,否则上克下。
	// 第一课之「下」为日干本身,克战以日干本气五行论(非寄宫支五行)——
	// 《六壬断案》案02 戊申日子将申时钉死:戊土上酉金无克,方成元首取辰。
	loElOf := func(i int) int {
		if i == 0 {
			return stemElement[dayStem]
		}
		return branchElement[courses[i][0]]
	}
	// 派生课(第二/第四课)与前课下上全同者不取其克——「干支阴阳已先据其位,
	// 不可重复取用」(《六壬断案》案07 乙巳日戌将亥时钉死:第四课辰卯与
	// 第一课重,其克作废,无贼克而走遥克弹射取辰)。干支两正课(第一/第三)
	// 不在此限:八专日课一课三支面同而克义异(课经「癸丑日俱有克」,
	// 案220 癸丑日酉将戌时钉死:课三丑克子发用,三传子亥戌)。
	dup := func(i int) bool {
		if i != 1 && i != 3 {
			return false
		}
		for j := 0; j < i; j++ {
			if courses[j] == courses[i] {
				return true
			}
		}
		return false
	}
	var zei, keUp []int // 存 course 索引
	for i, c := range courses {
		if dup(i) {
			continue
		}
		upEl := branchElement[c[1]]
		if keEl(loElOf(i), upEl) {
			zei = append(zei, i)
		} else if keEl(upEl, loElOf(i)) {
			keUp = append(keUp, i)
		}
	}
	cand, candZei := zei, true
	if len(cand) == 0 {
		cand, candZei = keUp, false
	}

	// 涉害(理法易简「不数深浅,先取孟上神」之见机式,孟仲季论上神所临
	// **地盘之位**,孟>仲>季>课序)。《六壬断案》四例钉死:案76(比者戌午,
	// 午临亥孟)、案74(午临申孟胜寅临辰季)、案188(无孟取仲:未临卯仲)、
	// 案189(戌临寅孟)、案191(丑临卯仲)。曾试历位数克之深浅法,
	// 与上列诸案不合,弃之(考据见 research/daliuren-survey.md)。
	pickSheHai := func(idxs []int) int {
		best := idxs[0]
		for _, i := range idxs[1:] {
			if classRank(courses[i][0]) < classRank(courses[best][0]) {
				best = i
			}
		}
		return best
	}
	pickCand := func(idxs []int, _ bool) (int, string) { // 贼克/比用/涉害
		if len(idxs) == 1 {
			return idxs[0], "贼克"
		}
		var same []int // 比用:上神阴阳同日干
		for _, i := range idxs {
			if (courses[i][1]%2 == 0) == yangDay {
				same = append(same, i)
			}
		}
		if len(same) == 1 {
			return same[0], "比用"
		}
		// 俱比则涉害于比者之中取;俱不比方于全体候选中取
		if len(same) > 1 {
			return pickSheHai(same), "涉害"
		}
		return pickSheHai(idxs), "涉害"
	}

	// 伏吟。中传取初传之刑;初传自刑者,阳日中取支上神、阴日中取干上神;
	// 中传复自刑者,末传冲取(课经伏吟诀,《六壬断案》案215 辛酉日丑将丑时
	// 钉死:初酉自刑、中取干上戌、末戌刑未 → 酉戌未)。
	if fuyin {
		var chu int
		if len(cand) > 0 {
			ci, _ := pickCand(cand, candZei)
			chu = courses[ci][1]
		} else if yangDay {
			chu = ganUpper
		} else {
			chu = zhiUpper
		}
		mid := xing[chu]
		if mid == chu { // 自刑:阳日取支上、阴日取干上
			if yangDay {
				mid = zhiUpper
			} else {
				mid = ganUpper
			}
		}
		end := xing[mid]
		if end == mid {
			end = (mid + 6) % 12 // 复自刑,冲取末传
		}
		return [3]int{chu, mid, end}, "伏吟"
	}
	// 返吟
	if fanyin {
		if len(cand) > 0 {
			ci, _ := pickCand(cand, candZei)
			return chuanFrom(courses[ci][1], tp), "返吟"
		}
		return [3]int{yiMa(dayBranch), zhiUpper, ganUpper}, "返吟"
	}

	// 贼克/比用/涉害
	if len(cand) > 0 {
		ci, name := pickCand(cand, candZei)
		return chuanFrom(courses[ci][1], tp), name
	}

	// 八专:干支同位(甲寅庚申丁未己未癸丑五日)无克,不论遥克——
	// 刚日干上神顺数三位、柔日第四课上神逆数三位为初传,中末俱并干上
	// (课经集走例钉死:甲寅日辰时丑将→丑亥亥;丁未日丑时辰将→亥戌戌)。
	if jiGong == dayBranch {
		var chu int
		if yangDay {
			chu = (ganUpper + 2) % 12
		} else {
			chu = ((courses[3][1]-2)%12 + 12) % 12
		}
		return [3]int{chu, ganUpper, ganUpper}, "八专"
	}

	// 遥克:二三四课上神与日干(本气五行)隔位相克——
	// 先取克干者(蒿矢),无则取干克者(弹射),再比用取舍。
	// 《六壬断案》案36 丁卯日子将卯时(子水克丁,先于丁克酉)、
	// 案95 己酉日卯将子时(书标「蒿矢」,卯木克己先于己克子)钉死。
	var keGan, ganKe []int
	ganEl := stemElement[dayStem]
	for i := 1; i < 4; i++ {
		upEl := branchElement[courses[i][1]]
		if keEl(upEl, ganEl) {
			keGan = append(keGan, i)
		} else if keEl(ganEl, upEl) {
			ganKe = append(ganKe, i)
		}
	}
	yao, yaoZei := keGan, true
	if len(yao) == 0 {
		yao, yaoZei = ganKe, false
	}
	if len(yao) > 0 {
		ci, _ := pickCand(yao, yaoZei)
		return chuanFrom(courses[ci][1], tp), "遥克"
	}

	// 别责:四课不全三课备、无克无遥——刚日取干合之干寄宫上神,
	// 柔日取日支三合前一辰(支+4),中末俱并干上
	// (课经集走例钉死:丙辰日卯时辰将→亥午午)。
	distinct := map[[2]int]bool{}
	for _, c := range courses {
		distinct[c] = true
	}
	if len(distinct) == 3 {
		var chu int
		if yangDay {
			heGan := (dayStem + 5) % 10 // 五合:甲己 乙庚 丙辛 丁壬 戊癸
			chu = tp[stemJiGong[heGan]]
		} else {
			chu = (dayBranch + 4) % 12
		}
		return [3]int{chu, ganUpper, ganUpper}, "别责"
	}

	// 昴星:阳日取酉上神(地盘酉之天盘),中支上末干上;阴日取酉下神(天盘酉临之地盘),中干上末支上
	if yangDay {
		chu := tp[9] // 地盘酉上之天盘
		return [3]int{chu, zhiUpper, ganUpper}, "昴星"
	}
	you := 0
	for i := 0; i < 12; i++ {
		if tp[i] == 9 {
			you = i
		}
	}
	return [3]int{you, ganUpper, zhiUpper}, "昴星"
}
