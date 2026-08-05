package daliuren

// 《六壬断案》全书扩批回归:以书校机(参数完整占例逐一复现三传)。
//
// 解析口径:每例自「NN)」起为一块;取块内首个「X将Y时」为占时/月将,
// 其前最近之「干支日」为日辰;三传取将时之后前三个「初/中/末」标记
// (排除正文「初传/中传/末传」行文),各标记向前回溯最近干支对,
// 取其支为该传(书载三传带旬遁干,如辛巳/甲戌/己卯)。
//
// 护栏:①干支阴阳不合(OCR 讹字)剔除;②书头「XY空亡」与日辰推算旬空
// 不符者剔除(空亡系干支纯算术,不符即日辰误提,非引擎问题);
// ③三传行残缺不全者剔除。剩余为可机验例,除存疑表外须全部吻合。

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

type duanAnCase struct {
	num       string
	line      int
	dayStem   int
	dayBranch int
	hour, gen int
	chuan     [3]int
	kongOK    bool // 书头空亡与日辰自洽(或书头无空亡记载)
}

var (
	daCaseStart = regexp.MustCompile(`^\s*([0-9]{1,3})[)\x{FF09}]`) // 案号后为全角/半角括号
	daJiangShi  = regexp.MustCompile(`([子丑寅卯辰巳午未申酉戌亥])将([子丑寅卯辰巳午未申酉戌亥])时`)
	daDayRe     = regexp.MustCompile(`([甲乙丙丁戊己庚辛壬癸])([子丑寅卯辰巳午未申酉戌亥])日`)
	daKongRe    = regexp.MustCompile(`([子丑寅卯辰巳午未申酉戌亥])\s*([子丑寅卯辰巳午未申酉戌亥])\s*空亡`)
)

func isStemRune(r rune) bool { return strings.ContainsRune("甲乙丙丁戊己庚辛壬癸", r) }
func isBranchRn(r rune) bool { return strings.ContainsRune("子丑寅卯辰巳午未申酉戌亥", r) }
func isMarkerRn(r rune) bool { return r == '初' || r == '中' || r == '末' }
func runeStem(r rune) int    { return strings.IndexRune("甲乙丙丁戊己庚辛壬癸", r) / 3 }
func runeBranch(r rune) int  { return strings.IndexRune("子丑寅卯辰巳午未申酉戌亥", r) / 3 }

// parseDuanAn 解析全书,返回可机验例与(原因→数量)统计。
func parseDuanAn(t *testing.T, text string) ([]duanAnCase, map[string]int) {
	t.Helper()
	lines := strings.Split(text, "\n")
	type block struct {
		num  string
		line int
		text string
	}
	var blocks []block
	for i, ln := range lines {
		if m := daCaseStart.FindStringSubmatch(ln); m != nil {
			blocks = append(blocks, block{num: m[1], line: i + 1})
		} else if len(blocks) > 0 {
			blocks[len(blocks)-1].text += ln + "\n"
		}
		if len(blocks) > 0 && blocks[len(blocks)-1].line == i+1 {
			blocks[len(blocks)-1].text = ln + "\n"
		}
	}

	skipped := map[string]int{}
	var cases []duanAnCase
	for _, b := range blocks {
		// 压平空白:书版干支常跨行断字(「辛巳⏎日」),先去全部空白再解析
		flat := strings.Map(func(r rune) rune {
			if r == '\n' || r == '\r' || r == '\t' || r == ' ' || r == '　' {
				return -1
			}
			return r
		}, b.text)
		js := daJiangShi.FindStringSubmatchIndex(flat)
		if js == nil {
			skipped["无将时参数(引前例/仅记年月)"]++
			continue
		}
		genR := []rune(flat[js[2]:js[3]])[0]
		hourR := []rune(flat[js[4]:js[5]])[0]

		// 日辰:将时之前最近的「干支日」
		var dayM []int
		for _, m := range daDayRe.FindAllStringSubmatchIndex(flat[:js[0]], -1) {
			dayM = m
		}
		if dayM == nil {
			skipped["将时前无日辰干支"]++
			continue
		}
		stemR := []rune(flat[dayM[2]:dayM[3]])[0]
		branchR := []rune(flat[dayM[4]:dayM[5]])[0]
		ds, db := runeStem(stemR), runeBranch(branchR)
		if ds%2 != db%2 {
			skipped["干支阴阳不合(OCR 讹字)"]++
			continue
		}

		// 三传:将时之后,前三个非「X传」的 初/中/末 标记,各回溯最近干支对。
		// 传落空亡时书不写遁干而记「空X」(如案67 初传「空巳」),一并识别。
		rest := []rune(flat[js[1]:])
		pairAt := func(j int) (int, bool) { // rest[j] 起是否 遁干支/空支 对
			if j+1 >= len(rest) || !isBranchRn(rest[j+1]) {
				return 0, false
			}
			if isStemRune(rest[j]) || rest[j] == '空' {
				return runeBranch(rest[j+1]), true
			}
			return 0, false
		}
		got := map[rune]int{} // 初/中/末 → 传支+1
		for i, r := range rest {
			if !isMarkerRn(r) || got[r] != 0 {
				continue
			}
			// 排除正文行文:「X传…」;初另排日期「初一~初十」(案85 之教训)。
			// 不可过宽:压平后表格「初」可紧跟下一行首之天将缩写(案104 之教训)
			var next rune
			if i+1 < len(rest) {
				next = rest[i+1]
			}
			if next == '传' {
				continue
			}
			if r == '初' && strings.ContainsRune("一二三四五六七八九十", next) {
				continue
			}
			// 回溯 ≤14 字找最近干支对,不跨越其他标记
			for j := i - 1; j >= 0 && j >= i-14; j-- {
				if isMarkerRn(rest[j]) {
					break
				}
				if br, ok := pairAt(j); ok {
					got[r] = br + 1 // +1 避开零值歧义
					break
				}
			}
		}
		chuan := [3]int{got['初'] - 1, got['中'] - 1, got['末'] - 1}
		if got['初'] == 0 || got['中'] == 0 || got['末'] == 0 {
			// 兜底:初/中/末 标记被 OCR 吞噬时,按「遁干支+天将缩写」行序取前三
			jiang := "贵蛇朱六陈龙空虎常玄阴后"
			var rows []int
			for j := 0; j+2 < len(rest) && len(rows) < 3; j++ {
				br, ok := pairAt(j)
				if !ok {
					continue
				}
				// 干支对后 ≤3 字内须现天将缩写(排除四课/天地盘裸支行)
				hit := false
				for k := j + 2; k < len(rest) && k <= j+4; k++ {
					if strings.ContainsRune(jiang, rest[k]) {
						hit = true
						break
					}
					if !(rest[k] == ' ' || rest[k] == '\t') {
						break
					}
				}
				if hit {
					rows = append(rows, br)
					j += 2
				}
			}
			if len(rows) == 3 {
				chuan = [3]int{rows[0], rows[1], rows[2]}
			} else {
				skipped["三传行残缺(OCR 断行)"]++
				continue
			}
		}

		c := duanAnCase{
			num: b.num, line: b.line,
			dayStem: ds, dayBranch: db,
			hour: runeBranch(hourR), gen: runeBranch(genR),
			chuan:  chuan,
			kongOK: true,
		}
		// 护栏:书头空亡两支与日辰旬空须自洽(不符即日辰误提)
		if km := daKongRe.FindStringSubmatch(flat); km != nil {
			xf := ((db-ds)%12 + 12) % 12
			want := map[int]bool{(xf + 10) % 12: true, (xf + 11) % 12: true}
			k1 := runeBranch([]rune(km[1])[0])
			k2 := runeBranch([]rune(km[2])[0])
			if !want[k1] || !want[k2] {
				skipped["书头空亡与日辰不符(参数存疑)"]++
				continue
			}
		}
		cases = append(cases, c)
	}
	return cases, skipped
}

