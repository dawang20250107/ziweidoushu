// caseverify:倪师《天纪》民国案例 → 排盘引擎约束校验。
//
// 案例只记「民国年 + 阴阳性别 + 五行局 + 紫微位置(+命宫位置/时辰)」,
// 无完整生辰。校验分两层:
//  1. 年干阴阳:民国 N 年 → 干支年,记录的「阴/阳」必须与年干一致;
//  2. 约束可满足性:枚举该干支年全部农历日 × 13 时辰,存在(日,时)使
//     引擎输出同时满足全部记录约束 → SAT(相容);无解 → UNSAT(口径冲突)。
//
// 用法:go run ./tools/caseverify
package main

import (
	"fmt"

	"time"

	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

type caseRec struct {
	ep       string // 出处(天纪第几集)
	minguo   int    // 民国纪年
	yinYang  string // 阴/阳(年干阴阳)
	gender   ziwei.Gender
	ju       string // 五行局名,空=未记录
	ziweiPos int    // 紫微地支索引,-1=未记录
	mingPos  int    // 命宫地支索引,-1=未记录
	hour     int    // 时辰索引,-1=未记录
	note     string
}

const branches = "子丑寅卯辰巳午未申酉戌亥"

var branchRunes = []rune(branches)

// b 地支字符 → 索引(rune 级;strings.IndexRune 返回字节偏移,不可用)。
func b(c rune) int {
	for i, r := range branchRunes {
		if r == c {
			return i
		}
	}
	return -1
}

var cases = []caseRec{
	{"第6集", 54, "阴", ziwei.Male, "木三局", b('丑'), -1, 10, "乙巳年戌时示范盘"},
	{"第6集", 36, "阴", ziwei.Male, "金四局", b('寅'), -1, 5, "丁亥年巳时示范盘"},
	{"第7集", 44, "阴", ziwei.Male, "水二局", b('卯'), -1, -1, "乙未年示范盘"},
	{"第7集", 56, "阴", ziwei.Female, "金四局", b('辰'), -1, -1, ""},
	{"第7集", 83, "阳", ziwei.Female, "金四局", b('巳'), -1, -1, ""},
	{"第7集", 58, "阴", ziwei.Male, "火六局", b('午'), -1, -1, ""},
	{"第7集", 46, "阴", ziwei.Female, "木三局", b('未'), -1, -1, ""},
	{"第7集", 40, "阴", ziwei.Male, "水二局", b('申'), -1, -1, ""},
	{"第8集", 50, "阴", ziwei.Male, "木三局", b('酉'), -1, -1, ""},
	{"第8集", 46, "阴", ziwei.Male, "木三局", b('戌'), -1, -1, ""},
	{"第8集", 38, "阴", ziwei.Female, "火六局", b('亥'), -1, -1, ""},
	{"第8集", 37, "阳", ziwei.Male, "木三局", b('子'), b('酉'), -1, "命例实批"},
	{"第9集", 42, "阴", ziwei.Male, "水二局", b('寅'), b('戌'), -1, "化忌冲命"},
	{"第10集", 46, "阴", ziwei.Female, "水二局", b('申'), -1, -1, ""},
	{"第10集", 51, "阳", ziwei.Male, "火六局", b('辰'), -1, -1, ""},
	{"第10集", 50, "阴", ziwei.Female, "金四局", b('子'), b('午'), -1, ""},
	{"第11集", 83, "阳", ziwei.Female, "金四局", b('巳'), b('酉'), -1, ""},
	{"第11集", 54, "阴", ziwei.Male, "", b('丑'), b('未'), -1, ""},
	{"第12集", 46, "阴", ziwei.Female, "金四局", b('寅'), b('亥'), -1, ""},
	{"第12集", 55, "阳", ziwei.Female, "木三局", b('戌'), b('亥'), -1, ""},
	{"第13集", 46, "阴", ziwei.Male, "木三局", b('戌'), b('子'), -1, ""},
	{"第13集", 33, "阳", ziwei.Female, "木三局", b('辰'), b('辰'), -1, ""},
	{"第13集", 28, "阴", ziwei.Male, "金四局", b('亥'), b('酉'), -1, ""},
	{"第13集", 41, "阳", ziwei.Female, "土五局", b('酉'), b('酉'), -1, ""},
	{"第13集", 41, "阳", ziwei.Female, "火六局", b('子'), b('辰'), -1, ""},
	{"第14集", 18, "阴", ziwei.Male, "水二局", b('巳'), b('丑'), -1, ""},
	{"第14集", 49, "阳", ziwei.Male, "水二局", b('丑'), b('申'), -1, ""},
	{"第14集", 31, "阳", ziwei.Female, "金四局", b('辰'), b('卯'), -1, ""},
	{"第14集", 38, "阴", ziwei.Female, "金四局", b('申'), b('酉'), -1, "孤鸾命"},
	{"第15集", 17, "阳", ziwei.Female, "土五局", b('寅'), b('巳'), -1, ""},
	{"第15集", 28, "阴", ziwei.Female, "金四局", b('卯'), b('申'), -1, ""},
	{"第15集", 30, "阴", ziwei.Female, "土五局", b('未'), b('子'), -1, ""},
	{"第15集", 28, "阴", ziwei.Female, "火六局", b('丑'), b('寅'), -1, "命无正曜格"},
}

var stems = []rune("甲乙丙丁戊己庚辛壬癸")

// yearStemYinYang 公历年 → 年干及阴阳(甲丙戊庚壬=阳)。
func yearStemYinYang(y int) (rune, string) {
	idx := ((y-4)%10 + 10) % 10
	if idx%2 == 0 {
		return stems[idx], "阳"
	}
	return stems[idx], "阴"
}

func ziweiBranch(c *ziwei.Chart) int {
	for i := range c.Palaces {
		for _, s := range c.Palaces[i].Stars {
			if s.Name == "紫微" {
				return c.Palaces[i].Branch
			}
		}
	}
	return -1
}

func main() {
	sat, unsat, stemBad := 0, 0, 0
	for _, cs := range cases {
		year := cs.minguo + 1911
		stem, yy := yearStemYinYang(year)
		stemMark := "✓"
		if yy != cs.yinYang {
			stemMark = "✗"
			stemBad++
		}

		// 枚举干支年覆盖的公历区间(春节漂移,取宽区间按年柱过滤)
		found := ""
		start := time.Date(year, 1, 15, 0, 0, 0, 0, time.UTC)
		end := time.Date(year+1, 2, 25, 0, 0, 0, 0, time.UTC)
		expectYear := string(stem) + zodiacBranch(year)
	outer:
		for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
			hours := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
			if cs.hour >= 0 {
				hours = []int{cs.hour}
			}
			for _, h := range hours {
				chart, err := ziwei.Generate(ziwei.BirthInfo{
					Year: d.Year(), Month: int(d.Month()), Day: d.Day(),
					Hour: h, Gender: cs.gender,
				}, ziwei.Options{ReferenceYear: 2026})
				if err != nil {
					continue
				}
				if chart.FourPillars.Year != expectYear {
					continue outer // 整天不属于该干支年,跳过当天
				}
				if cs.ju != "" && chart.WuxingJuName != cs.ju {
					continue
				}
				if cs.ziweiPos >= 0 && ziweiBranch(chart) != cs.ziweiPos {
					continue
				}
				if cs.mingPos >= 0 && chart.MingGongBranch != cs.mingPos {
					continue
				}
				found = fmt.Sprintf("%s %s时", d.Format("2006-01-02"), string([]rune(branches)[h%12]))
				break outer
			}
		}
		satMark := "SAT " + found
		if found == "" {
			satMark = "UNSAT ←← 无解,口径冲突"
			unsat++
		} else {
			sat++
		}
		fmt.Printf("%-5s 民国%02d(%d %s%s年) %s%s局%-4s 紫微%s 命宫%s | 年干%s | %s %s\n",
			cs.ep, cs.minguo, year, string(stem), zodiacBranch(year), cs.yinYang,
			string(cs.gender)[:0]+genderCN(cs.gender), cs.ju,
			posName(cs.ziweiPos), posName(cs.mingPos), stemMark, satMark, cs.note)
	}
	fmt.Printf("\n== 汇总:%d 例 | 约束相容 %d | 无解 %d | 年干阴阳不符 %d ==\n",
		len(cases), sat, unsat, stemBad)
}

func zodiacBranch(y int) string {
	return string([]rune(branches)[((y-4)%12+12)%12])
}

func genderCN(g ziwei.Gender) string {
	if g == ziwei.Male {
		return "男"
	}
	return "女"
}

func posName(i int) string {
	if i < 0 {
		return "-"
	}
	return string([]rune(branches)[i])
}
