// 农历生日输入:农历(含闰月)→ 公历换算与农历年月表。
// 委托 lunar-go(与四柱/大运同库同源,亦与黄金基准的 lunar-javascript 同作者同表,
// 1900-2100 全区间口径一致),引擎内部仍以公历为唯一事实源。
package ziwei

import (
	"fmt"

	"github.com/6tail/lunar-go/calendar"
)

// LunarMonthMeta 农历某年的一个月:Month 为 1-12,Leap 表示闰月,Days 为当月天数(29/30)。
type LunarMonthMeta struct {
	Month int  `json:"month"`
	Leap  bool `json:"leap"`
	Days  int  `json:"days"`
}

// LunarYearMonths 某农历年的逐月表(闰月按其在年内的实际位置插入)。
// 供前端渲染月/日选择器:哪年有闰几月、每月 29 还是 30 天,以此为准。
func LunarYearMonths(year int) ([]LunarMonthMeta, error) {
	if year < 1900 || year > 2100 {
		return nil, fmt.Errorf("农历年份须在 1900-2100 之间")
	}
	ly := calendar.NewLunarYear(year)
	var out []LunarMonthMeta
	for e := ly.GetMonthsInYear().Front(); e != nil; e = e.Next() {
		m := e.Value.(*calendar.LunarMonth)
		mm := m.GetMonth() // 负数 = 闰月
		meta := LunarMonthMeta{Month: mm, Leap: mm < 0, Days: m.GetDayCount()}
		if meta.Leap {
			meta.Month = -mm
		}
		out = append(out, meta)
	}
	return out, nil
}

// LunarToSolar 农历生日 → 公历(闰月以 leap 标记)。
// 逐项校验:年份区间、月份存在(含该年是否真有此闰月)、日不超过当月天数。
func LunarToSolar(year, month int, leap bool, day int) (sy, sm, sd int, err error) {
	months, err := LunarYearMonths(year)
	if err != nil {
		return 0, 0, 0, err
	}
	var meta *LunarMonthMeta
	for i := range months {
		if months[i].Month == month && months[i].Leap == leap {
			meta = &months[i]
			break
		}
	}
	if meta == nil {
		if leap {
			return 0, 0, 0, fmt.Errorf("%d 年没有闰%d月", year, month)
		}
		return 0, 0, 0, fmt.Errorf("农历月份无效: %d", month)
	}
	if day < 1 || day > meta.Days {
		return 0, 0, 0, fmt.Errorf("该月只有 %d 天", meta.Days)
	}
	lm := month
	if leap {
		lm = -month
	}
	solar := calendar.NewLunarFromYmd(year, lm, day).GetSolar()
	return solar.GetYear(), solar.GetMonth(), solar.GetDay(), nil
}
