// 紫微斗数格局识别（v2 严格化版本）
//
// 设计原则：
// 1. 古书条件优先：每个格局列出"必须 / 加分 / 破格"三层结构，出处可考
// 2. 倪师立场：不使用宫干自化、大限四化、来因宫等飞星派工具
// 3. 庙旺利陷：用 brightness 字段（bright=庙旺、normal=平、dim=陷）
// 4. 三方四正会照：命宫 + 财帛 + 官禄 + 迁移
// 5. 夹宫：命宫前后两宫
//
// 主要古籍出处：
//   - 《紫微斗数全集》（陈抟祖师传，明代刊本）
//   - 《紫微斗数全书》（罗洪先编，明代刊本）
//   - 《骨髓赋》《女命骨髓赋》《十二宫诸星得地合格诀》
//   - 倪海厦《天纪》紫微斗数讲义
package ziwei

import "strings"

// ────────────────── 类型 ──────────────────

// PatternCondition 成立条件分层（v2 新增）。
type PatternCondition struct {
	Required []string `json:"required"`           // 必须满足条件（已通过的）
	Bonus    []string `json:"bonus,omitempty"`    // 加分项（已触发）
	Breaking []string `json:"breaking,omitempty"` // 破格警示（已触发）
}

// Pattern 识别出的格局。Level 取值：excellent / good / neutral / caution。
type Pattern struct {
	Name        string            `json:"name"`
	Level       string            `json:"level"`
	Description string            `json:"description"`
	Palaces     []string          `json:"palaces"`              // 涉及宫位
	Conditions  *PatternCondition `json:"conditions,omitempty"` // 成立条件分层（v2 新增）
	Source      string            `json:"source,omitempty"`     // 古籍出处（v2 新增）
}

// ────────────────── 常量 ──────────────────

var (
	shaNames = []string{"擎羊", "陀罗", "火星", "铃星", "地空", "地劫"}
	shaHard  = []string{"擎羊", "陀罗", "火星", "铃星"} // 四煞
	shaKong  = []string{"地空", "地劫"}             // 空劫
	zuoYou   = []string{"左辅", "右弼"}
	changQu  = []string{"文昌", "文曲"}
	kuiYue   = []string{"天魁", "天钺"}
)

var patternBranchNames = []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

// ────────────────── 辅助函数 ──────────────────

func containsString(list []string, name string) bool {
	for _, s := range list {
		if s == name {
			return true
		}
	}
	return false
}

func getMajorStarNames(palace *Palace) []string {
	var out []string
	for i := range palace.Stars {
		if palace.Stars[i].Type == StarMajor {
			out = append(out, palace.Stars[i].Name)
		}
	}
	return out
}

func findStarInPalace(palace *Palace, name string) *Star {
	for i := range palace.Stars {
		if palace.Stars[i].Name == name {
			return &palace.Stars[i]
		}
	}
	return nil
}

func hasStarInPalace(palace *Palace, name string) bool {
	return findStarInPalace(palace, name) != nil
}

func findStarPalace(c *Chart, name string) *Palace {
	for i := range c.Palaces {
		for j := range c.Palaces[i].Stars {
			if c.Palaces[i].Stars[j].Name == name {
				return &c.Palaces[i]
			}
		}
	}
	return nil
}

func getPalaceByBranch(c *Chart, branch int) *Palace {
	b := ((branch % 12) + 12) % 12
	for i := range c.Palaces {
		if c.Palaces[i].Branch == b {
			return &c.Palaces[i]
		}
	}
	return nil
}

func shaCountInPalace(palace *Palace, list []string) int {
	count := 0
	for i := range palace.Stars {
		if containsString(list, palace.Stars[i].Name) {
			count++
		}
	}
	return count
}

func hasShaInPalace(palace *Palace, list []string) bool {
	for i := range palace.Stars {
		if containsString(list, palace.Stars[i].Name) {
			return true
		}
	}
	return false
}

func getSanFangPalaces(c *Chart) []*Palace {
	m := c.MingGongBranch
	branches := []int{m, (m + 4) % 12, (m + 8) % 12, (m + 6) % 12}
	var out []*Palace
	// 与 TS 原版一致:chart.palaces 在 TS 中按 iztro 宫位序(寅起)排列,
	// 此处按同一顺序遍历,保证收集到的宫位顺序与原版完全相同。
	for i := 0; i < 12; i++ {
		b := (i + 2) % 12
		for _, want := range branches {
			if b == want {
				if p := c.PalaceByBranch(b); p != nil {
					out = append(out, p)
				}
				break
			}
		}
	}
	return out
}

func isInSanFang(c *Chart, branch int) bool {
	m := c.MingGongBranch
	for _, b := range []int{m, (m + 4) % 12, (m + 8) % 12, (m + 6) % 12} {
		if b == branch {
			return true
		}
	}
	return false
}

func getDuiGong(c *Chart, branch int) *Palace {
	return getPalaceByBranch(c, (branch+6)%12)
}

func getJiaPalaces(c *Chart, branch int) (prev, next *Palace) {
	return getPalaceByBranch(c, (branch+11)%12), getPalaceByBranch(c, (branch+1)%12)
}

func sanFangAllStars(c *Chart) map[string]bool {
	set := make(map[string]bool)
	for _, p := range getSanFangPalaces(c) {
		for i := range p.Stars {
			set[p.Stars[i].Name] = true
		}
	}
	return set
}

func sanFangShaCount(c *Chart, list []string) int {
	sum := 0
	for _, p := range getSanFangPalaces(c) {
		sum += shaCountInPalace(p, list)
	}
	return sum
}

func isBright(palace *Palace, starName string) bool {
	s := findStarInPalace(palace, starName)
	return s != nil && s.BrightnessLevel == "bright"
}

func isDim(palace *Palace, starName string) bool {
	s := findStarInPalace(palace, starName)
	return s != nil && s.BrightnessLevel == "dim"
}

func getStarSiHua(palace *Palace, starName string) SiHua {
	s := findStarInPalace(palace, starName)
	if s == nil {
		return ""
	}
	return s.SiHua
}

// ────────────────── 正格识别器 ──────────────────

// detectJunChenQingHui 君臣庆会：紫微入命，左辅右弼同会（同宫或三方）
func detectJunChenQingHui(c *Chart, ming *Palace, patterns *[]Pattern) {
	if !hasStarInPalace(ming, "紫微") {
		return
	}
	sanFangSet := sanFangAllStars(c)
	hasZuo := sanFangSet["左辅"]
	hasYou := sanFangSet["右弼"]
	if !hasZuo || !hasYou {
		return
	}

	required := []string{"紫微入命", "左辅右弼同会三方四正"}
	var bonus []string
	var breaking []string
	if sanFangSet["文昌"] || sanFangSet["文曲"] {
		bonus = append(bonus, "再会文昌或文曲")
	}
	if sanFangSet["天魁"] || sanFangSet["天钺"] {
		bonus = append(bonus, "魁钺贵人加照")
	}
	if getStarSiHua(ming, "紫微") == HuaQuan {
		bonus = append(bonus, "紫微化权")
	}
	if sanFangShaCount(c, shaKong) >= 2 {
		breaking = append(breaking, "地空地劫双夹会照（紫微忌空劫）")
	}

	level := "excellent"
	if len(breaking) > 0 {
		level = "good"
	}
	*patterns = append(*patterns, Pattern{
		Name:        "君臣庆会",
		Level:       level,
		Description: "紫微入命，左辅右弼同会，帝王得贤臣辅佐，主大富大贵、统御之命。一生贵人不绝，宜走政商高位、跨界领袖之途。",
		Palaces:     []string{"命宫"},
		Conditions:  &PatternCondition{Required: required, Bonus: bonus, Breaking: breaking},
		Source:      "《紫微斗数全书·君臣庆会格》",
	})
}

// detectZiFu 紫府同宫：紫微+天府于命宫（限寅、申宫）
func detectZiFu(c *Chart, ming *Palace, patterns *[]Pattern) {
	ziwei := findStarPalace(c, "紫微")
	tianfu := findStarPalace(c, "天府")
	if ziwei == nil || tianfu == nil || ziwei.Branch != tianfu.Branch {
		return
	}

	inMing := ziwei.Branch == c.MingGongBranch
	var required []string
	if inMing {
		required = []string{"紫微天府同入命宫"}
	} else {
		required = []string{"紫微天府同宫（不在命宫，会照减力）"}
	}
	var bonus []string
	var breaking []string
	sanFangSet := sanFangAllStars(c)
	if sanFangSet["左辅"] && sanFangSet["右弼"] {
		bonus = append(bonus, "左辅右弼同会")
	}
	if sanFangSet["文昌"] || sanFangSet["文曲"] {
		bonus = append(bonus, "再会昌曲")
	}
	if hasShaInPalace(ziwei, shaKong) {
		breaking = append(breaking, "紫府宫坐空劫（破紫府之贵气）")
	}
	if shaCountInPalace(ziwei, shaHard) >= 2 {
		breaking = append(breaking, "紫府宫见双煞同坐")
	}

	level := "good"
	if inMing && len(breaking) == 0 {
		level = "excellent"
	}
	description := "紫微天府同宫但未坐命，主一生有贵人贵气依托，但本身不一定大富贵，需看会照吉煞而定。"
	if inMing {
		description = "紫微天府同入命宫，帝相并临，尊贵之命。主品行端正、衣食无忧、有领导才能，宜担任要职。需要左右辅弼来配合方为完整大格。"
	}
	*patterns = append(*patterns, Pattern{
		Name:        "紫府同宫",
		Level:       level,
		Description: description,
		Palaces:     []string{ziwei.Name},
		Conditions:  &PatternCondition{Required: required, Bonus: bonus, Breaking: breaking},
		Source:      "《紫微斗数全书·紫府同宫格》",
	})
	_ = ming
}

