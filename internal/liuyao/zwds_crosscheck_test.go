package liuyao

// 以第三方排盘软件(zwds.mdb)导出的六十四卦逐爻变卦表为独立基准,
// 校验卦名表与「动爻变卦」这一确定性机械运算。表为纯事实(卦名/爻变),
// 非断辞。溯源与结论见 research/zwds-mdb-crosscheck.md。

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/dawang20250107/ziweidoushu/internal/meihua"
)

type zwdsGua struct {
	Num  int      `json:"num"`
	Sgxg string   `json:"sgxg"` // 「上卦先天数-下卦先天数」
	Dy   []string `json:"dy"`   // 初爻…上爻 各变一爻后的卦名(不含序号)
}

// canonGua 归一异写:第三方用「垢」,工程用「姤」(同卦异写)。
func canonGua(s string) string { return strings.ReplaceAll(s, "垢", "姤") }

// hammingToName 目标卦名相对本卦六爻的变爻数(遍历 64 卦匹配名后按位比较)。
func hammingToName(base [6]bool, targetName string) int {
	for u := 1; u <= 8; u++ {
		for l := 1; l <= 8; l++ {
			if canonGua(meihua.HexagramNameByNums(u, l)) != targetName {
				continue
			}
			up, lo := meihua.TrigramByNum(u), meihua.TrigramByNum(l)
			tgt := [6]bool{lo.Lines[0], lo.Lines[1], lo.Lines[2], up.Lines[0], up.Lines[1], up.Lines[2]}
			d := 0
			for i := 0; i < 6; i++ {
				if base[i] != tgt[i] {
					d++
				}
			}
			return d
		}
	}
	return -1
}

// zwdsBianErrata 第三方(zwds.mdb)变卦表已知录入错误:键=「本卦名·变爻位」,
// 值=该表所填(经证为多爻变,不可能由单爻变得到)。跨书互证于此反向发现
// 商业软件的数据错误——引擎在这三处均正确(单爻变)。详见
// research/zwds-mdb-crosscheck.md。
var zwdsBianErrata = map[string]string{
	"水雷屯·2": "坎为水",  // 屯二爻变应得水泽节(距本卦2爻,误)
	"水雷屯·3": "水山蹇",  // 屯三爻变应得水火既济(距本卦2爻,误)
	"火风鼎·2": "火雷噬嗑", // 鼎二爻变应得火山旅(距本卦3爻,误)
}

// TestZWDSBianGuaCrosscheck 校验:① 上下卦先天数 → 卦名(全 64 卦名表);
// ② 逐爻取变 → 变卦名。引擎须与 zwds 变卦表逐项一致;唯 zwdsBianErrata
// 三处第三方表本身录入错误(填了多爻变卦),引擎正确,列为已知豁免并断言
// 「引擎单爻变 vs 基准多爻变」以锁定该结论、并防新分歧混入。
func TestZWDSBianGuaCrosscheck(t *testing.T) {
	raw, err := os.ReadFile("testdata/zwds_biangua.json")
	if err != nil {
		t.Fatal(err)
	}
	var want map[string]zwdsGua
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	if len(want) != 64 {
		t.Fatalf("变卦基准应 64 卦,实际 %d", len(want))
	}
	errataHit := map[string]bool{}

	for name, g := range want {
		parts := strings.Split(g.Sgxg, "-")
		if len(parts) != 2 {
			t.Fatalf("%s:sgxg 格式异常 %q", name, g.Sgxg)
		}
		upN, _ := strconv.Atoi(parts[0]) // 上卦先天数
		loN, _ := strconv.Atoi(parts[1]) // 下卦先天数

		// ① 上下卦 → 本卦名
		if got := meihua.HexagramNameByNums(upN, loN); canonGua(got) != canonGua(name) {
			t.Errorf("卦名:上%d下%d 引擎「%s」≠ 基准「%s」", upN, loN, got, name)
			continue
		}

		// 由上下卦重建六爻(自下而上:下卦三爻 + 上卦三爻)
		up := meihua.TrigramByNum(upN)
		lo := meihua.TrigramByNum(loN)
		var lines [6]bool
		lines[0], lines[1], lines[2] = lo.Lines[0], lo.Lines[1], lo.Lines[2]
		lines[3], lines[4], lines[5] = up.Lines[0], up.Lines[1], up.Lines[2]

		if len(g.Dy) != 6 {
			t.Fatalf("%s:变卦项应 6,实际 %d", name, len(g.Dy))
		}
		// ② 逐爻取变(pos 0=初爻…5=上爻)
		for pos := 0; pos < 6; pos++ {
			fl := lines
			fl[pos] = !fl[pos]
			bl := meihua.TrigramByLines([3]bool{fl[0], fl[1], fl[2]}).Num
			bu := meihua.TrigramByLines([3]bool{fl[3], fl[4], fl[5]}).Num
			got := meihua.HexagramNameByNums(bu, bl)
			if canonGua(got) == canonGua(g.Dy[pos]) {
				continue
			}
			key := name + "·" + strconv.Itoa(pos+1)
			if bad, ok := zwdsBianErrata[key]; ok && canonGua(bad) == canonGua(g.Dy[pos]) {
				// 已知第三方错误:断言其确为多爻变(距本卦>1),据此证明是表错非引擎错。
				if hd := hammingToName(lines, canonGua(g.Dy[pos])); hd <= 1 {
					t.Errorf("%s:errata「%s」竟为单爻变(距%d),豁免依据不成立", key, bad, hd)
				}
				errataHit[key] = true
				continue
			}
			hd := hammingToName(lines, canonGua(g.Dy[pos]))
			t.Errorf("%s 变第%d爻:引擎「%s」(单爻变) ≠ 基准「%s」(距本卦%d爻);非已知豁免",
				name, pos+1, got, g.Dy[pos], hd)
		}
	}

	// 已知豁免须全部命中——若某处第三方表被更正或引擎行为变化,errata 即成
	// 陈旧,提示应复核并收敛。
	for key := range zwdsBianErrata {
		if !errataHit[key] {
			t.Errorf("errata「%s」未命中:第三方表或引擎已变,请复核 zwdsBianErrata", key)
		}
	}
}
