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
	"交友": {"交友·部属", "朋友部属的助益或拖累"},
	"官禄": {"事业·官禄", "事业格局与成就取向"},
	"田宅": {"田宅·家运", "不动产、家宅与库藏"},
	"福德": {"福德·精神", "兴趣、心性与福分享受"},
	"父母": {"父母·长上", "双亲缘分与长辈、文书"},
}

// 维度呈现顺序(重要在前)。
var readingOrder = []string{"命宫", "财帛", "官禄", "夫妻", "迁移", "福德", "疾厄", "田宅", "子女", "兄弟", "交友", "父母"}

var sihuaNote = map[ziwei.SiHua]string{
	ziwei.HuaLu:   "化禄,进财顺遂、机遇增益",
	ziwei.HuaQuan: "化权,掌权任事、地位提升",
	ziwei.HuaKe:   "化科,名声贵人、文书之喜",
	ziwei.HuaJi:   "化忌,阻滞是非、防执着招损",
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
	rd := &Reading{Sections: make([]ReadingSection, 0, len(readingOrder)+1)}
	rd.Overview = buildOverview(chart, patterns)
	for _, pname := range readingOrder {
		p := chart.PalaceByName(pname)
		if p == nil {
			continue
		}
		rd.Sections = append(rd.Sections, sectionForPalace(pname, p))
	}
	if s := sectionForDaXian(chart); s != nil {
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

func sectionForPalace(pname string, p *ziwei.Palace) ReadingSection {
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
		clauses = append(clauses, name+"("+slant+")")
	}
	// 会照煞/吉
	score += len(lucky) - len(sha)

	var b strings.Builder
	if len(majors) == 0 {
		b.WriteString(fmt.Sprintf("%s宫无正曜,清静之宫,力量借他宫而定。", pname))
	} else {
		lead := "本宫"
		if borrowed {
			lead = fmt.Sprintf("本宫空,借对宫【%s】", p.BorrowedFromName)
		}
		b.WriteString(fmt.Sprintf("%s坐 %s,主%s。以 %s 之性投于此域:%s。",
			lead, strings.Join(starTags, "、"), lens[1], strings.Join(majors, "、"), strings.Join(clauses, ";")))
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
	b.WriteString(levelHint(pname, score))

	if len(starTags) == 0 && borrowed {
		starTags = majors
	}
	return ReadingSection{
		Key: palaceKey(pname), Title: lens[0], Palace: pname,
		Stars: starTags, Level: levelOf(score), Text: b.String(),
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
	} else {
		b.WriteString(who + "论命。")
	}
	b.WriteString(fmt.Sprintf("命属%s,身宫在%s宫。", chart.WuxingJuName, ziwei.Branches[chart.ShenGongBranch]))
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
	"财帛": "caibo", "疾厄": "jie", "迁移": "qianyi", "交友": "jiaoyou",
	"官禄": "guanlu", "田宅": "tianzhai", "福德": "fude", "父母": "fumu",
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
