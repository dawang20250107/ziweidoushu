// liuyaoverify:以书校机——从《增删卜易》语料批量提取占例
// (月建/日干支/旬空标注/本变卦名/装卦图文字),用引擎重新装卦后对撞:
//
//	旬空两支、本卦宫属、变卦名、六爻「六亲+纳甲干支+五行」串是否见于原文。
//
// 用法:go run ./tools/liuyaoverify research/books-json/r-bd9c7aaa.json
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/dawang20250107/ziweidoushu/internal/liuyao"
	"github.com/dawang20250107/ziweidoushu/internal/meihua"
)

type book struct {
	Chapters []struct {
		Paragraphs []struct {
			Text string `json:"text"`
		} `json:"paragraphs"`
	} `json:"chapters"`
}

var (
	stems    = []rune("甲乙丙丁戊己庚辛壬癸")
	branches = []rune("子丑寅卯辰巳午未申酉戌亥")
	// 注意:Go RE2 的 \W 仅 ASCII 口径,汉字亦属 \W,冒号必须显式列出
	caseRe = regexp.MustCompile(`干支[:：]\s*([子丑寅卯辰巳午未申酉戌亥])月\s*([甲乙丙丁戊己庚辛壬癸])([子丑寅卯辰巳午未申酉戌亥])日\s*\(旬空[:：]\s*([子丑寅卯辰巳午未申酉戌亥])([子丑寅卯辰巳午未申酉戌亥])\)`)
	guaRe  = regexp.MustCompile(`([乾坎艮震巽离坤兑])宫[:：]\s*([\x{4e00}-\x{9fff}]{2,6}?)(?:（[^）]{1,8}）)?[\s【]`)
)

func idx(r rune, set []rune) int {
	for i, x := range set {
		if x == r {
			return i
		}
	}
	return -1
}

// hexLines 卦名 → 六爻(自下而上);无此名返回 false。
var nameToLines = map[string][6]bool{}

func init() {
	for hi := 1; hi <= 8; hi++ {
		for lo := 1; lo <= 8; lo++ {
			name := meihua.HexagramNameByNums(hi, lo)
			u := meihua.TrigramByNum(hi).Lines
			l := meihua.TrigramByNum(lo).Lines
			nameToLines[name] = [6]bool{l[0], l[1], l[2], u[0], u[1], u[2]}
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: liuyaoverify <增删卜易.json>")
		os.Exit(1)
	}
	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var b book
	if err := json.Unmarshal(raw, &b); err != nil {
		panic(err)
	}

	total, kongOK, palaceOK, bianOK, bianTotal := 0, 0, 0, 0, 0
	najiaHit, najiaAll := 0, 0
	var fails []string

	for ci, ch := range b.Chapters {
		{
			// 章级拼接:语料按 420 字切段会把装卦图截断,拼回全文再逐例切片
			var sb strings.Builder
			for _, p := range ch.Paragraphs {
				sb.WriteString(p.Text)
				sb.WriteString("\n")
			}
			t := sb.String()
			// 一段可含多个占例:逐例切片,卦名只在本例区间内找(避免跨例错配)
			all := caseRe.FindAllStringSubmatchIndex(t, -1)
			for mi, m := range all {
				g := make([]string, 6)
				for k := 1; k <= 5; k++ {
					g[k] = t[m[2*k]:m[2*k+1]]
				}
				monthJian := []rune(g[1])[0]
				ds := idx([]rune(g[2])[0], stems)
				db := idx([]rune(g[3])[0], branches)
				kong1 := idx([]rune(g[4])[0], branches)
				kong2 := idx([]rune(g[5])[0], branches)

				end := len(t)
				if mi+1 < len(all) {
					end = all[mi+1][0]
				}
				rest := t[m[1]:end]
				caseText := t[m[0]:end]
				guas := guaRe.FindAllStringSubmatch(rest, 2)
				if len(guas) == 0 {
					continue
				}
				benPalace, benName := guas[0][1], strings.TrimSpace(guas[0][2])
				lines, ok := nameToLines[benName]
				if !ok {
					continue // 卦名残缺(排版噪声),不计
				}
				var moving []int
				bianName := ""
				if len(guas) == 2 {
					if bl, ok2 := nameToLines[strings.TrimSpace(guas[1][2])]; ok2 {
						bianName = strings.TrimSpace(guas[1][2])
						for i := 0; i < 6; i++ {
							if lines[i] != bl[i] {
								moving = append(moving, i+1)
							}
						}
					}
				}

				total++
				r, err := liuyao.AssembleForResearch(lines, moving, ds, db, monthJian)
				if err != nil {
					fails = append(fails, fmt.Sprintf("[ch%d]%s 装卦失败: %v", ci, benName, err))
					continue
				}

				// 1) 旬空对撞
				a, c := liuyao.XunKongForResearch(ds, db)
				if (a == kong1 && c == kong2) || (a == kong2 && c == kong1) {
					kongOK++
				} else {
					fails = append(fails, fmt.Sprintf("[ch%d]%s 旬空: 书%c%c 机%c%c(日%s%s)",
						ci, benName, branches[kong1], branches[kong2], branches[a], branches[c], g[2], g[3]))
				}
				// 2) 宫属对撞
				if strings.HasPrefix(r.Palace, benPalace) {
					palaceOK++
				} else {
					fails = append(fails, fmt.Sprintf("[ch%d]%s 宫属: 书%s宫 机%s", ci, benName, benPalace, r.Palace))
				}
				// 3) 变卦名对撞
				if bianName != "" {
					bianTotal++
					if r.BianName == bianName {
						bianOK++
					} else {
						fails = append(fails, fmt.Sprintf("[ch%d]%s 变卦: 书%s 机%s", ci, benName, bianName, r.BianName))
					}
				}
				// 4) 装卦图文字对撞:六爻「六亲纳甲五行」串在本例区间找(空白归一)
				caseNorm := strings.Map(func(r rune) rune {
					if r == ' ' || r == '\n' || r == '\t' || r == '　' {
						return -1
					}
					return r
				}, caseText)
				for _, y := range r.Yaos {
					najiaAll++
					if strings.Contains(caseNorm, y.LiuQin+y.Stem+y.Branch+y.Element) {
						najiaHit++
					}
				}
			}
		}
	}

	fmt.Printf("占例总数: %d\n", total)
	fmt.Printf("旬空对撞: %d/%d\n", kongOK, total)
	fmt.Printf("宫属对撞: %d/%d\n", palaceOK, total)
	fmt.Printf("变卦对撞: %d/%d\n", bianOK, bianTotal)
	fmt.Printf("装卦图六亲纳甲串命中: %d/%d(%.1f%%)\n", najiaHit, najiaAll, 100*float64(najiaHit)/float64(najiaAll))
	if len(fails) > 0 {
		fmt.Printf("\n分歧 %d 条:\n", len(fails))
		for i, f := range fails {
			if i >= 20 {
				fmt.Printf("…(余 %d 条)\n", len(fails)-20)
				break
			}
			fmt.Println("  " + f)
		}
	}
}
