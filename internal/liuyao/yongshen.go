// 用神建议层:按所问事类映射六亲(《增删卜易·用神章》为纲,
// 《卜筮正宗·用神问答》互证;考据见 research/liuyao-yongshen.md)。
// 仅作建议:事类识别为关键词层,AI 解卦可按事理改取(书死理活)。
package liuyao

import "strings"

// yongShenRule 一条事类规则:命中关键词 → 用神六亲与经义依据。
type yongShenRule struct {
	name     string
	basis    string
	keywords []string
}

// 匹配序:先人物(所占之人定用神,如「儿子考试」取子孙非官鬼),后事类。
var yongShenRules = []yongShenRule{
	// ── 人物 ──
	{"父母", "占父母、祖辈、师长及庇护我者,以父母爻为用神(增删卜易·用神章)",
		[]string{"父", "母", "爸", "妈", "双亲", "长辈", "老师", "师长", "祖父", "祖母", "爷爷", "奶奶", "外公", "外婆", "伯", "叔", "姑", "姨", "舅"}},
	{"官鬼", "妻占夫,以官鬼爻为用神(增删卜易·用神章)",
		[]string{"丈夫", "老公", "夫君", "未婚夫"}},
	{"妻财", "占妻妾及我驱使之人,以妻财爻为用神(增删卜易·用神章)",
		[]string{"妻", "老婆", "女友", "女朋友", "未婚妻", "保姆", "员工", "下属"}},
	{"子孙", "占子女晚辈、门徒,以子孙爻为用神(增删卜易·用神章)",
		[]string{"儿子", "女儿", "孩子", "小孩", "子女", "孙", "侄", "外甥", "女婿", "徒弟", "学生"}},
	{"兄弟", "占兄弟姊妹及结义弟兄,以兄弟爻为用神(增删卜易·用神章)",
		[]string{"兄", "弟", "姐", "妹", "哥"}},
	// ── 事类 ──
	{"官鬼", "占功名官讼,一切拘束我身者以官鬼爻为用神(增删卜易·用神章)",
		[]string{"工作", "事业", "求职", "跳槽", "升职", "升迁", "面试", "考试", "考编", "考公", "功名", "官司", "诉讼", "官非", "仕途", "职位", "岗位"}},
	{"妻财", "占货财买卖,一切使用之财物以妻财爻为用神(增删卜易·用神章)",
		[]string{"财", "钱", "投资", "生意", "买卖", "交易", "利润", "收益", "工资", "奖金", "货", "债", "股", "房价", "售", "购"}},
	{"父母", "占宅舍舟车、文书契约,皆以父母爻为用神;占雨亦然(增删卜易·用神章/天时章)",
		[]string{"房", "宅", "屋", "车", "船", "合同", "契约", "文书", "文章", "论文", "证书", "签证", "学业", "雨"}},
	{"子孙", "占医药、僧道、六畜,以子孙爻为用神;占晴亦然(增删卜易·用神章/天时章)",
		[]string{"医", "药", "手术", "宠物", "猫", "狗", "六畜", "晴"}},
	{"官鬼", "占鬼祟盗贼,以官鬼爻为用神(增删卜易·用神章)",
		[]string{"失窃", "被盗", "小偷", "盗贼", "邪祟", "怪事"}},
}

// SuggestYongShen 按问题文本建议用神;无法归类则依「自占吉凶,以世爻为用神」。
func SuggestYongShen(question string) (name, basis string) {
	for _, rule := range yongShenRules {
		for _, kw := range rule.keywords {
			if strings.Contains(question, kw) {
				return rule.name, rule.basis
			}
		}
	}
	return "世爻", "自占吉凶,以世爻为用神(增删卜易)"
}

// qinShengBy 生我者(六亲相生环:父母→兄弟→子孙→妻财→官鬼→父母)。
// 元忌仇链即沿环逆推:生用者元、生元者忌(克用)、生忌者仇(克元)——
// 增删卜易·元神章「假令金为用神,生金者土,土为元神;克金者火,火为忌神;
// 克土生火者木,木为仇神」。
var qinSheng = map[string]string{"父母": "兄弟", "兄弟": "子孙", "子孙": "妻财", "妻财": "官鬼", "官鬼": "父母"}

func qinShengBy(qin string) string {
	for a, b := range qinSheng {
		if b == qin {
			return a
		}
	}
	return ""
}

// yongShenWhitelist 显式取用允许值(问者指明所占之人事)。
var yongShenWhitelist = map[string]string{
	"世爻": "问自己之事,以世爻为用",
	"妻财": "所占关乎妻室、财帛、雇仆,取妻财为用",
	"官鬼": "所占关乎官职、官司、丈夫、病祟,取官鬼为用",
	"父母": "所占关乎父母长辈、文书屋宅、师长,取父母为用",
	"子孙": "所占关乎子女晚辈、僧道六畜、解忧之神,取子孙为用",
	"兄弟": "所占关乎兄弟姊妹、朋友同辈、合伙,取兄弟为用",
}

// applyYongShen 填入用神建议、爻位与元忌仇链(用神不上卦则位置为空,伏神之法由解卦层论)。
// 问者显式指明取用(YongShenOverride)优先于问辞关键词推断——所占之人事以问者自陈为准。
func (r *Result) applyYongShen() {
	name, basis := SuggestYongShen(r.Question)
	if r.YongShenOverride != "" {
		if b, ok := yongShenWhitelist[r.YongShenOverride]; ok {
			name, basis = r.YongShenOverride, b+"(问者指明)"
		}
	}
	r.YongShen = name
	r.YongShenBasis = basis
	r.YongShenPos = []int{}

	// 元忌仇以用神六亲论;世爻用神按世爻所临六亲推链
	chainQin := name
	for _, y := range r.Yaos {
		if (name == "世爻" && y.IsShi) || (name != "世爻" && y.LiuQin == name) {
			r.YongShenPos = append(r.YongShenPos, y.Pos)
		}
		if name == "世爻" && y.IsShi {
			chainQin = y.LiuQin
		}
	}
	r.YuanShen = qinShengBy(chainQin)
	r.JiShen = qinShengBy(r.YuanShen)
	r.ChouShen = qinShengBy(r.JiShen)
	r.YuanShenPos, r.JiShenPos = []int{}, []int{}
	for _, y := range r.Yaos {
		if y.LiuQin == r.YuanShen {
			r.YuanShenPos = append(r.YuanShenPos, y.Pos)
		}
		if y.LiuQin == r.JiShen {
			r.JiShenPos = append(r.JiShenPos, y.Pos)
		}
	}
}
