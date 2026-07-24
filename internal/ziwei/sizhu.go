// 四柱视角(八字附加层):在紫微盘四柱干支之上推导五行、十神、
// 藏干与纳音——八字不作独立产品,作为紫微盘的同源视角呈现。
package ziwei

// HiddenStem 地支藏干(本气在前)。
type HiddenStem struct {
	Stem    string `json:"stem"`
	Element string `json:"element"`
	ShiShen string `json:"shiShen"` // 相对日主的十神
}

// SiZhuPillar 一柱。
type SiZhuPillar struct {
	Name          string       `json:"name"` // 年柱/月柱/日柱/时柱
	Stem          string       `json:"stem"`
	Branch        string       `json:"branch"`
	StemElement   string       `json:"stemElement"`
	BranchElement string       `json:"branchElement"` // 地支本气五行
	StemShiShen   string       `json:"stemShiShen"`   // 日柱天干为「日主」
	Hidden        []HiddenStem `json:"hidden"`
	NaYin         string       `json:"naYin"`
	XunKong       bool         `json:"xunKong,omitempty"` // 此柱地支落日柱旬空
}

// SiZhuGeJu 月令取格(《子平真诠》法:八字用神专求月令)。
type SiZhuGeJu struct {
	Name   string `json:"name"`   // 正官格/七杀格/…/建禄格/阳刃格/月劫格/杂气月垣
	Basis  string `json:"basis"`  // 取格依据(本气秉令/本气不透取透干/禄刃)
	Note   string `json:"note"`   // 格局大意(据子平真诠,非吉凶断定)
	Source string `json:"source"` // 古籍出处
}

// SiZhuView 四柱视角。
type SiZhuView struct {
	DayMaster        string         `json:"dayMaster"` // 日干
	DayMasterElement string         `json:"dayMasterElement"`
	Pillars          [4]SiZhuPillar `json:"pillars"`
	// ElementCount 八字五行分布(四天干 + 四地支本气,共 8 字)。
	ElementCount map[string]int `json:"elementCount"`
	GeJu         *SiZhuGeJu     `json:"geJu,omitempty"`
	// ShenSha 神煞(三合/年支/日干/空亡),仅列命中柱者。
	ShenSha []ShenSha `json:"shenSha"`
	// DaYun 大运/流年(八字视角,与紫微同源)。
	DaYun *DaYunView `json:"daYun,omitempty"`
	// Note 历法口径说明(与紫微盘面四柱的分界差异)。
	Note string `json:"note,omitempty"`
}

// siZhuCaliberNote 四柱视角历法口径说明(前端展示)。
const siZhuCaliberNote = "四柱视角按子平节气分界:年柱起立春、月柱起节(精确到交接时刻)、晚子时日柱归次日。" +
	"紫微盘面四柱按正月初一分界(iztro 口径),岁首与节交前后两者或相差一柱,属两派口径并存,非计算歧误。"

var (
	szStems    = []rune("甲乙丙丁戊己庚辛壬癸")
	szBranches = []rune("子丑寅卯辰巳午未申酉戌亥")
	// 五行序:木0 火1 土2 金3 水4(顺生)
	elementNames = []string{"木", "火", "土", "金", "水"}
	// 天干五行:甲乙木 丙丁火 戊己土 庚辛金 壬癸水
	stemElement = []int{0, 0, 1, 1, 2, 2, 3, 3, 4, 4}
	// 地支本气五行
	// 基础表:与 liuyao/daliuren/ziwei(sizhu) 三处刻意各自持有(引擎解耦);口径须一致(子水始,木0火1土2金3水4),改此须同步余处,各引擎黄金基准会捕获不一致。
	branchElement = []int{4, 2, 0, 0, 2, 1, 1, 2, 3, 3, 2, 4}
	// 地支藏干(本气在前)
	hiddenStems = [][]rune{
		[]rune("癸"), []rune("己癸辛"), []rune("甲丙戊"), []rune("乙"),
		[]rune("戊乙癸"), []rune("丙庚戊"), []rune("丁己"), []rune("己丁乙"),
		[]rune("庚壬戊"), []rune("辛"), []rune("戊辛丁"), []rune("壬甲"),
	}
	// 六十甲子纳音(两柱一音,索引 = 甲子序/2)
	naYinNames = []string{
		"海中金", "炉中火", "大林木", "路旁土", "剑锋金", "山头火",
		"涧下水", "城头土", "白蜡金", "杨柳木", "泉中水", "屋上土",
		"霹雳火", "松柏木", "长流水", "沙中金", "山下火", "平地木",
		"壁上土", "金箔金", "覆灯火", "天河水", "大驿土", "钗钏金",
		"桑柘木", "大溪水", "沙中土", "天上火", "石榴木", "大海水",
	}
)

