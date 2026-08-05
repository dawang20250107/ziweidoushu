// 十二天将布盘(P34 补全,批次五语料跨书互证):
//
// 贵人歌(《精校本六壬大全·贵人歌》与《大六壬指南注解·神煞》一致):
// 甲戊庚丑未、乙己子申、丙丁亥酉、壬癸巳卯、六辛午寅——首支为昼贵、次支夜贵。
// 昼夜以占时分:卯辰巳午未申为昼,酉戌亥子丑寅为夜。
// 布序:贵人所临地盘位在亥子丑寅卯辰(背天门)顺布,巳午未申酉戌(向地户)逆布
// (大全:「以课之天盘起贵神之例,地盘定顺逆之序」)。
// 序次:贵人 螣蛇 朱雀 六合 勾陈 青龙 天空 白虎 太常 玄武 太阴 天后。
package daliuren

// TianJiangNames 十二天将序。
var TianJiangNames = []string{"贵人", "螣蛇", "朱雀", "六合", "勾陈", "青龙", "天空", "白虎", "太常", "玄武", "太阴", "天后"}

// 昼贵/夜贵(索引=日干 0甲…9癸)。
var dayGuiByStem = [10]int{1, 0, 11, 11, 1, 0, 1, 6, 5, 5}
var nightGuiByStem = [10]int{7, 8, 9, 9, 7, 8, 7, 2, 3, 3}

// jiangNote 天将简断(通行壬学类神义,原创短语)。
var jiangNote = map[string]string{
	"贵人": "得贵人之助,凶中有解",
	"螣蛇": "主惊恐怪异、虚惊缠绕",
	"朱雀": "主文书信息、口舌是非",
	"六合": "主和合成就、交易婚媾",
	"勾陈": "主勾连迟滞、田土争讼",
	"青龙": "主财喜生气、婚财之庆",
	"天空": "主虚诈不实、谋事落空",
	"白虎": "主道路凶丧、疾病刀兵",
	"太常": "主衣冠宴饮、平稳之庆",
	"玄武": "主盗失暗昧、阴私小人",
	"太阴": "主阴私庇佑、暗中得助",
	"天后": "主阴泽妇人、柔缓之情",
}

// placeTianJiang 按日干与占时布十二天将。tp 为天盘(地盘位 i 上所乘之神),
// 返回 jiang[i] = 地盘 i 位上所乘天将之序号,及贵人所临地盘位。
func placeTianJiang(tp [12]int, dayStem, hour int) (jiang [12]int, guiPos int, isDay bool) {
	isDay = hour >= 3 && hour <= 8 // 卯(3)~申(8)为昼
	gui := nightGuiByStem[dayStem]
	if isDay {
		gui = dayGuiByStem[dayStem]
	}
	guiPos = 0
	for i := 0; i < 12; i++ {
		if tp[i] == gui {
			guiPos = i
			break
		}
	}
	// 亥(11)子(0)丑(1)寅(2)卯(3)辰(4)顺布,余逆布
	forward := guiPos == 11 || guiPos <= 4
	for k := 0; k < 12; k++ {
		pos := (guiPos + k) % 12
		if !forward {
			pos = ((guiPos-k)%12 + 12) % 12
		}
		jiang[pos] = k
	}
	return jiang, guiPos, isDay
}
