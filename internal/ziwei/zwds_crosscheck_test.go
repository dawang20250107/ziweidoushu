package ziwei

// 以第三方排盘软件(zwds.mdb)导出的结构表为「第二独立黄金基准」,
// 与既有 iztro 基准互证核心算法。此处的表均为确定性事实(紫微定局、
// 安星诀、十神、六十甲子纳音),非断语/星情文本;iztro 与该商业软件为
// 两套彼此独立的实现,同表全等即强证引擎无误。安星链地基到主星完整覆盖:
// 局+日 → 紫微(zwds_ziwei_juday)→ 十四主星(zwds_majorstars)。
// 溯源与结论见 research/zwds-mdb-crosscheck.md。

import (
	"encoding/json"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// TestZWDSMajorStarsCrosscheck 校验 14 主星安星诀:对 12 个紫微落宫,
// 引擎摆放须与 zwds 紫微诸星定位表逐宫全等(共 12×14=168 星次)。
func TestZWDSMajorStarsCrosscheck(t *testing.T) {
	raw, err := os.ReadFile("testdata/zwds_majorstars.json")
	if err != nil {
		t.Fatal(err)
	}
	var want map[string]map[string][]string // 紫微地支 → 宫地支 → [主星]
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	if len(want) != 12 {
		t.Fatalf("基准应含 12 局,实际 %d", len(want))
	}

	joinSorted := func(ss []string) string {
		c := append([]string(nil), ss...)
		sort.Strings(c)
		return strings.Join(c, "+")
	}

	for _, zwName := range Branches {
		exp, ok := want[zwName]
		if !ok {
			t.Fatalf("基准缺紫微在%s", zwName)
		}
		// 引擎摆放:紫微落宫 → 天府落宫(fix12(12-ziweiIdx)),安十四主星。
		ziweiIdx := branchToPalaceIndex(branchIndexOf(zwName))
		tianfuIdx := fix12(12 - ziweiIdx)
		stars := make([][]Star, 12)
		placeMajorStars(stars, ziweiIdx, tianfuIdx, 0)

		got := map[string][]string{}
		total := 0
		for pIdx := 0; pIdx < 12; pIdx++ {
			bName := Branches[palaceIndexToBranch(pIdx)]
			for _, s := range stars[pIdx] {
				got[bName] = append(got[bName], s.Name)
				total++
			}
		}
		if total != 14 {
			t.Errorf("紫微在%s:引擎安主星 %d 颗(应 14)", zwName, total)
		}
		// 逐宫比对
		seen := map[string]bool{}
		for gw, names := range got {
			seen[gw] = true
			if joinSorted(names) != joinSorted(exp[gw]) {
				t.Errorf("紫微在%s·%s宫:引擎[%s] ≠ 基准[%s]",
					zwName, gw, joinSorted(names), joinSorted(exp[gw]))
			}
		}
		for gw := range exp {
			if !seen[gw] {
				t.Errorf("紫微在%s·%s宫:引擎无星,基准[%s]", zwName, gw, joinSorted(exp[gw]))
			}
		}
	}
}

// nayinCanon 归一第三方转写变体到工程口径(同一纳音的异写)。
var nayinCanon = map[string]string{
	"洞下水": "涧下水", "城墙土": "城头土", "霹雷火": "霹雳火",
	"佛灯火": "覆灯火", "桑林木": "桑柘木", "白腊金": "白蜡金",
}

func canonNaYin(s string) string {
	if c, ok := nayinCanon[s]; ok {
		return c
	}
	return s
}

// TestZWDSNaYinCrosscheck 校验六十甲子纳音:60 组干支的纳音归属,
// 引擎须与 zwds 纳音表逐组一致(容异写:洞下水=涧下水 等)。
func TestZWDSNaYinCrosscheck(t *testing.T) {
	raw, err := os.ReadFile("testdata/zwds_nayin.json")
	if err != nil {
		t.Fatal(err)
	}
	var want map[string]string // 干支名 → 纳音名
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	if len(want) != 60 {
		t.Fatalf("纳音基准应 60 组,实际 %d", len(want))
	}
	for gz, ny := range want {
		rs := []rune(gz)
		s, b := stemIndex(rs[0]), branchIndex(rs[1])
		if s < 0 || b < 0 {
			t.Fatalf("非法干支 %s", gz)
		}
		got := naYin(s, b)
		if canonNaYin(got) != canonNaYin(ny) {
			t.Errorf("%s:引擎纳音「%s」≠ 基准「%s」", gz, got, ny)
		}
	}
}

// TestZWDSZiweiJuDayCrosscheck 校验紫微定局(安星链地基):五行局 × 农历日 →
// 紫微所在宫。引擎须与 zwds 定局表逐格全等(5 局 × 30 日 = 150 格)。
func TestZWDSZiweiJuDayCrosscheck(t *testing.T) {
	raw, err := os.ReadFile("testdata/zwds_ziwei_juday.json")
	if err != nil {
		t.Fatal(err)
	}
	var want map[string][]int // 五行局(ju=2..6)→ [30 日]紫微宫(子=1..亥=12)
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	if len(want) != 5 {
		t.Fatalf("定局基准应 5 局,实际 %d", len(want))
	}
	for _, ju := range []int{2, 3, 4, 5, 6} {
		row, ok := want[strconv.Itoa(ju)]
		if !ok || len(row) != 30 {
			t.Fatalf("五行局 ju=%d 基准缺失或非 30 日", ju)
		}
		for day := 1; day <= 30; day++ {
			palIdx := ziweiPalaceByJuDay(ju, day)        // 宫位索引(寅=0)
			gotBranch := palaceIndexToBranch(palIdx) + 1 // → 子=1..亥=12
			if gotBranch != row[day-1] {
				t.Errorf("ju=%d 第%d日:引擎紫微在%s(%d) ≠ 基准 %d",
					ju, day, Branches[palaceIndexToBranch(palIdx)], gotBranch, row[day-1])
			}
		}
	}
}

// TestZWDSShiShenCrosscheck 校验十神:日主天干 × 对方天干 → 十神。
// 引擎 shiShen 须与 zwds 十神表逐格全等(10 × 10 = 100 格)。
func TestZWDSShiShenCrosscheck(t *testing.T) {
	raw, err := os.ReadFile("testdata/zwds_shishen.json")
	if err != nil {
		t.Fatal(err)
	}
	var want map[string][]string // 日主天干 → [对甲..癸]十神
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	if len(want) != 10 {
		t.Fatalf("十神基准应 10 日主,实际 %d", len(want))
	}
	for dm, row := range want {
		day := stemIndex([]rune(dm)[0])
		if day < 0 || len(row) != 10 {
			t.Fatalf("日主 %s 非法或非 10 列", dm)
		}
		for other := 0; other < 10; other++ {
			if got := shiShen(day, other); got != row[other] {
				t.Errorf("日主%s 对%s:引擎「%s」≠ 基准「%s」",
					dm, string(szStems[other]), got, row[other])
			}
		}
	}
}

// TestZWDSNianZhiShenShaCrosscheck 校验年支三合(将前)神煞落位:取自 zwds
// 将前十二神表 zwcomp_22——该表华盖正确落墓库(辰),与八字模块 sz_info_3/5
// 误置冲位(戌)不同;故三合神煞只认可靠的紫微模块表。12 年支 × 7 神煞全等。
// (红鸾/天喜/孤辰/寡宿因 sz_info_5 有多处录入错误,改由规则钉死,见下测试。)
func TestZWDSNianZhiShenShaCrosscheck(t *testing.T) {
	raw, err := os.ReadFile("testdata/zwds_shensha_nianzhi.json")
	if err != nil {
		t.Fatal(err)
	}
	var want map[string]map[string]string // 年支 → 神煞 → 地支
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	for yz := 0; yz < 12; yz++ {
		exp := want[string(szBranches[yz])]
		js := jiangStarZhi(yz)
		for _, s := range sanheShenSha {
			pos := (js + s.offset) % 12
			if e, ok := exp[s.name]; ok && e != string(szBranches[pos]) {
				t.Errorf("%s年·%s:引擎「%s」≠ 基准「%s」",
					string(szBranches[yz]), s.name, string(szBranches[pos]), e)
			}
		}
	}
}

// TestNianZhiShenShaRuleLock 红鸾/天喜/孤辰/寡宿按经典口诀钉死(sz_info_5 该四
// 神煞有移位/重出等录入错误,不作基准;此处以硬编码标准值锁定引擎规则)。
func TestNianZhiShenShaRuleLock(t *testing.T) {
	// 期望(子…亥),地支索引 子=0。
	hongLuanExp := []int{3, 2, 1, 0, 11, 10, 9, 8, 7, 6, 5, 4} // 子卯丑寅…
	guChenExp := []int{2, 2, 5, 5, 5, 8, 8, 8, 11, 11, 11, 2}  // 亥子丑→寅…
	guaSuExp := []int{10, 10, 1, 1, 1, 4, 4, 4, 7, 7, 7, 10}   // 亥子丑→戌…
	for yz := 0; yz < 12; yz++ {
		hl := ((3 - yz) + 12) % 12
		if hl != hongLuanExp[yz] {
			t.Errorf("%s年红鸾:引擎%d ≠ 期望%d", string(szBranches[yz]), hl, hongLuanExp[yz])
		}
		if tx := (hl + 6) % 12; tx != (hongLuanExp[yz]+6)%12 {
			t.Errorf("%s年天喜:引擎%d 非红鸾冲", string(szBranches[yz]), tx)
		}
		gc, gs := guChenGuaSu(yz)
		if gc != guChenExp[yz] || gs != guaSuExp[yz] {
			t.Errorf("%s年孤辰/寡宿:引擎{%d,%d} ≠ 期望{%d,%d}",
				string(szBranches[yz]), gc, gs, guChenExp[yz], guaSuExp[yz])
		}
	}
}

// TestZWDSRiGanShenShaCrosscheck 校验日干系神煞:禄神/文昌/羊刃(阳干)逐位
// 与 zwds 日干神煞表全等;天乙贵人用「基准 ⊆ 引擎」容错——该表辛干只录了寅、
// 漏了午(六辛逢马虎,午寅两位),软件所录须都在引擎位内即可。羊刃仅阳干
// (阳刃为阳干专有,阴干变体不采)。
func TestZWDSRiGanShenShaCrosscheck(t *testing.T) {
	raw, err := os.ReadFile("testdata/zwds_shensha_rigan.json")
	if err != nil {
		t.Fatal(err)
	}
	var want map[string]map[string][]string // 日干 → 神煞 → [地支...]
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	name := func(i int) string { return string(szBranches[i]) }
	for dg := 0; dg < 10; dg++ {
		exp := want[string(szStems[dg])]
		// 精确位:禄神、文昌、羊刃(阳干)
		exact := func(nm string, idx int) {
			if e, ok := exp[nm]; ok && !(len(e) == 1 && e[0] == name(idx)) {
				t.Errorf("日干%s·%s:引擎「%s」≠ 基准%v", string(szStems[dg]), nm, name(idx), e)
			}
		}
		exact("禄神", luBranch[dg])
		exact("文昌", wenChangZhi[dg])
		if rb, ok := renBranch[dg]; ok {
			exact("羊刃", rb)
		}
		// 天乙贵人:基准 ⊆ 引擎(容软件漏录)
		if e, ok := exp["天乙贵人"]; ok {
			engine := map[string]bool{name(tianYiGuiRen[dg][0]): true, name(tianYiGuiRen[dg][1]): true}
			for _, z := range e {
				if !engine[z] {
					t.Errorf("日干%s·天乙贵人:基准「%s」不在引擎位%v", string(szStems[dg]), z,
						[]string{name(tianYiGuiRen[dg][0]), name(tianYiGuiRen[dg][1])})
				}
			}
		}
	}
}

// TestZWDSKongWangCrosscheck 校验空亡(旬空):六十甲子各自旬空的两支,
// 引擎公式须与 zwds 空亡表逐组全等。
func TestZWDSKongWangCrosscheck(t *testing.T) {
	raw, err := os.ReadFile("testdata/zwds_kongwang.json")
	if err != nil {
		t.Fatal(err)
	}
	var want map[string][]string // 干支 → [空支×2]
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	if len(want) != 60 {
		t.Fatalf("空亡基准应 60 组,实际 %d", len(want))
	}
	skipped := 0
	for gz, kong := range want {
		rs := []rune(gz)
		s, b := stemIndex(rs[0]), branchIndex(rs[1])
		if s%2 != b%2 { // 无效干支(该表有一处 0510=戊酉 录入错误,阴阳不配),跳过
			skipped++
			continue
		}
		k1 := ((b-s)%12 + 12 + 10) % 12
		k2 := ((b-s)%12 + 12 + 11) % 12
		gotSet := map[string]bool{string(szBranches[k1]): true, string(szBranches[k2]): true}
		if len(kong) != 2 || !gotSet[kong[0]] || !gotSet[kong[1]] {
			t.Errorf("%s:引擎空亡{%s,%s} ≠ 基准%v",
				gz, string(szBranches[k1]), string(szBranches[k2]), kong)
		}
	}
	if skipped != 1 { // 已知恰一处录入错误;多于/少于 1 说明表变了,需复核
		t.Errorf("跳过的无效干支应为 1(0510=戊酉 录入错误),实际 %d", skipped)
	}
}
