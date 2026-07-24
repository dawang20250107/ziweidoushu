// 八字大运/流年层:在四柱视角之上排大运(十年一运)与当前大运流年。
// 起运节气精算、顺逆(阳年男/阴年女顺行)借助已依赖的开源历法库 lunar-go
// (EightChar/Yun,与四柱同源同库),再以我方 shiShen/naYin 注入运干十神与
// 纳音。此为「四柱视角」的深度补齐,与紫微同源不独立成盘。
package ziwei

import "github.com/6tail/lunar-go/calendar"

// LiuNianEntry 一年流年。
type LiuNianEntry struct {
	Year        int    `json:"year"`
	Age         int    `json:"age"` // 虚岁
	GanZhi      string `json:"ganZhi"`
	StemShiShen string `json:"stemShiShen"`
	NaYin       string `json:"naYin"`
	IsCurrent   bool   `json:"isCurrent"`
}

// DaYunEntry 一步大运(十年)。
type DaYunEntry struct {
	Index       int    `json:"index"`
	GanZhi      string `json:"ganZhi"`
	StartAge    int    `json:"startAge"` // 虚岁
	StartYear   int    `json:"startYear"`
	StemShiShen string `json:"stemShiShen"` // 运干十神(对日主)
	NaYin       string `json:"naYin"`
	XunKong     string `json:"xunKong"` // 大运旬空
	IsCurrent   bool   `json:"isCurrent"`
}

// DaYunView 大运排布。
type DaYunView struct {
	Forward        bool           `json:"forward"`   // 顺行(true)/逆行
	StartAge       int            `json:"startAge"`  // 起运虚岁(首步)
	StartDesc      string         `json:"startDesc"` // 「X 年 Y 月 Z 天后起运」
	List           []DaYunEntry   `json:"list"`
	CurrentLiuNian []LiuNianEntry `json:"currentLiuNian"` // 当前大运下十年流年
}

func genderCode(g Gender) int {
	if g == Male {
		return 1
	}
	return 0
}

// solarClockHour 时辰索引(0=子…11=亥)→ 排盘用钟点,与 takeCalendar 同口径。
func solarClockHour(timeIndex int) int {
	h := timeIndex*2 - 1
	if h < 0 {
		h = 0
	}
	return h
}

// enrichGanZhi 由干支字符串取运干十神与纳音(dayStem 为日主天干索引)。
func enrichGanZhi(ganZhi string, dayStem int) (shiShenStr, naYinStr string) {
	rs := []rune(ganZhi)
	if len(rs) != 2 {
		return "", ""
	}
	s, b := stemIndex(rs[0]), branchIndex(rs[1])
	if s < 0 || b < 0 {
		return "", ""
	}
	return shiShen(dayStem, s), naYin(s, b)
}

// buildDaYun 排大运与当前大运流年。refYear 为参照公历年,用于按「年」标注当前
// 运/流年(避免八字虚岁与紫微计龄口径不一致的歧义)。
// year/month/day 须传真太阳时校正后的公历日期(与四柱视角同源)。
func buildDaYun(gender Gender, year, month, day, timeIndex, refYear int) *DaYunView {
	solar := calendar.NewSolar(year, month, day, solarClockHour(timeIndex), 30, 0)
	lunar := solar.GetLunar()
	ec := lunar.GetEightChar()
	ec.SetSect(1) // 晚子时日柱归次日,与 takeSiZhuPillars/takeCalendar 日柱同口径
	dayStem := stemIndex([]rune(ec.GetDayGan())[0])
	if dayStem < 0 {
		return nil
	}
	yun := ec.GetYun(genderCode(gender))
	daYuns := yun.GetDaYun() // [0] 为起运前(童限),GanZhi 空

	view := &DaYunView{
		Forward:        yun.IsForward(),
		StartDesc:      formatQiYun(yun),
		List:           []DaYunEntry{},
		CurrentLiuNian: []LiuNianEntry{},
	}

	for _, dy := range daYuns {
		gz := dy.GetGanZhi()
		if gz == "" { // 跳过起运前童限段
			continue
		}
		ss, ny := enrichGanZhi(gz, dayStem)
		start := dy.GetStartAge()
		isCur := refYear >= dy.GetStartYear() && refYear <= dy.GetEndYear()
		if view.StartAge == 0 {
			view.StartAge = start // 首步大运起运虚岁
		}
		view.List = append(view.List, DaYunEntry{
			Index:       dy.GetIndex(),
			GanZhi:      gz,
			StartAge:    start,
			StartYear:   dy.GetStartYear(),
			StemShiShen: ss,
			NaYin:       ny,
			XunKong:     dy.GetXunKong(),
			IsCurrent:   isCur,
		})
		if isCur {
			for _, ln := range dy.GetLiuNian() {
				lgz := ln.GetGanZhi()
				lss, lny := enrichGanZhi(lgz, dayStem)
				view.CurrentLiuNian = append(view.CurrentLiuNian, LiuNianEntry{
					Year:        ln.GetYear(),
					Age:         ln.GetAge(),
					GanZhi:      lgz,
					StemShiShen: lss,
					NaYin:       lny,
					IsCurrent:   ln.GetYear() == refYear,
				})
			}
		}
	}
	return view
}

// formatQiYun 起运描述:「X 年 Y 月 Z 天后起运」。
func formatQiYun(yun *calendar.Yun) string {
	y, m, d := yun.GetStartYear(), yun.GetStartMonth(), yun.GetStartDay()
	s := ""
	if y > 0 {
		s += itoa(y) + " 年"
	}
	if m > 0 {
		s += itoa(m) + " 个月"
	}
	if d > 0 || s == "" {
		s += itoa(d) + " 天"
	}
	return s + "后起运"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
