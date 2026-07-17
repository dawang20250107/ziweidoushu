// Package meihua 梅花易数起卦与断卦引擎(邵康节体系)。
//
// 起卦:时间起卦(农历年支数+月+日 → 上卦;再加时辰数 → 下卦;总数取六余 → 动爻)
// 与数字起卦(两数/三数)。推导:互卦(二三四/三四五爻)、变卦(动爻翻转)、
// 体用(动爻所在为用、静者为体)、五行生克吉凶倾向。
// 断辞与类象由知识库/AI 层补充,本包只负责可验证的卦理计算。
//
// 正确性基准:邵康节《梅花易数》观梅占等经典占例单测钉死。
package meihua

import (
	"fmt"
	"time"

	"github.com/6tail/lunar-go/calendar"
)

// Trigram 八卦(先天数 1-8:乾兑离震巽坎艮坤)。
type Trigram struct {
	Num     int    `json:"num"`     // 先天卦数 1-8
	Name    string `json:"name"`    // 乾/兑/离/震/巽/坎/艮/坤
	Symbol  string `json:"symbol"`  // ☰☱☲☳☴☵☶☷
	Nature  string `json:"nature"`  // 天泽火雷风水山地
	Element string `json:"element"` // 金金火木木水土土
	// Lines 三爻,自下而上,true=阳爻。
	Lines [3]bool `json:"lines"`
}

// 先天八卦表(索引 = 先天数-1)。
var trigrams = []Trigram{
	{1, "乾", "☰", "天", "金", [3]bool{true, true, true}},
	{2, "兑", "☱", "泽", "金", [3]bool{true, true, false}},
	{3, "离", "☲", "火", "火", [3]bool{true, false, true}},
	{4, "震", "☳", "雷", "木", [3]bool{true, false, false}},
	{5, "巽", "☴", "风", "木", [3]bool{false, true, true}},
	{6, "坎", "☵", "水", "水", [3]bool{false, true, false}},
	{7, "艮", "☶", "山", "土", [3]bool{false, false, true}},
	{8, "坤", "☷", "地", "土", [3]bool{false, false, false}},
}

// TrigramByNum 先天数 → 八卦(1-8;0 视作 8)。
func TrigramByNum(n int) Trigram {
	n = ((n-1)%8 + 8) % 8
	return trigrams[n]
}

// trigramByLines 三爻 → 八卦。
func trigramByLines(lines [3]bool) Trigram {
	for _, t := range trigrams {
		if t.Lines == lines {
			return t
		}
	}
	panic("unreachable")
}

// hexagramNames 六十四卦名,[上卦先天数-1][下卦先天数-1]。
var hexagramNames = [8][8]string{
	{"乾为天", "天泽履", "天火同人", "天雷无妄", "天风姤", "天水讼", "天山遁", "天地否"},
	{"泽天夬", "兑为泽", "泽火革", "泽雷随", "泽风大过", "泽水困", "泽山咸", "泽地萃"},
	{"火天大有", "火泽睽", "离为火", "火雷噬嗑", "火风鼎", "火水未济", "火山旅", "火地晋"},
	{"雷天大壮", "雷泽归妹", "雷火丰", "震为雷", "雷风恒", "雷水解", "雷山小过", "雷地豫"},
	{"风天小畜", "风泽中孚", "风火家人", "风雷益", "巽为风", "风水涣", "风山渐", "风地观"},
	{"水天需", "水泽节", "水火既济", "水雷屯", "水风井", "坎为水", "水山蹇", "水地比"},
	{"山天大畜", "山泽损", "山火贲", "山雷颐", "山风蛊", "山水蒙", "艮为山", "山地剥"},
	{"地天泰", "地泽临", "地火明夷", "地雷复", "地风升", "地水师", "地山谦", "坤为地"},
}

// Hexagram 六十四卦。
type Hexagram struct {
	Name  string  `json:"name"`
	Upper Trigram `json:"upper"`
	Lower Trigram `json:"lower"`
	// Lines 六爻,自下而上(1-6 爻),true=阳爻。
	Lines [6]bool `json:"lines"`
}

func makeHexagram(upper, lower Trigram) Hexagram {
	var lines [6]bool
	copy(lines[:3], lower.Lines[:])
	copy(lines[3:], upper.Lines[:])
	return Hexagram{
		Name:  hexagramNames[upper.Num-1][lower.Num-1],
		Upper: upper,
		Lower: lower,
		Lines: lines,
	}
}

// Relation 体用生克关系。
type Relation string

const (
	RelYongShengTi Relation = "用生体" // 大吉
	RelBiHe        Relation = "比和"  // 吉
	RelTiKeYong    Relation = "体克用" // 小吉(费力得之)
	RelTiShengYong Relation = "体生用" // 小凶(泄耗)
	RelYongKeTi    Relation = "用克体" // 凶
)

// elementRelation a 对 b:生/克/同。五行:金水木火土相生;金木土水火相克。
func sheng(a, b string) bool {
	pairs := map[string]string{"金": "水", "水": "木", "木": "火", "火": "土", "土": "金"}
	return pairs[a] == b
}
func ke(a, b string) bool {
	pairs := map[string]string{"金": "木", "木": "土", "土": "水", "水": "火", "火": "金"}
	return pairs[a] == b
}

