package daliuren

// 由公历时刻起课:日干支(Exact)、占时(时支)、月将(按最近中气换将)。
// 月将换将口径依理法易简:自雨水起亥将,逐中气退一支(见 MonthGeneralByMidQi)。

import (
	"fmt"
	"time"

	"github.com/6tail/lunar-go/calendar"
)

// midQiIndex 中气名 → 中气序(0=雨水…11=大寒)。
var midQiIndex = map[string]int{
	"雨水": 0, "春分": 1, "谷雨": 2, "小满": 3, "夏至": 4, "大暑": 5,
	"处暑": 6, "秋分": 7, "霜降": 8, "小雪": 9, "冬至": 10, "大寒": 11,
}

func runeIndex(set []rune, r rune) int {
	for i, x := range set {
		if x == r {
			return i
		}
	}
	return 0
}

// CastByTime 由公历时刻(含时辰)起课。
func CastByTime(at time.Time) (*Result, error) {
	if at.Year() < 1902 || at.Year() > 2098 {
		return nil, fmt.Errorf("时间超出支持范围(1902-2098)")
	}
	lunar := calendar.NewSolarFromDate(at).GetLunar()
	// 日干支 Exact:夜子时(23 点后)归次日,与四柱/六爻日辰同口径。
	dgz := []rune(lunar.GetDayInGanZhiExact())
	if len(dgz) != 2 {
		return nil, fmt.Errorf("日干支解析失败")
	}
	ds := runeIndex(stems, dgz[0])
	db := runeIndex(branches, dgz[1])
	hour := lunar.GetTimeZhiIndex() // 占时地支索引(0=子…11=亥)

	qi := lunar.GetPrevQi() // 最近中气
	if qi == nil {
		return nil, fmt.Errorf("中气解析失败")
	}
	mi, ok := midQiIndex[qi.GetName()]
	if !ok {
		return nil, fmt.Errorf("月将中气未识别:%s", qi.GetName())
	}
	return Cast(ds, db, hour, MonthGeneralByMidQi(mi))
}
