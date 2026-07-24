package liuyao

// 经文层测试:zhouyi 组成表与 meihua 卦名表跨包互证 + 装卦经文接线。

import (
	"strings"
	"testing"

	"github.com/dawang20250107/ziweidoushu/internal/meihua"
	"github.com/dawang20250107/ziweidoushu/internal/zhouyi"
)

// TestZhouyiMeihuaCross 64 组合:周易经文表卦名须含于 meihua 卦名表
// (如「姤」⊂「天风姤」、「乾」⊂「乾为天」)——两表独立而互证。
func TestZhouyiMeihuaCross(t *testing.T) {
	for up := 1; up <= 8; up++ {
		for lo := 1; lo <= 8; lo++ {
			g := zhouyi.ByTrigrams(up, lo)
			full := meihua.HexagramNameByNums(up, lo)
			if g == nil || !strings.Contains(full, g.Name) {
				t.Fatalf("上%d下%d: zhouyi %q vs meihua %q", up, lo, g.Name, full)
			}
		}
	}
}

// TestJingWenWiring 装卦经文接线:天风姤初爻动 → 本卦辞「女壮」、
// 初六爻辞「系于金柅」、变卦乾为天卦辞「元亨。利贞。」。
func TestJingWenWiring(t *testing.T) {
	// 姤 = 巽下乾上,初爻动 → 变乾
	r, err := AssembleForResearch([6]bool{false, true, true, true, true, true}, []int{1}, 0, 0, '午')
	if err != nil {
		t.Fatal(err)
	}
	jw := r.JingWen
	if jw == nil {
		t.Fatal("经文层缺失")
	}
	if !strings.HasPrefix(jw.BenGuaCi, "女壮") {
		t.Fatalf("姤卦辞: %s", jw.BenGuaCi)
	}
	if len(jw.YaoCi) != 1 || !strings.Contains(jw.YaoCi[0], "系于金柅") || !strings.HasPrefix(jw.YaoCi[0], "初六") {
		t.Fatalf("动爻爻辞: %v", jw.YaoCi)
	}
	if jw.BianGuaCi != "元亨。利贞。" {
		t.Fatalf("变卦辞: %s", jw.BianGuaCi)
	}

	// 乾为天六爻皆动 → 用九
	r2, _ := AssembleForResearch([6]bool{true, true, true, true, true, true}, []int{1, 2, 3, 4, 5, 6}, 0, 0, '午')
	if r2.JingWen == nil || !strings.HasPrefix(r2.JingWen.Yong, "用九") {
		t.Fatalf("乾六爻皆动应见用九: %+v", r2.JingWen)
	}

	// 静卦:只有本卦辞
	r3, _ := AssembleForResearch([6]bool{true, true, true, false, false, false}, nil, 0, 0, '午')
	if r3.JingWen == nil || r3.JingWen.BianGuaCi != "" || len(r3.JingWen.YaoCi) != 0 {
		t.Fatalf("静卦经文: %+v", r3.JingWen)
	}
	if !strings.Contains(r3.JingWen.BenGuaCi, "小往大来") { // 地天泰
		t.Fatalf("泰卦辞: %s", r3.JingWen.BenGuaCi)
	}
}
