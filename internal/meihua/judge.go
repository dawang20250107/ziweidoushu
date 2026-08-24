// 断卦层(《梅花易数·体用总诀》口径,确定性推演):
//
//	①卦气旺衰:体卦五行对月令(正二月木旺、四五月火旺、七八月金旺、
//	  十月十一月水旺、三六九十二月土旺),旺相休囚死定体之根基;
//	②体党用党:遍数互变四卦生体/比和/克体之势;用克体时,生用之卦
//	  为助攻亦记不利(观梅占「互中巽木复生起离火,克体之卦气盛」之义);
//	③互卦断过程、变卦断结局(变卦取动爻所在之侧的新卦为事之归结);
//	④十八占事类口径:按所问事类给专断,不识别则以谋事通断;
//	⑤应期:吉应生体比和之气当令时节,凶防克体之气得令,数应参卦数。
//
// 正确性锚点:观梅占(用克体而体党有救 → 小损不危)、牡丹占(克体众而
// 无生意 → 尽毁)两占例回归钉死,推演结论与原书断语同向。
package meihua

import (
	"fmt"
	"strings"
)

// Judgment 确定性断卦结论(供盘面直出与 AI 贴卦发挥)。
type Judgment struct {
	Level string `json:"level"` // good / neutral / caution(与六爻同口径)
	Score int    `json:"score"`
	// Sections 分节深断(卦象总论/体用之辨/过程与结局/类象取应/应期)。
	Sections   []JudgeSection `json:"sections,omitempty"`
	TiQi       string         `json:"tiQi"` // 体卦月令旺衰:旺/相/休/囚/死
	Conclusion string         `json:"conclusion"`
	Points     []string       `json:"points"`
	Topic      string         `json:"topic"`
	TopicNote  string         `json:"topicNote,omitempty"`
	YingQi     string         `json:"yingQi"`
}

// wangElementByMonth 农历月 → 当令五行(梅花口径:三六九十二月土旺)。
func wangElementByMonth(month int) string {
	switch month {
	case 1, 2:
		return "木"
	case 4, 5:
		return "火"
	case 7, 8:
		return "金"
	case 10, 11:
		return "水"
	default: // 3 6 9 12
		return "土"
	}
}

// qiState 体卦五行对当令五行 → 旺相休囚死与力度分。
func qiState(ti, wang string) (string, int) {
	switch {
	case ti == wang:
		return "旺", 2
	case sheng(wang, ti):
		return "相", 1
	case sheng(ti, wang):
		return "休", 0
	case ke(ti, wang):
		return "囚", -1
	default: // 当令克体
		return "死", -2
	}
}

// baseScore 体用生克基础分。
var baseScore = map[Relation]int{
	RelYongShengTi: 2, RelBiHe: 1, RelTiKeYong: 1, RelTiShengYong: -1, RelYongKeTi: -2,
}

// relateToTi 某卦对体:生体/比和/克体/泄体/耗体(体生之为泄、体克之为耗)。
func relateToTi(x, ti string) string {
	switch {
	case x == ti:
		return "比和"
	case sheng(x, ti):
		return "生体"
	case ke(x, ti):
		return "克体"
	case sheng(ti, x):
		return "泄体"
	default:
		return "耗体"
	}
}

// seasonOf 五行 → 当令时节(应期用)。
var seasonOf = map[string]string{
	"木": "春(寅卯月日)", "火": "夏(巳午月日)", "金": "秋(申酉月日)",
	"水": "冬(亥子月日)", "土": "四季土月(辰戌丑未月日)",
}

// shengSrc 生我者;keSrc 克我者。
var shengSrc = map[string]string{"金": "土", "木": "水", "水": "金", "火": "木", "土": "火"}
var keSrc = map[string]string{"金": "火", "木": "金", "水": "土", "火": "水", "土": "木"}

// topicRules 事类识别(问题关键词 → 十八占类目)。
var topicRules = []struct {
	topic string
	kws   []string
}{
	{"求财", []string{"财", "钱", "投资", "生意", "收益", "盈利", "买卖", "股", "债", "薪"}},
	{"婚恋", []string{"婚", "恋", "感情", "对象", "相亲", "复合", "分手", "喜欢", "脱单"}},
	{"疾病", []string{"病", "健康", "手术", "身体", "医院", "体检", "康复"}},
	{"官讼", []string{"官司", "诉讼", "仲裁", "纠纷", "起诉", "官非"}},
	{"求名", []string{"考试", "考研", "面试", "升职", "晋升", "竞聘", "录取", "评职", "求职", "offer"}},
	{"出行", []string{"出行", "旅行", "出差", "搬家", "迁", "出国", "远行"}},
	{"行人", []string{"归来", "回来", "失联", "音信", "消息", "何时回"}},
	{"失物", []string{"丢", "失物", "遗失", "找回", "被盗", "不见"}},
	{"家宅", []string{"家宅", "房", "住", "装修", "风水", "邻里"}},
}

