// 万物类象(《梅花易数·八卦万物类占》义,结构化速查表):
// 断辞落到具体人事物时取象所需——盘面展示与 AI 提示词共用。
// 为速查节选常用类目,非原文全录;取象为传统类比,展示层须注明参考性质。
package meihua

// TrigramLore 一卦之常用类象。
type TrigramLore struct {
	Renlun  string `json:"renlun"`  // 人物
	Shenti  string `json:"shenti"`  // 身体
	Dongwu  string `json:"dongwu"`  // 动物
	Jingwu  string `json:"jingwu"`  // 静物/器物
	Fangwei string `json:"fangwei"` // 方位(后天)
	Tianshi string `json:"tianshi"` // 天时
	Xing    string `json:"xing"`    // 性情意象
}

// LoreOf 卦名 → 类象。
func LoreOf(name string) (TrigramLore, bool) {
	l, ok := trigramLore[name]
	return l, ok
}

var trigramLore = map[string]TrigramLore{
	"乾": {
		Renlun: "父、长者、官贵、名人", Shenti: "首、骨、肺",
		Dongwu: "马、天鹅、狮", Jingwu: "金玉、珠宝、圆物、冠镜",
		Fangwei: "西北", Tianshi: "晴、冰、寒", Xing: "刚健、尊贵、动而不息",
	},
	"兑": {
		Renlun: "少女、口舌之人、伶人、译人", Shenti: "口、舌、肺、痰涎",
		Dongwu: "羊、泽中之物", Jingwu: "金刃、金类、乐器、缺器",
		Fangwei: "西", Tianshi: "雨泽、新月", Xing: "喜悦、口舌、毁折、谗说",
	},
	"离": {
		Renlun: "中女、文人、目疾人、甲胄之士", Shenti: "目、心、上焦",
		Dongwu: "雉、龟、蟹、蚌", Jingwu: "火、书、文书、干戈、干燥物",
		Fangwei: "南", Tianshi: "晴、电、虹霓", Xing: "文明、亮丽、性急、依附",
	},
	"震": {
		Renlun: "长男、好动之人", Shenti: "足、肝、发、声音",
		Dongwu: "龙、蛇", Jingwu: "木竹、乐器之竹、萑苇",
		Fangwei: "东", Tianshi: "雷", Xing: "震动、起发、性急、多动少静",
	},
	"巽": {
		Renlun: "长女、僧尼道姑、商旅之人", Shenti: "股、气、风疾",
		Dongwu: "鸡、百禽、山林之鸟", Jingwu: "木香、绳索、扇、长物、竹木",
		Fangwei: "东南", Tianshi: "风", Xing: "柔顺、进退不果、鼓舞号令",
	},
	"坎": {
		Renlun: "中男、江湖之人、盗", Shenti: "耳、血、肾",
		Dongwu: "豕、鱼、水族", Jingwu: "水带子之物、弓轮、酒器",
		Fangwei: "北", Tianshi: "雨、雪、霜露", Xing: "陷险、内刚外柔、忧、劳",
	},
	"艮": {
		Renlun: "少男、闲人、山中人、童子", Shenti: "手指、鼻、背、脾胃",
		Dongwu: "犬、鼠、虎、黔喙之属", Jingwu: "土石、瓜果、门阙、径路",
		Fangwei: "东北", Tianshi: "云雾、山岚", Xing: "静止、笃实、迟滞、保守",
	},
	"坤": {
		Renlun: "母、老妇、农人、众人", Shenti: "腹、脾胃、肌肉",
		Dongwu: "牛、百兽、牝马", Jingwu: "布帛、五谷、舆、釜、瓦器",
		Fangwei: "西南", Tianshi: "阴云、雾气", Xing: "柔顺、安静、吝啬、承载",
	},
}
