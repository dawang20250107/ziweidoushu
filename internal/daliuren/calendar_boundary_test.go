package daliuren

import (
	"testing"
	"time"
)

// TestCastByTimeLateZi 夜子时(23 点后)日干支归次日,占时为子:
// 23:30 起课的日干支应与次日上午起课一致。
func TestCastByTimeLateZi(t *testing.T) {
	late, err := CastByTime(time.Date(2024, 6, 15, 23, 30, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	next, err := CastByTime(time.Date(2024, 6, 16, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if late.DayStem != next.DayStem || late.DayBranch != next.DayBranch {
		t.Errorf("夜子时日干支(%s%s)应与次日(%s%s)一致",
			late.DayStem, late.DayBranch, next.DayStem, next.DayBranch)
	}
	if late.HourBranch != "子" {
		t.Errorf("23:30 占时应为子,得 %s", late.HourBranch)
	}
}
