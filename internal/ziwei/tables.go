package ziwei

// 本文件为星曜静态数据表,全部由 iztro 2.5.8 源码直录:
//   - lib/data/stars.js         (STARS_INFO 亮度表)
//   - lib/data/heavenlyStems.js (十天干四化表)
//   - lib/star/majorStar.js     (十四主星顺序与亮度取用方式)
//   - lib/star/minorStar.js     (十四辅星 iztro type)
//   - lib/i18n/locales/zh-CN/star.js / brightness.js (key→中文对照)
// 业务分类口径(sha/lucky 集合与三档亮度)直录自本仓库 lib/ziwei/algorithm.ts。

import "strings"

// ─── 1. 十四主星 ────────────────────────────────────────────────
//
// MajorStars 十四主星中文名,顺序按 iztro majorStar.js 源码:
//
//	紫微系 ziweiGroup  = ['ziweiMaj','tianjiMaj','','taiyangMaj','wuquMaj','tiantongMaj','','','lianzhenMaj']
//	天府系 tianfuGroup = ['tianfuMaj','taiyinMaj','tanlangMaj','jumenMaj','tianxiangMaj',
//	                     'tianliangMaj','qishaMaj','','','','pojunMaj']
//
// (跳过空位,星名按 zh-CN 字典翻译)
var MajorStars = []string{
	"紫微", "天机", "太阳", "武曲", "天同", "廉贞",
	"天府", "太阴", "贪狼", "巨门", "天相", "天梁", "七杀", "破军",
}

// ─── 2. 亮度表 ──────────────────────────────────────────────────
//
// brightnessTable 星曜亮度表,直录自 iztro lib/data/stars.js STARS_INFO。
//
// ⚠️ 下标基准结论:数组下标 0 对应【寅宫】,顺行至下标 11 = 丑宫。证据:
//  1. stars.js 第 9 行注释:「1. 亮度(bright), 按照宫位地支排序(从寅开始)」;
//  2. majorStar.js 第 15 行注释:「安主星,寅宫下标为0」,且第 66 行调用
//     getBrightness((0,i18n_1.t)(s), (0,utils_1.fixIndex)(ziweiIndex - i)),
//     传入的正是「寅=0」的宫位索引;
//  3. utils/index.js 第 66 行 getBrightness 实现:
//     return t(targetBrightness[fixIndex(index)]) —— 用宫位索引直查数组,无偏移。
//
// 亮度码 zh-CN 翻译(lib/i18n/locales/zh-CN/brightness.js):
//
//	miao=庙 wang=旺 de=得 li=利 ping=平 bu=不 xian=陷,空串表示该宫无亮度记载。
//
// 以下 20 颗星按源码 STARS_INFO 原始顺序、原始数组顺序直录(下标 0=寅):
var brightnessTable = map[string][12]string{
	// ziweiMaj
	"紫微": {"旺", "旺", "得", "旺", "庙", "庙", "旺", "旺", "得", "旺", "平", "庙"},
	// tianjiMaj
	"天机": {"得", "旺", "利", "平", "庙", "陷", "得", "旺", "利", "平", "庙", "陷"},
	// taiyangMaj
	"太阳": {"旺", "庙", "旺", "旺", "旺", "得", "得", "陷", "不", "陷", "陷", "不"},
	// wuquMaj
	"武曲": {"得", "利", "庙", "平", "旺", "庙", "得", "利", "庙", "平", "旺", "庙"},
	// tiantongMaj
	"天同": {"利", "平", "平", "庙", "陷", "不", "旺", "平", "平", "庙", "旺", "不"},
	// lianzhenMaj
	"廉贞": {"庙", "平", "利", "陷", "平", "利", "庙", "平", "利", "陷", "平", "利"},
	// tianfuMaj
	"天府": {"庙", "得", "庙", "得", "旺", "庙", "得", "旺", "庙", "得", "庙", "庙"},
	// taiyinMaj
	"太阴": {"旺", "陷", "陷", "陷", "不", "不", "利", "不", "旺", "庙", "庙", "庙"},
	// tanlangMaj
	"贪狼": {"平", "利", "庙", "陷", "旺", "庙", "平", "利", "庙", "陷", "旺", "庙"},
	// jumenMaj
	"巨门": {"庙", "庙", "陷", "旺", "旺", "不", "庙", "庙", "陷", "旺", "旺", "不"},
	// tianxiangMaj
	"天相": {"庙", "陷", "得", "得", "庙", "得", "庙", "陷", "得", "得", "庙", "庙"},
	// tianliangMaj
	"天梁": {"庙", "庙", "庙", "陷", "庙", "旺", "陷", "得", "庙", "陷", "庙", "旺"},
	// qishaMaj
	"七杀": {"庙", "旺", "庙", "平", "旺", "庙", "庙", "庙", "庙", "平", "旺", "庙"},
	// pojunMaj
	"破军": {"得", "陷", "旺", "平", "庙", "旺", "得", "陷", "旺", "平", "庙", "旺"},
	// wenchangMin
	"文昌": {"陷", "利", "得", "庙", "陷", "利", "得", "庙", "陷", "利", "得", "庙"},
	// wenquMin
	"文曲": {"平", "旺", "得", "庙", "陷", "旺", "得", "庙", "陷", "旺", "得", "庙"},
	// huoxingMin
	"火星": {"庙", "利", "陷", "得", "庙", "利", "陷", "得", "庙", "利", "陷", "得"},
	// lingxingMin
	"铃星": {"庙", "利", "陷", "得", "庙", "利", "陷", "得", "庙", "利", "陷", "得"},
	// qingyangMin
	"擎羊": {"", "陷", "庙", "", "陷", "庙", "", "陷", "庙", "", "陷", "庙"},
	// tuoluoMin
	"陀罗": {"陷", "", "庙", "陷", "", "庙", "陷", "", "庙", "陷", "", "庙"},
}

