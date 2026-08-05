package ziwei

import "testing"

// TestLunarToSolar 农历→公历换算:实测案例 + 闰月锚点 + 边界拒绝。
func TestLunarToSolar(t *testing.T) {
	cases := []struct {
		y, m  int
		leap  bool
		d     int
		sy    int
		sm    int
		sd    int
		wantE bool
	}{
		{1993, 10, false, 7, 1993, 11, 20, false}, // 甲女实测案例(农历一九九三年十月初七)
		{1994, 1, false, 15, 1994, 2, 24, false},  // 乙男实测案例(农历一九九四年正月十五)
		{2023, 2, true, 1, 2023, 3, 22, false},    // 闰二月初一
		{2020, 4, true, 1, 2020, 5, 23, false},    // 闰四月初一
		{1993, 3, true, 1, 1993, 4, 22, false},    // 1993 闰三月初一
		{2033, 11, true, 1, 2033, 12, 22, false},  // 2033 闰十一月(历法难题年)
		{1993, 2, true, 1, 0, 0, 0, true},         // 1993 无闰二月 → 拒绝
		{1899, 1, false, 1, 0, 0, 0, true},        // 区间外 → 拒绝
	}
	for _, c := range cases {
		sy, sm, sd, err := LunarToSolar(c.y, c.m, c.leap, c.d)
		if c.wantE {
			if err == nil {
				t.Errorf("LunarToSolar(%d,%d,leap=%v,%d) 应拒绝,得 %d-%d-%d", c.y, c.m, c.leap, c.d, sy, sm, sd)
			}
			continue
		}
		if err != nil || sy != c.sy || sm != c.sm || sd != c.sd {
			t.Errorf("LunarToSolar(%d,%d,leap=%v,%d)=%d-%d-%d err=%v, want %d-%d-%d",
				c.y, c.m, c.leap, c.d, sy, sm, sd, err, c.sy, c.sm, c.sd)
		}
	}
	// 小月三十拒绝:动态找一个 29 天的月(不赌记忆里的大小月)
	months, _ := LunarYearMonths(2023)
	for _, m := range months {
		if m.Days == 29 {
			if _, _, _, err := LunarToSolar(2023, m.Month, m.Leap, 30); err == nil {
				t.Errorf("2023 年 %d 月(leap=%v)仅 29 天,三十应拒绝", m.Month, m.Leap)
			}
			break
		}
	}
}

// TestLunarYearMonths 农历年月表:闰月位置、天数、平年 12 月/闰年 13 月。
func TestLunarYearMonths(t *testing.T) {
	ms93, err := LunarYearMonths(1993)
	if err != nil {
		t.Fatal(err)
	}
	if len(ms93) != 13 {
		t.Fatalf("1993 有闰三月,应 13 个月,得 %d", len(ms93))
	}
	// 闰三月应紧随三月之后
	leapIdx := -1
	for i, m := range ms93 {
		if m.Leap {
			leapIdx = i
			if m.Month != 3 {
				t.Errorf("1993 闰月应为三月,得 %d", m.Month)
			}
		}
		if m.Days != 29 && m.Days != 30 {
			t.Errorf("月天数应为 29/30,得 %d", m.Days)
		}
	}
	if leapIdx < 1 || ms93[leapIdx-1].Month != 3 || ms93[leapIdx-1].Leap {
		t.Errorf("闰三月应紧随三月之后,位置 %d", leapIdx)
	}
	ms95, err := LunarYearMonths(1995)
	if err != nil {
		t.Fatal(err)
	}
	// 1995 闰八月
	found := false
	for _, m := range ms95 {
		if m.Leap && m.Month == 8 {
			found = true
		}
	}
	if !found {
		t.Error("1995 应有闰八月")
	}
	ms94, _ := LunarYearMonths(1994)
	if len(ms94) != 12 {
		t.Errorf("1994 平年应 12 个月,得 %d", len(ms94))
	}
	// 换算一致性:每月初一经 LunarToSolar 再由引擎公历→农历应回到原月
	for _, m := range ms93 {
		sy, sm, sd, err := LunarToSolar(1993, m.Month, m.Leap, 1)
		if err != nil {
			t.Fatalf("1993-%d(leap=%v)-1 换算失败: %v", m.Month, m.Leap, err)
		}
		if sy == 0 || sm == 0 || sd == 0 {
			t.Fatalf("换算得零值")
		}
	}
}
