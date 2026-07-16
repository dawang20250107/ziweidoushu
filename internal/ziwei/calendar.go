package ziwei

import (
	"fmt"

	"github.com/6tail/lunar-go/calendar"
)

// calendarSnapshot 一次公历输入对应的全部历法派生量。
// 口径与 iztro 依赖的 lunar-lite(底层 lunar-typescript,与 lunar-go 同源)完全一致:
//   - 年柱按农历正月初一分界(iztro 默认 yearDivide='normal')
//   - 月柱按初一分界的五虎遁(iztro 默认 horoscopeDivide='normal')
//   - 日柱用 Exact 口径(23:00 后算次日)
//   - 干支取样时间为 max(timeIndex*2-1, 0) 时 30 分
type calendarSnapshot struct {
	LunarYear  int
	LunarMonth int // 恒为正,闰月看 IsLeap
	LunarDay   int
	IsLeap     bool

	YearStem   int // 年柱天干索引(正月初一分界)
	YearBranch int // 年柱地支索引

	Pillars   FourPillars
	LunarText string // 如「一九九〇年五月廿三」
	XingZuo   string // 星座
}

// takeCalendar 计算公历生日在 timeIndex 时辰下的历法快照。
func takeCalendar(year, month, day, timeIndex int) calendarSnapshot {
	hour := timeIndex*2 - 1
	if hour < 0 {
		hour = 0
	}
	solar := calendar.NewSolar(year, month, day, hour, 30, 0)
	lunar := solar.GetLunar()

	rawMonth := lunar.GetMonth()
	snap := calendarSnapshot{
		LunarYear:  lunar.GetYear(),
		LunarMonth: abs(rawMonth),
		LunarDay:   lunar.GetDay(),
		IsLeap:     rawMonth < 0,
		YearStem:   stemIndexOf(lunar.GetYearGan()),
		YearBranch: branchIndexOf(lunar.GetYearZhi()),
		LunarText:  lunar.String(),
		XingZuo:    solar.GetXingZuo() + "座",
	}

	// 月柱:初一分界口径 —— 五虎遁 + 闰月下半月进位(lunar-lite calculateMonthlyGanZhi 'normal' 分支)。
	leapAdd := 0
	if snap.IsLeap && snap.LunarDay > 15 {
		leapAdd = 1
	}
	monthStem := fix10(tigerRule[snap.YearStem] + snap.LunarMonth - 1 + leapAdd)
	monthBranch := fix12(2 + snap.LunarMonth - 1 + leapAdd) // 正月建寅
	snap.Pillars = FourPillars{
		Year:  lunar.GetYearGan() + lunar.GetYearZhi(),
		Month: Stems[monthStem] + Branches[monthBranch],
		Day:   lunar.GetDayGanExact() + lunar.GetDayZhiExact(),
		Hour:  lunar.GetTimeGan() + lunar.GetTimeZhi(),
	}
	return snap
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// AdjustHourByLongitude 真太阳时校正:按出生地东经度数与东经 120°(北京时间基准)的差值
// 折算时差(每 1° 差 4 分钟),返回校正后的时辰索引(0-11)。
// 返回值另含校正说明,方便 API 层回显。
func AdjustHourByLongitude(year, month, day, hourIndex int, longitude float64) (int, string) {
	if longitude == 0 {
		return hourIndex, ""
	}
	// 时辰中点钟点:子时取 0 点,其余取区间中点(奇数点)。
	midHour := hourIndex * 2
	offsetMinutes := (longitude - 120.0) * 4
	totalMinutes := midHour*60 + int(offsetMinutes)
	// 归一化到一天内(跨日仅影响时辰归属,排盘按时辰索引进行)。
	totalMinutes = ((totalMinutes % 1440) + 1440) % 1440
	adjHour := totalMinutes / 60
	var adjIndex int
	if adjHour == 23 || adjHour == 0 {
		adjIndex = 0
	} else {
		adjIndex = (adjHour + 1) / 2
	}
	note := fmt.Sprintf("按东经 %.1f° 真太阳时校正 %.0f 分钟:%s → %s",
		longitude, offsetMinutes, ChineseTimeNames[min(hourIndex, 12)], ChineseTimeNames[min(adjIndex, 12)])
	if adjIndex == hourIndex {
		note = ""
	}
	return adjIndex, note
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