// brightnessOf 按地支索引(0=子)查星曜亮度,无亮度记载返回空串。
//
// 换算推导:brightnessTable 下标基准是「寅=0」的宫位索引(见上),
// 而本包对外统一用「子=0」的地支索引。寅的地支索引为 2,即
//
//	地支索引 branch 对应的源码下标 = (branch - 2 + 12) % 12
//
// (branch=2 寅 → 0;branch=3 卯 → 1;……;branch=1 丑 → 11),
// 恰为 constants.go 中 branchToPalaceIndex 的定义,与 iztro
// utils/index.js fixEarthlyBranchIndex(减去寅的下标再 fixIndex)语义一致。
func brightnessOf(starName string, branch int) string {
	row, ok := brightnessTable[starName]
	if !ok {
		return ""
	}
	return row[branchToPalaceIndex(branch)]
}

// ─── 3. 十天干四化表 ────────────────────────────────────────────
//
// mutagenTable 十天干(0=甲 ... 9=癸)→ [化禄, 化权, 化科, 化忌] 星名,
// 直录自 iztro lib/data/heavenlyStems.js 各天干的 mutagen 数组
// (顺序即源码 MUTAGEN = ['sihuaLu','sihuaQuan','sihuaKe','sihuaJi'],
// 星名 key 按 zh-CN 字典翻译):
//
//	甲:廉破武阳  乙:机梁紫月  丙:同机昌廉  丁:阴同机巨  戊:贪月弼机
//	己:武贪梁曲  庚:阳武阴同  辛:巨日曲昌  壬:梁紫左武  癸:破巨阴贪
var mutagenTable = [10][4]string{
	{"廉贞", "破军", "武曲", "太阳"}, // 甲 jiaHeavenly
	{"天机", "天梁", "紫微", "太阴"}, // 乙 yiHeavenly
	{"天同", "天机", "文昌", "廉贞"}, // 丙 bingHeavenly
	{"太阴", "天同", "天机", "巨门"}, // 丁 dingHeavenly
	{"贪狼", "太阴", "右弼", "天机"}, // 戊 wuHeavenly
	{"武曲", "贪狼", "天梁", "文曲"}, // 己 jiHeavenly
	{"太阳", "武曲", "太阴", "天同"}, // 庚 gengHeavenly
	{"巨门", "太阳", "文曲", "文昌"}, // 辛 xinHeavenly
	{"天梁", "紫微", "左辅", "武曲"}, // 壬 renHeavenly
	{"破军", "巨门", "太阴", "贪狼"}, // 癸 guiHeavenly
}

// ─── 4. 业务口径:煞星/吉星集合 ─────────────────────────────────
//
// shaStarNames / luckyStarNames 原样直录自 lib/ziwei/algorithm.ts 的
// SHA_STARS / LUCKY_STARS(仓库既有业务口径,优先于 iztro type)。

var shaStarNames = map[string]bool{
	"擎羊": true, "陀罗": true, "火星": true, "铃星": true, "地空": true, "地劫": true,
	"天空": true, "旬空": true, "截路": true, "大耗": true, "天使": true, "天伤": true,
}

var luckyStarNames = map[string]bool{
	"文昌": true, "文曲": true, "左辅": true, "右弼": true, "天魁": true, "天钺": true,
	"禄存": true, "天马": true, "天官": true, "天福": true, "天才": true, "天寿": true,
	"三台": true, "八座": true, "恩光": true, "天贵": true, "台辅": true, "龙池": true,
	"凤阁": true, "红鸾": true, "天喜": true, "孤辰": true, "寡宿": true,
}

// ─── 5. 星曜分类 ────────────────────────────────────────────────
//
// classifyStar 移植 algorithm.ts mapStarType:
// 先查 SHA_STARS 集合,再查 LUCKY_STARS 集合,最后按 iztro type 归类。
func classifyStar(name string, iztroType string) StarType {
	if shaStarNames[name] {
		return StarSha
	}
	if luckyStarNames[name] {
		return StarLucky
	}
	t := strings.ToLower(iztroType)
	switch t {
	case "主星", "major":
		return StarMajor
	case "煞星", "tough":
		return StarSha
	case "吉星", "soft", "禄存", "天马":
		return StarLucky
	}
	return StarMinor
}

// ─── 6. 三档亮度 ────────────────────────────────────────────────
//
// mapBrightnessLevel 移植 algorithm.ts mapBrightness:
// 庙/旺 → "bright",陷/不 → "dim",其余(得/利/平/空)→ "normal"。
// 源码返回类型为 'bright' | 'normal' | 'dim'。
func mapBrightnessLevel(b string) string {
	switch b {
	case "庙", "旺":
		return "bright"
	case "陷", "不":
		return "dim"
	}
	return "normal"
}

// ─── 7. 十四辅星 iztro type 对照表 ──────────────────────────────
//
// minorStarIztroType 星名 → iztro 原生 type,直录自 lib/star/minorStar.js
// 各 FunctionalStar 构造参数的 type 字段。
var minorStarIztroType = map[string]string{
	"左辅": "soft",
	"右弼": "soft",
	"文昌": "soft",
	"文曲": "soft",
	"天魁": "soft",
	"天钺": "soft",
	"禄存": "lucun",
	"天马": "tianma",
	"地空": "tough",
	"地劫": "tough",
	"火星": "tough",
	"铃星": "tough",
	"擎羊": "tough",
	"陀罗": "tough",
}
