package meihua

import (
	"testing"
	"time"

	"github.com/6tail/lunar-go/calendar"
)

// TestGuanMeiZhan 观梅占(《梅花易数》开卷第一占,黄金基准):
// 辰年十二月十七日申时:年辰5+月12+日17=34 → 上卦兑;34+申9=43 → 下卦离;
// 43%6=1 → 初爻动。得「泽火革」,互见乾巽(互卦天风姤),变「泽山咸」;
// 动爻在下卦,体=兑金,用=离火,火克金 → 用克体,凶(典载:女子折股应之)。
func TestGuanMeiZhan(t *testing.T) {
	// 壬辰年农历十二月十七申时 → 公历(经 lunar-go 反推,申时取 16 点)
	lunar := calendar.NewLunar(2012, 12, 17, 16, 0, 0)
	solar := lunar.GetSolar()
	tm := time.Date(solar.GetYear(), time.Month(solar.GetMonth()), solar.GetDay(), 16, 0, 0, 0, time.Local)

	r, err := ByTime(tm, "观梅")
	if err != nil {
		t.Fatal(err)
	}
	if r.Ben.Name != "泽火革" {
		t.Fatalf("本卦: got %s want 泽火革(农历 %s)", r.Ben.Name, r.LunarText)
	}
	if r.Moving != 1 {
		t.Fatalf("动爻: got %d want 1", r.Moving)
	}
	if r.Hu.Name != "天风姤" {
		t.Fatalf("互卦: got %s want 天风姤(互见乾巽)", r.Hu.Name)
	}
	if r.Bian.Name != "泽山咸" {
		t.Fatalf("变卦: got %s want 泽山咸", r.Bian.Name)
	}
	if r.TiTrigram.Name != "兑" || r.YongTrigram.Name != "离" {
		t.Fatalf("体用: got 体%s 用%s want 体兑 用离", r.TiTrigram.Name, r.YongTrigram.Name)
	}
	if r.Relation != RelYongKeTi {
		t.Fatalf("生克: got %s want 用克体", r.Relation)
	}
}

// TestHexagramNamesComplete 六十四卦名表完整且互不重复。
func TestHexagramNamesComplete(t *testing.T) {
	seen := map[string]bool{}
	for u := 0; u < 8; u++ {
		for l := 0; l < 8; l++ {
			name := hexagramNames[u][l]
			if name == "" {
				t.Fatalf("卦名缺失: 上%d 下%d", u+1, l+1)
			}
			if seen[name] {
				t.Fatalf("卦名重复: %s", name)
			}
			seen[name] = true
		}
	}
	if len(seen) != 64 {
		t.Fatalf("卦名总数: got %d want 64", len(seen))
	}
}

// TestPureHexagramHu 八纯卦结构自检:乾坤两纯卦的互卦仍为自身。
func TestPureHexagramHu(t *testing.T) {
	r := derive(1, 1, 1) // 乾为天
	if r.Hu.Name != "乾为天" {
		t.Fatalf("乾为天互卦: got %s want 乾为天", r.Hu.Name)
	}
	r = derive(8, 8, 1) // 坤为地
	if r.Hu.Name != "坤为地" {
		t.Fatalf("坤为地互卦: got %s want 坤为地", r.Hu.Name)
	}
}

// TestByNumbers 数字起卦:8/8 取余作 8(坤),动爻和取六余。
func TestByNumbers(t *testing.T) {
	r, err := ByNumbers([]int{8, 16}, time.Time{}, "")
	if err != nil {
		t.Fatal(err)
	}
	if r.Ben.Name != "坤为地" {
		t.Fatalf("8/16 起卦: got %s want 坤为地", r.Ben.Name)
	}
	if r.Moving != 6 { // (8+16)%6=0,余零作六
		t.Fatalf("动爻: got %d want 6", r.Moving)
	}
	// 三数式:第三数定动爻
	r, _ = ByNumbers([]int{1, 2, 6}, time.Time{}, "")
	if r.Ben.Name != "乾为天" && r.Ben.Upper.Name != "乾" {
		t.Fatalf("上卦: got %s want 乾", r.Ben.Upper.Name)
	}
	if r.Moving != 6 {
		t.Fatalf("三数动爻: got %d want 6", r.Moving)
	}
	if _, err := ByNumbers([]int{0, 5}, time.Time{}, ""); err == nil {
		t.Fatal("非正整数应拒绝")
	}
}

// TestTiYongJudge 体用五行生克断五种关系。
func TestTiYongJudge(t *testing.T) {
	cases := []struct {
		ti, yong string
		want     Relation
	}{
		{"金", "土", RelYongShengTi}, // 土生金
		{"金", "金", RelBiHe},
		{"金", "木", RelTiKeYong},
		{"金", "水", RelTiShengYong},
		{"金", "火", RelYongKeTi},
	}
	for _, c := range cases {
		got, _ := judge(c.ti, c.yong)
		if got != c.want {
			t.Fatalf("体%s 用%s: got %s want %s", c.ti, c.yong, got, c.want)
		}
	}
}

// TestXiaoLiuRen 小六壬:正月初一子时三步皆落大安(起点自检)。
func TestXiaoLiuRen(t *testing.T) {
	// 农历 2024 正月初一子时(0点)
	lunar := calendar.NewLunar(2024, 1, 1, 0, 0, 0)
	solar := lunar.GetSolar()
	tm := time.Date(solar.GetYear(), time.Month(solar.GetMonth()), solar.GetDay(), 0, 30, 0, 0, time.Local)
	r, err := XiaoLiuRen(tm, "")
	if err != nil {
		t.Fatal(err)
	}
	if r.Result.Name != "大安" {
		t.Fatalf("正月初一子时: got %s want 大安(农历 %s)", r.Result.Name, r.LunarText)
	}
	if r.Steps != [3]string{"大安", "大安", "大安"} {
		t.Fatalf("三步落位: %v", r.Steps)
	}
}