func inferTopic(question string) string {
	for _, r := range topicRules {
		for _, kw := range r.kws {
			if strings.Contains(question, kw) {
				return r.topic
			}
		}
	}
	return "谋事"
}

// topicNotes 十八占义:事类 × 体用关系 → 专断(原创白话,依《梅花易数》
// 卷二分占之义归纳,非原文摘录)。
var topicNotes = map[string]map[Relation]string{
	"求财": {
		RelYongShengTi: "用生体,财自外来、得之顺遂,可进",
		RelBiHe:        "比和,财路通畅、同侪相助,快意之财",
		RelTiKeYong:    "体克用,财可得而须自取,勤经营乃有成",
		RelTiShengYong: "体生用,财源外泄、多支出少进项,宜守不宜攻",
		RelYongKeTi:    "用克体,财上受制、防破财倒账,不宜举债投机",
	},
	"婚恋": {
		RelYongShengTi: "用生体,对方有情有益于我,婚恋易成",
		RelBiHe:        "比和,两情相合、门户相当,吉配",
		RelTiKeYong:    "体克用,可成而我方主动费力,宜诚不宜迫",
		RelTiShengYong: "体生用,我付出多而回应少,勉强则倦",
		RelYongKeTi:    "用克体,对方强势或外阻挠,难成、成亦受制",
	},
	"疾病": {
		RelYongShengTi: "用生体,医药得力、调养见效,渐安",
		RelBiHe:        "比和,病势平缓,将息可复",
		RelTiKeYong:    "体克用,正气足以胜邪,愈期不远",
		RelTiShengYong: "体生用,元气暗耗、病势缠绵,重在养",
		RelYongKeTi:    "用克体,邪势克身、病重防变,速就良医",
	},
	"官讼": {
		RelYongShengTi: "用生体,讼得理助、对方让步,于我有利",
		RelBiHe:        "比和,宜和解息讼,两平收场",
		RelTiKeYong:    "体克用,理在我方、可胜,然费时费力",
		RelTiShengYong: "体生用,徒耗财力精神,不如早了",
		RelYongKeTi:    "用克体,势不在我、防受屈,慎讼避讼",
	},
	"求名": {
		RelYongShengTi: "用生体,有提携之力、名位可就",
		RelBiHe:        "比和,同侪并进、水到渠成",
		RelTiKeYong:    "体克用,力争可得,主动出击",
		RelTiShengYong: "体生用,虚耗奔忙、名难副实,调整路径",
		RelYongKeTi:    "用克体,压制在上、时机未至,养实力待时",
	},
	"出行": {
		RelYongShengTi: "用生体,出行得助、所至有获,可行",
		RelBiHe:        "比和,一路平顺",
		RelTiKeYong:    "体克用,可行,所往之地我为主动",
		RelTiShengYong: "体生用,劳顿破费、收获平平,非必要缓行",
		RelYongKeTi:    "用克体,途有阻碍、防意外损失,改期为宜",
	},
	"行人": {
		RelYongShengTi: "用生体,行人将至、且携喜讯",
		RelBiHe:        "比和,归期不日",
		RelTiKeYong:    "体克用,归迟,催促可动",
		RelTiShengYong: "体生用,行人在外滞留,音信渐通",
		RelYongKeTi:    "用克体,行人有阻,未宜远盼,先通消息",
	},
	"失物": {
		RelYongShengTi: "用生体,失物可见、多在近处",
		RelBiHe:        "比和,旧处寻之可得",
		RelTiKeYong:    "体克用,可寻回,须费周折",
		RelTiShengYong: "体生用,物已转手他处,寻回费力",
		RelYongKeTi:    "用克体,恐为人所据,难复得",
	},
	"家宅": {
		RelYongShengTi: "用生体,宅气生身、居之安泰",
		RelBiHe:        "比和,家宅平和,守成即吉",
		RelTiKeYong:    "体克用,宅事由我做主,修整有利",
		RelTiShengYong: "体生用,宅耗人财,量入为出",
		RelYongKeTi:    "用克体,宅有不安之处,宜察修缮、缓置产",
	},
}

