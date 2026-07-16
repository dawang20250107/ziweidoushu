package ziwei

// 本文件为安星诀实现,全部索引为宫位索引(寅=0),口径与 iztro location.js 一致。

// bp 地支名速记 → 宫位索引。
func bp(branchName string) int { return branchToPalaceIndex(branchIndexOf(branchName)) }

// lucunIndex 禄存(按年干):甲寅乙卯,丙戊巳,丁己午,庚申辛酉,壬亥癸子。
func lucunIndex(yearStem int) int {
	table := [...]string{"寅", "卯", "巳", "午", "巳", "午", "申", "酉", "亥", "子"}
	return bp(table[yearStem])
}

// tianmaIndex 天马(按年支三合):寅午戌在申,申子辰在寅,巳酉丑在亥,亥卯未在巳。
func tianmaIndex(yearBranch int) int {
	switch yearBranch {
	case 2, 6, 10:
		return bp("申")
	case 8, 0, 4:
		return bp("寅")
	case 5, 9, 1:
		return bp("亥")
	default:
		return bp("巳")
	}
}

// kuiYueIndex 天魁天钺(按年干)。
func kuiYueIndex(yearStem int) (int, int) {
	switch yearStem {
	case 0, 4, 6: // 甲戊庚 丑未
		return bp("丑"), bp("未")
	case 1, 5: // 乙己 子申
		return bp("子"), bp("申")
	case 7: // 辛 午寅
		return bp("午"), bp("寅")
	case 2, 3: // 丙丁 亥酉
		return bp("亥"), bp("酉")
	default: // 壬癸 卯巳
		return bp("卯"), bp("巳")
	}
}

// huoLingIndex 火星铃星(按年支三合起子时,顺数至生时)。
func huoLingIndex(yearBranch, timeBranch int) (int, int) {
	var huoStart, lingStart int
	switch yearBranch {
	case 2, 6, 10: // 寅午戌:火起丑,铃起卯
		huoStart, lingStart = bp("丑"), bp("卯")
	case 8, 0, 4: // 申子辰:火起寅,铃起戌
		huoStart, lingStart = bp("寅"), bp("戌")
	case 5, 9, 1: // 巳酉丑:火起卯,铃起戌
		huoStart, lingStart = bp("卯"), bp("戌")
	default: // 亥卯未:火起酉,铃起戌
		huoStart, lingStart = bp("酉"), bp("戌")
	}
	return fix12(huoStart + timeBranch), fix12(lingStart + timeBranch)
}

// placeMajorStars 安十四主星。
// 紫微系(逆行):紫微 天机 隔一 太阳 武曲 天同 隔二 廉贞;
// 天府系(顺行):天府 太阴 贪狼 巨门 天相 天梁 七杀 隔三 破军。
func placeMajorStars(stars [][]Star, ziweiIdx, tianfuIdx, yearStem int) {
	ziweiGroup := [...]string{"紫微", "天机", "", "太阳", "武曲", "天同", "", "", "廉贞"}
	tianfuGroup := [...]string{"天府", "太阴", "贪狼", "巨门", "天相", "天梁", "七杀", "", "", "", "破军"}
	for i, name := range ziweiGroup {
		if name == "" {
			continue
		}
		idx := fix12(ziweiIdx - i)
		stars[idx] = append(stars[idx], newStar(name, StarMajor, idx, yearStem))
	}
	for i, name := range tianfuGroup {
		if name == "" {
			continue
		}
		idx := fix12(tianfuIdx + i)
		stars[idx] = append(stars[idx], newStar(name, StarMajor, idx, yearStem))
	}
}

