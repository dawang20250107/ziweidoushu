package ziwei

// 以第三方排盘软件(zwds.mdb)导出的结构表为「第二独立黄金基准」,
// 与既有 iztro 基准互证核心算法。此处的表均为确定性事实(安星诀、
// 六十甲子纳音),非断语/星情文本;iztro 与该商业软件为两套彼此独立
// 的实现,同表全等即强证引擎无误。溯源与结论见
// research/zwds-mdb-crosscheck.md。

import (
	"encoding/json"
	"os"
	"sort"
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
