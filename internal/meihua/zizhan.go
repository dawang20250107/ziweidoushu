// 测字起卦(字占):以汉字笔画数起卦,《梅花易数》端法之义。
//
// 口径(义引归纳):
//   - 一字:字画起上卦(总画数除八取余),字画加时辰数起下卦并取动爻
//     ——「一字动乎心,加时以变」;
//   - 二字:首字画数起上卦、次字画数起下卦(二字为两仪平分),
//     二字总画加时辰数取动爻。
//
// 笔画数据:Unicode Unihan kTotalStrokes(简体首值,U+4E00-9FFF 全覆盖,
// 随二进制嵌入 21KB;Unicode License,允许再分发)。
package meihua

import (
	_ "embed"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/6tail/lunar-go/calendar"
)

//go:embed strokes_cjk.bin
var strokesCJK []byte

// StrokesOf 汉字总笔画(Unihan 简体口径);非 CJK 基本区或无数据返回 0。
func StrokesOf(r rune) int {
	if r < 0x4E00 || r > 0x9FFF {
		return 0
	}
	return int(strokesCJK[r-0x4E00])
}

// ByZi 测字起卦:text 须为一或二个汉字,at 供时辰数与断卦月令。
func ByZi(text string, at time.Time, question string) (Result, error) {
	chars := []rune(strings.TrimSpace(text))
	if len(chars) == 0 || len(chars) > 2 {
		return Result{}, fmt.Errorf("测字请写一或两个汉字")
	}
	var strokes []int
	for _, c := range chars {
		if !unicode.Is(unicode.Han, c) {
			return Result{}, fmt.Errorf("「%c」不是汉字", c)
		}
		s := StrokesOf(c)
		if s == 0 {
			return Result{}, fmt.Errorf("「%c」暂无笔画数据,请换一字", c)
		}
		strokes = append(strokes, s)
	}
	if at.Year() < 1902 || at.Year() > 2098 {
		return Result{}, fmt.Errorf("时间超出支持范围(1902-2098)")
	}
	lunar := calendar.NewSolarFromDate(at).GetLunar()
	monthN := lunar.GetMonth()
	if monthN < 0 {
		monthN = -monthN
	}
	hourN := branchNum(lunar.GetTimeZhi())

	norm8 := func(n int) int {
		if m := n % 8; m != 0 {
			return m
		}
		return 8
	}
	norm6 := func(n int) int {
		if m := n % 6; m != 0 {
			return m
		}
		return 6
	}

	var upperN, lowerN, moving int
	var basis string
	if len(strokes) == 1 {
		upperN = norm8(strokes[0])
		lowerN = norm8(strokes[0] + hourN)
		moving = norm6(strokes[0] + hourN)
		basis = fmt.Sprintf("测字「%c」%d画起上卦,加%s时数%d配下卦并取动爻(一字动乎心)",
			chars[0], strokes[0], lunar.GetTimeZhi(), hourN)
	} else {
		upperN = norm8(strokes[0])
		lowerN = norm8(strokes[1])
		moving = norm6(strokes[0] + strokes[1] + hourN)
		basis = fmt.Sprintf("测字「%c%c」:「%c」%d画起上卦、「%c」%d画起下卦(两仪平分),总画加%s时数%d取动爻",
			chars[0], chars[1], chars[0], strokes[0], chars[1], strokes[1], lunar.GetTimeZhi(), hourN)
	}

	r := derive(upperN, lowerN, moving)
	r.Method = "zi"
	r.Question = question
	r.ZiText = string(chars)
	r.ZiStrokes = strokes
	r.CastBasis = basis
	r.LunarText = fmt.Sprintf("%s年%s月%s日%s时",
		lunar.GetYearZhi(), lunar.GetMonthInChinese(), lunar.GetDayInChinese(), lunar.GetTimeZhi())
	r.Judgment = r.Judge(monthN, question)
	return r, nil
}
