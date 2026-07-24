package liuyao

import (
	"strings"
	"testing"
	"time"
)

// 历法边界回归:日辰夜子时归次日(Exact)、月建以节交接时刻分界。

// TestDayGanZhiLateZi 夜子时(23 点后)日辰归次日:
// 23:30 起卦的日辰应与次日上午起卦一致,且比 22:30 的日辰进一位。
func TestDayGanZhiLateZi(t *testing.T) {
	before, _ := time.Parse("2006-01-02 15:04", "2024-06-15 22:30")
	late, _ := time.Parse("2006-01-02 15:04", "2024-06-15 23:30")
	next, _ := time.Parse("2006-01-02 15:04", "2024-06-16 10:00")

	s1, b1, _, _, err := dayGanZhi(before)
	if err != nil {
		t.Fatal(err)
	}
	s2, b2, _, text2, err := dayGanZhi(late)
	if err != nil {
		t.Fatal(err)
	}
	s3, b3, _, _, err := dayGanZhi(next)
	if err != nil {
		t.Fatal(err)
	}
	if s2 != s3 || b2 != b3 {
		t.Errorf("夜子时日辰(%d,%d)应与次日(%d,%d)一致", s2, b2, s3, b3)
	}
	if s2 != (s1+1)%10 || b2 != (b1+1)%12 {
		t.Errorf("夜子时日辰(%d,%d)应比 22:30(%d,%d)进一位", s2, b2, s1, b1)
	}
	if !strings.Contains(text2, "夜子时") {
		t.Errorf("夜子时起卦应有归次日说明,得 %q", text2)
	}
}

// TestMonthJianJieQiExact 月建以节交接时刻分界:
// 2024 年立春在 2 月 4 日 16:27——当日上午仍丑月建,入夜已寅月建。
func TestMonthJianJieQiExact(t *testing.T) {
	morning, _ := time.Parse("2006-01-02 15:04", "2024-02-04 10:00")
	evening, _ := time.Parse("2006-01-02 15:04", "2024-02-04 18:00")
	_, _, m1, _, err := dayGanZhi(morning)
	if err != nil {
		t.Fatal(err)
	}
	_, _, m2, _, err := dayGanZhi(evening)
	if err != nil {
		t.Fatal(err)
	}
	if m1 != '丑' {
		t.Errorf("立春交节前月建: got %c want 丑", m1)
	}
	if m2 != '寅' {
		t.Errorf("立春交节后月建: got %c want 寅", m2)
	}
}
