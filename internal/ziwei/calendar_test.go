package ziwei

import "testing"

// 历法回归:闰月、年界、晚子时、节气月柱、真太阳时跨日。
// 断言值均为公开历表事实(万年历可查),非由被测代码自证。

// TestLeapMonthConversion 闰月识别:2023 癸卯年闰二月、2020 庚子年闰四月。
func TestLeapMonthConversion(t *testing.T) {
	cases := []struct {
		y, m, d    int
		wantMonth  int
		wantDay    int
		wantIsLeap bool
	}{
		{2023, 3, 22, 2, 1, true},  // 闰二月初一
		{2023, 4, 6, 2, 16, true},  // 闰二月十六
		{2023, 2, 20, 2, 1, false}, // 正常二月初一
		{2020, 5, 23, 4, 1, true},  // 闰四月初一
		{2020, 4, 23, 4, 1, false}, // 正常四月初一
	}
	for _, c := range cases {
		snap := takeCalendar(c.y, c.m, c.d, 6)
		if snap.LunarMonth != c.wantMonth || snap.LunarDay != c.wantDay || snap.IsLeap != c.wantIsLeap {
			t.Errorf("%d-%d-%d: got 月%d日%d闰%v, want 月%d日%d闰%v",
				c.y, c.m, c.d, snap.LunarMonth, snap.LunarDay, snap.IsLeap, c.wantMonth, c.wantDay, c.wantIsLeap)
		}
	}
}

// TestLeapMonthPillarCarry 闰月月柱进位(iztro normal 口径):
// 闰月上半月同本月柱,过半(>15)折下月柱。癸卯年五虎遁甲寅起:二月乙卯、三月丙辰。
func TestLeapMonthPillarCarry(t *testing.T) {
	if got := takeCalendar(2023, 3, 31, 6).Pillars.Month; got != "乙卯" { // 闰二月初十
		t.Errorf("闰二月初十月柱: got %s want 乙卯", got)
	}
	if got := takeCalendar(2023, 4, 6, 6).Pillars.Month; got != "丙辰" { // 闰二月十六
		t.Errorf("闰二月十六月柱: got %s want 丙辰", got)
	}
}

// TestYearDivideChuYiVsLiChun 年界两派口径并存:
// 1987 年春节 1 月 29 日、立春 2 月 4 日。2 月 1 日介于其间——
// 紫微(正月初一分界)已入丁卯,四柱视角(立春分界)仍属丙寅。
func TestYearDivideChuYiVsLiChun(t *testing.T) {
	snap := takeCalendar(1987, 2, 1, 6)
	if snap.Pillars.Year != "丁卯" {
		t.Errorf("紫微年柱(初一分界): got %s want 丁卯", snap.Pillars.Year)
	}
	sz := takeSiZhuPillars(1987, 2, 1, 6)
	if sz.Year != "丙寅" {
		t.Errorf("四柱视角年柱(立春分界): got %s want 丙寅", sz.Year)
	}
	// 初一之前两派一致:1 月 28 日均为丙寅。
	if got := takeCalendar(1987, 1, 28, 6).Pillars.Year; got != "丙寅" {
		t.Errorf("初一前紫微年柱: got %s want 丙寅", got)
	}
	// 立春之后两派再度一致:2 月 5 日均为丁卯。
	if got := takeSiZhuPillars(1987, 2, 5, 6).Year; got != "丁卯" {
		t.Errorf("立春后四柱年柱: got %s want 丁卯", got)
	}
}

// TestMonthPillarJieQiDivide 四柱视角月柱以节分界:
// 2024 年惊蛰在 3 月 5 日 10:23——3 月 4 日仍寅月(丙寅)、3 月 6 日已卯月(丁卯)。
func TestMonthPillarJieQiDivide(t *testing.T) {
	if got := takeSiZhuPillars(2024, 3, 4, 6).Month; got != "丙寅" {
		t.Errorf("惊蛰前月柱: got %s want 丙寅", got)
	}
	if got := takeSiZhuPillars(2024, 3, 6, 6).Month; got != "丁卯" {
		t.Errorf("惊蛰后月柱: got %s want 丁卯", got)
	}
	// 惊蛰当日交节时刻前(辰时 07-09 点)仍属寅月:节气分界精确到时刻。
	if got := takeSiZhuPillars(2024, 3, 5, 4).Month; got != "丙寅" {
		t.Errorf("惊蛰日交节前月柱: got %s want 丙寅", got)
	}
	// 交节时刻后(午时)入卯月。
	if got := takeSiZhuPillars(2024, 3, 5, 6).Month; got != "丁卯" {
		t.Errorf("惊蛰日交节后月柱: got %s want 丁卯", got)
	}
}