// duanAnSuspect 存疑例(复现不符,经查为传抄讹误/文本自相矛盾,不计失败):
// 值为存疑原因,逐例人工核对原文后录入。
var duanAnSuspect = map[string]string{
	"120": "三传行遁干「壬寅/戊午」与庚辰日甲戌旬不合(应戊寅/壬午),且与同课骨之案74(同 tp、同四课,书取午辰寅)矛盾——行序紊乱,真三传当为午辰寅(引擎所出)",
	"137": "头行「酉将亥时」系将时倒置:盘面四课干上酉、支上亦酉,证实为亥将酉时;从盘面参数则引擎正复现书载三传酉酉酉(八专独足格)",
	"208": "头行「戌将午时」占时讹误:书面四课(辛上午/午上寅/卯上亥/亥上未)证实天盘为戌将寅时;以盘面参数则引擎正复现书载三传未卯亥(比用)",
}

// TestDuanAnBatchRegression 全书参数完整占例批量回归。
func TestDuanAnBatchRegression(t *testing.T) {
	data, err := os.ReadFile("../../research/corpus/00bf403d-六壬断案01.txt")
	if err != nil {
		t.Fatalf("研究语料缺失: %v", err)
	}
	cases, skipped := parseDuanAn(t, string(data))
	t.Logf("解析: 可机验 %d 例,剔除 %v", len(cases), skipped)

	match, mismatch := 0, 0
	for _, c := range cases {
		r, err := Cast(c.dayStem, c.dayBranch, c.hour, c.gen)
		if err != nil {
			t.Fatalf("案%s Cast: %v", c.num, err)
		}
		want := fmt.Sprintf("%c%c%c", branches[c.chuan[0]], branches[c.chuan[1]], branches[c.chuan[2]])
		gotS := r.Chuan[0] + r.Chuan[1] + r.Chuan[2]
		if gotS == want {
			match++
			continue
		}
		if reason, ok := duanAnSuspect[c.num]; ok {
			t.Logf("案%s(行%d)存疑跳过: %s", c.num, c.line, reason)
			continue
		}
		mismatch++
		t.Errorf("案%s(行%d)%s%s日%s将%s时: 书载三传%s,引擎%s(%s)",
			c.num, c.line,
			string(stems[c.dayStem]), string(branches[c.dayBranch]),
			string(branches[c.gen]), string(branches[c.hour]),
			want, gotS, r.KeType)
	}
	t.Logf("回归: 吻合 %d / 存疑 %d / 不符 %d", match, len(duanAnSuspect), mismatch)
	if match < 190 { // 基线 205/208(2026-07),下滑即解析器或引擎回归
		t.Fatalf("吻合例过少(%d),疑解析器回归", match)
	}
}
