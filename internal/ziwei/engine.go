package ziwei

import (
	"errors"
	"fmt"
	"time"

	"github.com/6tail/lunar-go/calendar"
)

// Options 排盘选项。
type Options struct {
	// ReferenceYear 计算「当前年龄/当前大限」的参考年份;0 表示取系统当前年。
	ReferenceYear int
	// TrueSolarTime 是否按出生地经度做真太阳时校正(需要 BirthInfo.Longitude)。
	TrueSolarTime bool
}

// Generate 依据公历生辰生成完整命盘。
//
// 算法口径与 iztro 2.5.8 默认配置(《紫微斗数全书》安星法、年界正月初一、
// fixLeap=true 闰月过半折下月)逐字段对齐,由黄金基准回归测试保障。
func Generate(b BirthInfo, opt Options) (*Chart, error) {
	if err := validate(b); err != nil {
		return nil, err
	}

	timeIndex := b.Hour
	var solarTimeNote string
	if opt.TrueSolarTime && b.Longitude != 0 {
		timeIndex, solarTimeNote = AdjustHourByLongitude(b.Year, b.Month, b.Day, b.Hour, b.Longitude)
		_ = solarTimeNote
	}

	snap := takeCalendar(b.Year, b.Month, b.Day, timeIndex)

	// ── 月支索引(寅=0)与命身宫 ──────────────────────────────
	// 闰月过半折算下月;晚子时(timeIndex=12)不折算。
	leapAdd := 0
	if snap.IsLeap && snap.LunarDay > 15 && timeIndex != 12 {
		leapAdd = 1
	}
	monthIndex := fix12(snap.LunarMonth - 1 + leapAdd)
	timeBranch := fix12(timeIndex) // 时辰地支索引(12=晚子时→子)

	soulIndex := fix12(monthIndex - timeBranch) // 命宫(宫位索引,寅=0)
	bodyIndex := fix12(monthIndex + timeBranch) // 身宫

	// ── 十二宫天干(五虎遁)──────────────────────────────────
	startStem := tigerRule[snap.YearStem]
	soulStem := fix10(startStem + soulIndex)
	soulBranch := palaceIndexToBranch(soulIndex)

	// ── 五行局(命宫干支纳音口诀)────────────────────────────
	ju := fiveElementsJu(soulStem, soulBranch)

	// ── 紫微 / 天府 定位 ─────────────────────────────────────
	ziweiIdx, tianfuIdx := ziweiTianfuIndex(snap, timeIndex, ju)

	// ── 逐宫安星 ─────────────────────────────────────────────
	stars := make([][]Star, 12)
	placeMajorStars(stars, ziweiIdx, tianfuIdx, snap.YearStem)
	placeMinorStars(stars, snap, monthIndex, timeBranch)
	placeAdjectiveStars(stars, snap, soulIndex, bodyIndex, monthIndex, timeBranch, timeIndex)

	// ── 长生十二神 / 博士十二神 ─────────────────────────────
	clockwise := genderYinYang(b.Gender) == snap.YearBranch%2 // 阳男阴女顺行
	cs12 := changsheng12Seq(ju, clockwise)
	bs12 := boshi12Seq(snap.YearStem, clockwise)

	// ── 大限与小限 ───────────────────────────────────────────
	decadalStart := make([]int, 12) // 每宫大限起始虚岁
	for i := 0; i < 12; i++ {
		var idx int
		if clockwise {
			idx = fix12(soulIndex + i)
		} else {
			idx = fix12(soulIndex - i)
		}
		decadalStart[idx] = ju + 10*i
	}
	agesByPalace := make([][]int, 12)
	ageStart := xiaoxianStartIndex(snap.YearBranch)
	for i := 0; i < 12; i++ {
		seq := make([]int, 10)
		for j := 0; j < 10; j++ {
			seq[j] = 12*j + i + 1
		}
		var idx int
		if b.Gender == Male {
			idx = fix12(ageStart + i)
		} else {
			idx = fix12(ageStart - i)
		}
		agesByPalace[idx] = seq
	}

	// ── 组装十二宫(输出按地支索引 0-11 排序)────────────────
	palaces := make([]Palace, 12)
	for i := 0; i < 12; i++ { // i 为宫位索引(寅=0)
		branch := palaceIndexToBranch(i)
		if stars[i] == nil {
			stars[i] = []Star{} // 契约:空宫输出 [] 而非 null(前端可直接迭代)
		}
		p := Palace{
			Branch:         branch,
			Stem:           fix10(startStem + i),
			Name:           PalaceNamesClockwise[fix12(i-soulIndex)],
			Stars:          stars[i],
			DaXianStart:    decadalStart[i],
			DaXianEnd:      decadalStart[i] + 9,
			IsMingGong:     i == soulIndex,
			IsShenGong:     i == bodyIndex,
			OppositeBranch: fix12(branch + 6),
			Changsheng12:   cs12[i],
			Boshi12:        bs12[i],
			Ages:           agesByPalace[i],
		}
		palaces[branch] = p
	}

	// ── 空宫借对宫主星 ───────────────────────────────────────
	for i := range palaces {
		p := &palaces[i]
		majors := p.MajorStarNames()
		p.IsEmpty = len(majors) == 0
		if p.IsEmpty {
			opp := &palaces[p.OppositeBranch]
			p.BorrowedFromBranch = opp.Branch
			p.BorrowedFromName = opp.Name
			p.BorrowedStars = opp.MajorStarNames()
		}
	}

	// ── 大限数组(按起始年龄升序)────────────────────────────
	daXians := make([]DaXian, 0, 12)
	for i := 0; i < 12; i++ {
		var idx int
		if clockwise {
			idx = fix12(soulIndex + i)
		} else {
			idx = fix12(soulIndex - i)
		}
		branch := palaceIndexToBranch(idx)
		daXians = append(daXians, DaXian{
			StartAge:     ju + 10*i,
			EndAge:       ju + 10*i + 9,
			PalaceBranch: branch,
			PalaceName:   palaces[branch].Name,
		})
	}

	refYear := opt.ReferenceYear
	if refYear == 0 {
		refYear = time.Now().Year()
	}
	currentAge := refYear - b.Year
	currentDX := -1
	for i, dx := range daXians {
		if currentAge >= dx.StartAge && currentAge <= dx.EndAge {
			currentDX = i
			break
		}
	}

	chart := &Chart{
		BirthInfo:     b,
		LunarInfo:     LunarInfo{LunarYear: snap.LunarYear, LunarMonth: snap.LunarMonth, LunarDay: snap.LunarDay, YearStem: snap.YearStem, YearBranch: snap.YearBranch, IsLeapMonth: snap.IsLeap},
		LunarDateText: snap.LunarText,
		FourPillars:   snap.Pillars,
		TimeName:      ChineseTimeNames[min(timeIndex, 12)],
		Zodiac:        zodiacByBranch[snap.YearBranch],
		Sign:          snap.XingZuo,

		MingGongBranch: soulBranch,
		ShenGongBranch: palaceIndexToBranch(bodyIndex),
		MingZhu:        soulStarByBranch[soulBranch],
		ShenZhu:        bodyStarByBranch[snap.YearBranch],
		WuxingJu:       ju,
		WuxingJuName:   juNames[ju],
		ZiweiPos:       palaceIndexToBranch(ziweiIdx),

		Palaces: palaces,
		DaXians: daXians,

		ReferenceYear:      refYear,
		CurrentAge:         currentAge,
		CurrentDaXianIndex: currentDX,
	}
	return chart, nil
}