func stemIndex(r rune) int {
	for i, s := range szStems {
		if s == r {
			return i
		}
	}
	return -1
}

func branchIndex(r rune) int {
	for i, b := range szBranches {
		if b == r {
			return i
		}
	}
	return -1
}

// shiShen 日主天干 → 另一天干的十神。
// 同五行:同阴阳比肩/异劫财;我生:同食神/异伤官;我克:同偏财/异正财;
// 克我:同七杀/异正官;生我:同偏印/异正印。
func shiShen(dayStem, other int) string {
	de, oe := stemElement[dayStem], stemElement[other]
	same := dayStem%2 == other%2
	pick := func(a, b string) string {
		if same {
			return a
		}
		return b
	}
	switch {
	case de == oe:
		return pick("比肩", "劫财")
	case (de+1)%5 == oe: // 我生
		return pick("食神", "伤官")
	case (de+2)%5 == oe: // 我克
		return pick("偏财", "正财")
	case (oe+2)%5 == de: // 克我
		return pick("七杀", "正官")
	default: // 生我
		return pick("偏印", "正印")
	}
}

// ── 月令取格(子平真诠) ─────────────────────────────

// 禄位:甲寅 乙卯 丙巳 丁午 戊巳 己午 庚申 辛酉 壬亥 癸子
var luBranch = []int{2, 3, 5, 6, 5, 6, 8, 9, 11, 0}

// 阳刃位(阳干):甲卯 丙午 戊午 庚酉 壬子
var renBranch = map[int]int{0: 3, 2: 6, 4: 6, 6: 9, 8: 0}

// geJuNotes 各格大意与出处(据《子平真诠》诸论,描述取用喜忌倾向,非吉凶断定)。
var geJuNotes = map[string][2]string{
	"正官格":  {"官星卫身,贵气所系;喜财印相辅、日主健旺,忌伤官克官、刑冲月令。", "《子平真诠·论正官》"},
	"七杀格":  {"杀以攻身,制合得宜反成大贵;喜食神制杀、印绶化杀,忌财党杀而无制。", "《子平真诠·论偏官》"},
	"正财格":  {"财为养命之源;喜身强任财、食伤生财,忌比劫分夺、月令逢冲。", "《子平真诠·论财》"},
	"偏财格":  {"众人之财,善营豪爽;亦须身旺,喜食伤相生,忌劫刃夺财。", "《子平真诠·论财》"},
	"正印格":  {"印绶生身,主庇荫文贵;喜官杀相生、身弱得印,忌财星坏印。", "《子平真诠·论印绶》"},
	"偏印格":  {"枭印机敏孤介;喜以财损印取贵、杀印相生,忌枭神夺食同透。", "《子平真诠·论印绶》"},
	"食神格":  {"食神吐秀,福寿之征;喜生财、身旺,忌偏印夺食、财官并夺其秀。", "《子平真诠·论食神》"},
	"伤官格":  {"伤官秀气所钟,配印或生财则贵;最忌伤官见官,为祸百端。", "《子平真诠·论伤官》"},
	"建禄格":  {"月令建禄,身旺不以禄为用,须于财官食伤中别求用神。", "《子平真诠·论建禄月劫》"},
	"阳刃格":  {"刃者身强之极,喜官杀制刃成威权之贵,忌刃重无制、刃冲岁运。", "《子平真诠·论阳刃》"},
	"月劫格":  {"月劫同建禄论,身旺帮扶有余,别求财官食伤为用。", "《子平真诠·论建禄月劫》"},
	"杂气月垣": {"四库杂气,财官印所藏;喜所藏之神透干清格,忌壅塞不透。", "《子平真诠·论杂格》"},
}