// Judge 按月令与所问事类作确定性断卦。month 为农历月(1-12,闰月按当月)。
func (r *Result) Judge(month int, question string) *Judgment {
	ti := r.TiTrigram.Element
	yong := r.YongTrigram.Element
	j := &Judgment{Topic: inferTopic(question)}

	// ① 体气旺衰
	wang := wangElementByMonth(month)
	state, qiScore := qiState(ti, wang)
	j.TiQi = state
	score := baseScore[r.Relation] + qiScore
	j.Points = append(j.Points, fmt.Sprintf("体卦%s属%s,月令%s当权,体气【%s】;%s",
		r.TiTrigram.Name, ti, wang, state, r.Verdict))

	// ② 体党用党:互变四卦向背(用克体时,生用之卦助攻亦记不利)
	others := []Trigram{r.Hu.Upper, r.Hu.Lower, r.Bian.Upper, r.Bian.Lower}
	helpers, harmers := []string{}, []string{}
	for _, x := range others {
		switch relateToTi(x.Element, ti) {
		case "生体", "比和":
			score++
			helpers = append(helpers, x.Name)
		case "克体":
			score--
			harmers = append(harmers, x.Name)
		default:
			if r.Relation == RelYongKeTi && sheng(x.Element, yong) {
				score--
				harmers = append(harmers, x.Name+"(生起克体之"+r.YongTrigram.Name+")")
			}
		}
	}
	party := "互变之中"
	if len(helpers) > 0 {
		party += fmt.Sprintf(",生扶比和体卦者:%s", strings.Join(helpers, "、"))
	}
	if len(harmers) > 0 {
		party += fmt.Sprintf(",克伐体卦之势:%s", strings.Join(harmers, "、"))
	}
	if len(helpers) == 0 && len(harmers) == 0 {
		party += ",于体无甚生克,吉凶专看体用与卦气"
	}
	j.Points = append(j.Points, party+"。")

	// ③ 互卦过程、变卦结局
	huRel := relateToTi(r.Hu.Upper.Element, ti)
	huRel2 := relateToTi(r.Hu.Lower.Element, ti)
	j.Points = append(j.Points, fmt.Sprintf("过程(互卦%s):上互%s%s、下互%s%s——中段之助阻在此。",
		r.Hu.Name, r.Hu.Upper.Name, huRel, r.Hu.Lower.Name, huRel2))
	bianSide := r.Bian.Lower
	if !r.TiIsUpper { // 体在下,动在上 → 变卦看上
		bianSide = r.Bian.Upper
	}
	endRel := relateToTi(bianSide.Element, ti)
	endText := map[string]string{
		"生体": "终得成全、渐入佳境", "比和": "终局平顺如意",
		"克体": "结尾防变故反复,须留后手", "泄体": "末段费神耗力,收束宜早", "耗体": "收尾平平,及时知止",
	}[endRel]
	j.Points = append(j.Points, fmt.Sprintf("结局(变卦%s):动侧化出%s,%s——%s。",
		r.Bian.Name, bianSide.Name, endRel, endText))

	// ④ 事类专断
	if notes, ok := topicNotes[j.Topic]; ok {
		j.TopicNote = notes[r.Relation]
		j.Points = append(j.Points, fmt.Sprintf("所问【%s】:%s。", j.Topic, j.TopicNote))
	}

	// ⑤ 应期
	sumNum := r.Ben.Upper.Num + r.Ben.Lower.Num + r.Moving
	shengTi := shengSrc[ti]
	if score >= 0 {
		j.YingQi = fmt.Sprintf("应期:吉应看生体之%s气当令——%s或体气(%s)得时;数应参卦数 %d(近事以日计、远事以月计)。",
			shengTi, seasonOf[shengTi], seasonOf[ti], sumNum)
	} else {
		j.YingQi = fmt.Sprintf("应期:所防在克体之%s气得令——%s;过此气衰则缓;数应参卦数 %d(近事以日计、远事以月计)。",
			keSrc[ti], seasonOf[keSrc[ti]], sumNum)
	}

	// 综断
	j.Score = score
	switch {
	case score >= 2:
		j.Level = "good"
		j.Conclusion = "体得生扶、其势盛,所谋可成,顺势而为。"
	case score <= -3:
		j.Level = "caution"
		j.Conclusion = "克体之势众而生意薄,所谋多阻,宜守宜避,勿强图。"
	default:
		j.Level = "neutral"
		if r.Relation == RelYongKeTi {
			j.Conclusion = "用克体本为不利,幸盘中体党有力、凶中有救——事或有小损而不至大败,谨慎可过。"
		} else {
			j.Conclusion = "体用之势相持,成败参半,谋事在人,把握应期为要。"
		}
	}
	r.buildSections(j)
	return j
}
