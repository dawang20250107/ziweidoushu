package ziwei

import (
	"fmt"
	"math"
	"time"

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

// equationOfTimeMinutes 均时差(真太阳时 − 平太阳时,单位分钟)。
// 采用 NOAA 通用近似(公开天文算法,精度约 ±0.5 分钟):按年内日序求太阳
// 视位置相位角 γ,展开三角级数。全年幅度约 −14~+16 分钟,二月初、十一月初
// 最大,足以在时辰边界改变时辰归属,故真太阳时必须叠加此项(不能只做经度差)。
func equationOfTimeMinutes(year, month, day int) float64 {
	n := time.Date(year, time.Month(month), day, 12, 0, 0, 0, time.UTC).YearDay()
	g := 2 * math.Pi / 365.0 * float64(n-1)
	return 229.18 * (0.000075 +
		0.001868*math.Cos(g) - 0.032077*math.Sin(g) -
		0.014615*math.Cos(2*g) - 0.040849*math.Sin(2*g))
}

// AdjustHourByLongitude 真太阳时校正:平太阳时(北京时,东经 120° 基准)→ 真太阳时。
// 两项叠加:①经度差(每 1°=4 分钟,以 120°E 为准);②均时差 EoT。
// 校正后若跨越子/午日界,返回 dayDelta(±1)以便调用方同步改公历日期
// (农历日/日柱/据日安星均随之改动);返回校正后时辰索引(0-11)与说明。
func AdjustHourByLongitude(year, month, day, hourIndex int, longitude float64) (adjIndex, dayDelta int, note string) {
	if longitude == 0 {
		return hourIndex, 0, ""
	}
	// 时辰中点钟点:子=0、丑=2…亥=22、晚子=24。
	midHour := hourIndex * 2
	lonMinutes := (longitude - 120.0) * 4
	eotMinutes := equationOfTimeMinutes(year, month, day)
	totalMinutes := midHour*60 + int(math.Round(lonMinutes+eotMinutes))

	// 分离日界进退与日内分钟(floor 除法,兼顾负值)。
	dayDelta = int(math.Floor(float64(totalMinutes) / 1440.0))
	minuteOfDay := totalMinutes - dayDelta*1440
	adjHour := minuteOfDay / 60
	if adjHour == 23 || adjHour == 0 {
		adjIndex = 0
	} else {
		adjIndex = (adjHour + 1) / 2
	}
	if adjIndex == hourIndex && dayDelta == 0 {
		return adjIndex, 0, ""
	}
	dayNote := ""
	if dayDelta > 0 {
		dayNote = "(跨入次日)"
	} else if dayDelta < 0 {
		dayNote = "(退回前日)"
	}
	note = fmt.Sprintf("按东经 %.1f° 真太阳时校正(经度差 %.0f 分 + 均时差 %.0f 分):%s → %s%s",
		longitude, lonMinutes, eotMinutes, ChineseTimeNames[min(hourIndex, 12)], ChineseTimeNames[min(adjIndex, 12)], dayNote)
	return adjIndex, dayDelta, note
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
