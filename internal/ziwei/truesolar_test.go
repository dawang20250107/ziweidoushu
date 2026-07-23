package ziwei

import (
	"math"
	"testing"
)

// TestEquationOfTime 均时差幅度:二月初约 −14 分、十一月初约 +16 分(全年极值)。
func TestEquationOfTime(t *testing.T) {
	feb := equationOfTimeMinutes(2024, 2, 11)
	nov := equationOfTimeMinutes(2024, 11, 3)
	if feb > -10 {
		t.Errorf("二月初均时差应约 −14 分,得 %.1f", feb)
	}
	if nov < 10 {
		t.Errorf("十一月初均时差应约 +16 分,得 %.1f", nov)
	}
	if math.Abs(equationOfTimeMinutes(2024, 4, 16)) > 3 { // 四月中约 0
		t.Errorf("四月中均时差应近 0,得 %.1f", equationOfTimeMinutes(2024, 4, 16))
	}
}

// TestAdjustSolarTime 真太阳时校正:经度差 + 均时差,并处理跨日界。
func TestAdjustSolarTime(t *testing.T) {
	// 经度 0:不校正
	if idx, dd, note := AdjustHourByLongitude(2024, 6, 1, 3, 0); idx != 3 || dd != 0 || note != "" {
		t.Errorf("经度0应原样返回,得 idx=%d dd=%d", idx, dd)
	}
	// 新疆(东经~87°)早子时:真太阳时约退回前一日亥时。
	idx, dd, note := AdjustHourByLongitude(2024, 6, 1, 0, 87.0)
	if idx != 11 || dd != -1 {
		t.Errorf("新疆早子真太阳时应退前日亥时(idx=11,dd=-1),得 idx=%d dd=%d", idx, dd)
	}
	if note == "" {
		t.Error("跨日校正应有说明文字")
	}
}

// TestTrueSolarDateCarry 端到端:真太阳时跨日界时,日柱随公历日期进退而改。
func TestTrueSolarDateCarry(t *testing.T) {
	b := BirthInfo{Year: 2024, Month: 6, Day: 1, Hour: 0, Gender: Male, Longitude: 87.0}
	naive, err := Generate(b, Options{})
	if err != nil {
		t.Fatal(err)
	}
	corrected, err := Generate(b, Options{TrueSolarTime: true})
	if err != nil {
		t.Fatal(err)
	}
	// 校正后落前一日亥时:四柱日柱应与原盘不同(日期退一天)。
	if corrected.FourPillars.Day == naive.FourPillars.Day {
		t.Errorf("真太阳时跨日应改日柱:naive=%s corrected=%s",
			naive.FourPillars.Day, corrected.FourPillars.Day)
	}
	// 与直接排「前一日亥时(hour=11)」应一致(日柱与命宫)。
	prev := BirthInfo{Year: 2024, Month: 5, Day: 31, Hour: 11, Gender: Male}
	ref, err := Generate(prev, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if corrected.FourPillars.Day != ref.FourPillars.Day {
		t.Errorf("校正盘日柱应等于前日亥时盘:corrected=%s ref=%s",
			corrected.FourPillars.Day, ref.FourPillars.Day)
	}
}
