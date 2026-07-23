package ai

// 结构化多维断语:不依赖 LLM,由命盘各宫「主星 × 庙旺 × 生年四化 × 煞吉会照」
// 组合出逐宫、逐维度的断语。断语随命盘星曜实配而变——两张不同的盘会得到
// 显著不同的解读,解决「不同人断语雷同」的问题。内容为我方原创 + 公版紫微
// 通则,不含任何第三方软件的断语文本。LLM 存在时,本结构作为「事实骨架」注入
// 提示词,令 LLM 输出同样贴盘;不存在时,直接渲染为规则版解读。

import (
	"fmt"
	"strings"

	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// ReadingSection 一个维度(通常对应一宫)的断语。
type ReadingSection struct {
	Key    string   `json:"key"`    // ming/cai/guan/fu/...
	Title  string   `json:"title"`  // 事业·官禄
	Palace string   `json:"palace"` // 宫名
	Stars  []string `json:"stars"`  // 该宫主星(带庙旺/四化标记),空宫为借星
	Level  string   `json:"level"`  // good/caution/neutral
	Text   string   `json:"text"`   // 组合断语
}

// Reading 多维结构化断语。
type Reading struct {
	Overview string           `json:"overview"` // 命格总论
	Sections []ReadingSection `json:"sections"`
}

// starTrait 十四主星特质:{核心, 得地(庙旺)正面, 落陷负面}。原创提炼自公版通则。
var starTrait = map[string][3]string{
	"紫微": {"尊贵领导、主观自重", "格局高、有威权、堪当大任", "孤高刚愎、眼高手低、喜奉承"},
	"天机": {"机敏善谋、心思灵动", "谋略过人、企划才佳", "多思少成、心绪不宁、善变"},
	"太阳": {"博爱光明、付出施予", "事业光明、有名望、利男亲", "劳而费力、性急、男性缘损"},
	"武曲": {"刚毅果决、财星务实", "财官双美、执行力强", "孤克刚硬、财来财去、欠圆融"},
	"天同": {"温和知足、福星享受", "有福有寿、人缘和顺", "惰性依赖、情绪化、欠冲劲"},
	"廉贞": {"次桃花、能干多变", "干练圆滑、异途显达", "是非情困、易招血光官非"},
	"天府": {"稳重保守、财库之星", "富足稳当、有理财守成", "空库外强中干、过于自保"},
	"太阴": {"内敛柔美、田宅财星", "富而秀、细腻、利女亲", "情绪起伏、母缘薄、暗损"},
	"贪狼": {"欲望才艺、交际桃花", "多才多艺、善交际应酬", "贪多务得、易溺酒色赌博"},
	"巨门": {"口才研究、明辨深究", "专业口才、以口维生佳", "口舌是非、多疑招小人"},
	"天相": {"辅佐忠厚、印星衣食", "衣食无缺、贵人相助", "懦弱受制、优柔缺主见"},
	"天梁": {"清高荫庇、老成之星", "逢凶化吉、有寿、得庇荫", "孤高飘荡、爱管闲事"},
	"七杀": {"肃杀开创、将星独当", "威权、能独当一面", "冲动成败、六亲缘薄"},
	"破军": {"破旧立新、耗星开创", "开创魄力、先破后成", "破耗动荡、变动不安"},
}

// palaceLens 十二宫透镜:{维度标题, 领域说法}。
var palaceLens = map[string][2]string{
	"命宫": {"性格·命格", "性格气质与人生基调"},
	"兄弟": {"手足·同侪", "兄弟手足与平辈助力"},
	"夫妻": {"婚姻·感情", "配偶特质与婚姻缘分"},
	"子女": {"子女·桃花", "子嗣缘分与晚辈、桃花"},
	"财帛": {"财帛·财运", "求财方式与财富厚薄"},
	"疾厄": {"健康·体质", "身体弱点与情绪健康"},
	"迁移": {"迁移·外出", "在外际遇与远行人脉"},
	"仆役": {"交友·部属", "朋友部属的助益或拖累"},
	"官禄": {"事业·官禄", "事业格局与成就取向"},
	"田宅": {"田宅·家运", "不动产、家宅与库藏"},
	"福德": {"福德·精神", "兴趣、心性与福分享受"},
	"父母": {"父母·长上", "双亲缘分与长辈、文书"},
}

// 维度呈现顺序(重要在前)。
var readingOrder = []string{"命宫", "财帛", "官禄", "夫妻", "迁移", "福德", "疾厄", "田宅", "子女", "兄弟", "仆役", "父母"}

var sihuaNote = map[ziwei.SiHua]string{
	ziwei.HuaLu:   "化禄,进财顺遂、机遇增益",
	ziwei.HuaQuan: "化权,掌权任事、地位提升",
	ziwei.HuaKe:   "化科,名声贵人、文书之喜",
	ziwei.HuaJi:   "化忌,阻滞是非、防执着招损",
}

// majorRank 十四主星规范排序(紫微系在前、天府系随后),用于「双主星组合」定名。
var majorRank = map[string]int{
	"紫微": 0, "天机": 1, "太阳": 2, "武曲": 3, "天同": 4, "廉贞": 5, "天府": 6,
	"太阴": 7, "贪狼": 8, "巨门": 9, "天相": 10, "天梁": 11, "七杀": 12, "破军": 13,
}

// starPairTrait 双主星同宫组合专断。同宫两主星并非各自特质相加,而自成一格——
// 这是紫微斗数星系断语「贴盘准头」的关键一层:紫破与武贪、廉杀,格性天差地别。
// 内容为三合派公版星系通则的原创提炼(非任何第三方软件文本)。键为规范组合名。
var starPairTrait = map[string]string{
	// 紫微系六组
	"紫微天府": "帝星得府库,尊贵又能守成,富贵安泰之组;喜左右昌曲夹拱则贵显,孤君无辅则空有其表、流于自负专断。",
	"紫微贪狼": "尊星逢桃花,才艺风流、交际手腕高、好物质享受;见桃花煞易溺酒色,得空曜火铃反主宗教哲思、转化为进取。",
	"紫微天相": "帝星得印星辅佐,忠厚有格、处事得体、宜从公从政,多贵人相扶;性温和而主观仍在,须防耳根软受制。",
	"紫微七杀": "帝星驾杀、化气为权,威权肃杀、能独当一面、开创力强,宜掌实业兵权;过刚则孤,须防独断专行。",
	"紫微破军": "帝星遇耗星,敢破敢立、不安现状、开创求变;一生成败起伏大,宜先破后立、忌再见煞星则更动荡。",
	// 天机系
	"天机太阴": "智星配柔星,心思细密、善策划文墨,宜幕僚企划文教财经;情绪较敏感、迁动多,近于谋士而非主帅。",
	"天机巨门": "机变加口才,善辩多智,宜口才研究技术传播为生;是非口舌较多、心多疑虑,须防聪明反被聪明误。",
	"天机天梁": "智星逢荫星,善谈玄理、有谋略又清高,宜宗教哲学五术、参谋顾问;近「善谈兵」而少实战,宜沉潜务实。",
	// 太阳系
	"太阳太阴": "日月同宫、阴阳双秉,才华洋溢而心性多变;庙旺则名利双收,落陷则内心矛盾、劳碌起伏,丑未之地尤须细察日月得失。",
	"太阳巨门": "阳星配暗星,以口舌名声立业,宜外语法律传播异族生意;能化暗为明、名播远方,然口舌是非难全免。",
	"太阳天梁": "光明配荫星,正直有名望、宜公职文教清贵;逢文昌禄存成「阳梁昌禄」,主贵显、利考试功名科甲。",
	// 武曲系
	"武曲天府": "财星坐财库,善理财、精算务实,富格之组,宜金融财会实业;较重物质、稍欠情趣,守成有余而开创稍缓。",
	"武曲贪狼": "财星配欲星,「武贪不发少年人」,中晚年横发之格,行动力企图心强,宜市场业务军警;见火铃成火贪铃贪暴发格。",
	"武曲天相": "财星得印,稳健务实、有信誉,宜财经实业,执行力强;处事持重而少变通,宜广纳雅言。",
	"武曲七杀": "双刚会聚,决断刚毅、开创力猛,宜军警外科技术等竞争性行业;性刚易孤克,财上须防因刚失和、当机立断亦当留转圜。",
	"武曲破军": "财星遇耗,大破大立、白手创业、财来财去;冲劲十足而不耐守成,宜动中求财、忌投机豪赌。",
	// 天同系
	"天同太阴": "双柔福星,温和恬淡、感情细腻、近艺文享受,宜服务文创女性市场;庙旺清秀有福,落陷则耽逸少冲劲。",
	"天同巨门": "福星配暗星,心软而多虑、口才佳却易招是非;宜以专业口才立身,情感多波折,须防钻牛角尖。",
	"天同天梁": "福荫相会,温和厚道、逢凶化吉、人缘佳,宜服务教育医护公职;安稳有余而进取不足,宜借煞星激发志气。",
	// 廉贞系
	"廉贞天府": "囚星得库,能干而守成、外圆内方,宜行政管理财务;理性务实、感情较含蓄,处世圆融中自有分寸。",
	"廉贞贪狼": "双桃花,才艺交际、感情丰富、应酬手腕强;巳亥落陷易流放纵酒色,须以正业收束、专注方能成器。",
	"廉贞天相": "囚星得印,能干圆融、善协调,宜从政从商;若见煞成「刑囚夹印」防官非牢狱,须守法自持、慎签契约。",
	"廉贞七杀": "「路上埋尸」之烈组,刚烈果决、拼搏力强,宜军警技术冒险之业;逢煞主血光意外,须化刚为柔、慎防伤灾。",
	"廉贞破军": "囚星遇耗,变动激烈、敢冲敢闯、破旧立新;人生起伏大,宜投身开创变革之业、忌感情用事招祸惹非。",
}

// pairTraitOf 宫内恰为两颗已知主星时,返回其组合规范名与专断,否则空。
func pairTraitOf(majors []string) (name, trait string) {
	if len(majors) != 2 {
		return "", ""
	}
	a, b := majors[0], majors[1]
	ra, oka := majorRank[a]
	rb, okb := majorRank[b]
	if !oka || !okb {
		return "", ""
	}
	if ra > rb {
		a, b = b, a
	}
	key := a + b
	if t, ok := starPairTrait[key]; ok {
		return key, t
	}
	return "", ""
}

// BuildReading 由命盘组合多维断语(确定性,不走 LLM)。
func (it *Interpreter) BuildReading(chart *ziwei.Chart, patterns []ziwei.Pattern) *Reading {
	return buildReading(chart, patterns)
}

// buildReading 纯函数版:仅依赖命盘与格局。
func buildReading(chart *ziwei.Chart, patterns []ziwei.Pattern) *Reading {
	if chart == nil {
		return nil
	}
	rd := &Reading{Sections: make([]ReadingSection, 0, len(readingOrder)+2)}
	rd.Overview = buildOverview(chart, patterns)
	if s := sectionForSihua(chart); s != nil { // 四化落宫置前:最贴盘的差异层
		rd.Sections = append(rd.Sections, *s)
	}
	for _, pname := range readingOrder {
		p := chart.PalaceByName(pname)
		if p == nil {
			continue
		}
		rd.Sections = append(rd.Sections, sectionForPalace(chart, pname, p))
	}
	if s := sectionForDaXian(chart); s != nil {
		rd.Sections = append(rd.Sections, *s)
	}
	if s := sectionForLiuNian(chart); s != nil { // 流年:当年一岁之气
		rd.Sections = append(rd.Sections, *s)
	}
	return rd
}

// palaceMajors 取宫主星(空宫返回借星名 + borrowed=true)。
func palaceMajors(p *ziwei.Palace) (names []string, borrowed bool) {
	names = p.MajorStarNames()
	if len(names) == 0 && len(p.BorrowedStars) > 0 {
		return p.BorrowedStars, true
	}
	return names, false
}

// starMods 从宫内星曜取某主星的庙旺档与四化,及全宫煞/吉会照。
func scanPalace(p *ziwei.Palace) (bright, dim map[string]bool, sihua map[string]ziwei.SiHua, sha, lucky []string) {
	bright, dim = map[string]bool{}, map[string]bool{}
	sihua = map[string]ziwei.SiHua{}
	for i := range p.Stars {
		s := &p.Stars[i]
		switch s.Type {
		case ziwei.StarMajor:
			if s.BrightnessLevel == "bright" {
				bright[s.Name] = true
			} else if s.BrightnessLevel == "dim" {
				dim[s.Name] = true
			}
			if s.SiHua != "" {
				sihua[s.Name] = s.SiHua
			}
		case ziwei.StarSha:
			sha = append(sha, s.Name)
		case ziwei.StarLucky:
			lucky = append(lucky, s.Name)
		}
	}
	return
}

// triadMajors 三方四正(对宫 +6、三合 +4/+8)三宫的主星名(去重),供会照合参。
func triadMajors(chart *ziwei.Chart, branch int) []string {
	var out []string
	seen := map[string]bool{}
	for _, off := range []int{4, 6, 8} {
		tp := chart.PalaceByBranch(branch + off)
		if tp == nil {
			continue
		}
		for _, n := range tp.MajorStarNames() {
			if !seen[n] {
				seen[n] = true
				out = append(out, n)
			}
		}
	}
	return out
}

func sectionForPalace(chart *ziwei.Chart, pname string, p *ziwei.Palace) ReadingSection {
	lens := palaceLens[pname]
	majors, borrowed := palaceMajors(p)
	bright, dim, sihua, sha, lucky := scanPalace(p)

	var starTags []string
	var clauses []string
	score := 0
	for _, name := range majors {
		tag := name
		trait := starTrait[name]
		slant := trait[0] // 核心
		if bright[name] {
			tag += "庙旺"
			slant = trait[1]
			score += 2
		} else if dim[name] {
			tag += "落陷"
			slant = trait[2]
			score -= 2
		} else {
			score++
		}
		if h, ok := sihua[name]; ok {
			tag += "化" + string(h)
			if h == ziwei.HuaJi {
				score -= 3
			} else {
				score += 2
			}
		}
		starTags = append(starTags, tag)
		// 宫位透镜按星细化:优先取「星×宫」专属断语,缺项回退通用星性。
		if sp := starInPalace(name, pname); sp != "" {
			clauses = append(clauses, name+"："+sp)
		} else {
			clauses = append(clauses, name+"("+slant+")")
		}
	}
	// 会照煞/吉
	score += len(lucky) - len(sha)

	var b strings.Builder
	if len(majors) == 0 {
		b.WriteString(fmt.Sprintf("%s无正曜,清静之宫,力量借他宫而定。", palaceLabel(pname)))
	} else {
		lead := "本宫"
		if borrowed {
			lead = fmt.Sprintf("本宫空,借对宫【%s】", p.BorrowedFromName)
		}
		b.WriteString(fmt.Sprintf("%s坐 %s,主%s。以 %s 之性投于此域:%s。",
			lead, strings.Join(starTags, "、"), lens[1], strings.Join(majors, "、"), strings.Join(clauses, ";")))
	}
	// 双主星同宫组合专断(准头核心:组合自成一格,非特质相加)
	if pn, pt := pairTraitOf(majors); pt != "" {
		b.WriteString(fmt.Sprintf("此为【%s】同宫:%s", pn, pt))
	}
	// 四化点睛
	for name, h := range sihua {
		b.WriteString(fmt.Sprintf("%s%s。", name, sihuaNote[h]))
	}
	if len(sha) > 0 {
		b.WriteString(fmt.Sprintf("会煞星%s,主此域多波折冲击,宜制化、忌强求。", strings.Join(sha, "、")))
	}
	if len(lucky) > 0 {
		b.WriteString(fmt.Sprintf("得吉星%s相扶,添助力、逢难有救。", strings.Join(lucky, "、")))
	}
	// 三方四正合参(紫微铁律:本宫须连对宫、三合两宫同看)
	if triad := triadMajors(chart, p.Branch); len(triad) > 0 {
		b.WriteString(fmt.Sprintf("三方四正会照 %s,%s一域非独看本宫,须合此数曜之势通断。", strings.Join(triad, "、"), pname))
	}
	b.WriteString(levelHint(pname, score))

	if len(starTags) == 0 && borrowed {
		starTags = majors
	}
	return ReadingSection{
		Key: palaceKey(pname), Title: lens[0], Palace: pname,
		Stars: starTags, Level: levelOf(score), Text: b.String(),
	}
}

// sihuaLanding 生年四化落宫:某化 → 承化之星与其坐落之宫。
type sihuaLanding struct {
	Hua    ziwei.SiHua
	Star   string
	Palace string
	Branch int
}

// sihuaLandings 扫全盘,取生年四化各自落于何宫(禄权科忌顺序)。
func sihuaLandings(chart *ziwei.Chart) []sihuaLanding {
	byHua := map[ziwei.SiHua]sihuaLanding{}
	for pi := range chart.Palaces {
		p := &chart.Palaces[pi]
		for si := range p.Stars {
			if s := &p.Stars[si]; s.SiHua != "" {
				byHua[s.SiHua] = sihuaLanding{Hua: s.SiHua, Star: s.Name, Palace: p.Name, Branch: p.Branch}
			}
		}
	}
	var out []sihuaLanding
	for _, h := range []ziwei.SiHua{ziwei.HuaLu, ziwei.HuaQuan, ziwei.HuaKe, ziwei.HuaJi} {
		if l, ok := byHua[h]; ok {
			out = append(out, l)
		}
	}
	return out
}

// sihuaDomainEffect 四化落宫对该宫领域的定性(福泽 vs 罣碍)。
var sihuaDomainEffect = map[ziwei.SiHua]string{
	ziwei.HuaLu:   "为福泽进财所在,该域机遇顺遂、易得实惠",
	ziwei.HuaQuan: "为掌权任事所在,该域能力凸显、宜主导进取",
	ziwei.HuaKe:   "为名声贵人所在,该域平顺有助、逢难有救",
	ziwei.HuaJi:   "为此生功课罣碍所在,该域易生执着牵绊、须多费心化解",
}

// sectionForSihua 生年四化落宫串联:倪师体系四化固定不动,看其坐落何宫,
// 即知一生福泽(禄权科)与罣碍(忌)分布——这是最贴盘、最具个人差异的一层。
func sectionForSihua(chart *ziwei.Chart) *ReadingSection {
	ls := sihuaLandings(chart)
	if len(ls) == 0 {
		return nil
	}
	var b strings.Builder
	var tags []string
	jiBranch, luBranch := -1, -1
	b.WriteString("生年四化固定不动,其坐落之宫定一生福泽与罣碍:")
	for _, l := range ls {
		domain := ""
		if lens := palaceLens[l.Palace]; lens[1] != "" {
			domain = "(" + lens[1] + ")"
		}
		b.WriteString(fmt.Sprintf("%s化%s入【%s】%s,%s;", l.Star, string(l.Hua), palaceLabel(l.Palace), domain, sihuaDomainEffect[l.Hua]))
		tags = append(tags, l.Star+"化"+string(l.Hua)+"·"+l.Palace)
		switch l.Hua {
		case ziwei.HuaJi:
			jiBranch = l.Branch
		case ziwei.HuaLu:
			luBranch = l.Branch
		}
	}
	// 化忌冲对宫:忌宫对面之域连带受牵。
	if jiBranch >= 0 {
		if opp := chart.PalaceByBranch(jiBranch + 6); opp != nil {
			b.WriteString(fmt.Sprintf("化忌并冲对宫【%s】,该域亦连带牵动、宜防其失;", palaceLabel(opp.Name)))
		}
	}
	// 禄忌交战:禄忌同宫或对宫,主福祸相倚、成败起伏。
	if jiBranch >= 0 && luBranch >= 0 {
		if jiBranch == luBranch {
			b.WriteString("禄忌同宫,福祸相倚、得失同门,成败皆系于此域;")
		} else if (jiBranch+6)%12 == luBranch {
			b.WriteString("禄忌对拱交战,一进一退、起伏相随,此两域宜通盘权衡;")
		}
	}
	level := "neutral"
	if jiBranch >= 0 {
		if jp := chart.PalaceByBranch(jiBranch); jp != nil {
			switch jp.Name {
			case "命宫", "疾厄", "夫妻", "财帛":
				level = "caution" // 忌落切身之宫,功课较重
			}
		}
	}
	return &ReadingSection{
		Key: "sihua", Title: "生年四化·福泽罣碍", Palace: "", Stars: tags,
		Level: level, Text: b.String(),
	}
}

func sectionForDaXian(chart *ziwei.Chart) *ReadingSection {
	if chart.CurrentDaXianIndex < 0 || chart.CurrentDaXianIndex >= len(chart.DaXians) {
		return nil
	}
	dx := chart.DaXians[chart.CurrentDaXianIndex]
	p := chart.PalaceByBranch(dx.PalaceBranch)
	if p == nil {
		return nil
	}
	majors, borrowed := palaceMajors(p)
	bright, dim, sihua, sha, lucky := scanPalace(p)
	score := len(lucky) - len(sha)
	var tags, clauses []string
	for _, name := range majors {
		trait := starTrait[name]
		slant := trait[0]
		tag := name
		if bright[name] {
			tag, slant, score = tag+"庙旺", trait[1], score+2
		} else if dim[name] {
			tag, slant, score = tag+"落陷", trait[2], score-2
		}
		if h, ok := sihua[name]; ok {
			tag += "化" + string(h)
		}
		tags = append(tags, tag)
		clauses = append(clauses, name+"("+slant+")")
	}
	prefix := fmt.Sprintf("现行 %d-%d 岁【%s】大限(%s宫)。", dx.StartAge, dx.EndAge, dx.PalaceName, ziwei.Branches[dx.PalaceBranch])
	body := "此十年以本宫星曜为主题。"
	if len(majors) > 0 {
		lead := "限内"
		if borrowed {
			lead = "限宫空借,"
		}
		body = fmt.Sprintf("%s坐 %s,十年主题:%s。", lead, strings.Join(tags, "、"), strings.Join(clauses, ";"))
	}
	extra := ""
	if len(sha) > 0 {
		extra += fmt.Sprintf("逢煞%s,防波动。", strings.Join(sha, "、"))
	}
	if len(lucky) > 0 {
		extra += fmt.Sprintf("得吉%s,宜把握。", strings.Join(lucky, "、"))
	}
	return &ReadingSection{
		Key: "daxian", Title: "当前大限", Palace: dx.PalaceName,
		Stars: tags, Level: levelOf(score), Text: prefix + body + extra + "(倪师体系四化固定,大限重在宫位星曜的十年主题切换)",
	}
}

// findStarPalace 全盘定位某星所在之宫(十四主星与六吉恒在盘上)。
func findStarPalace(chart *ziwei.Chart, star string) *ziwei.Palace {
	for i := range chart.Palaces {
		if chart.Palaces[i].HasStar(star) {
			return &chart.Palaces[i]
		}
	}
	return nil
}

// liuNianHuaWord 流年四化飞入本命某宫,对本年该域的定性(时效性,不同于生年四化的一生定盘)。
var liuNianHuaWord = map[ziwei.SiHua]string{
	ziwei.HuaLu:   "本年该域进财顺遂、多喜庆机遇",
	ziwei.HuaQuan: "本年该域宜掌事任权、主动可成",
	ziwei.HuaKe:   "本年该域有贵人名声、逢难有解",
	ziwei.HuaJi:   "本年该域易生波折是非,宜谨慎守成、忌冲动强求",
}

// sectionForLiuNian 流年维度断语:以 ReferenceYear 为当年,叠本命盘——
// 流年命宫落于本命何宫(定本年主题)+ 流年四化飞入本命何宫(定本年动向)。
// 倪师体系:生年四化仍固定不动,流年只叠加当年年干四化这一层动态。
func sectionForLiuNian(chart *ziwei.Chart) *ReadingSection {
	year := chart.ReferenceYear
	if year <= 0 {
		return nil
	}
	yb := ziwei.YearBranchIndex(year)
	ys := ziwei.YearStemIndex(year)
	ganzhi := ziwei.Stems[ys] + ziwei.Branches[yb]
	lnMing := chart.PalaceByBranch(yb)
	if lnMing == nil {
		return nil
	}
	set := ziwei.LiuNianSiHua(year)

	var b strings.Builder
	var tags []string
	b.WriteString(fmt.Sprintf("%d 年为%s年,流年命宫落于本命【%s】", year, ganzhi, palaceLabel(lnMing.Name)))
	majors, _ := palaceMajors(lnMing)
	if len(majors) > 0 {
		var cl []string
		for _, n := range majors {
			cl = append(cl, n+"("+starTrait[n][0]+")")
		}
		b.WriteString(fmt.Sprintf("(坐 %s),本年整体气象以此为主题:%s。", strings.Join(majors, "、"), strings.Join(cl, ";")))
		if pn, pt := pairTraitOf(majors); pt != "" {
			b.WriteString(fmt.Sprintf("【%s】:%s", pn, firstSentence(pt)))
		}
	} else {
		b.WriteString(",本宫无正曜,借对宫参看,本年宜守常、随三方之势。")
	}
	// 流年四化飞入本命宫位。
	b.WriteString("流年四化动向:")
	jiInHeavy := false
	for _, hv := range []struct {
		h    ziwei.SiHua
		star string
	}{{ziwei.HuaLu, set.Lu}, {ziwei.HuaQuan, set.Quan}, {ziwei.HuaKe, set.Ke}, {ziwei.HuaJi, set.Ji}} {
		if hv.star == "" {
			continue
		}
		where := "本命盘外"
		if p := findStarPalace(chart, hv.star); p != nil {
			where = "本命" + palaceLabel(p.Name)
			if hv.h == ziwei.HuaJi {
				switch p.Name {
				case "命宫", "疾厄", "夫妻", "财帛", "官禄":
					jiInHeavy = true
				}
			}
		}
		b.WriteString(fmt.Sprintf("%s化%s飞入%s,%s;", hv.star, string(hv.h), where, liuNianHuaWord[hv.h]))
		tags = append(tags, hv.star+"化"+string(hv.h))
	}
	level := "neutral"
	if jiInHeavy {
		level = "caution"
	}
	return &ReadingSection{
		Key: "liunian", Title: fmt.Sprintf("流年·%d %s", year, ganzhi), Palace: "",
		Stars: tags, Level: level,
		Text: b.String() + "(流年为当年一岁之气,与大限十年、本命一生分层合看)",
	}
}

func buildOverview(chart *ziwei.Chart, patterns []ziwei.Pattern) string {
	ming := chart.MingGong()
	majors, borrowed := palaceMajors(ming)
	bright, dim, sihua, sha, lucky := scanPalace(ming)
	var b strings.Builder
	who := "命宫"
	if borrowed {
		who = fmt.Sprintf("命宫无正曜,借【%s】", ming.BorrowedFromName)
	}
	if len(majors) > 0 {
		var parts []string
		for _, name := range majors {
			t := starTrait[name]
			s := t[0]
			if bright[name] {
				s = t[1]
			} else if dim[name] {
				s = t[2]
			}
			if h, ok := sihua[name]; ok {
				s += "、化" + string(h)
			}
			parts = append(parts, name+"("+s+")")
		}
		b.WriteString(fmt.Sprintf("%s坐 %s,立命之本:%s。", who, strings.Join(majors, "、"), strings.Join(parts, ";")))
		if pn, pt := pairTraitOf(majors); pt != "" {
			b.WriteString(fmt.Sprintf("命宫双主星【%s】——%s", pn, pt))
		}
	} else {
		b.WriteString(who + "论命。")
	}
	b.WriteString(fmt.Sprintf("命属%s,身宫在%s宫。", chart.WuxingJuName, ziwei.Branches[chart.ShenGongBranch]))
	if triad := triadMajors(chart, ming.Branch); len(triad) > 0 {
		b.WriteString(fmt.Sprintf("命宫三方四正会 %s,当合参定高低。", strings.Join(triad, "、")))
	}
	// 生年化忌坐宫:一生功课所在,最能点出命主的牵绊主题。
	for _, l := range sihuaLandings(chart) {
		if l.Hua == ziwei.HuaJi {
			b.WriteString(fmt.Sprintf("生年%s化忌坐【%s】,此生功课多在%s一域,宜早修此心;", l.Star, palaceLabel(l.Palace), l.Palace))
			break
		}
	}
	if len(sha) > 0 {
		b.WriteString(fmt.Sprintf("命逢%s,性格带冲劲亦需修养;", strings.Join(sha, "")))
	}
	if len(lucky) > 0 {
		b.WriteString(fmt.Sprintf("命会%s,一生多贵人助;", strings.Join(lucky, "")))
	}
	if len(patterns) > 0 {
		top := patterns[0]
		b.WriteString(fmt.Sprintf("格局见【%s】(%s):%s", top.Name, levelLabel(top.Level), firstSentence(top.Description)))
	}
	return b.String()
}

func levelWord(level string) string {
	switch level {
	case "good":
		return "(吉)"
	case "caution":
		return "(慎)"
	default:
		return "(平)"
	}
}

func levelOf(score int) string {
	switch {
	case score >= 3:
		return "good"
	case score <= -2:
		return "caution"
	default:
		return "neutral"
	}
}

func levelHint(pname string, score int) string {
	switch levelOf(score) {
	case "good":
		return fmt.Sprintf("综论:%s一域根基佳,宜顺势进取。", pname)
	case "caution":
		return fmt.Sprintf("综论:%s一域压力较显,宜守成、化解为先。", pname)
	default:
		return fmt.Sprintf("综论:%s一域中平,成事在人为。", pname)
	}
}

var palaceKeyMap = map[string]string{
	"命宫": "ming", "兄弟": "xiongdi", "夫妻": "fuqi", "子女": "zinv",
	"财帛": "caibo", "疾厄": "jie", "迁移": "qianyi", "仆役": "jiaoyou",
	"官禄": "guanlu", "田宅": "tianzhai", "福德": "fude", "父母": "fumu",
}

// palaceLabel 宫名规范为「…宫」显示;命宫本身已含「宫」则不叠加。
func palaceLabel(name string) string {
	if strings.HasSuffix(name, "宫") {
		return name
	}
	return name + "宫"
}

func palaceKey(pname string) string {
	if k, ok := palaceKeyMap[pname]; ok {
		return k
	}
	return pname
}

func firstSentence(s string) string {
	for i, r := range s {
		if r == '。' || r == '；' || r == ';' {
			return s[:i] + "。"
		}
	}
	if r := []rune(s); len(r) > 40 {
		return string(r[:40]) + "…"
	}
	return s
}