// detectFuXiangChaoYuan 府相朝垣：天府、天相分别坐守命宫的三方四正
func detectFuXiangChaoYuan(c *Chart, ming *Palace, patterns *[]Pattern) {
	tianfu := findStarPalace(c, "天府")
	tianxiang := findStarPalace(c, "天相")
	if tianfu == nil || tianxiang == nil {
		return
	}
	if !isInSanFang(c, tianfu.Branch) || !isInSanFang(c, tianxiang.Branch) {
		return
	}
	if tianfu.Branch == c.MingGongBranch && tianxiang.Branch == c.MingGongBranch {
		return
	}
	if tianfu.Branch == tianxiang.Branch {
		return
	}

	required := []string{"天府坐命三方", "天相坐命三方", "两星不同宫"}
	var bonus []string
	var breaking []string
	if hasStarInPalace(ming, "禄存") || hasStarInPalace(ming, "化禄") {
		bonus = append(bonus, "命宫见禄")
	}
	if sanFangAllStars(c)["左辅"] {
		bonus = append(bonus, "再会左辅")
	}
	if hasShaInPalace(ming, shaHard) {
		breaking = append(breaking, "命宫坐煞星")
	}
	if sanFangShaCount(c, shaHard) >= 3 {
		breaking = append(breaking, "三方四正煞星过多")
	}

	level := "excellent"
	if len(breaking) > 0 {
		level = "good"
	}
	*patterns = append(*patterns, Pattern{
		Name:        "府相朝垣",
		Level:       level,
		Description: "天府天相分守命宫三方四正，文武并济、权印双辉，主一生衣食丰足、地位崇高。古书云\"府相朝垣千钟食禄\"，常见于政界、企业管理者。",
		Palaces:     []string{tianfu.Name, tianxiang.Name},
		Conditions:  &PatternCondition{Required: required, Bonus: bonus, Breaking: breaking},
		Source:      "《紫微斗数全书·府相朝垣格》",
	})
}

// detectYangLiangChangLu 阳梁昌禄：太阳+天梁+文昌+禄存四星会命宫，大贵格
func detectYangLiangChangLu(c *Chart, ming *Palace, patterns *[]Pattern) {
	sanFangSet := sanFangAllStars(c)
	if !sanFangSet["太阳"] || !sanFangSet["天梁"] ||
		!sanFangSet["文昌"] || !sanFangSet["禄存"] {
		return
	}

	sun := findStarPalace(c, "太阳")
	liang := findStarPalace(c, "天梁")
	required := []string{
		"太阳会命宫三方",
		"天梁会命宫三方",
		"文昌会命宫三方",
		"禄存会命宫三方",
	}
	var bonus []string
	var breaking []string
	if isBright(sun, "太阳") {
		bonus = append(bonus, "太阳庙旺")
	}
	if isBright(liang, "天梁") {
		bonus = append(bonus, "天梁庙旺")
	}
	if sanFangSet["化科"] {
		bonus = append(bonus, "再会化科")
	}
	if isDim(sun, "太阳") {
		breaking = append(breaking, "太阳落陷（阳梁失辉）")
	}
	if sanFangShaCount(c, shaHard) >= 2 {
		breaking = append(breaking, "三方煞重")
	}

	level := "excellent"
	if len(breaking) > 0 {
		level = "good"
	}
	*patterns = append(*patterns, Pattern{
		Name:        "阳梁昌禄",
		Level:       level,
		Description: "太阳、天梁、文昌、禄存四星齐会命宫三方，号称\"科举之星\"，主清贵显达、考运极佳，宜走学术、文教、研究、专业认证之路，一生功名易就。",
		Palaces:     []string{sun.Name, liang.Name},
		Conditions:  &PatternCondition{Required: required, Bonus: bonus, Breaking: breaking},
		Source:      "《紫微斗数全书·阳梁昌禄格》",
	})
	_ = ming
}

// detectHuoTanLingTan 火贪格 / 铃贪格：贪狼+火星 或 贪狼+铃星 同宫或会照
func detectHuoTanLingTan(c *Chart, ming *Palace, patterns *[]Pattern) {
	tan := findStarPalace(c, "贪狼")
	if tan == nil {
		return
	}
	huo := findStarPalace(c, "火星")
	ling := findStarPalace(c, "铃星")

	for _, pair := range []struct {
		shaName   string
		shaPalace *Palace
	}{{"火星", huo}, {"铃星", ling}} {
		shaName, shaPalace := pair.shaName, pair.shaPalace
		if shaPalace == nil {
			continue
		}
		sameOrTrine := tan.Branch == shaPalace.Branch ||
			(tan.Branch+4)%12 == shaPalace.Branch ||
			(tan.Branch+8)%12 == shaPalace.Branch ||
			(tan.Branch+6)%12 == shaPalace.Branch
		if !sameOrTrine {
			continue
		}
		if !isInSanFang(c, tan.Branch) {
			continue
		}

		sameWord := "会照"
		descWord := "三方会照"
		if tan.Branch == shaPalace.Branch {
			sameWord = "同宫"
			descWord = "同宫"
		}
		required := []string{"贪狼" + sameWord + shaName, "贪狼会照命宫三方"}
		var bonus []string
		var breaking []string
		if isBright(tan, "贪狼") {
			bonus = append(bonus, "贪狼庙旺")
		}
		if getStarSiHua(tan, "贪狼") == HuaLu || getStarSiHua(tan, "贪狼") == HuaQuan {
			bonus = append(bonus, "贪狼化禄/化权")
		}
		if hasShaInPalace(tan, []string{"擎羊", "陀罗"}) {
			breaking = append(breaking, "贪狼宫又见羊陀（破横发之力）")
		}
		if hasShaInPalace(tan, shaKong) {
			breaking = append(breaking, "贪狼遇空劫（财来财去）")
		}

		name := "铃贪格"
		if shaName == "火星" {
			name = "火贪格"
		}
		level := "excellent"
		breakingSuffix := ""
		if len(breaking) > 0 {
			level = "good"
			breakingSuffix = "本盘破格条件已触发，发力打折。"
		}
		*patterns = append(*patterns, Pattern{
			Name:        name,
			Level:       level,
			Description: "贪狼遇" + shaName + descWord + "，主突发横财、突如其来的机遇。古书云“贪狼遇火铃，必发横财”，但来得快去得也快，宜见好就收。" + breakingSuffix,
			Palaces:     []string{tan.Name, shaPalace.Name},
			Conditions:  &PatternCondition{Required: required, Bonus: bonus, Breaking: breaking},
			Source:      "《紫微斗数骨髓赋》",
		})
	}
	_ = ming
}

// detectWuTan 武贪格：武曲+贪狼 同宫（丑、未） 或 对照
func detectWuTan(c *Chart, ming *Palace, patterns *[]Pattern) {
	wu := findStarPalace(c, "武曲")
	tan := findStarPalace(c, "贪狼")
	if wu == nil || tan == nil {
		return
	}
	sameOrOppose := wu.Branch == tan.Branch || (wu.Branch+6)%12 == tan.Branch
	if !sameOrOppose {
		return
	}
	if !isInSanFang(c, wu.Branch) && !isInSanFang(c, tan.Branch) {
		return
	}

	first := "武曲贪狼对宫拱照"
	if wu.Branch == tan.Branch {
		first = "武曲贪狼同宫（丑/未）"
	}
	required := []string{first, "会照命宫三方"}
	var bonus []string
	var breaking []string
	if sanFangAllStars(c)["火星"] || sanFangAllStars(c)["铃星"] {
		bonus = append(bonus, "再遇火星/铃星（火贪/铃贪叠加）")
	}
	if getStarSiHua(wu, "武曲") == HuaLu {
		bonus = append(bonus, "武曲化禄")
	}
	if hasShaInPalace(wu, []string{"擎羊", "陀罗"}) {
		breaking = append(breaking, "武贪宫见羊陀")
	}
	if hasShaInPalace(wu, shaKong) {
		breaking = append(breaking, "武贪宫遇空劫")
	}

	level := "excellent"
	if len(breaking) > 0 {
		level = "good"
	}
	*patterns = append(*patterns, Pattern{
		Name:        "武贪格",
		Level:       level,
		Description: "武曲贪狼会命，财星与桃花欲望星交辉，古书云\"武贪不发少年人\"——三十岁后方能厚积薄发。主中年以后大富大贵，财源由人脉、应酬、欲望管理而来，适合金融、投机、销售、娱乐业。",
		Palaces:     []string{wu.Name, tan.Name},
		Conditions:  &PatternCondition{Required: required, Bonus: bonus, Breaking: breaking},
		Source:      "《紫微斗数骨髓赋》",
	})
	_ = ming
}

