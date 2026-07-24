package ai

// 应验分档:双星/主星组合再叠煞吉会照,从「格性」推进到「应验档位」。
//
// 同一组合遇不同煞吉,吉凶天差地别:廉杀本刚烈,加羊陀则应「路上埋尸」血光档;
// 贪狼逢火铃则成暴发格。此层以本宫 + 三方四正的煞吉为触发条件给出应验断语,
// 并区分「同宫」(力足、应验重)与「会照」(力减、应验缓)——同宫之煞贴身,
// 三方之煞隔位,古法轻重有别。内容为三合派公版通则的原创提炼(非第三方文本)。

import (
	"fmt"

	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// palaceStarSet 本宫全部星曜名集合。
func palaceStarSet(p *ziwei.Palace) map[string]bool {
	set := map[string]bool{}
	for i := range p.Stars {
		set[p.Stars[i].Name] = true
	}
	return set
}

// triadStarSet 本宫 + 三方四正(对宫+三合)全部星曜名集合,用于会照判定。
func triadStarSet(chart *ziwei.Chart, branch int) map[string]bool {
	set := map[string]bool{}
	for _, off := range []int{0, 4, 6, 8} {
		if p := chart.PalaceByBranch(branch + off); p != nil {
			for i := range p.Stars {
				set[p.Stars[i].Name] = true
			}
		}
	}
	return set
}

func hasMajor(majors []string, name string) bool {
	for _, m := range majors {
		if m == name {
			return true
		}
	}
	return false
}

// hasAnyMajor 主星列表是否含指定之一。
func hasAnyMajor(majors []string, names ...string) bool {
	for _, n := range names {
		if hasMajor(majors, n) {
			return true
		}
	}
	return false
}

// aspect 煞星相对本宫的方位强弱:同宫(贴身,力足)优先于会照(隔位,力减)。
type aspect struct {
	present bool
	inHouse bool // true=同宫;false=仅三方四正会照
}

// aspectOf 判定 names 中任一星相对本宫的方位:先看同宫,再看会照。
func aspectOf(local, triad map[string]bool, names ...string) aspect {
	for _, n := range names {
		if local[n] {
			return aspect{present: true, inHouse: true}
		}
	}
	for _, n := range names {
		if triad[n] {
			return aspect{present: true, inHouse: false}
		}
	}
	return aspect{}
}

// word 方位说法;heavy 力足档时的分值权重(同宫重、会照轻)。
func (a aspect) word() string {
	if a.inHouse {
		return "同宫"
	}
	return "会照"
}

// escalations 由本宫主星 + 煞吉方位推出应验档断语与吉凶增减。
// dim 为落陷主星集合,sihua 为主星四化(均来自 scanPalace)。
func escalations(chart *ziwei.Chart, p *ziwei.Palace, majors []string, dim map[string]bool, sihua map[string]ziwei.SiHua) (clauses []string, scoreDelta int) {
	local := palaceStarSet(p)
	triad := triadStarSet(chart, p.Branch)

	// 1. 贪狼逢火铃:暴发格(同宫成正格力足,会照力缓)。
	if hasMajor(majors, "贪狼") {
		if a := aspectOf(local, triad, "火星", "铃星"); a.present {
			if a.inHouse {
				clauses = append(clauses, "【应验·暴发格】贪狼同宫火/铃,成火贪、铃贪正格:主突发横财、机遇骤至,暴发力足;然骤起亦骤落,宜见好即收")
				scoreDelta += 2
			} else {
				clauses = append(clauses, "【应验·暴发格】贪狼会照火/铃,暴发之兆(隔位力缓):偶有骤得之机,然不如同宫之烈,宜稳中待时")
				scoreDelta++
			}
		}
	}

	// 2. 廉杀逢羊陀:「路上埋尸」血光档(同宫最烈,独占)。
	lianSha := hasMajor(majors, "廉贞") && hasMajor(majors, "七杀")
	yt := aspectOf(local, triad, "擎羊", "陀罗")
	hl := aspectOf(local, triad, "火星", "铃星")
	if lianSha && yt.present {
		if yt.inHouse {
			clauses = append(clauses, "【应验·血光档】廉杀同宫羊陀,古云『路上埋尸』应验重:主意外、血光、伤灾,尤忌驾驶与高危,宜信仰行善、慎行以化")
			scoreDelta -= 3
		} else {
			clauses = append(clauses, "【应验·血光档】廉杀会照羊陀,『路上埋尸』之虞(隔位稍减):仍防意外血光、伤灾,忌高危、宜慎行")
			scoreDelta -= 2
		}
	} else if hasMajor(majors, "七杀") || hasMajor(majors, "破军") {
		// 3. 杀破逢煞:冲击破耗加剧(廉杀羊陀已独占则不叠)。
		if yt.present || hl.present {
			a := yt
			if !a.present {
				a = hl
			}
			extra, sd := "力足", -2
			if !a.inHouse {
				extra, sd = "隔位力缓", -1
			}
			clauses = append(clauses, fmt.Sprintf("【应验·冲击档】杀破%s羊陀火铃(%s),冲击破耗加剧:成败起伏大、防伤灾与耗财,宜以稳制动、忌逞强", a.word(), extra))
			scoreDelta += sd
		}
	}

	// 4. 巨门逢煞忌:是非官非加剧。
	if hasMajor(majors, "巨门") {
		a := aspectOf(local, triad, "擎羊", "陀罗")
		if a.present || sihua["巨门"] == ziwei.HuaJi {
			trig := "羊陀" + a.word()
			if !a.present {
				trig = "化忌"
			}
			clauses = append(clauses, fmt.Sprintf("【应验·是非档】巨门逢%s,是非口舌、官非构陷加剧:慎签约、防小人,宜守口修心", trig))
			scoreDelta--
		}
	}

	// 5. 空劫夹会财动之星:财空成空(同宫最忌)。
	if a := aspectOf(local, triad, "地空", "地劫"); a.present && hasAnyMajor(majors, "武曲", "贪狼", "七杀", "破军", "天府", "太阴") {
		clauses = append(clauses, fmt.Sprintf("【应验·成空档】财动之星%s地空地劫,谋财求进易付诸流水:宜脚踏实地、忌投机豪赌与合伙重财", a.word()))
		if a.inHouse {
			scoreDelta -= 2
		} else {
			scoreDelta--
		}
	}

	// 6/7. 日月落陷化忌:损耗档(本宫落陷、本宫化忌,天然同宫)。
	if hasMajor(majors, "太阳") && dim["太阳"] && sihua["太阳"] == ziwei.HuaJi {
		clauses = append(clauses, "【应验·损耗档】太阳落陷又化忌,损男亲、伤眼目、事业阻滞:宜养目、缓进、勿强出头")
		scoreDelta--
	}
	if hasMajor(majors, "太阴") && dim["太阴"] && sihua["太阴"] == ziwei.HuaJi {
		clauses = append(clauses, "【应验·损耗档】太阴落陷又化忌,损女亲、情绪暗耗、暗处失财:宜调心养血、防阴私之失")
		scoreDelta--
	}

	// 8. 会昌曲(吉向):文星拱照;同宫夹命更佳。
	if ca := aspectOf(local, triad, "文昌"); triad["文昌"] && triad["文曲"] {
		q := "会"
		if ca.inHouse && local["文曲"] {
			q = "同宫"
		}
		clauses = append(clauses, fmt.Sprintf("【应验·文贵档】%s文昌文曲,文星拱照:利文墨、考试、名声与专业,主聪慧秀发、锦上添花", q))
		scoreDelta++
	}
	// 9. 财星会禄存(吉向):财禄丰盈。
	if la := aspectOf(local, triad, "禄存"); la.present && hasAnyMajor(majors, "武曲", "天府", "太阴") {
		clauses = append(clauses, fmt.Sprintf("【应验·财禄档】财星%s禄存,财禄丰盈:进财稳而厚、善积累,宜置产守成", la.word()))
		scoreDelta++
	}
	return clauses, scoreDelta
}
