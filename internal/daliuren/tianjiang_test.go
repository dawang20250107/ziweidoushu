package daliuren

import "testing"

// 天将布盘:贵人歌两书互证锚点 + 顺逆布规则。
// 伏吟课(月将==占时)天地盘重合,便于手算核对。
func TestTianJiang(t *testing.T) {
	// 甲子日午时占(昼贵丑):伏吟 → 丑临地盘丑(1 位,亥~辰间)→ 顺布。
	r, err := Cast(0, 0, 6, 6)
	if err != nil {
		t.Fatal(err)
	}
	if !r.GuiIsDay {
		t.Error("午时应用昼贵")
	}
	if r.TianJiang[1] != "贵人" || r.TianJiang[2] != "螣蛇" || r.TianJiang[0] != "天后" {
		t.Errorf("昼贵顺布: 丑%s 寅%s 子%s, want 贵人/螣蛇/天后", r.TianJiang[1], r.TianJiang[2], r.TianJiang[0])
	}
	// 甲子日子时占(夜贵未):未临地盘未(7 位,巳~戌间)→ 逆布。
	r2, err := Cast(0, 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if r2.GuiIsDay {
		t.Error("子时应用夜贵")
	}
	if r2.TianJiang[7] != "贵人" || r2.TianJiang[6] != "螣蛇" || r2.TianJiang[8] != "天后" {
		t.Errorf("夜贵逆布: 未%s 午%s 申%s, want 贵人/螣蛇/天后", r2.TianJiang[7], r2.TianJiang[6], r2.TianJiang[8])
	}
	// 三传乘将齐备,断语逐传附乘将
	for i, jg := range r.ChuanJiang {
		if jg == "" {
			t.Errorf("三传第 %d 传乘将缺失", i+1)
		}
	}
	if r.Judgment == nil || len(r.Judgment.SanChuan) == 0 {
		t.Fatal("断语缺失")
	}
	for _, ln := range r.Judgment.SanChuan {
		if !contains(ln, "乘") {
			t.Errorf("逐传解应附乘将: %s", ln)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
