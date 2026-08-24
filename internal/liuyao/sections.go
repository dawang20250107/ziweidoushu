// 六爻分节深断层:取用/旺衰/元忌/动变/世应/逐爻/应期(确定性,不走 LLM)。
// 义理依《增删卜易》《卜筮正宗》通行断法归纳、原创行文;只做呈现纵深,
// 不改 judgeCore 的任何推演与吉凶评分(占例批回归口径不动)。
package liuyao

import (
	"fmt"
	"strings"
)

// JudgeSection 断语分节(免费确定性层的呈现单元)。
type JudgeSection struct {
	Key   string `json:"key"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

// Judge 断语入口:推演本体 + 分节深断。
func (r *Result) Judge() *Judgment {
	j := r.judgeCore()
	r.buildSections(j)
	return j
}

// monthStateDeep 月建旺衰义理。
var monthStateDeep = map[string]string{
	"旺": "当月建之气,如树当春,根基最牢——纵遇刑伤亦有抵挡",
	"相": "受月建所生,气势方长,得时之助",
	"休": "生月建而自泄,气息偏弱,须赖日辰动爻扶持",
	"囚": "克月建而自困,力有不逮",
	"死": "受月建之克,如草遇霜,全赖他处生扶方能起用",
}

// dayRelDeep 日辰作用义理(日辰为六爻之主宰)。
var dayRelDeep = map[string]string{
	"临": "日辰亲临,即时之力最盛",
	"冲": "日辰冲之,旺者冲动、衰者冲散",
	"合": "日辰合之,合者绊也——吉者合牢,凶者亦被缠住",
	"扶": "日辰同气相扶,添一分底气",
	"生": "日辰生之,如逢甘露",
	"克": "日辰克之,当下受制",
	"泄": "生日辰而自泄,力有外耗",
	"耗": "克日辰而自耗,费力少功",
}

// bianRelDeep 动变作用义理(动爻之变,吉凶自变中来)。
var bianRelDeep = []struct{ key, text string }{
	{"化进神", "化进神——愈变愈进,其势方张,吉者吉进、凶者凶进"},
	{"化退神", "化退神——愈变愈退,其势渐消,凶者可解、吉者难久"},
	{"回头生", "变爻回头生——自变中得援,如困而遇友,其力可恃"},
	{"回头克", "变爻回头克——自变中生患,动而自伤,最须提防"},
	{"化长生", "化长生——变而植根,后劲绵长,事从此起"},
	{"化帝旺", "化帝旺——变至极盛,一时风头无两,盛极亦当思守"},
	{"化墓", "化入墓——动而入墓,如人入库,须待冲墓之日方出"},
	{"化绝", "化入绝——变至绝地,其力将尽,非重生之日难起"},
	{"化合", "化合——变而成合,事有缔结缠绊之象"},
	{"伏吟", "化伏吟——动而不变、呻吟不前,进退两难之象"},
	{"反吟", "化反吟——变而对冲,反复颠倒,事多往返"},
}

// yongYao 取用神爻(首位);无则 nil。
func (r *Result) yongYao() *Yao {
	if len(r.YongShenPos) == 0 {
		return nil
	}
	p := r.YongShenPos[0]
	if p < 1 || p > 6 {
		return nil
	}
	return &r.Yaos[p-1]
}

func posList(pos []int) string {
	if len(pos) == 0 {
		return ""
	}
	parts := make([]string, len(pos))
	for i, p := range pos {
		parts[i] = fmt.Sprintf("%d", p)
	}
	return "第 " + strings.Join(parts, "、") + " 爻"
}

// yaoBrief 一爻的白话速览(逐爻细览用)。
func yaoBrief(y *Yao) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s%s %s%s(%s·%s", y.LiuShen, y.LiuQin, y.Stem, y.Branch, y.Element, y.MonthState))
	if y.DayRelation != "" {
		b.WriteString("、日" + y.DayRelation)
	}
	b.WriteString(")")
	var marks []string
	if y.IsShi {
		marks = append(marks, "世")
	}
	if y.IsYing {
		marks = append(marks, "应")
	}
	if y.YuePo {
		marks = append(marks, "月破")
	}
	if y.XunKong {
		marks = append(marks, "旬空")
	}
	if y.AnDong {
		marks = append(marks, "暗动")
	}
	if y.RiPo {
		marks = append(marks, "日破")
	}
	if len(marks) > 0 {
		b.WriteString("〔" + strings.Join(marks, "·") + "〕")
	}
	if y.Moving {
		b.WriteString("动")
		if y.BianYao != nil {
			b.WriteString(fmt.Sprintf("→变%s%s", y.BianYao.LiuQin, y.BianYao.Branch))
		}
		if y.BianRelation != "" {
			b.WriteString("(" + y.BianRelation + ")")
		}
	}
	return b.String()
}

// buildSections 组装分节深断。
func (r *Result) buildSections(j *Judgment) {
	// ── 取用神 ──
	var qy strings.Builder
	qy.WriteString(fmt.Sprintf("凡占必先取用神。所测之事以【%s】为用", r.YongShen))
	if len(r.YongShenPos) > 0 {
		qy.WriteString("(" + posList(r.YongShenPos) + ")")
	}
	if r.YongShenBasis != "" {
		qy.WriteString("——" + r.YongShenBasis + "。")
	} else {
		qy.WriteString("。")
	}
	if fs := r.FuShen; fs != nil {
		qy.WriteString(fmt.Sprintf("用神不上本卦,取伏神:%s%s伏于第 %d 爻飞神%s之下,%s;出伏之期在%s。",
			fs.LiuQin, fs.Branch, fs.Pos, fs.Fei, fs.Note, fs.ChuFuRi))
	}
	j.Sections = append(j.Sections, JudgeSection{Key: "yongshen", Title: "取用神", Text: qy.String()})

	// ── 用神旺衰(月建定根基、日辰主当下) ──
	if y := r.yongYao(); y != nil {
		var ws strings.Builder
		ws.WriteString(fmt.Sprintf("用神%s%s属%s。月建%s之下其气【%s】——%s。",
			y.Stem, y.Branch, y.Element, r.MonthJian, y.MonthState, monthStateDeep[y.MonthState]))
		if y.DayRelation != "" {
			ws.WriteString(fmt.Sprintf("日辰%s%s对用神为「%s」:%s。", r.DayStem, r.DayBranch, y.DayRelation, dayRelDeep[y.DayRelation]))
		}
		if y.DayStage != "" {
			ws.WriteString(fmt.Sprintf("对日辰居【%s】之地(野鹤口径)。", y.DayStage))
		}
		if y.YuePo {
			ws.WriteString("用神月破——破者散也,月内难成,出月填合之日方可再论。")
		}
		if y.XunKong {
			ws.WriteString("用神旬空——空非无用,待出空、冲空实空之日,其事乃应。")
		}
		if y.AnDong {
			ws.WriteString("用神暗动——静而旺相逢日冲,是暗中已动,事在悄然推进。")
		}
		j.Sections = append(j.Sections, JudgeSection{Key: "wangshuai", Title: "用神旺衰", Text: ws.String()})
	}

	// ── 元忌之势 ──
	if r.YuanShen != "" {
		var yj strings.Builder
		yj.WriteString(fmt.Sprintf("生用者为元神【%s】%s,克用者为忌神【%s】%s,生忌克元者为仇神【%s】。",
			r.YuanShen, posName(r.YuanShenPos), r.JiShen, posName(r.JiShenPos), r.ChouShen))
		yj.WriteString(fmt.Sprintf("按《增删卜易》有力/无力条目机械对照:元神%s;忌神%s。",
			PowerText(r.YuanShenPower), PowerText(r.JiShenPower)))
		yj.WriteString("总贵用神有气:用神无根则元神有力亦难生,忌神无力亦休喜——元忌之势须合用神旺衰同参。")
		j.Sections = append(j.Sections, JudgeSection{Key: "yuanji", Title: "元忌之势", Text: yj.String()})
	}

	// ── 动变作用 ──
	var db strings.Builder
	for i := range r.Yaos {
		y := &r.Yaos[i]
		if !y.Moving {
			continue
		}
		db.WriteString(fmt.Sprintf("第 %d 爻%s%s发动", y.Pos, y.LiuQin, y.Branch))
		if y.BianYao != nil {
			db.WriteString(fmt.Sprintf(",化出%s%s", y.BianYao.LiuQin, y.BianYao.Branch))
		}
		matched := false
		for _, br := range bianRelDeep {
			if y.BianRelation != "" && strings.Contains(y.BianRelation, br.key) {
				db.WriteString(":" + br.text + "。")
				matched = true
				break
			}
		}
		if !matched {
			db.WriteString("。")
		}
	}
	if db.Len() == 0 {
		db.WriteString("六爻安静,无一爻发动——静卦以用神旺衰日月生克论之,事势平缓、少突变;安静而用神有气,是不动声色中自有定数。")
	}
	j.Sections = append(j.Sections, JudgeSection{Key: "dongbian", Title: "动变作用", Text: db.String()})

	// ── 世应格局(世为自己、应为彼端) ──
	var shi, ying *Yao
	for i := range r.Yaos {
		if r.Yaos[i].IsShi {
			shi = &r.Yaos[i]
		}
		if r.Yaos[i].IsYing {
			ying = &r.Yaos[i]
		}
	}
	if shi != nil && ying != nil {
		var sy strings.Builder
		sy.WriteString(fmt.Sprintf("世爻%s%s(%s·%s)为自己之位,应爻%s%s(%s·%s)为彼端之位。",
			shi.LiuQin, shi.Branch, shi.Element, shi.MonthState, ying.LiuQin, ying.Branch, ying.Element, ying.MonthState))
		rel := elementRelation(shi.Element, ying.Element)
		sy.WriteString("世应之间" + rel + "。")
		if shi.XunKong {
			sy.WriteString("世爻旬空——自心未定、己方有虚。")
		}
		if ying.XunKong {
			sy.WriteString("应爻旬空——彼意难凭、对方存虚。")
		}
		j.Sections = append(j.Sections, JudgeSection{Key: "shiying", Title: "世应格局", Text: sy.String()})
	}

	// ── 逐爻细览(自上而下,爻名依传统:初/二/三/四/五/上) ──
	yaoNames := []string{"初", "二", "三", "四", "五", "上"}
	var zy []string
	for i := 5; i >= 0; i-- {
		zy = append(zy, yaoNames[i]+"爻 "+yaoBrief(&r.Yaos[i]))
	}
	j.Sections = append(j.Sections, JudgeSection{Key: "zhuyao", Title: "逐爻细览", Text: strings.Join(zy, "\n")})

	// ── 应期 ──
	if j.YingQi != "" {
		j.Sections = append(j.Sections, JudgeSection{Key: "yingqi", Title: "应期",
			Text: j.YingQi + "。应期之法:旬空者候出空冲空,月破者候出月填合,入墓者候冲墓,伏藏者候出伏——皆以卦中标注为准。"})
	}
}

// posName 爻位串(空则「不上卦」)。
func posName(pos []int) string {
	if len(pos) == 0 {
		return "(不上卦)"
	}
	return "(" + posList(pos) + ")"
}

// elementRelation 世应五行关系白话。
func elementRelation(a, b string) string {
	idx := map[string]int{"木": 0, "火": 1, "土": 2, "金": 3, "水": 4}
	ai, aok := idx[a]
	bi, bok := idx[b]
	if !aok || !bok {
		return "各安其位"
	}
	switch {
	case ai == bi:
		return "比和——彼此同气,商量得通"
	case (ai+1)%5 == bi:
		return "世生应——我就彼势,主动在我、耗力亦在我"
	case (bi+1)%5 == ai:
		return "应生世——彼来就我,得对方之助"
	case (ai+2)%5 == bi:
		return "世克应——我制彼端,事可掌握"
	default:
		return "应克世——彼势压我,谋之费力"
	}
}
