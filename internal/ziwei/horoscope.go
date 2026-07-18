package ziwei

import (
	"fmt"
	"time"
)

// 运限引擎:大限 / 小限(童限) / 流年 / 流月 / 流日 / 流时。
// 口径与 iztro 2.5.8 horoscope() 一致(默认配置:运限年月界均为初一,虚岁按自然年+1),
// 由 1200+ 运限黄金基准用例回归保障。
//
// 倪师《天纪》立场说明:生年四化永远固定(已在本命盘星曜上标注);
// 流年四化(当年年干)为倪师认可的动态层;大限/流月等其余各层的 Mutagen
// 字段属飞星派口径,仅作研究参考,产品展示层可自行取舍。

// HoroscopeScope 运限中的一层(大限/小限/流年/流月/流日/流时)。
// 所有按宫位的数组均以地支索引(0=子 ... 11=亥)为下标,与 Chart.Palaces 一致。
type HoroscopeScope struct {
	// Name 层名:大限/童限/小限/流年/流月/流日/流时。
	Name string `json:"name"`
	// PalaceBranch 该层命宫所在地支索引。
	PalaceBranch int `json:"palaceBranch"`
	// Stem/Branch 该层干支(大限小限为宫干支,流年月日时为对应柱干支)。
	Stem   string `json:"stem"`
	Branch string `json:"branch"`
	// PalaceNames 以该层命宫重排的十二宫名,下标为地支索引。
	PalaceNames [12]string `json:"palaceNames"`
	// Mutagen 该层天干四化 [禄,权,科,忌]。
	// ⚠️ 倪师体系只认生年四化与流年四化;其余层为飞星派研究字段。
	Mutagen [4]string `json:"mutagen"`
	// Stars 该层流曜(魁钺昌曲禄羊陀马鸾喜,流年另含年解),下标为地支索引。
	Stars [][]Star `json:"stars,omitempty"`
	// NominalAge 虚岁(仅小限层填写)。
	NominalAge int `json:"nominalAge,omitempty"`
}

// Horoscope 某目标日期对一张本命盘的完整运限叠加。
type Horoscope struct {
	// TargetSolarDate 目标公历日期 "YYYY-M-D"。
	TargetSolarDate string `json:"targetSolarDate"`
	// TargetLunarText 目标农历中文。
	TargetLunarText string `json:"targetLunarText"`
	// NominalAge 虚岁。
	NominalAge int `json:"nominalAge"`

	Decadal HoroscopeScope `json:"decadal"` // 大限(或童限)
	Age     HoroscopeScope `json:"age"`     // 小限
	Yearly  HoroscopeScope `json:"yearly"`  // 流年
	Monthly HoroscopeScope `json:"monthly"` // 流月
	Daily   HoroscopeScope `json:"daily"`   // 流日
	Hourly  HoroscopeScope `json:"hourly"`  // 流时

	// Suiqian12 流年岁前十二神、Jiangqian12 将前十二神,下标为地支索引。
	Suiqian12   [12]string `json:"suiqian12"`
	Jiangqian12 [12]string `json:"jiangqian12"`
}

