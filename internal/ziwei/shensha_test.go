package ziwei

import (
	"sort"
	"strings"
	"testing"
)

// TestBuildShenSha 锁定神煞命中柱与空亡回填(手算基准)。
// 四柱:年甲子 月丙寅 日甲子 时甲戌。
//
//	年支子→将星在子:命中年、日(皆子);日干甲→禄在寅:命中月(寅);
//	日柱甲子旬空戌亥:时柱戌落空。
func TestBuildShenSha(t *testing.T) {
	stems := [4]int{0, 2, 0, 0}     // 甲 丙 甲 甲
	branches := [4]int{0, 2, 0, 10} // 子 寅 子 戌
	ss, kong := buildShenSha(stems, branches)

	got := map[string]string{}
	for _, s := range ss {
		c := append([]string(nil), s.Pillars...)
		sort.Strings(c)
		got[s.Name] = strings.Join(c, "")
	}
	want := map[string]string{
		"将星": "年日", // 排序后(年<日<时,Unicode 序)
		"禄神": "月",
		"空亡": "时",
	}
	for name, pillars := range want {
		if got[name] != pillars {
			t.Errorf("神煞%s:引擎命中「%s」≠ 期望「%s」", name, got[name], pillars)
		}
	}
	if !kong[3] || kong[0] || kong[1] || kong[2] {
		t.Errorf("空亡回填错误:kong=%v(应仅时柱 true)", kong)
	}
}
