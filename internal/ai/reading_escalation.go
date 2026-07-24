package ai

// 应验分档:双星/主星组合再叠煞吉会照,从「格性」推进到「应验档位」。
//
// 同一组合遇不同煞吉,吉凶天差地别:廉杀本刚烈,加羊陀则应「路上埋尸」血光档;
// 贪狼逢火铃则成暴发格。此层以本宫 + 三方四正的煞吉会照为触发条件,给出应验断语。
// 内容为三合派公版通则的原创提炼(非任何第三方文本)。

import "github.com/dawang20250107/ziweidoushu/internal/ziwei"

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

func anyIn(set map[string]bool, names ...string) bool {
	for _, n := range names {
		if set[n] {
			return true
		}
	}
	return false
}

func hasMajor(majors []string, name string) bool {
	for _, m := range majors {
		if m == name {
			return true
		}
	}
	return false
}

// escalations 由本宫主星 + 煞吉会照推出应验档断语与吉凶增减。
// dim 为落陷主星集合,sihua 为主星四化(均来自 scanPalace)。
func escalations(chart *ziwei.Chart, p *ziwei.Palace, majors []string, dim map[string]bool, sihua map[string]ziwei.SiHua) (clauses []string, scoreDelta int) {
	set := triadStarSet(chart, p.Branch)
	huoLing := anyIn(set, "火星", "铃星")
	yangTuo := anyIn(set, "擎羊", "陀罗")
	kongJie := set["地空"] && set["地劫"]

	// 1. 贪狼逢火铃:暴发格应验(吉中带戒)。
	if hasMajor(majors, "贪狼") && huoLing {
		clauses = append(clauses, "【应验·暴发格】贪狼逢火/铃,成火贪、铃贪:主突发横财、机遇骤至;然骤起亦骤落,宜见好即收、忌得陇望蜀")
		scoreDelta++
	}

	// 2. 廉杀逢羊陀:「路上埋尸」血光重档(最烈,独占)。
	lianSha := hasMajor(majors, "廉贞") && hasMajor(majors, "七杀")
	if lianSha && yangTuo {
		clauses = append(clauses, "【应验·血光档】廉杀逢羊陀,古云『路上埋尸』:主意外、血光、伤灾,尤忌驾驶与高危作业,宜信仰行善、慎行以化")
		scoreDelta -= 2
	} else if (hasMajor(majors, "七杀") || hasMajor(majors, "破军")) && (yangTuo || huoLing) {
		// 3. 杀破逢煞:冲击破耗加剧(廉杀羊陀已独占则不再叠)。
		clauses = append(clauses, "【应验·冲击档】杀破逢羊陀火铃,冲击破耗加剧:成败起伏大、防伤灾意外与耗财,宜以稳制动、忌逞强")
		scoreDelta--
	}

	// 4. 巨门逢煞忌:是非官非加剧。
	if hasMajor(majors, "巨门") && (yangTuo || sihua["巨门"] == ziwei.HuaJi) {
		clauses = append(clauses, "【应验·是非档】巨门逢羊陀或化忌,是非口舌、官非构陷加剧:慎签约、防小人,宜守口修心")
		scoreDelta--
	}

	// 5. 空劫会照财动之星:财空成空。
	if kongJie && hasAnyMajor(majors, "武曲", "贪狼", "七杀", "破军", "天府", "太阴") {
		clauses = append(clauses, "【应验·成空档】财动之星逢地空地劫,谋财求进易付诸流水:宜脚踏实地、忌投机豪赌与合伙重财")
		scoreDelta--
	}

	// 6. 太阳落陷化忌:损男亲、伤眼目。
	if hasMajor(majors, "太阳") && dim["太阳"] && sihua["太阳"] == ziwei.HuaJi {
		clauses = append(clauses, "【应验·损耗档】太阳落陷又化忌,损男亲、伤眼目、事业阻滞:宜养目、缓进、勿强出头")
		scoreDelta--
	}
	// 7. 太阴落陷化忌:损女亲、情绪暗耗。
	if hasMajor(majors, "太阴") && dim["太阴"] && sihua["太阴"] == ziwei.HuaJi {
		clauses = append(clauses, "【应验·损耗档】太阴落陷又化忌,损女亲、情绪暗耗、暗处失财:宜调心养血、防阴私之失")
		scoreDelta--
	}

	// 8. 会昌曲(吉向):文星拱照、锦上添花。
	if set["文昌"] && set["文曲"] {
		clauses = append(clauses, "【应验·文贵档】会文昌文曲,文星拱照:利文墨、考试、名声与专业,主聪慧秀发、锦上添花")
		scoreDelta++
	}
	// 9. 财星会禄存(吉向):财禄丰盈。
	if set["禄存"] && hasAnyMajor(majors, "武曲", "天府", "太阴") {
		clauses = append(clauses, "【应验·财禄档】财星会禄存,财禄丰盈:进财稳而厚、善积累,宜置产守成")
		scoreDelta++
	}
	return clauses, scoreDelta
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
