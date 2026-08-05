// 四柱神煞层:在四柱干支之上标注常用神煞与空亡——为「四柱视角」补齐
// 神煞维度(此前只有十神/藏干/纳音/格局)。规则取子平通行口径(《三命通会》
// 《渊海子平》),并以 zwds.mdb 结构表反向互证(见 research/sizhu-shensha.md);
// 凡与该软件八字表分歧处,均系其表之误(如华盖列偏冲位),我方按正确规则实现。
package ziwei

// ShenSha 一个命中的神煞:名 + 所在柱(年/月/日/时)+ 查法依据。
type ShenSha struct {
	Name    string   `json:"name"`
	Pillars []string `json:"pillars"` // 命中柱:年/月/日/时
	Basis   string   `json:"basis"`   // 查法:年支三合 / 年支 / 日干 / 日柱旬
}

// pillarShort 柱序 → 短名。
var pillarShort = []string{"年", "月", "日", "时"}

// jiangStarZhi 年支三合局的将星位(旺神):申子辰→子、亥卯未→卯、寅午戌→午、
// 巳酉丑→酉。三合以地支 mod 4 归局。
func jiangStarZhi(yz int) int {
	switch yz % 4 {
	case 0: // 申子辰(水)
		return 0 // 子
	case 1: // 巳酉丑(金)
		return 9 // 酉
	case 2: // 寅午戌(火)
		return 6 // 午
	default: // 亥卯未(木)
		return 3 // 卯
	}
}

// 将前十二神中我方采用的常用者:自将星顺行的偏移。
// 将星0 攀鞍1 岁驿(驿马)2 息神3 华盖4 劫煞5 灾煞6 天煞7 指背8 咸池(桃花)9 月煞10 亡神11。
var sanheShenSha = []struct {
	name   string
	offset int
}{
	{"将星", 0}, {"驿马", 2}, {"华盖", 4}, {"劫煞", 5},
	{"灾煞", 6}, {"桃花", 9}, {"亡神", 11},
}

// 天乙贵人(按日干,两位):甲戊庚丑未、乙己子申、丙丁亥酉、辛午寅、壬癸卯巳。
var tianYiGuiRen = [10][2]int{
	{1, 7}, {0, 8}, {11, 9}, {11, 9}, {1, 7},
	{0, 8}, {1, 7}, {6, 2}, {3, 5}, {3, 5},
}

// 文昌(按日干):甲巳 乙午 丙戊申 丁己酉 庚亥 辛子 壬寅 癸卯。
var wenChangZhi = []int{5, 6, 8, 9, 8, 9, 11, 0, 2, 3}

// guChenGuaSu 年支方位 → 孤辰、寡宿位。
func guChenGuaSu(yz int) (int, int) {
	switch yz {
	case 11, 0, 1: // 亥子丑
		return 2, 10 // 孤辰寅、寡宿戌
	case 2, 3, 4: // 寅卯辰
		return 5, 1 // 孤辰巳、寡宿丑
	case 5, 6, 7: // 巳午未
		return 8, 4 // 孤辰申、寡宿辰
	default: // 申酉戌
		return 11, 7 // 孤辰亥、寡宿未
	}
}

// buildShenSha 由四柱(stem,branch 索引,子=0)推神煞与空亡。
// stems/branches[0..3] = 年月日时;返回命中神煞,并按柱回填空亡标记 kong[柱]。
func buildShenSha(stems, branches [4]int) ([]ShenSha, [4]bool) {
	yz := branches[0] // 年支
	dg := stems[2]    // 日干
	ds, db := stems[2], branches[2]

	// 收集某组目标地支命中的柱名。
	hits := func(targets ...int) []string {
		set := map[int]bool{}
		for _, t := range targets {
			set[((t%12)+12)%12] = true
		}
		var ps []string
		for i, b := range branches {
			if set[b] {
				ps = append(ps, pillarShort[i])
			}
		}
		return ps
	}

	out := []ShenSha{} // 非 nil:遵守 JSON 契约(Go nil slice→null 会致前端崩)
	add := func(name, basis string, targets ...int) {
		if ps := hits(targets...); len(ps) > 0 {
			out = append(out, ShenSha{Name: name, Pillars: ps, Basis: basis})
		}
	}

	// ① 年支三合(将前)神煞
	js := jiangStarZhi(yz)
	for _, s := range sanheShenSha {
		add(s.name, "年支三合", js+s.offset)
	}
	// ② 年支:红鸾、天喜、孤辰、寡宿
	hongLuan := ((3 - yz) + 12) % 12
	add("红鸾", "年支", hongLuan)
	add("天喜", "年支", hongLuan+6)
	gc, gs := guChenGuaSu(yz)
	add("孤辰", "年支", gc)
	add("寡宿", "年支", gs)
	// ③ 日干:天乙贵人、禄神、羊刃、文昌
	add("天乙贵人", "日干", tianYiGuiRen[dg][0], tianYiGuiRen[dg][1])
	add("禄神", "日干", luBranch[dg])
	if rb, ok := renBranch[dg]; ok { // 阳刃(阳干专有)
		add("羊刃", "日干", rb)
	}
	add("文昌", "日干", wenChangZhi[dg])

	// ④ 空亡:以日柱旬(旬首支=db-ds),空亡为旬首后第 10、11 位;
	// 标记年/月/时柱地支落空(日柱自身不标)。
	var kong [4]bool
	k1 := ((db-ds)%12 + 12 + 10) % 12
	k2 := ((db-ds)%12 + 12 + 11) % 12
	var kongPillars []string
	for i, b := range branches {
		if i == 2 {
			continue // 日柱自身不论空
		}
		if b == k1 || b == k2 {
			kong[i] = true
			kongPillars = append(kongPillars, pillarShort[i])
		}
	}
	if len(kongPillars) > 0 {
		out = append(out, ShenSha{Name: "空亡", Pillars: kongPillars, Basis: "日柱旬"})
	}
	return out, kong
}