func validate(b BirthInfo) error {
	if b.Year < 1900 || b.Year > 2100 {
		return fmt.Errorf("出生年份 %d 超出支持范围(1900-2100)", b.Year)
	}
	if b.Month < 1 || b.Month > 12 {
		return fmt.Errorf("月份 %d 非法", b.Month)
	}
	if b.Day < 1 || b.Day > 31 {
		return fmt.Errorf("日期 %d 非法", b.Day)
	}
	t := time.Date(b.Year, time.Month(b.Month), b.Day, 0, 0, 0, 0, time.UTC)
	if t.Year() != b.Year || int(t.Month()) != b.Month || t.Day() != b.Day {
		return fmt.Errorf("公历日期 %d-%d-%d 不存在", b.Year, b.Month, b.Day)
	}
	if b.Hour < 0 || b.Hour > 12 {
		return errors.New("时辰索引需在 0-12 之间(0=早子时,11=亥时,12=晚子时)")
	}
	if b.Gender != Male && b.Gender != Female {
		return errors.New("性别需为 male 或 female")
	}
	return nil
}

// genderYinYang 男=0(阳),女=1(阴);与地支索引奇偶(偶=阳)相比对。
func genderYinYang(g Gender) int {
	if g == Male {
		return 0
	}
	return 1
}