// placeMinorStars 安十四辅星(左右昌曲魁钺禄马空劫火铃羊陀)。
func placeMinorStars(stars [][]Star, snap calendarSnapshot, monthIndex, timeBranch int) {
	zuo := fix12(bp("辰") + monthIndex)   // 辰上顺正寻左辅
	you := fix12(bp("戌") - monthIndex)   // 戌上逆正右弼当
	chang := fix12(bp("戌") - timeBranch) // 戌上逆时觅文昌
	qu := fix12(bp("辰") + timeBranch)    // 辰上顺时文曲位
	kui, yue := kuiYueIndex(snap.YearStem)
	lu := lucunIndex(snap.YearStem)
	yang, tuo := fix12(lu+1), fix12(lu-1)
	ma := tianmaIndex(snap.YearBranch)
	kong := fix12(bp("亥") - timeBranch) // 亥上子时逆安地空
	jie := fix12(bp("亥") + timeBranch)  // 亥上子时顺安地劫
	huo, ling := huoLingIndex(snap.YearBranch, timeBranch)

	// 与 iztro minorStar.js 相同的安放顺序(影响宫内星曜排序)。
	push := func(idx int, name string, withMutagen bool) {
		s := Star{Name: name, Type: classifyStar(name, minorStarIztroType[name])}
		s.Brightness = brightnessOf(name, palaceIndexToBranch(idx))
		if s.Brightness != "" {
			s.BrightnessLevel = mapBrightnessLevel(s.Brightness)
		}
		if withMutagen {
			s.SiHua = mutagenOf(name, snap.YearStem)
		}
		stars[idx] = append(stars[idx], s)
	}
	push(zuo, "左辅", true)
	push(you, "右弼", true)
	push(chang, "文昌", true)
	push(qu, "文曲", true)
	push(kui, "天魁", false)
	push(yue, "天钺", false)
	push(lu, "禄存", false)
	push(ma, "天马", false)
	push(kong, "地空", false)
	push(jie, "地劫", false)
	push(huo, "火星", false)
	push(ling, "铃星", false)
	push(yang, "擎羊", false)
	push(tuo, "陀罗", false)
}

