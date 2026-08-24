package meihua

import (
	"strings"
	"testing"
	"time"
)

// TestByZi 测字起卦:笔画锚点、一字/二字口径、拒绝路径。
func TestByZi(t *testing.T) {
	// 笔画锚点(Unihan 简体)
	for ch, want := range map[rune]int{'梅': 11, '花': 7, '易': 8, '数': 13, '转': 8, '职': 11} {
		if got := StrokesOf(ch); got != want {
			t.Errorf("StrokesOf(%c)=%d want %d", ch, got, want)
		}
	}
	// 2026-08-05 10:30 CST = 巳时(时辰数6)
	at := time.Date(2026, 8, 5, 10, 30, 0, 0, time.FixedZone("CST", 8*3600))

	// 一字:「梅」11画 → 上卦 11%8=3 离;下卦 (11+6)%8=1 乾;动爻 (11+6)%6=5
	r, err := ByZi("梅", at, "问事")
	if err != nil {
		t.Fatal(err)
	}
	if r.Ben.Upper.Num != 3 || r.Ben.Lower.Num != 1 || r.Moving != 5 {
		t.Errorf("「梅」巳时应 上离3/下乾1/动5,得 %d/%d/%d", r.Ben.Upper.Num, r.Ben.Lower.Num, r.Moving)
	}
	if r.Method != "zi" || r.ZiText != "梅" || len(r.ZiStrokes) != 1 || r.ZiStrokes[0] != 11 {
		t.Errorf("测字溯源字段: %+v %v", r.ZiText, r.ZiStrokes)
	}
	if !strings.Contains(r.CastBasis, "11画") {
		t.Errorf("起数依据应含笔画: %s", r.CastBasis)
	}
	if r.Judgment == nil || len(r.Judgment.Sections) == 0 {
		t.Error("测字卦亦应有分节深断")
	}

	// 二字:「转职」8+11 → 上8坤 下 11%8=3 离;动 (8+11+6)%6=1
	r2, err := ByZi("转职", at, "转职")
	if err != nil {
		t.Fatal(err)
	}
	if r2.Ben.Upper.Num != 8 || r2.Ben.Lower.Num != 3 || r2.Moving != 1 {
		t.Errorf("「转职」巳时应 上坤8/下离3/动1,得 %d/%d/%d", r2.Ben.Upper.Num, r2.Ben.Lower.Num, r2.Moving)
	}

	// 拒绝:空、超两字、非汉字
	for _, bad := range []string{"", "转职梅", "ab", "梅a"} {
		if _, err := ByZi(bad, at, ""); err == nil {
			t.Errorf("ByZi(%q) 应拒绝", bad)
		}
	}
}
