package daliuren

// 活时报数起课测试:报数定占时(自子顺数),日干支/月将仍按实时;正时不受影响。

import (
	"testing"
	"time"
)

func TestBaoShuCast(t *testing.T) {
	at := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	zheng, err := CastByTime(at)
	if err != nil {
		t.Fatal(err)
	}
	if zheng.BaoShu != 0 || zheng.HourNote != "" {
		t.Fatalf("正时不应带报数溯源: %+v", zheng)
	}

	// 报数 7:自子顺数至午((7-1)%12=6)
	r7, err := CastByTimeBaoShu(at, 7)
	if err != nil {
		t.Fatal(err)
	}
	if r7.HourBranch != "午" || r7.BaoShu != 7 || r7.HourNote != "活时·报数7" {
		t.Fatalf("报数7应占午时: %s %d %q", r7.HourBranch, r7.BaoShu, r7.HourNote)
	}
	// 日干支与月将不随报数变
	if r7.DayStem != zheng.DayStem || r7.DayBranch != zheng.DayBranch || r7.MonthGen != zheng.MonthGen {
		t.Fatalf("活时只改占时: %s%s %s vs %s%s %s",
			r7.DayStem, r7.DayBranch, r7.MonthGen, zheng.DayStem, zheng.DayBranch, zheng.MonthGen)
	}
	// 报数 13 绕圈回子
	r13, _ := CastByTimeBaoShu(at, 13)
	if r13.HourBranch != "子" {
		t.Fatalf("报数13应占子时: %s", r13.HourBranch)
	}
	// 相邻报数占时异 → 课骨必异
	r8, _ := CastByTimeBaoShu(at, 8)
	if r7.HourBranch == r8.HourBranch {
		t.Fatal("报数7/8占时不应相同")
	}
	// 非法报数
	if _, err := CastByTimeBaoShu(at, 0); err == nil {
		t.Fatal("报数0应报错")
	}
}