// detectShaPoLang 杀破狼：七杀、破军、贪狼三方齐聚
func detectShaPoLang(c *Chart, ming *Palace, patterns *[]Pattern) {
	sanFangSet := sanFangAllStars(c)
	var has []string
	for _, s := range []string{"七杀", "破军", "贪狼"} {
		if sanFangSet[s] {
			has = append(has, s)
		}
	}
	if len(has) < 3 {
		return
	}

	required := []string{"七杀、破军、贪狼三星齐入命宫三方四正"}
	var bonus []string
	var breaking []string
	if sanFangSet["化禄"] || sanFangSet["化权"] {
		bonus = append(bonus, "三方有化禄或化权（动得有力）")
	}
	if sanFangSet["左辅"] && sanFangSet["右弼"] {
		bonus = append(bonus, "辅弼同会（变动中得贵人）")
	}
	if sanFangShaCount(c, shaHard) >= 3 {
		breaking = append(breaking, "煞星过重（动而无成）")
	}
	if hasShaInPalace(ming, shaKong) {
		breaking = append(breaking, "命坐空劫（动得辛苦）")
	}

	var palaceNames []string
	for _, p := range getSanFangPalaces(c) {
		majors := getMajorStarNames(p)
		if len(majors) > 0 && containsString(has, majors[0]) {
			palaceNames = append(palaceNames, p.Name)
		}
	}

	level := "good"
	if len(breaking) > 0 {
		level = "caution"
	}
	*patterns = append(*patterns, Pattern{
		Name:        "杀破狼",
		Level:       level,
		Description: "七杀、破军、贪狼三星会命，开创闯荡之命格。一生变动多、不甘平凡，宜创业、军警、业务、销售。中年后才能稳定守成，年轻时易因冲动失利。",
		Palaces:     palaceNames,
		Conditions:  &PatternCondition{Required: required, Bonus: bonus, Breaking: breaking},
		Source:      "《紫微斗数全书·杀破狼》",
	})
}

// detectJiYueTongLiang 机月同梁：天机、太阴、天同、天梁四星齐入命迁财官
func detectJiYueTongLiang(c *Chart, ming *Palace, patterns *[]Pattern) {
	sanFangSet := sanFangAllStars(c)
	var has []string
	for _, s := range []string{"天机", "太阴", "天同", "天梁"} {
		if sanFangSet[s] {
			has = append(has, s)
		}
	}
	if len(has) < 4 {
		return
	}

	required := []string{"天机、太阴、天同、天梁四星齐入命宫三方四正"}
	var bonus []string
	var breaking []string
	if sanFangSet["文昌"] || sanFangSet["文曲"] {
		bonus = append(bonus, "再会昌曲")
	}
	if sanFangSet["化科"] {
		bonus = append(bonus, "再会化科")
	}
	if sanFangShaCount(c, shaHard) >= 3 {
		breaking = append(breaking, "煞星过多（机月同梁忌煞）")
	}
	if hasShaInPalace(ming, shaHard) {
		breaking = append(breaking, "命宫坐煞")
	}

	var palaceNames []string
	for _, p := range getSanFangPalaces(c) {
		majors := getMajorStarNames(p)
		matched := false
		for _, s := range has {
			if containsString(majors, s) {
				matched = true
				break
			}
		}
		if matched {
			palaceNames = append(palaceNames, p.Name)
		}
	}

	level := "excellent"
	if len(breaking) > 0 {
		level = "good"
	}
	*patterns = append(*patterns, Pattern{
		Name:        "机月同梁",
		Level:       level,
		Description: "天机太阴天同天梁四星齐入命迁财官，文质彬彬、聪慧善谋。最适合公职、学术、文艺、医疗、服务等需稳定累积的行业，不宜大冒险大投机。",
		Palaces:     palaceNames,
		Conditions:  &PatternCondition{Required: required, Bonus: bonus, Breaking: breaking},
		Source:      "《紫微斗数全书·机月同梁格》",
	})
}

// detectLianXiang 廉贞天相：同宫
func detectLianXiang(c *Chart, patterns *[]Pattern) {
	lian := findStarPalace(c, "廉贞")
	xiang := findStarPalace(c, "天相")
	if lian == nil || xiang == nil || lian.Branch != xiang.Branch {
		return
	}

	inMing := lian.Branch == c.MingGongBranch
	required := []string{"廉贞天相同宫"}
	var bonus []string
	var breaking []string
	if hasStarInPalace(lian, "禄存") || getStarSiHua(lian, "廉贞") == HuaLu {
		bonus = append(bonus, "见禄存或廉贞化禄")
	}
	if sanFangAllStars(c)["左辅"] {
		bonus = append(bonus, "左辅会照")
	}
	if hasShaInPalace(lian, []string{"擎羊"}) {
		breaking = append(breaking, "廉相宫坐擎羊（廉杀羊倾向）")
	}
	if getStarSiHua(lian, "廉贞") == HuaJi {
		breaking = append(breaking, "廉贞化忌")
	}

	level := "neutral"
	if len(breaking) > 0 {
		level = "caution"
	} else if inMing {
		level = "good"
	}
	*patterns = append(*patterns, Pattern{
		Name:        "廉贞天相格",
		Level:       level,
		Description: "廉贞天相同宫，印绶格局，主秉公处事、清廉之名，宜任公职、行政管理、法务、企划。怕见擎羊化忌，则反主官非。",
		Palaces:     []string{lian.Name},
		Conditions:  &PatternCondition{Required: required, Bonus: bonus, Breaking: breaking},
		Source:      "《紫微斗数全书》",
	})
}

// detectWuQiSha 武曲七杀：同宫，将星配财星
func detectWuQiSha(c *Chart, patterns *[]Pattern) {
	wu := findStarPalace(c, "武曲")
	qi := findStarPalace(c, "七杀")
	if wu == nil || qi == nil || wu.Branch != qi.Branch {
		return
	}

	inMing := wu.Branch == c.MingGongBranch
	required := []string{"武曲七杀同宫"}
	var bonus []string
	var breaking []string
	if getStarSiHua(wu, "武曲") == HuaQuan {
		bonus = append(bonus, "武曲化权")
	}
	if getStarSiHua(wu, "武曲") == HuaLu {
		bonus = append(bonus, "武曲化禄")
	}
	if getStarSiHua(wu, "武曲") == HuaJi {
		breaking = append(breaking, "武曲化忌（武曲化忌为财劫之兆）")
	}
	if hasShaInPalace(wu, []string{"擎羊", "陀罗", "火星", "铃星"}) {
		breaking = append(breaking, "武杀宫煞星过多")
	}

	level := "good"
	if len(breaking) > 0 {
		level = "caution"
	} else if inMing {
		level = "excellent"
	}
	*patterns = append(*patterns, Pattern{
		Name:        "武曲七杀",
		Level:       level,
		Description: "武曲七杀同宫，将星配财星，主果决刚毅、理财能力强，适合金融、军警、创业。但忌见化忌煞星，否则凶险。一生奋斗、积财但操心。",
		Palaces:     []string{wu.Name},
		Conditions:  &PatternCondition{Required: required, Bonus: bonus, Breaking: breaking},
		Source:      "《紫微斗数全书》",
	})
}

// detectTongLiang 天同天梁：同宫
func detectTongLiang(c *Chart, patterns *[]Pattern) {
	tong := findStarPalace(c, "天同")
	liang := findStarPalace(c, "天梁")
	if tong == nil || liang == nil || tong.Branch != liang.Branch {
		return
	}

	required := []string{"天同天梁同宫"}
	var bonus []string
	var breaking []string
	if sanFangAllStars(c)["文昌"] {
		bonus = append(bonus, "文昌会照")
	}
	if getStarSiHua(tong, "天同") == HuaLu {
		bonus = append(bonus, "天同化禄")
	}
	if hasShaInPalace(tong, shaHard) {
		breaking = append(breaking, "煞星同坐")
	}

	level := "good"
	if len(breaking) > 0 {
		level = "neutral"
	}
	*patterns = append(*patterns, Pattern{
		Name:        "天同天梁格",
		Level:       level,
		Description: "天同天梁同宫，福星与荫星共会，主宽厚和善、乐于助人，宜医疗、教育、宗教、社会公益。但偏温和保守，难成大富大贵之局。",
		Palaces:     []string{tong.Name},
		Conditions:  &PatternCondition{Required: required, Bonus: bonus, Breaking: breaking},
		Source:      "《紫微斗数全书》",
	})
}