// deriveGeJu 依月支藏干定格:本气非比劫则取本气(本气不透而他藏透干者取透者);
// 本气为比劫则依禄位/刃位论建禄/阳刃/月劫;土日主逢四库月按杂气透干取格。
func deriveGeJu(dayStem, monthBranch int, transparent map[rune]bool) *SiZhuGeJu {
	hidden := hiddenStems[monthBranch]
	benQi := stemIndex(hidden[0])
	benSS := shiShen(dayStem, benQi)

	mk := func(name, basis string) *SiZhuGeJu {
		n := geJuNotes[name]
		return &SiZhuGeJu{Name: name, Basis: basis, Note: n[0], Source: n[1]}
	}

	if benSS == "比肩" || benSS == "劫财" {
		if luBranch[dayStem] == monthBranch {
			return mk("建禄格", "月支为日主禄位")
		}
		if rb, ok := renBranch[dayStem]; ok && rb == monthBranch {
			return mk("阳刃格", "月支为日主刃位")
		}
		// 土日主逢辰戌丑未等:杂气,取透干之藏神为格
		for _, hr := range hidden[1:] {
			if transparent[hr] {
				g := mk(shiShen(dayStem, stemIndex(hr))+"格", "杂气月垣,"+string(hr)+"透干取格")
				return g
			}
		}
		if benSS == "劫财" {
			return mk("月劫格", "月支本气为劫财")
		}
		return mk("杂气月垣", "所藏之神皆不透干")
	}

	// 本气透干,或月令唯一藏干:径取本气
	if transparent[hidden[0]] || len(hidden) == 1 {
		basis := "月支本气秉令"
		if transparent[hidden[0]] {
			basis = "月支本气透干"
		}
		return mk(benSS+"格", basis)
	}
	// 本气不透:取透出之中气/余气(子平真诠:以透出者为格)
	for _, hr := range hidden[1:] {
		hs := stemIndex(hr)
		ss := shiShen(dayStem, hs)
		if transparent[hr] && ss != "比肩" && ss != "劫财" {
			return mk(ss+"格", "本气不透,取透干之"+string(hr)+"为格")
		}
	}
	// 皆不透:仍以本气论
	return mk(benSS+"格", "藏神皆不透,仍以本气论")
}

// naYin 干支 → 纳音(六十甲子序)。
func naYin(stem, branch int) string {
	for j := 0; j < 60; j++ {
		if j%10 == stem && j%12 == branch {
			return naYinNames[j/2]
		}
	}
	return ""
}

// buildSiZhu 由四柱干支字符串构建四柱视角。
func buildSiZhu(fp FourPillars) *SiZhuView {
	pillarStrs := []string{fp.Year, fp.Month, fp.Day, fp.Hour}
	names := []string{"年柱", "月柱", "日柱", "时柱"}

	dayRunes := []rune(fp.Day)
	if len(dayRunes) != 2 {
		return nil
	}
	dayStem := stemIndex(dayRunes[0])
	if dayStem < 0 {
		return nil
	}

	view := &SiZhuView{
		DayMaster:        string(szStems[dayStem]),
		DayMasterElement: elementNames[stemElement[dayStem]],
		ElementCount:     map[string]int{"木": 0, "火": 0, "土": 0, "金": 0, "水": 0},
	}

	var stemsIdx, branchesIdx [4]int
	for i, ps := range pillarStrs {
		rs := []rune(ps)
		if len(rs) != 2 {
			return nil
		}
		s, b := stemIndex(rs[0]), branchIndex(rs[1])
		if s < 0 || b < 0 {
			return nil
		}
		stemsIdx[i], branchesIdx[i] = s, b
		p := SiZhuPillar{
			Name:          names[i],
			Stem:          string(rs[0]),
			Branch:        string(rs[1]),
			StemElement:   elementNames[stemElement[s]],
			BranchElement: elementNames[branchElement[b]],
			NaYin:         naYin(s, b),
		}
		if i == 2 {
			p.StemShiShen = "日主"
		} else {
			p.StemShiShen = shiShen(dayStem, s)
		}
		for _, hr := range hiddenStems[b] {
			hs := stemIndex(hr)
			p.Hidden = append(p.Hidden, HiddenStem{
				Stem:    string(hr),
				Element: elementNames[stemElement[hs]],
				ShiShen: shiShen(dayStem, hs),
			})
		}
		view.ElementCount[p.StemElement]++
		view.ElementCount[p.BranchElement]++
		view.Pillars[i] = p
	}

	// 月令取格:透干以年/月/时三天干论(日干为日主自身,不计)
	monthBranch := branchIndex([]rune(fp.Month)[1])
	transparent := map[rune]bool{}
	for _, i := range []int{0, 1, 3} {
		transparent[[]rune(pillarStrs[i])[0]] = true
	}
	view.GeJu = deriveGeJu(dayStem, monthBranch, transparent)

	// 神煞层 + 空亡回填
	shenSha, kong := buildShenSha(stemsIdx, branchesIdx)
	view.ShenSha = shenSha
	for i := range view.Pillars {
		view.Pillars[i].XunKong = kong[i]
	}
	return view
}
