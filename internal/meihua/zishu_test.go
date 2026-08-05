package meihua

// 字数起卦(声音占义)测试:同刻不同问辞卦各不同,无问辞回退观梅旧法。

import (
	"strings"
	"testing"
	"time"

	"github.com/6tail/lunar-go/calendar"
)

// 观梅占同刻(申时,时数9)为基准时刻。
func shenShiFixture() time.Time {
	lunar := calendar.NewLunar(2012, 12, 17, 16, 0, 0)
	solar := lunar.GetSolar()
	return time.Date(solar.GetYear(), time.Month(solar.GetMonth()), solar.GetDay(), 16, 0, 0, 0, time.Local)
}

// TestZiShuQiGua 十二字问辞申时:上卦 12%8=4 震,下卦 (12+9)%8=5 巽,
// 动爻 21%6=3 → 雷风恒之三爻动。
func TestZiShuQiGua(t *testing.T) {
	tm := shenShiFixture()
	r, err := ByTimeAndText(tm, strings.Repeat("问", 12))
	if err != nil {
		t.Fatal(err)
	}
	if r.Ben.Name != "雷风恒" || r.Moving != 3 {
		t.Fatalf("12字申时应得雷风恒三爻动,got %s 动%d(basis %s)", r.Ben.Name, r.Moving, r.CastBasis)
	}
	if r.CastBasis == "" {
		t.Fatal("应记起数依据")
	}

	// 同刻不同字数 → 卦不同(5 字 vs 6 字上卦异)
	r5, _ := ByTimeAndText(tm, strings.Repeat("问", 5))
	r6, _ := ByTimeAndText(tm, strings.Repeat("问", 6))
	if r5.Ben.Name == r6.Ben.Name && r5.Moving == r6.Moving {
		t.Fatalf("同刻不同问辞不应同卦同爻: %s动%d", r5.Ben.Name, r5.Moving)
	}

	// 无问辞(含全空白)回退观梅旧法(与 ByTime 全同)
	rEmpty, err := ByTimeAndText(tm, "  ")
	if err != nil {
		t.Fatal(err)
	}
	rOld, _ := ByTime(tm, "")
	if rEmpty.Ben.Name != rOld.Ben.Name || rEmpty.Moving != rOld.Moving {
		t.Fatalf("无问辞应回退年月日时法: %s vs %s", rEmpty.Ben.Name, rOld.Ben.Name)
	}
}