// detectRiYueTongGong 日月同宫：太阳太阴丑或未宫同宫
func detectRiYueTongGong(c *Chart, patterns *[]Pattern) {
	sun := findStarPalace(c, "太阳")
	moon := findStarPalace(c, "太阴")
	if sun == nil || moon == nil || sun.Branch != moon.Branch {
		return
	}
	if sun.Branch != 1 && sun.Branch != 7 { // 必须丑(1) 或 未(7)
		return
	}

	inMing := sun.Branch == c.MingGongBranch
	required := []string{"太阳太阴同入" + patternBranchNames[sun.Branch] + "宫"}
	var bonus []string
	var breaking []string
	if sun.Branch == 7 {
		bonus = append(bonus, "未宫日月同辉（古书云未宫日月双美）")
	}
	if sanFangAllStars(c)["文昌"] && sanFangAllStars(c)["文曲"] {
		bonus = append(bonus, "昌曲会照")
	}
	if hasShaInPalace(sun, shaHard) {
		breaking = append(breaking, "日月宫煞星同坐")
	}

	level := "good"
	if len(breaking) == 0 && inMing {
		level = "excellent"
	}
	descSuffix := "丑宫日月同宫力量较平。"
	if sun.Branch == 7 {
		descSuffix = "未宫日月双美尤佳。"
	}
	*patterns = append(*patterns, Pattern{
		Name:        "日月同宫",
		Level:       level,
		Description: "太阳太阴于" + patternBranchNames[sun.Branch] + "宫同宫，阴阳平衡，文武兼备。主异性缘佳、事业顺遂、名声远播。" + descSuffix,
		Palaces:     []string{sun.Name},
		Conditions:  &PatternCondition{Required: required, Bonus: bonus, Breaking: breaking},
		Source:      "《紫微斗数全书》",
	})
}

// detectRiYueJiaMing 日月夹命：太阳太阴在命宫前后两宫
func detectRiYueJiaMing(c *Chart, patterns *[]Pattern) {
	prev, next := getJiaPalaces(c, c.MingGongBranch)
	if prev == nil || next == nil {
		return
	}
	prevHasSun := hasStarInPalace(prev, "太阳")
	prevHasMoon := hasStarInPalace(prev, "太阴")
	nextHasSun := hasStarInPalace(next, "太阳")
	nextHasMoon := hasStarInPalace(next, "太阴")
	ok := (prevHasSun && nextHasMoon) || (prevHasMoon && nextHasSun)
	if !ok {
		return
	}

	sunPalace := next
	if prevHasSun {
		sunPalace = prev
	}
	moonPalace := next
	if prevHasMoon {
		moonPalace = prev
	}
	required := []string{"太阳太阴分居命宫前后两宫"}
	var bonus []string
	var breaking []string
	if isBright(sunPalace, "太阳") {
		bonus = append(bonus, "太阳庙旺")
	}
	if isBright(moonPalace, "太阴") {
		bonus = append(bonus, "太阴庙旺")
	}
	if isDim(sunPalace, "太阳") || isDim(moonPalace, "太阴") {
		breaking = append(breaking, "日月落陷（夹命无光）")
	}

	level := "excellent"
	if len(breaking) > 0 {
		level = "good"
	}
	*patterns = append(*patterns, Pattern{
		Name:        "日月夹命",
		Level:       level,
		Description: "太阳太阴分居命宫两侧夹照，光明磊落，一生贵人相助，事业蓬勃。男主官贵，女主旺夫兴家。日月须不落陷方为真夹。",
		Palaces:     []string{sunPalace.Name, moonPalace.Name},
		Conditions:  &PatternCondition{Required: required, Bonus: bonus, Breaking: breaking},
		Source:      "《紫微斗数全书·日月夹命》",
	})
}

// detectJuRiTongGong 巨日同宫：巨门太阳同入寅或申
func detectJuRiTongGong(c *Chart, patterns *[]Pattern) {
	ju := findStarPalace(c, "巨门")
	sun := findStarPalace(c, "太阳")
	if ju == nil || sun == nil || ju.Branch != sun.Branch {
		return
	}
	if ju.Branch != 2 && ju.Branch != 8 { // 必须寅(2) 或 申(8)
		return
	}

	inMing := ju.Branch == c.MingGongBranch
	required := []string{"巨门太阳同入" + patternBranchNames[ju.Branch] + "宫"}
	var bonus []string
	var breaking []string
	if ju.Branch == 2 {
		bonus = append(bonus, "寅宫太阳庙旺，巨门得日光化解是非")
	}
	if getStarSiHua(ju, "巨门") == HuaLu || getStarSiHua(ju, "巨门") == HuaQuan {
		bonus = append(bonus, "巨门化禄/化权（口才生财）")
	}
	if getStarSiHua(ju, "巨门") == HuaJi {
		breaking = append(breaking, "巨门化忌（口舌官非）")
	}
	if ju.Branch == 8 {
		breaking = append(breaking, "申宫太阳偏西，巨门暗曜更显")
	}

	level := "good"
	if len(breaking) > 0 {
		level = "caution"
	} else if inMing && ju.Branch == 2 {
		level = "excellent"
	}
	*patterns = append(*patterns, Pattern{
		Name:        "巨日同宫",
		Level:       level,
		Description: "巨门太阳同" + patternBranchNames[ju.Branch] + "宫，太阳化解巨门暗曜，主以口才、传媒、外语、专业立业。寅宫为佳，申宫力减。怕巨门化忌则官非。",
		Palaces:     []string{ju.Name},
		Conditions:  &PatternCondition{Required: required, Bonus: bonus, Breaking: breaking},
		Source:      "《紫微斗数全书·巨日同宫》",
	})
}

// detectShiZhongYinYu 石中隐玉：巨门入命于子午宫
func detectShiZhongYinYu(c *Chart, ming *Palace, patterns *[]Pattern) {
	if !hasStarInPalace(ming, "巨门") {
		return
	}
	if ming.Branch != 0 && ming.Branch != 6 { // 子(0) 或 午(6)
		return
	}

	required := []string{"巨门入命于" + patternBranchNames[ming.Branch] + "宫"}
	var bonus []string
	var breaking []string
	if getStarSiHua(ming, "巨门") == HuaLu || getStarSiHua(ming, "巨门") == HuaQuan {
		bonus = append(bonus, "巨门化禄/化权")
	}
	if sanFangAllStars(c)["文昌"] {
		bonus = append(bonus, "文昌会照（石中隐玉得明）")
	}
	if getStarSiHua(ming, "巨门") == HuaJi {
		breaking = append(breaking, "巨门化忌（玉藏深泥）")
	}
	if hasShaInPalace(ming, shaHard) {
		breaking = append(breaking, "命坐煞星")
	}

	level := "excellent"
	if len(breaking) > 0 {
		level = "caution"
	}
	*patterns = append(*patterns, Pattern{
		Name:        "石中隐玉",
		Level:       level,
		Description: "巨门坐命子午，外表平凡而内蕴才学。早年默默无闻、中年方显贵气，宜走专业、研究、口才、传媒。需有禄权或文昌相助方能\"凿石见玉\"。",
		Palaces:     []string{"命宫"},
		Conditions:  &PatternCondition{Required: required, Bonus: bonus, Breaking: breaking},
		Source:      "《紫微斗数骨髓赋·石中隐玉》",
	})
}

// detectMingZhuChuHai 明珠出海：命宫在未空宫，对宫丑宫为太阳太阴
func detectMingZhuChuHai(c *Chart, ming *Palace, patterns *[]Pattern) {
	if ming.Branch != 7 { // 命在未
		return
	}
	if len(getMajorStarNames(ming)) > 0 { // 命宫为空宫
		return
	}
	dui := getDuiGong(c, ming.Branch)
	if dui == nil {
		return
	}
	if !hasStarInPalace(dui, "太阳") || !hasStarInPalace(dui, "太阴") {
		return
	}

	required := []string{"命宫在未为空宫", "对宫丑宫为太阳太阴同度"}
	var bonus []string
	var breaking []string
	if sanFangAllStars(c)["文昌"] || sanFangAllStars(c)["文曲"] {
		bonus = append(bonus, "再会昌曲")
	}
	if sanFangAllStars(c)["左辅"] || sanFangAllStars(c)["右弼"] {
		bonus = append(bonus, "辅弼相助")
	}
	if sanFangShaCount(c, shaHard) >= 2 {
		breaking = append(breaking, "煞星会照（珠光黯淡）")
	}

	level := "excellent"
	if len(breaking) > 0 {
		level = "good"
	}
	*patterns = append(*patterns, Pattern{
		Name:        "明珠出海",
		Level:       level,
		Description: "命未空宫，对宫丑宫日月同辉拱照，号\"明珠出海\"。主出生平凡、后天努力出头，宜远赴他乡、学术研究或大公司高位，主大富大贵。",
		Palaces:     []string{"命宫", dui.Name},
		Conditions:  &PatternCondition{Required: required, Bonus: bonus, Breaking: breaking},
		Source:      "《紫微斗数全集·明珠出海》",
	})
}