// judge 体用断吉凶。
func judge(ti, yong string) (Relation, string) {
	switch {
	case ti == yong:
		return RelBiHe, "体用比和,谋事顺遂,吉。"
	case sheng(yong, ti):
		return RelYongShengTi, "用生体,得外力相助,事易成,大吉。"
	case ke(ti, yong):
		return RelTiKeYong, "体克用,事可成但需费力经营,小吉。"
	case sheng(ti, yong):
		return RelTiShengYong, "体生用,气耗于外,劳而少功,小凶。"
	default:
		return RelYongKeTi, "用克体,外势压身,谋事多阻,凶。"
	}
}

// Result 一次占卜的完整卦象。
type Result struct {
	Method   string `json:"method"` // time | number
	Question string `json:"question,omitempty"`

	// 起卦参数溯源(可复现)
	LunarText string `json:"lunarText,omitempty"` // 时间起卦:农历「辰年十二月十七日申时」
	Numbers   []int  `json:"numbers,omitempty"`   // 数字起卦的原始数

	Ben    Hexagram `json:"ben"`    // 本卦
	Hu     Hexagram `json:"hu"`     // 互卦
	Bian   Hexagram `json:"bian"`   // 变卦
	Moving int      `json:"moving"` // 动爻 1-6

	// 体用
	TiTrigram   Trigram  `json:"tiTrigram"`
	YongTrigram Trigram  `json:"yongTrigram"`
	TiIsUpper   bool     `json:"tiIsUpper"`
	Relation    Relation `json:"relation"`
	Verdict     string   `json:"verdict"` // 吉凶倾向一句话(卦理层,非断辞)
}

// derive 由上卦数/下卦数/动爻组装完整结果。
func derive(upperN, lowerN, moving int) Result {
	upper := TrigramByNum(upperN)
	lower := TrigramByNum(lowerN)
	ben := makeHexagram(upper, lower)

	// 互卦:二三四爻为下互,三四五爻为上互
	huLower := trigramByLines([3]bool{ben.Lines[1], ben.Lines[2], ben.Lines[3]})
	huUpper := trigramByLines([3]bool{ben.Lines[2], ben.Lines[3], ben.Lines[4]})
	hu := makeHexagram(huUpper, huLower)

	// 变卦:动爻阴阳翻转
	bianLines := ben.Lines
	bianLines[moving-1] = !bianLines[moving-1]
	bian := makeHexagram(
		trigramByLines([3]bool{bianLines[3], bianLines[4], bianLines[5]}),
		trigramByLines([3]bool{bianLines[0], bianLines[1], bianLines[2]}),
	)

	// 体用:动爻所在之卦为用,静者为体
	movingInUpper := moving >= 4
	ti, yong := lower, upper
	if !movingInUpper {
		ti, yong = upper, lower
	}
	rel, verdict := judge(ti.Element, yong.Element)

	return Result{
		Ben: ben, Hu: hu, Bian: bian, Moving: moving,
		TiTrigram: ti, YongTrigram: yong, TiIsUpper: !movingInUpper,
		Relation: rel, Verdict: verdict,
	}
}

// branchNum 地支序数:子1 丑2 … 亥12(梅花年支/时辰数)。
func branchNum(branch string) int {
	order := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	for i, b := range order {
		if b == branch {
			return i + 1
		}
	}
	return 0
}

// ByTime 时间起卦:公历时刻 → 农历(年支数+月+日)为上卦,加时辰数为下卦,
// 总和取六余为动爻(余零作六)。
func ByTime(t time.Time, question string) (Result, error) {
	if t.Year() < 1902 || t.Year() > 2098 {
		return Result{}, fmt.Errorf("时间超出支持范围(1902-2098)")
	}
	solar := calendar.NewSolarFromDate(t)
	lunar := solar.GetLunar()

	yearN := branchNum(lunar.GetYearZhi())
	monthN := lunar.GetMonth()
	if monthN < 0 { // lunar-go 闰月为负,按当月数起卦
		monthN = -monthN
	}
	dayN := lunar.GetDay()
	hourN := branchNum(lunar.GetTimeZhi())

	upperSum := yearN + monthN + dayN
	lowerSum := upperSum + hourN
	upperN := upperSum % 8
	if upperN == 0 {
		upperN = 8
	}
	lowerN := lowerSum % 8
	if lowerN == 0 {
		lowerN = 8
	}
	moving := lowerSum % 6
	if moving == 0 {
		moving = 6
	}

	r := derive(upperN, lowerN, moving)
	r.Method = "time"
	r.Question = question
	r.LunarText = fmt.Sprintf("%s年%s月%s日%s时",
		lunar.GetYearZhi(), lunar.GetMonthInChinese(), lunar.GetDayInChinese(), lunar.GetTimeZhi())
	return r, nil
}

// ByNumbers 数字起卦:两数(前上后下,和取动爻)或三数(第三数定动爻)。
func ByNumbers(nums []int, question string) (Result, error) {
	if len(nums) != 2 && len(nums) != 3 {
		return Result{}, fmt.Errorf("数字起卦需两个或三个正整数")
	}
	for _, n := range nums {
		if n <= 0 || n > 999 {
			return Result{}, fmt.Errorf("起卦数须为 1-999 的正整数")
		}
	}
	upperN := nums[0] % 8
	if upperN == 0 {
		upperN = 8
	}
	lowerN := nums[1] % 8
	if lowerN == 0 {
		lowerN = 8
	}
	movingBase := nums[0] + nums[1]
	if len(nums) == 3 {
		movingBase = nums[2]
	}
	moving := movingBase % 6
	if moving == 0 {
		moving = 6
	}
	r := derive(upperN, lowerN, moving)
	r.Method = "number"
	r.Question = question
	r.Numbers = nums
	return r, nil
}