// fiveElementsJu 由命宫干支按纳音口诀定五行局。
// 干支取数相加,超五减五:1木三局 2金四局 3水二局 4火六局 5土五局。
func fiveElementsJu(stem, branch int) int {
	stemNumber := stem/2 + 1
	branchNumber := (branch%6)/2 + 1
	n := stemNumber + branchNumber
	for n > 5 {
		n -= 5
	}
	return [...]int{3, 4, 2, 6, 5}[n-1]
}

// ziweiTianfuIndex 起紫微星诀 + 天府镜像。返回宫位索引(寅=0)。
func ziweiTianfuIndex(snap calendarSnapshot, timeIndex, ju int) (int, int) {
	// 晚子时加一天(默认 dayDivide='forward'),跨月归位。
	day := snap.LunarDay
	if timeIndex == 12 {
		day++
	}
	m := snap.LunarMonth
	if snap.IsLeap {
		m = -m
	}
	maxDays := calendar.NewLunarMonthFromYm(snap.LunarYear, m).GetDayCount()
	if day > maxDays {
		day -= maxDays
	}

	offset := -1
	quotient := 0
	remainder := -1
	for remainder != 0 {
		offset++
		divisor := day + offset
		quotient = divisor / ju
		remainder = divisor % ju
	}
	quotient %= 12
	ziwei := quotient - 1
	if offset%2 == 0 {
		ziwei += offset
	} else {
		ziwei -= offset
	}
	ziwei = fix12(ziwei)
	return ziwei, fix12(12 - ziwei)
}

// xiaoxianStartIndex 小限起宫(宫位索引):
// 寅午戌人辰上起,申子辰人自戌宫,巳酉丑人未宫始,亥卯未人起丑宫。
func xiaoxianStartIndex(yearBranch int) int {
	switch yearBranch {
	case 2, 6, 10:
		return branchToPalaceIndex(4) // 辰
	case 8, 0, 4:
		return branchToPalaceIndex(10) // 戌
	case 5, 9, 1:
		return branchToPalaceIndex(7) // 未
	default: // 亥卯未
		return branchToPalaceIndex(1) // 丑
	}
}

// changsheng12Seq 长生十二神:按五行局起长生,阳男阴女顺行。
// 水二局申、木三局亥、金四局巳、土五局申、火六局寅。
func changsheng12Seq(ju int, clockwise bool) [12]string {
	names := [...]string{"长生", "沐浴", "冠带", "临官", "帝旺", "衰", "病", "死", "墓", "绝", "胎", "养"}
	var start int
	switch ju {
	case 2:
		start = branchToPalaceIndex(8) // 申
	case 3:
		start = branchToPalaceIndex(11) // 亥
	case 4:
		start = branchToPalaceIndex(5) // 巳
	case 5:
		start = branchToPalaceIndex(8) // 申
	case 6:
		start = branchToPalaceIndex(2) // 寅
	}
	var out [12]string
	for i, name := range names {
		var idx int
		if clockwise {
			idx = fix12(start + i)
		} else {
			idx = fix12(start - i)
		}
		out[idx] = name
	}
	return out
}

// boshi12Seq 博士十二神:从禄存起,阳男阴女顺行。
func boshi12Seq(yearStem int, clockwise bool) [12]string {
	names := [...]string{"博士", "力士", "青龙", "小耗", "将军", "奏书", "飞廉", "喜神", "病符", "大耗", "伏兵", "官府"}
	lu := lucunIndex(yearStem)
	var out [12]string
	for i, name := range names {
		var idx int
		if clockwise {
			idx = fix12(lu + i)
		} else {
			idx = fix12(lu - i)
		}
		out[idx] = name
	}
	return out
}