// detectZiWeiInMing 紫微独坐入命
func detectZiWeiInMing(c *Chart, ming *Palace, patterns *[]Pattern) {
	if !hasStarInPalace(ming, "紫微") || hasStarInPalace(ming, "天府") {
		return
	}

	required := []string{"紫微独坐命宫（无天府同坐）"}
	var bonus []string
	var breaking []string
	sanFangSet := sanFangAllStars(c)
	if sanFangSet["左辅"] && sanFangSet["右弼"] {
		bonus = append(bonus, "左辅右弼同会")
	}
	if sanFangSet["文昌"] && sanFangSet["文曲"] {
		bonus = append(bonus, "文昌文曲同会")
	}
	if !sanFangSet["左辅"] && !sanFangSet["右弼"] {
		breaking = append(breaking, "无辅弼（孤君无臣）")
	}
	if hasShaInPalace(ming, shaKong) {
		breaking = append(breaking, "紫微遇空劫（古书最忌）")
	}

	level := "good"
	if len(breaking) > 0 {
		level = "caution"
	} else if len(bonus) > 0 {
		level = "excellent"
	}
	*patterns = append(*patterns, Pattern{
		Name:        "紫微入命",
		Level:       level,
		Description: "紫微独坐命宫，帝王之星，自尊心强、有领导魅力。但紫微最忌\"在野孤君\"——若无左右辅弼相会，反成孤高自傲、易招毁谤。",
		Palaces:     []string{"命宫"},
		Conditions:  &PatternCondition{Required: required, Bonus: bonus, Breaking: breaking},
		Source:      "《紫微斗数全书》",
	})
}

// detectFuBiJiaMing 辅弼夹命
func detectFuBiJiaMing(c *Chart, patterns *[]Pattern) {
	prev, next := getJiaPalaces(c, c.MingGongBranch)
	if prev == nil || next == nil {
		return
	}
	prevHasZuo := hasStarInPalace(prev, "左辅")
	prevHasYou := hasStarInPalace(prev, "右弼")
	nextHasZuo := hasStarInPalace(next, "左辅")
	nextHasYou := hasStarInPalace(next, "右弼")
	if !((prevHasZuo && nextHasYou) || (prevHasYou && nextHasZuo)) {
		return
	}

	required := []string{"左辅右弼分居命宫前后两宫"}
	var bonus []string
	var breaking []string
	if sanFangAllStars(c)["天魁"] || sanFangAllStars(c)["天钺"] {
		bonus = append(bonus, "再会魁钺")
	}

	*patterns = append(*patterns, Pattern{
		Name:        "辅弼夹命",
		Level:       "excellent",
		Description: "左辅右弼夹命，一生贵人不断、逢凶化吉。适合走仕途、大企业管理，有贵人提携之命。古书云\"左辅右弼，终身福厚\"。",
		Palaces:     []string{"命宫", prev.Name, next.Name},
		Conditions:  &PatternCondition{Required: required, Bonus: bonus, Breaking: breaking},
		Source:      "《紫微斗数全书·辅弼夹命》",
	})
}

// detectChangQuJiaMing 昌曲夹命
func detectChangQuJiaMing(c *Chart, patterns *[]Pattern) {
	prev, next := getJiaPalaces(c, c.MingGongBranch)
	if prev == nil || next == nil {
		return
	}
	prevHasChang := hasStarInPalace(prev, "文昌")
	prevHasQu := hasStarInPalace(prev, "文曲")
	nextHasChang := hasStarInPalace(next, "文昌")
	nextHasQu := hasStarInPalace(next, "文曲")
	if !((prevHasChang && nextHasQu) || (prevHasQu && nextHasChang)) {
		return
	}

	*patterns = append(*patterns, Pattern{
		Name:        "昌曲夹命",
		Level:       "excellent",
		Description: "文昌文曲夹命宫，主聪明俊秀、文采斐然，宜走文教、学术、艺术、写作。古书云\"昌曲夹命主科甲\"，最利考运。",
		Palaces:     []string{"命宫", prev.Name, next.Name},
		Conditions:  &PatternCondition{Required: []string{"文昌文曲分居命宫前后两宫"}},
		Source:      "《紫微斗数全书》",
	})
}

// detectKuiYueJiaMing 魁钺夹命
func detectKuiYueJiaMing(c *Chart, patterns *[]Pattern) {
	prev, next := getJiaPalaces(c, c.MingGongBranch)
	if prev == nil || next == nil {
		return
	}
	okA := hasStarInPalace(prev, "天魁") && hasStarInPalace(next, "天钺")
	okB := hasStarInPalace(prev, "天钺") && hasStarInPalace(next, "天魁")
	if !okA && !okB {
		return
	}

	*patterns = append(*patterns, Pattern{
		Name:        "魁钺夹命",
		Level:       "good",
		Description: "天魁天钺夹命，男称天乙、女称玉堂，一生贵人提携。考试、求职、关键时刻常有意外贵人相助。",
		Palaces:     []string{"命宫", prev.Name, next.Name},
		Conditions:  &PatternCondition{Required: []string{"天魁天钺分居命宫前后两宫"}},
		Source:      "《紫微斗数全书》",
	})
}

// detectShuangLuChaoYuan 双禄朝垣：化禄 + 禄存 同会三方
func detectShuangLuChaoYuan(c *Chart, ming *Palace, patterns *[]Pattern) {
	sanFang := getSanFangPalaces(c)
	huaLuFound := false
	luCunFound := false
	for _, p := range sanFang {
		for i := range p.Stars {
			if p.Stars[i].SiHua == HuaLu {
				huaLuFound = true
			}
		}
		if hasStarInPalace(p, "禄存") {
			luCunFound = true
		}
	}
	if !huaLuFound || !luCunFound {
		return
	}

	var palaceNames []string
	for _, p := range sanFang {
		palaceNames = append(palaceNames, p.Name)
	}
	var breaking []string
	if hasShaInPalace(ming, shaKong) {
		breaking = []string{"命坐空劫（双禄遇空，财来财去）"}
	}
	*patterns = append(*patterns, Pattern{
		Name:        "双禄朝垣",
		Level:       "excellent",
		Description: "化禄、禄存同会命宫三方四正，财源涌动、衣食丰足。古书云\"双禄朝垣，富比陶朱\"，主一生不愁财，多有正财横财兼得。",
		Palaces:     palaceNames,
		Conditions: &PatternCondition{
			Required: []string{"化禄会照三方四正", "禄存会照三方四正"},
			Breaking: breaking,
		},
		Source: "《紫微斗数全书·双禄朝垣》",
	})
}

// detectSanQiJiaHui 三奇加会：化禄 化权 化科 同会三方
func detectSanQiJiaHui(c *Chart, patterns *[]Pattern) {
	sanFangPalaces := getSanFangPalaces(c)
	lu, quan, ke := false, false, false
	for _, p := range sanFangPalaces {
		for i := range p.Stars {
			if p.Stars[i].SiHua == HuaLu {
				lu = true
			}
			if p.Stars[i].SiHua == HuaQuan {
				quan = true
			}
			if p.Stars[i].SiHua == HuaKe {
				ke = true
			}
		}
	}
	if !(lu && quan && ke) {
		return
	}

	var palaceNames []string
	for _, p := range sanFangPalaces {
		palaceNames = append(palaceNames, p.Name)
	}
	*patterns = append(*patterns, Pattern{
		Name:        "三奇加会",
		Level:       "excellent",
		Description: "化禄、化权、化科三吉化齐会命宫三方四正，号称\"三奇加会\"。主一生功名、财富、贵人三全，是紫微斗数最高吉格之一。",
		Palaces:     palaceNames,
		Conditions:  &PatternCondition{Required: []string{"化禄、化权、化科三吉化齐会命宫三方四正"}},
		Source:      "《紫微斗数全书·三奇加会》",
	})
}

// detectHuaLuRuMing 化禄入命/官/财
func detectHuaLuRuMing(c *Chart, ming *Palace, patterns *[]Pattern) {
	var huaLuStar *Star
	for i := range ming.Stars {
		if ming.Stars[i].SiHua == HuaLu && ming.Stars[i].Type == StarMajor {
			huaLuStar = &ming.Stars[i]
			break
		}
	}
	if huaLuStar == nil {
		return
	}

	extra := ""
	switch huaLuStar.Name {
	case "武曲":
		extra = "武曲化禄属正财，宜实业、金融。"
	case "太阴":
		extra = "太阴化禄属阴财、不动产。"
	case "贪狼":
		extra = "贪狼化禄属人脉财、桃花财。"
	}
	*patterns = append(*patterns, Pattern{
		Name:        huaLuStar.Name + "化禄入命",
		Level:       "good",
		Description: huaLuStar.Name + "化禄坐命，主生财顺利、人缘佳、机缘多。" + extra,
		Palaces:     []string{"命宫"},
		Conditions:  &PatternCondition{Required: []string{huaLuStar.Name + "化禄坐命宫"}},
		Source:      "《紫微斗数全书》",
	})
	_ = c
}

// ────────────────── 恶格识别器 ──────────────────

