// 年支系补充杂曜:大耗、龙德、劫煞——《紫微斗数全书》诸星安法有载,
// iztro 星表未收,故独立放入 Palace.ExtraStars(不入 Stars,黄金基准不受影响)。
// 断语层与盘面「大师」密度档展示;跨软件对照(文墨天机)已核对落宫一致。
package ziwei

// daHaoByYearBranch 大耗(年支):子未 丑午 寅酉 卯申 辰亥 巳戌
// 午丑 未子 申卯 酉寅 戌巳 亥辰(阳支顺七、阴支顺五)。主耗散、破财。
var daHaoByYearBranch = [12]int{7, 6, 9, 8, 11, 10, 1, 0, 3, 2, 5, 4}

// jieShaByYearBranch 劫煞(年支三合):申子辰在巳、亥卯未在申、
// 寅午戌在亥、巳酉丑在寅(三合五行绝地)。主劫夺、盗失、小人。
var jieShaByYearBranch = [12]int{5, 2, 11, 8, 5, 2, 11, 8, 5, 2, 11, 8}

// placeExtraStars 按生年支安补充杂曜。龙德 = 年支顺行第八位(与岁前十二神
// 龙德同法定盘),主逢凶化解、贵气缓冲。
func placeExtraStars(palaces []Palace, yearBranch int) {
	put := func(branch int, name string) {
		for i := range palaces {
			if palaces[i].Branch == branch {
				palaces[i].ExtraStars = append(palaces[i].ExtraStars, Star{Name: name, Type: StarMinor})
				return
			}
		}
	}
	put(daHaoByYearBranch[fix12(yearBranch)], "大耗")
	put(fix12(yearBranch+7), "龙德")
	put(jieShaByYearBranch[fix12(yearBranch)], "劫煞")
}
