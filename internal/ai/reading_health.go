package ai

// 疾厄宫健康专层:主星 → 身体部位/病症系统,会照煞 → 具体病厄。
//
// 紫微斗数疾厄宫按星断健康有成熟公版通则:太阳主心血压眼目、武曲主呼吸鼻、
// 巨门主口腔肠胃暗疾……煞星再指其应验(擎羊主刀刃外伤、火星主炎症发热)。
// 本层据此给疾厄宫专属健康断语。内容为公版通则的原创提炼(非第三方文本)。
// 注:命理健康提示仅供养生参考,不构成医疗诊断,身体不适应就医。

import (
	"fmt"
	"strings"

	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// starHealth 主星在疾厄宫主司的身体部位/系统与养生提示。
var starHealth = map[string]string{
	"紫微": "脾胃与头部,宜节饮食、防血压脑压偏高",
	"天机": "肝胆、神经与四肢,多思伤神、宜防失眠与筋骨劳损",
	"太阳": "心脏、血压与眼目(尤应男性亲长),宜防心血管与目疾、忌操劳上火",
	"武曲": "呼吸系统、鼻、大肠与牙齿,宜防外感、金属外伤",
	"天同": "膀胱、内分泌与耳,情绪化易致水肿、肥胖,宜疏解情志",
	"廉贞": "血液、心火与妇科泌尿,易有暗疾、脓血之症,宜清心解郁",
	"天府": "脾胃与消化,宜防饮食积滞、慢性肠胃",
	"太阴": "阴分、内分泌与妇科、眼目(尤应女性亲长),宜防失眠、暗耗气血",
	"贪狼": "肝胆、性腺与内分泌,宜节欲戒酒、防纵欲伤身",
	"巨门": "口腔、食道与肠胃,多暗疾潜伏,宜早查、慎口腹",
	"天相": "皮肤、泌尿与膀胱,体质大致平和,宜规律作息",
	"天梁": "脾胃与乳部;天梁荫星主逢凶化吉、大病每有转机,仍宜养生防慢病",
	"七杀": "肺、呼吸与筋骨,易有外伤、急症,宜防意外与开刀之厄",
	"破军": "生殖泌尿与消耗性疾病,易有外伤,宜防因劳因欲损身",
}

// shaHealthOrder 煞星健康应验(固定顺序,保证断语稳定)。
var shaHealthOrder = []struct{ name, eff string }{
	{"擎羊", "刀刃金创、手术、头面外伤"},
	{"陀罗", "牙齿、暗疾与旧疾拖延、跌打损伤"},
	{"火星", "炎症、发热、烫伤与皮肤之疾"},
	{"铃星", "神经、皮肤与隐疾"},
	{"地空", "虚症、查无实据之不适"},
	{"地劫", "耗损性病症与意外之伤"},
}

// healthClause 疾厄宫专属健康断语:本宫主星部位 + 会照煞病厄 + 庙陷体质。
func healthClause(chart *ziwei.Chart, p *ziwei.Palace, majors []string, bright, dim map[string]bool) string {
	var b strings.Builder
	var parts []string
	for _, m := range majors {
		if h, ok := starHealth[m]; ok {
			parts = append(parts, m+"主"+h)
		}
	}
	if len(parts) > 0 {
		b.WriteString("健康专断——" + strings.Join(parts, ";") + "。")
	}
	// 庙陷定体质强弱。
	strong, weak := 0, 0
	for _, m := range majors {
		if bright[m] {
			strong++
		} else if dim[m] {
			weak++
		}
	}
	if strong > 0 && weak == 0 {
		b.WriteString("主星庙旺,先天体质较佳、病亦易愈。")
	} else if weak > 0 && strong == 0 {
		b.WriteString("主星落陷,体质偏弱、宜积极调养、勿讳疾忌医。")
	}
	// 会照煞星指其应验。
	set := triadStarSet(chart, p.Branch)
	var shaParts []string
	for _, s := range shaHealthOrder {
		if set[s.name] {
			shaParts = append(shaParts, s.name+"防"+s.eff)
		}
	}
	if len(shaParts) > 0 {
		b.WriteString("会煞星:" + strings.Join(shaParts, ";") + ",宜预防、定期检查。")
	}
	if b.Len() == 0 {
		return ""
	}
	return b.String() + "(命理健康提示仅供养生参考,不代医疗诊断,不适应就医)"
}

// healthOverview 供合盘/报告引用的一句话健康概览(疾厄宫主星部位)。
func healthOverview(chart *ziwei.Chart) string {
	p := chart.PalaceByName("疾厄")
	if p == nil {
		return ""
	}
	majors, _ := palaceMajors(p)
	var parts []string
	for _, m := range majors {
		if h, ok := starHealth[m]; ok {
			parts = append(parts, fmt.Sprintf("%s(%s)", m, firstSentence(h)))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "健康须留意:" + strings.Join(parts, "、")
}