// detectHuaJiRuMingQian 化忌入命/迁
func detectHuaJiRuMingQian(c *Chart, patterns *[]Pattern) {
	qianBranch := (c.MingGongBranch + 6) % 12
	for pi := range c.Palaces {
		palace := &c.Palaces[pi]
		if palace.Branch != c.MingGongBranch && palace.Branch != qianBranch {
			continue
		}
		var jiStar *Star
		for i := range palace.Stars {
			if palace.Stars[i].SiHua == HuaJi && palace.Stars[i].Type == StarMajor {
				jiStar = &palace.Stars[i]
				break
			}
		}
		if jiStar == nil {
			continue
		}

		inMing := palace.Branch == c.MingGongBranch
		mingOrQian := "迁"
		if inMing {
			mingOrQian = "命"
		}
		description := jiStar.Name + "化忌坐迁移宫，外出、远行、人际关系易有波折，宜守不宜动。"
		if inMing {
			description = jiStar.Name + "化忌坐命宫，需留意自身固执、心理障碍或健康隐患，凡事退一步思考。化忌不一定坏，代表此星能量需要特别关注。"
		}
		*patterns = append(*patterns, Pattern{
			Name:        jiStar.Name + "化忌入" + mingOrQian,
			Level:       "caution",
			Description: description,
			Palaces:     []string{palace.Name},
			Conditions:  &PatternCondition{Required: []string{jiStar.Name + "化忌坐" + mingOrQian + "宫"}},
			Source:      "《紫微斗数全书》",
		})
	}
}

// detectYangTuoJiaJi 羊陀夹忌：化忌坐宫，左右被擎羊陀罗夹
func detectYangTuoJiaJi(c *Chart, patterns *[]Pattern) {
	for pi := range c.Palaces {
		palace := &c.Palaces[pi]
		var jiStar *Star
		for i := range palace.Stars {
			if palace.Stars[i].SiHua == HuaJi {
				jiStar = &palace.Stars[i]
				break
			}
		}
		if jiStar == nil {
			continue
		}
		if palace.Branch != c.MingGongBranch { // 只看命宫被夹
			continue
		}

		prev, next := getJiaPalaces(c, palace.Branch)
		if prev == nil || next == nil {
			continue
		}
		aPrev := hasStarInPalace(prev, "擎羊") && hasStarInPalace(next, "陀罗")
		aNext := hasStarInPalace(prev, "陀罗") && hasStarInPalace(next, "擎羊")
		if !aPrev && !aNext {
			continue
		}

		*patterns = append(*patterns, Pattern{
			Name:        "羊陀夹忌",
			Level:       "caution",
			Description: "化忌坐命，左右擎羊陀罗夹命，古书云\"羊陀夹忌为败局\"，主一生劳碌奔波、坎坷不顺、身心俱疲。需以德行修养与积极做事化解，凡事谨慎为上。",
			Palaces:     []string{"命宫", prev.Name, next.Name},
			Conditions:  &PatternCondition{Required: []string{"化忌坐命", "擎羊陀罗分居命宫前后两宫"}},
			Source:      "《紫微斗数骨髓赋·羊陀夹忌》",
		})
		return
	}
}

// detectHuoLingJiaMing 火铃夹命：火星铃星分居命宫前后
func detectHuoLingJiaMing(c *Chart, patterns *[]Pattern) {
	prev, next := getJiaPalaces(c, c.MingGongBranch)
	if prev == nil || next == nil {
		return
	}
	okA := hasStarInPalace(prev, "火星") && hasStarInPalace(next, "铃星")
	okB := hasStarInPalace(prev, "铃星") && hasStarInPalace(next, "火星")
	if !okA && !okB {
		return
	}

	*patterns = append(*patterns, Pattern{
		Name:        "火铃夹命",
		Level:       "caution",
		Description: "火星铃星分居命宫前后两宫夹命，主性急、易冲动、突发意外或纠纷。需培养耐性、避免冲动决策。",
		Palaces:     []string{"命宫", prev.Name, next.Name},
		Conditions:  &PatternCondition{Required: []string{"火星铃星分居命宫前后两宫"}},
		Source:      "《紫微斗数全书》",
	})
}

// detectKongJieJiaMing 空劫夹命：地空地劫分居命宫前后
func detectKongJieJiaMing(c *Chart, patterns *[]Pattern) {
	prev, next := getJiaPalaces(c, c.MingGongBranch)
	if prev == nil || next == nil {
		return
	}
	okA := hasStarInPalace(prev, "地空") && hasStarInPalace(next, "地劫")
	okB := hasStarInPalace(prev, "地劫") && hasStarInPalace(next, "地空")
	if !okA && !okB {
		return
	}

	*patterns = append(*patterns, Pattern{
		Name:        "空劫夹命",
		Level:       "caution",
		Description: "地空地劫夹命，主财来财去、思想脱俗、易遁入宗教哲学。古书云\"空劫夹命，财不聚\"。宜技艺、宗教、研究等不重物质之业。",
		Palaces:     []string{"命宫", prev.Name, next.Name},
		Conditions:  &PatternCondition{Required: []string{"地空地劫分居命宫前后两宫"}},
		Source:      "《紫微斗数全书》",
	})
}

// detectLianShaYang 廉杀羊：廉贞、七杀、擎羊三星会照（流年大限最凶）
func detectLianShaYang(c *Chart, patterns *[]Pattern) {
	sanFangSet := sanFangAllStars(c)
	if !(sanFangSet["廉贞"] && sanFangSet["七杀"] && sanFangSet["擎羊"]) {
		return
	}

	*patterns = append(*patterns, Pattern{
		Name:        "廉杀羊",
		Level:       "caution",
		Description: "廉贞、七杀、擎羊三星会照命宫三方，古书警示之凶格。主血光、官非、意外。本命有此格不必惊慌，但流年大限再触发时需特别谨慎驾驶、避免冲突、注意手术风险。",
		Palaces:     []string{"命宫"},
		Conditions:  &PatternCondition{Required: []string{"廉贞、七杀、擎羊三星会照三方四正"}},
		Source:      "《紫微斗数全书·廉杀羊》",
	})
}

// detectJuHuoYang 巨火羊：巨门、火星、擎羊会照
func detectJuHuoYang(c *Chart, patterns *[]Pattern) {
	sanFangSet := sanFangAllStars(c)
	if !(sanFangSet["巨门"] && sanFangSet["火星"] && sanFangSet["擎羊"]) {
		return
	}

	*patterns = append(*patterns, Pattern{
		Name:        "巨火羊",
		Level:       "caution",
		Description: "巨门、火星、擎羊三星会照，古书云\"巨火羊，终身缢死\"——古时凶格。现代理解为：易因口舌、激烈冲突而招大祸。需修身养性、慎言慎行，避免极端情绪。",
		Palaces:     []string{"命宫"},
		Conditions:  &PatternCondition{Required: []string{"巨门、火星、擎羊三星会照三方四正"}},
		Source:      "《紫微斗数骨髓赋·巨火羊》",
	})
}

// detectLingChangTuoWu 铃昌陀武：铃星、文昌、陀罗、武曲会照（限至投河）
func detectLingChangTuoWu(c *Chart, patterns *[]Pattern) {
	sanFangSet := sanFangAllStars(c)
	if !(sanFangSet["铃星"] && sanFangSet["文昌"] && sanFangSet["陀罗"] && sanFangSet["武曲"]) {
		return
	}

	*patterns = append(*patterns, Pattern{
		Name:        "铃昌陀武",
		Level:       "caution",
		Description: "铃星、文昌、陀罗、武曲四星齐会，古书云\"铃昌陀武，限至投河\"——古时大凶格。本命有此组合本身不必恐慌，但流年大限触发时需高度警觉重大决策、情绪起伏、水边活动。",
		Palaces:     []string{"命宫"},
		Conditions:  &PatternCondition{Required: []string{"铃星、文昌、陀罗、武曲四星会照三方四正"}},
		Source:      "《紫微斗数骨髓赋·铃昌陀武》",
	})
}

// detectMaTouDaiJian 马头带箭：擎羊在午宫坐命
func detectMaTouDaiJian(c *Chart, ming *Palace, patterns *[]Pattern) {
	if ming.Branch != 6 { // 必须午
		return
	}
	if !hasStarInPalace(ming, "擎羊") {
		return
	}

	required := []string{"擎羊于午宫坐命"}
	var bonus []string
	if sanFangAllStars(c)["七杀"] || sanFangAllStars(c)["破军"] {
		bonus = append(bonus, "再会七杀或破军（武职大贵）")
	}
	if sanFangAllStars(c)["天魁"] || sanFangAllStars(c)["天钺"] {
		bonus = append(bonus, "魁钺加照")
	}

	level := "caution"
	if len(bonus) > 0 {
		level = "good"
	}
	*patterns = append(*patterns, Pattern{
		Name:        "马头带箭",
		Level:       level,
		Description: "擎羊于午宫坐命，号\"马头带箭\"。古书云\"威镇边疆\"——主刚毅果决、有冲杀之力，宜军警武职、运动员、外科医师。但同时主危险与意外，需配合杀破狼或贵人方为大格，否则反主血光。",
		Palaces:     []string{"命宫"},
		Conditions:  &PatternCondition{Required: required, Bonus: bonus},
		Source:      "《紫微斗数骨髓赋·马头带箭》",
	})
}

