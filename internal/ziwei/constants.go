package ziwei

// Stems 十天干。
var Stems = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}

// Branches 十二地支(0=子)。
var Branches = []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

// PalaceNamesClockwise 从命宫起顺时针的宫名序列(iztro zh-CN 口径)。
// 顺时针:命宫→父母→福德→田宅→官禄→仆役→迁移→疾厄→财帛→子女→夫妻→兄弟。
var PalaceNamesClockwise = []string{
	"命宫", "父母", "福德", "田宅", "官禄", "仆役",
	"迁移", "疾厄", "财帛", "子女", "夫妻", "兄弟",
}

// ChineseTimeNames 时辰名,下标即 timeIndex(12=晚子时)。
var ChineseTimeNames = []string{
	"早子时", "丑时", "寅时", "卯时", "辰时", "巳时",
	"午时", "未时", "申时", "酉时", "戌时", "亥时", "晚子时",
}

// TimeRanges 时辰对应钟点区间,下标即 timeIndex。
var TimeRanges = []string{
	"00:00~01:00", "01:00~03:00", "03:00~05:00", "05:00~07:00", "07:00~09:00", "09:00~11:00",
	"11:00~13:00", "13:00~15:00", "15:00~17:00", "17:00~19:00", "19:00~21:00", "21:00~23:00", "23:00~00:00",
}

// tigerRule 五虎遁:年干 → 寅月/寅宫起始天干索引。
// 甲己起丙,乙庚起戊,丙辛起庚,丁壬起壬,戊癸起甲。
var tigerRule = [10]int{2, 4, 6, 8, 0, 2, 4, 6, 8, 0}

// fiveElementsJu 五行 → 局数。
var juNames = map[int]string{2: "水二局", 3: "木三局", 4: "金四局", 5: "土五局", 6: "火六局"}

// soulStarByBranch 命主星(通用派:按命宫地支,0=子)。
var soulStarByBranch = [12]string{
	"贪狼", "巨门", "禄存", "文曲", "廉贞", "武曲",
	"破军", "武曲", "廉贞", "文曲", "禄存", "巨门",
}

// bodyStarByBranch 身主星(按年支,0=子)。
var bodyStarByBranch = [12]string{
	"火星", "天相", "天梁", "天同", "文昌", "天机",
	"火星", "天相", "天梁", "天同", "文昌", "天机",
}

// zodiacByBranch 生肖(按年支,0=子)。
var zodiacByBranch = [12]string{"鼠", "牛", "虎", "兔", "龙", "蛇", "马", "羊", "猴", "鸡", "狗", "猪"}

// fixIndex 将索引规整到 [0, max) 区间。
func fixIndex(index, max int) int {
	return ((index % max) + max) % max
}

// fix12 十二宫索引规整。
func fix12(index int) int { return fixIndex(index, 12) }

// fix10 天干索引规整。
func fix10(index int) int { return fixIndex(index, 10) }

// stemIndexOf 天干名 → 索引,未知返回 -1。
func stemIndexOf(name string) int {
	for i, s := range Stems {
		if s == name {
			return i
		}
	}
	return -1
}

// branchIndexOf 地支名 → 索引,未知返回 -1。
func branchIndexOf(name string) int {
	for i, b := range Branches {
		if b == name {
			return i
		}
	}
	return -1
}

// palaceIndexToBranch iztro 宫位索引(0=寅)→ 地支索引(0=子)。
func palaceIndexToBranch(palaceIdx int) int { return fix12(palaceIdx + 2) }

// branchToPalaceIndex 地支索引(0=子)→ iztro 宫位索引(0=寅)。
func branchToPalaceIndex(branch int) int { return fix12(branch - 2) }
