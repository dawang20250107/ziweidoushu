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

// CastByTime 由公历时刻(含时辰)正时起课。
// 注意:正时之课同一时辰人人相同(古法本然,断案诸例皆正时,以年命分断);
// 产品层众人同刻求各异之课,用 CastByTimeBaoShu 活时报数。
func CastByTime(at time.Time) (*Result, error) {
	return castByTimeHour(at, -1, 0)
}

// CastByTimeBaoShu 活时报数起课:占时不取正时,以报数自子顺数所至之支
// ((n-1)%12)为占时;日干支(Exact)与月将仍按实时。
func CastByTimeBaoShu(at time.Time, n int) (*Result, error) {
	if n <= 0 {
		return nil, fmt.Errorf("报数须为正整数")
	}
	return castByTimeHour(at, (n-1)%12, n)
}

// castByTimeHour 起课内核;hourOverride<0 取正时,否则以其为占时支(活时)。
func castByTimeHour(at time.Time, hourOverride, baoShu int) (*Result, error) {
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
	if hourOverride >= 0 {
		hour = hourOverride
	}

	qi := lunar.GetPrevQi() // 最近中气
	if qi == nil {
		return nil, fmt.Errorf("中气解析失败")
	}
	mi, ok := midQiIndex[qi.GetName()]
	if !ok {
		return nil, fmt.Errorf("月将中气未识别:%s", qi.GetName())
	}
	r, err := Cast(ds, db, hour, MonthGeneralByMidQi(mi))
	if err == nil && baoShu > 0 {
		r.BaoShu = baoShu
		r.HourNote = fmt.Sprintf("活时·报数%d", baoShu)
	}
	return r, err
}
