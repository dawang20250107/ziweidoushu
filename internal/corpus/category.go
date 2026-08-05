// 板块归类:按书名/slug 启发式给语料分板块,供 AI 引用检索按术数门类过滤。
// 起因(实测教训):紫微深度报告曾引《梅花心易》脉诀 —— 全域检索不分门类,
// 他门口诀混入本门断语,内容失准。归类后各解读线只引本门 + 通用语料。
package corpus

import "strings"

// 板块常量:显式 category 字段与启发式归类共用此口径。
const (
	CatZiwei    = "ziwei"    // 紫微斗数
	CatMeihua   = "meihua"   // 梅花易数
	CatLiuYao   = "liuyao"   // 六爻纳甲
	CatLiuRen   = "liuren"   // 大六壬
	CatZhouYi   = "zhouyi"   // 周易经传(梅花/六爻共用)
	CatBazi     = "bazi"     // 四柱子平
	CatXiang    = "xiang"    // 相法(紫微形神层引用)
	CatFengshui = "fengshui" // 堪舆(暂无解读线引用)
	CatQimen    = "qimen"    // 奇门(暂无解读线引用)
	CatNi       = "ni"       // 倪师通论(天纪/人间道等,各线通用)
	CatMisc     = "misc"     // 未归类杂项(不被任何解读线引用)
)

// categoryRules 首中即止;主题类目在前、倪师通论殿后:
// 「倪海厦紫微斗数」应归 ziwei 而非 ni。
var categoryRules = []struct {
	cat      string
	keywords []string
}{
	{CatZiwei, []string{"紫微", "斗数", "斗數", "骨髓赋", "形性赋", "太微赋"}},
	{CatMeihua, []string{"梅花", "心易", "观梅"}},
	{CatLiuYao, []string{"六爻", "卜筮", "卜易", "增删", "纳甲", "火珠林", "易冒", "易隐"}},
	{CatLiuRen, []string{"六壬", "壬归", "壬学", "课经", "毕法"}},
	{CatQimen, []string{"奇门"}},
	{CatBazi, []string{"八字", "子平", "四柱", "滴天", "穷通", "渊海"}},
	{CatXiang, []string{"相法", "麻衣", "冰鉴", "公笃"}},
	{CatFengshui, []string{"风水", "地理", "阳宅", "撼龙", "堪舆", "罗盘", "地脉", "地纪"}},
	{CatZhouYi, []string{"周易", "易经", "易傳", "易传", "伊川"}},
	{CatNi, []string{"倪海厦", "倪师", "天纪", "人间道", "天机道"}},
}

// bookCategory 取书的板块:显式 category 字段优先,否则按书名+slug 启发式,兜底 misc。
func bookCategory(b *Book) string {
	if b.Category != "" {
		return b.Category
	}
	t := b.Title + " " + b.Slug
	for _, rule := range categoryRules {
		for _, kw := range rule.keywords {
			if strings.Contains(t, kw) {
				return rule.cat
			}
		}
	}
	return CatMisc
}