// TestLateZiDayPillar 晚子时日柱归次日(Exact 口径),两套四柱一致:
// 1990-6-15 日柱辛亥;晚子时(timeIndex=12)进为壬子;早子时(0)不进。
func TestLateZiDayPillar(t *testing.T) {
	if got := takeCalendar(1990, 6, 15, 0).Pillars.Day; got != "辛亥" {
		t.Errorf("早子时日柱: got %s want 辛亥", got)
	}
	if got := takeCalendar(1990, 6, 15, 12).Pillars.Day; got != "壬子" {
		t.Errorf("晚子时日柱(紫微): got %s want 壬子", got)
	}
	if got := takeSiZhuPillars(1990, 6, 15, 12).Day; got != "壬子" {
		t.Errorf("晚子时日柱(四柱视角): got %s want 壬子", got)
	}
}

// TestTrueSolarCrossMidnight 真太阳时跨日退回:乌鲁木齐(东经 87.6°)子夜 0 时半,
// 经度差约 -130 分钟 → 退回前日亥时;整盘等价于按前一日亥时排。
func TestTrueSolarCrossMidnight(t *testing.T) {
	adjIdx, dayDelta, note := AdjustHourByLongitude(2024, 6, 15, 0, 87.6, 0)
	if dayDelta != -1 || adjIdx != 11 {
		t.Fatalf("真太阳时校正: got 时辰%d 日差%d(%s), want 时辰11(亥) 日差-1", adjIdx, dayDelta, note)
	}
	adjusted, err := Generate(BirthInfo{Year: 2024, Month: 6, Day: 15, Hour: 0, Gender: Male, Longitude: 87.6},
		Options{ReferenceYear: 2024, TrueSolarTime: true})
	if err != nil {
		t.Fatal(err)
	}
	plain, err := Generate(BirthInfo{Year: 2024, Month: 6, Day: 14, Hour: 11, Gender: Male}, Options{ReferenceYear: 2024})
	if err != nil {
		t.Fatal(err)
	}
	if adjusted.FourPillars != plain.FourPillars {
		t.Errorf("跨日校正后四柱 %v ≠ 前日亥时四柱 %v", adjusted.FourPillars, plain.FourPillars)
	}
	if adjusted.LunarDateText != plain.LunarDateText {
		t.Errorf("跨日校正后农历 %q ≠ 前日亥时农历 %q", adjusted.LunarDateText, plain.LunarDateText)
	}
}

// TestSiZhuUsesJieQiPillars 四柱视角整链:1987-2-1 生人取格月令应为丑月
// (小寒后大寒中,立春未到),而非紫微初一口径的正月;大运年柱基准同为丙寅。
func TestSiZhuUsesJieQiPillars(t *testing.T) {
	c, err := Generate(BirthInfo{Year: 1987, Month: 2, Day: 1, Hour: 6, Gender: Male}, Options{ReferenceYear: 2024})
	if err != nil {
		t.Fatal(err)
	}
	if c.SiZhu == nil {
		t.Fatal("SiZhu 为空")
	}
	if got := c.SiZhu.Pillars[0].Stem + c.SiZhu.Pillars[0].Branch; got != "丙寅" {
		t.Errorf("四柱视角年柱: got %s want 丙寅(立春分界)", got)
	}
	if got := c.SiZhu.Pillars[1].Branch; got != "丑" {
		t.Errorf("四柱视角月支: got %s want 丑(立春前属丑月)", got)
	}
	// 紫微盘面年柱保持初一口径不受影响(黄金基准锚定)。
	if c.FourPillars.Year != "丁卯" {
		t.Errorf("紫微年柱: got %s want 丁卯", c.FourPillars.Year)
	}
	if c.SiZhu.Note == "" {
		t.Error("四柱视角应携带历法口径说明")
	}
}