// GenerateHoroscope 计算 chart 在目标日期(公历 y/m/d + 时辰索引 0-12)的运限叠加。
func GenerateHoroscope(c *Chart, year, month, day, timeIndex int) (*Horoscope, error) {
	if c == nil {
		return nil, fmt.Errorf("命盘为空")
	}
	if year < 1900 || year > 2100 {
		return nil, fmt.Errorf("目标年份 %d 超出支持范围(1900-2100)", year)
	}
	if timeIndex < 0 || timeIndex > 12 {
		return nil, fmt.Errorf("目标时辰索引需在 0-12 之间")
	}
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if t.Year() != year || int(t.Month()) != month || t.Day() != day {
		return nil, fmt.Errorf("目标日期 %d-%d-%d 不存在", year, month, day)
	}
	b := c.BirthInfo
	birth := time.Date(b.Year, time.Month(b.Month), b.Day, 0, 0, 0, 0, time.UTC)
	if t.Before(birth) {
		return nil, fmt.Errorf("目标日期早于出生日期")
	}

	target := takeCalendar(year, month, day, timeIndex)
	// 虚岁(自然年口径):目标农历年 - 出生农历年 + 1。
	nominalAge := target.LunarYear - c.LunarInfo.LunarYear + 1
	// 大限最高覆盖至局数+119,小限序列覆盖至 120;超出即拒绝
	// (iztro 对此返回退化值,属其边界缺陷,不予对齐)。
	if nominalAge < 1 || nominalAge > 120 {
		return nil, fmt.Errorf("目标日期对应虚岁 %d 超出运限支持范围(1-120)", nominalAge)
	}

	h := &Horoscope{
		TargetSolarDate: fmt.Sprintf("%d-%d-%d", year, month, day),
		TargetLunarText: target.LunarText,
		NominalAge:      nominalAge,
	}

	// ── 大限(或童限)────────────────────────────────────────
	decadalBranch := -1
	decadalName := "大限"
	for i := range c.Palaces {
		p := &c.Palaces[i]
		if nominalAge >= p.DaXianStart && nominalAge <= p.DaXianEnd {
			decadalBranch = p.Branch
			break
		}
	}
	if decadalBranch < 0 {
		// 未起运 → 童限:一命二财三疾厄,四岁夫妻五福德,六岁事业。
		childhood := []string{"命宫", "财帛", "疾厄", "夫妻", "福德", "官禄"}
		if nominalAge >= 1 && nominalAge <= len(childhood) {
			if p := c.PalaceByName(childhood[nominalAge-1]); p != nil {
				decadalBranch = p.Branch
				decadalName = "童限"
			}
		}
		if decadalBranch < 0 {
			return nil, fmt.Errorf("虚岁 %d 无法确定大限/童限", nominalAge)
		}
	}
	decadalPalace := c.PalaceByBranch(decadalBranch)
	h.Decadal = buildScope(decadalName, decadalBranch, decadalPalace.Stem, decadalBranch, true)

	// ── 小限 ─────────────────────────────────────────────────
	ageBranch := -1
	for i := range c.Palaces {
		for _, a := range c.Palaces[i].Ages {
			if a == nominalAge {
				ageBranch = c.Palaces[i].Branch
				break
			}
		}
		if ageBranch >= 0 {
			break
		}
	}
	if ageBranch < 0 {
		return nil, fmt.Errorf("虚岁 %d 无法确定小限", nominalAge)
	}
	agePalace := c.PalaceByBranch(ageBranch)
	h.Age = buildScope("小限", ageBranch, agePalace.Stem, ageBranch, false)
	h.Age.NominalAge = nominalAge

	// ── 流年 ─────────────────────────────────────────────────
	yearStem := stemIndexOf(string([]rune(target.Pillars.Year)[0]))
	yearBranch := target.YearBranch
	h.Yearly = buildScope("流年", yearBranch, yearStem, yearBranch, true)
	// 流年另安年解。
	nj := palaceIndexToBranch(nianjieIndex(yearBranch))
	h.Yearly.Stars[nj] = append([]Star{{Name: "年解", Type: StarMinor}}, h.Yearly.Stars[nj]...)

	// ── 流月 ─────────────────────────────────────────────────
	// 流年地支逆数到生月,再顺数到生时,为正月所在宫,之后每月一宫;
	// 出生月与目标月的闰月过半均需进位。
	birthLeapAdd := 0
	if c.LunarInfo.IsLeapMonth && c.LunarInfo.LunarDay > 15 {
		birthLeapAdd = 1
	}
	targetLeapAdd := 0
	if target.IsLeap && target.LunarDay > 15 {
		targetLeapAdd = 1
	}
	birthHourBranch := fix12(b.Hour)
	yearlyIdx := branchToPalaceIndex(yearBranch)
	monthlyIdx := fix12(yearlyIdx - (c.LunarInfo.LunarMonth + birthLeapAdd) + birthHourBranch + (target.LunarMonth + targetLeapAdd))
	monthStem, monthBranch := pillarIndices(target.Pillars.Month)
	h.Monthly = buildScope("流月", palaceIndexToBranch(monthlyIdx), monthStem, monthBranch, true)

	// ── 流日 ─────────────────────────────────────────────────
	dailyIdx := fix12(monthlyIdx + target.LunarDay - 1)
	dayStem, dayBranch := pillarIndices(target.Pillars.Day)
	h.Daily = buildScope("流日", palaceIndexToBranch(dailyIdx), dayStem, dayBranch, true)

	// ── 流时 ─────────────────────────────────────────────────
	hourlyIdx := fix12(dailyIdx + fix12(timeIndex))
	hourStem, hourBranch := pillarIndices(target.Pillars.Hour)
	h.Hourly = buildScope("流时", palaceIndexToBranch(hourlyIdx), hourStem, hourBranch, true)

	// ── 岁前/将前十二神(按流年地支)──────────────────────────
	suiqianNames := [...]string{"岁建", "晦气", "丧门", "贯索", "官符", "小耗", "大耗", "龙德", "白虎", "天德", "吊客", "病符"}
	for i, name := range suiqianNames {
		h.Suiqian12[palaceIndexToBranch(fix12(yearlyIdx+i))] = name
	}
	jiangqianNames := [...]string{"将星", "攀鞍", "岁驿", "息神", "华盖", "劫煞", "灾煞", "天煞", "指背", "咸池", "月煞", "亡神"}
	jqStart := jiangqianStartIndex(yearBranch)
	for i, name := range jiangqianNames {
		h.Jiangqian12[palaceIndexToBranch(fix12(jqStart+i))] = name
	}

	return h, nil
}