// ────────────────── 基础格局（提升识别覆盖率）──────────────────
// 设计：让普通命盘也能识别出 1-3 个常见格局，而不是 30+ 严格古书格局都不匹配。
// 这些都是单一条件触发的轻量识别，level 多为 neutral / good。

// detectLuCunShouShen 禄存守身：禄存入身宫（或命宫与身宫同宫）
func detectLuCunShouShen(c *Chart, patterns *[]Pattern) {
	luCunPalace := findStarPalace(c, "禄存")
	if luCunPalace == nil {
		return
	}
	inMing := luCunPalace.Branch == c.MingGongBranch
	inShen := luCunPalace.Branch == c.ShenGongBranch
	if !inMing && !inShen {
		return
	}
	name := "禄存守身"
	description := "禄存入身宫，主中年后财源稳定、得禄自享。倪师说「禄存入身，财气近身」——配偶或事业方向能带来稳定财禄。"
	palace := "身宫"
	requiredItem := "禄存入身宫"
	if inMing {
		name = "禄存守命"
		description = "禄存坐命，主一生衣食无忧、财禄稳定。性格保守，善积累，但羊陀夹禄须防小人。最宜配化禄、左辅右弼方为大格。"
		palace = "命宫"
		requiredItem = "禄存入命宫"
	}
	*patterns = append(*patterns, Pattern{
		Name:        name,
		Level:       "good",
		Description: description,
		Palaces:     []string{palace},
		Conditions:  &PatternCondition{Required: []string{requiredItem}},
		Source:      "《紫微斗数全书·禄存星》",
	})
}

// detectTianMaRuMing 天马入命/迁：驿马星动
func detectTianMaRuMing(c *Chart, patterns *[]Pattern) {
	tianMaPalace := findStarPalace(c, "天马")
	if tianMaPalace == nil {
		return
	}
	inMing := tianMaPalace.Branch == c.MingGongBranch
	inQian := tianMaPalace.Branch == (c.MingGongBranch+6)%12
	if !inMing && !inQian {
		return
	}
	name := "天马在迁"
	description := "天马在迁移宫，主外出有利、远行得财，宜异乡发展。配化禄主异地生财，配煞星则旅途多波折。"
	requiredItem := "天马入迁移宫"
	if inMing {
		name = "天马入命"
		description = "天马坐命，主一生奔波、动中得财，宜走商旅、外勤、跨界发展。倪师说「天马入命，无禄不发」——若再会禄存或化禄即「禄马交驰」之富格。"
		requiredItem = "天马入命宫"
	}
	*patterns = append(*patterns, Pattern{
		Name:        name,
		Level:       "neutral",
		Description: description,
		Palaces:     []string{tianMaPalace.Name},
		Conditions:  &PatternCondition{Required: []string{requiredItem}},
		Source:      "《紫微斗数全书·天马星》",
	})
}

// detectHuaLuRuCai 化禄入财：财帛宫主星化禄
func detectHuaLuRuCai(c *Chart, patterns *[]Pattern) {
	var cai *Palace
	for i := range c.Palaces {
		if c.Palaces[i].Name == "财帛" {
			cai = &c.Palaces[i]
			break
		}
	}
	if cai == nil {
		return
	}
	var luStar *Star
	for i := range cai.Stars {
		if cai.Stars[i].Type == StarMajor && cai.Stars[i].SiHua == HuaLu {
			luStar = &cai.Stars[i]
			break
		}
	}
	if luStar == nil {
		return
	}
	*patterns = append(*patterns, Pattern{
		Name:        "化禄入财",
		Level:       "good",
		Description: luStar.Name + "化禄入财帛宫，主财源畅通、收入稳定。倪师讲化禄是「正财」象征——这个化禄星所代表的能力（" + luStar.Name + "的核心特质）是你赚钱的主轴。配禄存或天马则财源更广。",
		Palaces:     []string{"财帛"},
		Conditions:  &PatternCondition{Required: []string{luStar.Name + "化禄入财帛宫"}},
		Source:      "《紫微斗数全书·四化论》",
	})
}

// detectHuaQuanRuGuan 化权入官：官禄宫主星化权
func detectHuaQuanRuGuan(c *Chart, patterns *[]Pattern) {
	var guan *Palace
	for i := range c.Palaces {
		if c.Palaces[i].Name == "官禄" {
			guan = &c.Palaces[i]
			break
		}
	}
	if guan == nil {
		return
	}
	var quanStar *Star
	for i := range guan.Stars {
		if guan.Stars[i].Type == StarMajor && guan.Stars[i].SiHua == HuaQuan {
			quanStar = &guan.Stars[i]
			break
		}
	}
	if quanStar == nil {
		return
	}
	*patterns = append(*patterns, Pattern{
		Name:        "化权入官",
		Level:       "good",
		Description: quanStar.Name + "化权入官禄宫，主事业有掌控力、能担当独当一面的职位。化权代表权力与执行力——" + quanStar.Name + "化权说明你在事业上能成为决策者或核心执行者，宜走管理或技术权威路线。",
		Palaces:     []string{"官禄"},
		Conditions:  &PatternCondition{Required: []string{quanStar.Name + "化权入官禄宫"}},
		Source:      "《紫微斗数全书·四化论》",
	})
}

// detectHuaKeRuMingShen 化科入命/身：科名加身
func detectHuaKeRuMingShen(c *Chart, patterns *[]Pattern) {
	ming := getPalaceByBranch(c, c.MingGongBranch)
	shen := getPalaceByBranch(c, c.ShenGongBranch)
	var target []*Palace
	for _, p := range []*Palace{ming, shen} {
		if p != nil {
			target = append(target, p)
		}
	}
	for _, p := range target {
		var keStar *Star
		for i := range p.Stars {
			if p.Stars[i].Type == StarMajor && p.Stars[i].SiHua == HuaKe {
				keStar = &p.Stars[i]
				break
			}
		}
		if keStar == nil {
			continue
		}
		isMing := p.Branch == c.MingGongBranch
		mingOrShen := "身"
		if isMing {
			mingOrShen = "命"
		}
		*patterns = append(*patterns, Pattern{
			Name:        "化科入" + mingOrShen,
			Level:       "good",
			Description: keStar.Name + "化科入" + mingOrShen + "宫，主名声、文书、学术运。倪师讲化科是「贵人星」——" + keStar.Name + "化科带来的是被人看重的特质，宜从事文书、教育、研究、咨询、文创等“以名取利”的方向。",
			Palaces:     []string{mingOrShen + "宫"},
			Conditions:  &PatternCondition{Required: []string{keStar.Name + "化科入" + mingOrShen + "宫"}},
			Source:      "《紫微斗数全书·四化论》",
		})
		return // 命和身重复时只识别一次
	}
}

// detectJiYueTongLiangPartial 机月同梁三星会（降级版）：天机/太阴/天同/天梁 任 3 星齐入三方四正
func detectJiYueTongLiangPartial(c *Chart, ming *Palace, patterns *[]Pattern) {
	sanFangSet := sanFangAllStars(c)
	var has []string
	for _, s := range []string{"天机", "太阴", "天同", "天梁"} {
		if sanFangSet[s] {
			has = append(has, s)
		}
	}
	if len(has) != 3 { // 4 星齐由 detectJiYueTongLiang 处理
		return
	}
	// 避免和上面 detectJiYueTongLiang 重复（4 星齐的不进这里）
	var missing []string
	for _, s := range []string{"天机", "太阴", "天同", "天梁"} {
		if !sanFangSet[s] {
			missing = append(missing, s)
		}
	}
	var palaceNames []string
	for _, p := range getSanFangPalaces(c) {
		majors := getMajorStarNames(p)
		matched := false
		for _, s := range has {
			if containsString(majors, s) {
				matched = true
				break
			}
		}
		if matched {
			palaceNames = append(palaceNames, p.Name)
		}
	}
	hasJoined := strings.Join(has, "、")
	missingJoined := strings.Join(missing, "、")
	*patterns = append(*patterns, Pattern{
		Name:        "机月同梁三星会",
		Level:       "neutral",
		Description: "三方四正会齐" + hasJoined + "，差" + missingJoined + "未会。机月同梁不全格，文质带谋，但稳定度不如四星齐。仍宜公职、教研、医疗、服务等需要积累与稳定的行业，关键看缺位星与四化的配合。",
		Palaces:     palaceNames,
		Conditions:  &PatternCondition{Required: []string{"三方四正会" + hasJoined + "（机月同梁缺" + missingJoined + "）"}},
		Source:      "《紫微斗数全书·机月同梁格》（降级版）",
	})
	_ = ming
}