// placeAdjectiveStars 安 38 杂曜,push 顺序与 iztro adjectiveStar.js 严格一致。
func placeAdjectiveStars(stars [][]Star, snap calendarSnapshot, soulIndex, bodyIndex, monthIndex, timeBranch, timeIndex int) {
	yearBranch := snap.YearBranch
	yearStem := snap.YearStem

	// ── 年系 ──
	hongluan := fix12(bp("卯") - yearBranch)
	tianxi := fix12(hongluan + 6)
	var huagai, xianchi int
	switch yearBranch {
	case 2, 6, 10:
		huagai, xianchi = bp("戌"), bp("卯")
	case 8, 0, 4:
		huagai, xianchi = bp("辰"), bp("酉")
	case 5, 9, 1:
		huagai, xianchi = bp("丑"), bp("午")
	default:
		huagai, xianchi = bp("未"), bp("子")
	}
	var guchen, guasu int
	switch yearBranch {
	case 2, 3, 4: // 寅卯辰 → 巳丑
		guchen, guasu = bp("巳"), bp("丑")
	case 5, 6, 7: // 巳午未 → 申辰
		guchen, guasu = bp("申"), bp("辰")
	case 8, 9, 10: // 申酉戌 → 亥未
		guchen, guasu = bp("亥"), bp("未")
	default: // 亥子丑 → 寅戌
		guchen, guasu = bp("寅"), bp("戌")
	}
	tiancai := fix12(soulIndex + yearBranch)
	tianshou := fix12(bodyIndex + yearBranch)
	tianchu := bp([...]string{"巳", "午", "子", "巳", "午", "申", "寅", "午", "酉", "亥"}[yearStem])
	posui := bp([...]string{"巳", "丑", "酉"}[yearBranch%3])
	feilian := bp([...]string{"申", "酉", "戌", "巳", "午", "未", "寅", "卯", "辰", "亥", "子", "丑"}[yearBranch])
	longchi := fix12(bp("辰") + yearBranch)
	fengge := fix12(bp("戌") - yearBranch)
	tianku := fix12(bp("午") - yearBranch)
	tianxu := fix12(bp("午") + yearBranch)
	tianguan := bp([...]string{"未", "辰", "巳", "寅", "卯", "酉", "亥", "酉", "戌", "午"}[yearStem])
	tianfuAdj := bp([...]string{"酉", "申", "子", "亥", "卯", "寅", "午", "巳", "午", "巳"}[yearStem])
	tiande := fix12(bp("酉") + yearBranch)
	yuede := fix12(bp("巳") + yearBranch)
	tiankong := fix12(branchToPalaceIndex(yearBranch) + 1)
	jielu := bp([...]string{"申", "午", "辰", "寅", "子"}[yearStem%5])
	kongwang := bp([...]string{"酉", "未", "巳", "卯", "丑"}[yearStem%5])
	xunkong := fix12(branchToPalaceIndex(yearBranch) + 9 - yearStem + 1)
	if yearBranch%2 != xunkong%2 {
		xunkong = fix12(xunkong + 1)
	}
	nianjie := bp([...]string{"戌", "酉", "申", "未", "午", "巳", "辰", "卯", "寅", "丑", "子", "亥"}[yearBranch])
	tianshang := fix12(5 + soulIndex) // 仆役宫
	tianshi := fix12(7 + soulIndex)   // 疾厄宫

	// ── 月系 ──
	yuejie := bp([...]string{"申", "戌", "子", "寅", "辰", "午"}[monthIndex/2])
	tianyao := fix12(bp("丑") + monthIndex)
	tianxing := fix12(bp("酉") + monthIndex)
	yinsha := bp([...]string{"寅", "子", "戌", "申", "午", "辰"}[monthIndex%6])
	tianyue := bp([...]string{"戌", "巳", "辰", "寅", "未", "卯", "亥", "未", "寅", "午", "戌", "寅"}[monthIndex])
	tianwu := bp([...]string{"巳", "申", "寅", "亥"}[monthIndex%4])

	// ── 日系(依左辅右弼昌曲起宫)──
	zuo := fix12(bp("辰") + monthIndex)
	you := fix12(bp("戌") - monthIndex)
	chang := fix12(bp("戌") - timeBranch)
	qu := fix12(bp("辰") + timeBranch)
	dayIndex := snap.LunarDay - 1
	if timeIndex >= 12 {
		dayIndex = snap.LunarDay
	}
	santai := fix12((zuo + dayIndex) % 12)
	bazuo := fix12((you - dayIndex) % 12)
	enguang := fix12((chang+dayIndex)%12 - 1)
	tiangui := fix12((qu+dayIndex)%12 - 1)

	// ── 时系 ──
	taifu := fix12(bp("午") + timeBranch)
	fenggao := fix12(bp("寅") + timeBranch)

	// push 顺序与 iztro adjectiveStar.js 一致。
	push := func(idx int, name string) {
		stars[idx] = append(stars[idx], Star{Name: name, Type: StarMinor})
	}
	push(hongluan, "红鸾")
	push(tianxi, "天喜")
	push(tianyao, "天姚")
	push(xianchi, "咸池")
	push(yuejie, "解神")
	push(santai, "三台")
	push(bazuo, "八座")
	push(enguang, "恩光")
	push(tiangui, "天贵")
	push(longchi, "龙池")
	push(fengge, "凤阁")
	push(tiancai, "天才")
	push(tianshou, "天寿")
	push(taifu, "台辅")
	push(fenggao, "封诰")
	push(tianwu, "天巫")
	push(huagai, "华盖")
	push(tianguan, "天官")
	push(tianfuAdj, "天福")
	push(tianchu, "天厨")
	push(tianyue, "天月")
	push(tiande, "天德")
	push(yuede, "月德")
	push(tiankong, "天空")
	push(xunkong, "旬空")
	push(jielu, "截路")
	push(kongwang, "空亡")
	push(guchen, "孤辰")
	push(guasu, "寡宿")
	push(feilian, "蜚廉")
	push(posui, "破碎")
	push(tianxing, "天刑")
	push(yinsha, "阴煞")
	push(tianku, "天哭")
	push(tianxu, "天虚")
	push(tianshi, "天使")
	push(tianshang, "天伤")
	push(nianjie, "年解")
}

// newStar 构造主星(带亮度与生年四化)。
func newStar(name string, typ StarType, palaceIdx, yearStem int) Star {
	s := Star{Name: name, Type: typ}
	s.Brightness = brightnessOf(name, palaceIndexToBranch(palaceIdx))
	s.BrightnessLevel = mapBrightnessLevel(s.Brightness)
	s.SiHua = mutagenOf(name, yearStem)
	return s
}

// mutagenOf 生年四化:星名在年干四化表中的位置 → 禄权科忌。
func mutagenOf(starName string, yearStem int) SiHua {
	table := mutagenTable[yearStem]
	switch starName {
	case table[0]:
		return HuaLu
	case table[1]:
		return HuaQuan
	case table[2]:
		return HuaKe
	case table[3]:
		return HuaJi
	}
	return ""
}