// buildScope 组装一层运限:重排宫名、该层四化、流曜。
// scopeStem 为天干索引,ganzhiBranch 为该层干支的地支索引(大限小限=宫支)。
func buildScope(name string, palaceBranch, scopeStem, ganzhiBranch int, withStars bool) HoroscopeScope {
	sc := HoroscopeScope{
		Name:         name,
		PalaceBranch: palaceBranch,
		Stem:         Stems[scopeStem],
		Branch:       Branches[ganzhiBranch],
	}
	scopeIdx := branchToPalaceIndex(palaceBranch)
	for i := 0; i < 12; i++ { // i 为宫位索引(寅=0)
		sc.PalaceNames[palaceIndexToBranch(i)] = PalaceNamesClockwise[fix12(i-scopeIdx)]
	}
	sc.Mutagen = mutagenTable[scopeStem]
	if withStars {
		sc.Stars = horoscopeStars(name, scopeStem, ganzhiBranch)
	}
	return sc
}

// horoscopeStars 流曜十颗:魁钺昌曲禄羊陀马鸾喜,按层前缀命名(运/流/月/日/时)。
// 安放顺序与 iztro horoscopeStar.js 一致。
func horoscopeStars(scopeName string, stem, branch int) [][]Star {
	prefix := map[string]string{
		"大限": "运", "童限": "运", "流年": "流", "流月": "月", "流日": "日", "流时": "时",
	}[scopeName]

	kui, yue := kuiYueIndex(stem)
	chang, qu := changQuIndexByStem(stem)
	lu := lucunIndex(stem)
	yang, tuo := fix12(lu+1), fix12(lu-1)
	ma := tianmaIndex(branch)
	hongluan := fix12(bp("卯") - branch)
	tianxi := fix12(hongluan + 6)

	stars := make([][]Star, 12)
	for i := range stars {
		stars[i] = []Star{} // JSON 契约:空宫格输出 [] 而非 null
	}
	push := func(palaceIdx int, suffix string, iztroType string) {
		b := palaceIndexToBranch(palaceIdx)
		stars[b] = append(stars[b], Star{Name: prefix + suffix, Type: classifyStar(prefix+suffix, iztroType)})
	}
	push(kui, "魁", "soft")
	push(yue, "钺", "soft")
	push(chang, "昌", "soft")
	push(qu, "曲", "soft")
	push(lu, "禄", "lucun")
	push(yang, "羊", "tough")
	push(tuo, "陀", "tough")
	push(ma, "马", "tianma")
	push(hongluan, "鸾", "flower")
	push(tianxi, "喜", "flower")
	return stars
}

// changQuIndexByStem 流昌流曲(按天干):
// 流昌起巳位,甲乙顺流去,不用四墓宫;流曲起酉位,甲乙逆行踪,亦不用四墓。
func changQuIndexByStem(stem int) (int, int) {
	changTable := [...]string{"巳", "午", "申", "酉", "申", "酉", "亥", "子", "寅", "卯"}
	quTable := [...]string{"酉", "申", "午", "巳", "午", "巳", "卯", "寅", "子", "亥"}
	return bp(changTable[stem]), bp(quTable[stem])
}

// nianjieIndex 年解(按年支):解神从戌上起子,逆数至当生年太岁上是也。
func nianjieIndex(yearBranch int) int {
	table := [...]string{"戌", "酉", "申", "未", "午", "巳", "辰", "卯", "寅", "丑", "子", "亥"}
	return bp(table[yearBranch])
}

// jiangqianStartIndex 将前十二神起宫(按年支三合):
// 寅午戌年将星午,申子辰年子将星,巳酉丑将酉上驻,亥卯未将卯上停。
func jiangqianStartIndex(yearBranch int) int {
	switch yearBranch {
	case 2, 6, 10:
		return bp("午")
	case 8, 0, 4:
		return bp("子")
	case 5, 9, 1:
		return bp("酉")
	default:
		return bp("卯")
	}
}

// pillarIndices 干支字符串(如「癸酉」)→ (天干索引, 地支索引)。
func pillarIndices(pillar string) (int, int) {
	runes := []rune(pillar)
	if len(runes) != 2 {
		return 0, 0
	}
	return stemIndexOf(string(runes[0])), branchIndexOf(string(runes[1]))
}