// detectChangQuTongHui 昌曲同会：文昌+文曲都在命三方四正
func detectChangQuTongHui(c *Chart, patterns *[]Pattern) {
	sanFangSet := sanFangAllStars(c)
	if !sanFangSet["文昌"] || !sanFangSet["文曲"] {
		return
	}
	ming := getPalaceByBranch(c, c.MingGongBranch)
	if ming == nil {
		return
	}
	inMing := hasStarInPalace(ming, "文昌") && hasStarInPalace(ming, "文曲")
	name := "昌曲同会"
	description := "文昌文曲同会三方四正，主才华横溢、口才文笔俱佳。宜走需要表达与文采的行业，化科加持则名声大显。"
	if inMing {
		name = "昌曲坐命"
		description = "文昌文曲同入命宫，主聪明俊秀、文采斐然，宜文学、教育、写作、咨询。最忌化忌——昌曲化忌主文书契约暗亏。"
	}
	*patterns = append(*patterns, Pattern{
		Name:        name,
		Level:       "good",
		Description: description,
		Palaces:     []string{"命宫"},
		Conditions:  &PatternCondition{Required: []string{"文昌、文曲同会命宫三方四正"}},
		Source:      "《紫微斗数全书·文星论》",
	})
}

// detectFuBiTongHui 辅弼同会：左辅+右弼都在命三方四正
func detectFuBiTongHui(c *Chart, patterns *[]Pattern) {
	sanFangSet := sanFangAllStars(c)
	if !sanFangSet["左辅"] || !sanFangSet["右弼"] {
		return
	}
	*patterns = append(*patterns, Pattern{
		Name:        "辅弼同会",
		Level:       "good",
		Description: "左辅右弼同会命宫三方四正，主一生贵人不绝、人缘极佳。最宜领导岗位与团队合作型工作。倪师说「辅弼夹命，平生贵人多」——你不是单打独斗的命，要善用人际网络。",
		Palaces:     []string{"命宫"},
		Conditions:  &PatternCondition{Required: []string{"左辅、右弼同会命宫三方四正"}},
		Source:      "《紫微斗数全书·辅弼论》",
	})
}

// detectKuiYueTongHui 魁钺同会：天魁+天钺都在命三方四正
func detectKuiYueTongHui(c *Chart, patterns *[]Pattern) {
	sanFangSet := sanFangAllStars(c)
	if !sanFangSet["天魁"] || !sanFangSet["天钺"] {
		return
	}
	*patterns = append(*patterns, Pattern{
		Name:        "魁钺同会",
		Level:       "good",
		Description: "天魁天钺同会命宫三方四正，主\"天乙贵人\"加持，关键时刻总有贵人提携。倪师说「魁钺夹命，必为贵人」——遇到困难时身边会出现得力相助者，宜主动维护人脉。",
		Palaces:     []string{"命宫"},
		Conditions:  &PatternCondition{Required: []string{"天魁、天钺同会命宫三方四正"}},
		Source:      "《紫微斗数全书·魁钺论》",
	})
}

// detectKeQuanShuangHui 科权双会：化科 + 化权 同会三方四正
func detectKeQuanShuangHui(c *Chart, patterns *[]Pattern) {
	sfPalaces := getSanFangPalaces(c)
	hasKe, hasQuan := false, false
	for _, p := range sfPalaces {
		for i := range p.Stars {
			if p.Stars[i].Type == StarMajor && p.Stars[i].SiHua == HuaKe {
				hasKe = true
			}
			if p.Stars[i].Type == StarMajor && p.Stars[i].SiHua == HuaQuan {
				hasQuan = true
			}
		}
	}
	if !hasKe || !hasQuan {
		return
	}
	*patterns = append(*patterns, Pattern{
		Name:        "科权双会",
		Level:       "good",
		Description: "化科 + 化权 同会三方四正，主名权双美——既有学识/名声（科），又有掌控力（权），宜走\"专业权威\"路线（如医生、律师、教授、技术骨干），名利双收且根基扎实。",
		Palaces:     []string{"命宫"},
		Conditions:  &PatternCondition{Required: []string{"化科、化权同会命宫三方四正"}},
		Source:      "《紫微斗数全书·四化会照》",
	})
}

// ────────────────── 主入口 ──────────────────

// DetectPatterns 识别命盘全部格局（对应 TS detectPatterns）。
func DetectPatterns(c *Chart) []Pattern {
	patterns := []Pattern{}
	ming := getPalaceByBranch(c, c.MingGongBranch)
	if ming == nil {
		return patterns
	}

	// 上格
	detectJunChenQingHui(c, ming, &patterns)
	detectZiFu(c, ming, &patterns)
	detectFuXiangChaoYuan(c, ming, &patterns)
	detectYangLiangChangLu(c, ming, &patterns)
	detectHuoTanLingTan(c, ming, &patterns)
	detectWuTan(c, ming, &patterns)
	detectShaPoLang(c, ming, &patterns)
	detectJiYueTongLiang(c, ming, &patterns)

	// 中格
	detectLianXiang(c, &patterns)
	detectWuQiSha(c, &patterns)
	detectTongLiang(c, &patterns)
	detectRiYueTongGong(c, &patterns)
	detectRiYueJiaMing(c, &patterns)
	detectJuRiTongGong(c, &patterns)
	detectShiZhongYinYu(c, ming, &patterns)
	detectMingZhuChuHai(c, ming, &patterns)
	detectZiWeiInMing(c, ming, &patterns)

	// 助力格
	detectFuBiJiaMing(c, &patterns)
	detectChangQuJiaMing(c, &patterns)
	detectKuiYueJiaMing(c, &patterns)
	detectShuangLuChaoYuan(c, ming, &patterns)
	detectSanQiJiaHui(c, &patterns)
	detectHuaLuRuMing(c, ming, &patterns)

	// 恶格
	detectHuaJiRuMingQian(c, &patterns)
	detectYangTuoJiaJi(c, &patterns)
	detectHuoLingJiaMing(c, &patterns)
	detectKongJieJiaMing(c, &patterns)
	detectLianShaYang(c, &patterns)
	detectJuHuoYang(c, &patterns)
	detectLingChangTuoWu(c, &patterns)
	detectMaTouDaiJian(c, ming, &patterns)

	// 基础格局（提升识别覆盖率，让普通命盘也能识别 1-3 个）
	detectLuCunShouShen(c, &patterns)
	detectTianMaRuMing(c, &patterns)
	detectHuaLuRuCai(c, &patterns)
	detectHuaQuanRuGuan(c, &patterns)
	detectHuaKeRuMingShen(c, &patterns)
	detectJiYueTongLiangPartial(c, ming, &patterns)
	detectChangQuTongHui(c, &patterns)
	detectFuBiTongHui(c, &patterns)
	detectKuiYueTongHui(c, &patterns)
	detectKeQuanShuangHui(c, &patterns)

	return patterns
}

// ────────────────── 命宫摘要（保持向后兼容）──────────────────

// MingGongSummary 命宫摘要（对应 TS getMingGongSummary 返回值）。
type MingGongSummary struct {
	Stars    []string `json:"stars"`
	Keywords []string `json:"keywords"`
	Nature   string   `json:"nature"`
}

var mingGongKeywordMap = map[string][]string{
	"紫微": {"尊贵", "独立", "领导"},
	"天机": {"智慧", "机变", "善谋"},
	"太阳": {"阳刚", "官贵", "慷慨"},
	"武曲": {"财富", "刚毅", "果断"},
	"天同": {"温和", "享福", "随缘"},
	"廉贞": {"才艺", "桃花", "多变"},
	"天府": {"财库", "稳重", "保守"},
	"太阴": {"柔美", "财富", "细腻"},
	"贪狼": {"欲望", "桃花", "多才"},
	"巨门": {"善辩", "多思", "口才"},
	"天相": {"辅佐", "行政", "稳健"},
	"天梁": {"荫护", "医药", "长辈"},
	"七杀": {"将星", "果决", "孤克"},
	"破军": {"开创", "变动", "破旧"},
}

var mingGongNatureMap = map[string]string{
	"紫微": "帝王星", "天机": "智慧星", "太阳": "贵人星",
	"武曲": "财帛星", "天同": "福德星", "廉贞": "桃花星",
	"天府": "财库星", "太阴": "财富星", "贪狼": "桃花星",
	"巨门": "是非星", "天相": "印绶星", "天梁": "荫庇星",
	"七杀": "将帅星", "破军": "变动星",
}

// GetMingGongSummary 命宫摘要（对应 TS getMingGongSummary）。
func GetMingGongSummary(c *Chart) MingGongSummary {
	mingPalace := getPalaceByBranch(c, c.MingGongBranch)
	if mingPalace == nil {
		return MingGongSummary{Stars: []string{}, Keywords: []string{}, Nature: ""}
	}

	starNames := []string{}
	for i := range mingPalace.Stars {
		if mingPalace.Stars[i].Type == StarMajor {
			starNames = append(starNames, mingPalace.Stars[i].Name)
		}
	}

	keywords := []string{}
	for _, n := range starNames {
		keywords = append(keywords, mingGongKeywordMap[n]...)
	}
	if len(keywords) > 5 {
		keywords = keywords[:5]
	}

	nature := "空宫"
	if len(starNames) > 0 {
		nature = mingGongNatureMap[starNames[0]]
	}

	return MingGongSummary{Stars: starNames, Keywords: keywords, Nature: nature}
}
