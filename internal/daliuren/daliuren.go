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
	return r, nil
}

// ── 九宗门三传推导(理法易简简化口径)──────────────────────

// keSheng a 生 b / a 克 b(五行索引)。
func ke(a, b int) bool { return (branchElement[a]+2)%5 == branchElement[b] } // a 克 b

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

	// 收集克:下贼上(下克上)优先,否则上克下
	var zei, keUp []int // 存 course 索引
	for i, c := range courses {
		lo, up := c[0], c[1]
		if ke(lo, up) {
			zei = append(zei, i)
		} else if ke(up, lo) {
			keUp = append(keUp, i)
		}
	}
	cand := zei
	if len(cand) == 0 {
		cand = keUp
	}

	pickByClass := func(idxs []int) int { // 涉害:孟>仲>季,同类取课序小
		best := idxs[0]
		for _, i := range idxs[1:] {
			bi, bb := classRank(courses[i][1]), classRank(courses[best][1])
			if bi < bb {
				best = i
			}
		}
		return best
	}
	pickCand := func(idxs []int) (int, string) { // 贼克/比用/涉害
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
		return pickByClass(idxs), "涉害"
	}

	// 伏吟
	if fuyin {
		var chu int
		if len(cand) > 0 {
			ci, _ := pickCand(cand)
			chu = courses[ci][1]
		} else if yangDay {
			chu = ganUpper
		} else {
			chu = zhiUpper
		}
		mid := xing[chu]
		if mid == chu {
			mid = (chu + 6) % 12 // 自刑取冲
		}
		end := xing[mid]
		if end == mid {
			end = (mid + 6) % 12
		}
		return [3]int{chu, mid, end}, "伏吟"
	}
	// 返吟
	if fanyin {
		if len(cand) > 0 {
			ci, _ := pickCand(cand)
			return chuanFrom(courses[ci][1], tp), "返吟"
		}
		return [3]int{yiMa(dayBranch), zhiUpper, ganUpper}, "返吟"
	}

	// 贼克/比用/涉害
	if len(cand) > 0 {
		ci, name := pickCand(cand)
		return chuanFrom(courses[ci][1], tp), name
	}

	// 遥克:日干(以寄宫论)与上神相克
	var yao []int
	jiEl := branchElement[jiGong]
	for i, c := range courses {
		up := c[1]
		if (jiEl+2)%5 == branchElement[up] || (branchElement[up]+2)%5 == jiEl {
			yao = append(yao, i)
		}
	}
	if len(yao) > 0 {
		// 先取克干者(上克下之遥),再比用/孟仲季;简化取比用/课序
		ci, _ := pickCand(yao)
		return chuanFrom(courses[ci][1], tp), "遥克"
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
