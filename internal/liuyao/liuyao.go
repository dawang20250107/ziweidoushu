// Package liuyao 六爻(火珠林法)装卦引擎:以钱代蓍摇卦、京房纳甲、
// 八宫世应、六亲、六神、动爻变卦。断辞由 AI 层引语料
// (火珠林/增删卜易/卜筮正宗均在研究语料)补充,本包只负责可验证的装卦计算。
//
// 正确性:八宫六十四卦按变爻规律程序生成,与《卜筮正宗》定表逐卦对照(单测)。
package liuyao

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/6tail/lunar-go/calendar"
	"github.com/dawang20250107/ziweidoushu/internal/meihua"
)

// ── 基础表 ───────────────────────────────────────────────────

var (
	stems    = []rune("甲乙丙丁戊己庚辛壬癸")
	branches = []rune("子丑寅卯辰巳午未申酉戌亥")
	// 地支五行:木0 火1 土2 金3 水4(与 sizhu 同序)
	branchElement = []int{4, 2, 0, 0, 2, 1, 1, 2, 3, 3, 2, 4}
	elementNames  = []string{"木", "火", "土", "金", "水"}
	// 八卦宫五行(先天数序 乾兑离震巽坎艮坤)
	trigramElement = []int{3, 3, 1, 0, 0, 4, 2, 2}
)

// 京房纳甲:每纯卦内/外三爻的天干与三支(自下而上)。索引=先天数-1。
var najia = [8]struct {
	innerStem rune
	outerStem rune
	inner     [3]rune
	outer     [3]rune
}{
	{'甲', '壬', [3]rune{'子', '寅', '辰'}, [3]rune{'午', '申', '戌'}}, // 乾
	{'丁', '丁', [3]rune{'巳', '卯', '丑'}, [3]rune{'亥', '酉', '未'}}, // 兑
	{'己', '己', [3]rune{'卯', '丑', '亥'}, [3]rune{'酉', '未', '巳'}}, // 离
	{'庚', '庚', [3]rune{'子', '寅', '辰'}, [3]rune{'午', '申', '戌'}}, // 震
	{'辛', '辛', [3]rune{'丑', '亥', '酉'}, [3]rune{'未', '巳', '卯'}}, // 巽
	{'戊', '戊', [3]rune{'寅', '辰', '午'}, [3]rune{'申', '戌', '子'}}, // 坎
	{'丙', '丙', [3]rune{'辰', '午', '申'}, [3]rune{'戌', '子', '寅'}}, // 艮
	{'乙', '癸', [3]rune{'未', '巳', '卯'}, [3]rune{'丑', '亥', '酉'}}, // 坤
}

// 六神(按日干起,自初爻循环向上)。
var liuShen = []string{"青龙", "朱雀", "勾陈", "腾蛇", "白虎", "玄武"}

// liuShenStart 日干 → 初爻六神索引:甲乙青龙 丙丁朱雀 戊勾陈 己腾蛇 庚辛白虎 壬癸玄武。
func liuShenStart(dayStem int) int {
	switch {
	case dayStem <= 1:
		return 0
	case dayStem <= 3:
		return 1
	case dayStem == 4:
		return 2
	case dayStem == 5:
		return 3
	case dayStem <= 7:
		return 4
	default:
		return 5
	}
}

// ── 八宫与世应(规律生成,单测与定表对照)────────────────────

// palaceHexagrams 生成某宫八卦的六爻爻象序列(自下而上)。
// 宫序:0 纯卦,1-5 一至五世(依次再变初~五爻),6 游魂(变四爻),7 归魂(内卦复原)。
func palaceHexagrams(pure [6]bool) [8][6]bool {
	var out [8][6]bool
	cur := pure
	out[0] = cur
	for i := 1; i <= 5; i++ {
		cur[i-1] = !cur[i-1]
		out[i] = cur
	}
	// 游魂:在五世卦基础上再变四爻
	cur[3] = !cur[3]
	out[6] = cur
	// 归魂:游魂内卦复原为本宫内卦
	copy(cur[:3], pure[:3])
	out[7] = cur
	return out
}

// shiPositions 宫序 → 世爻位置(1-6)。
var shiPositions = [8]int{6, 1, 2, 3, 4, 5, 4, 3}

type palaceEntry struct {
	palace int // 宫(先天数-1)
	seq    int // 宫内序 0-7
	shi    int // 世爻 1-6
}

// hexIndex 六爻爻象 → 64 序(下卦*8+上卦 各 0-7,用先天数-1)。
func hexKey(lines [6]bool) int {
	lo := meihua.TrigramByLines([3]bool{lines[0], lines[1], lines[2]}).Num - 1
	hi := meihua.TrigramByLines([3]bool{lines[3], lines[4], lines[5]}).Num - 1
	return hi*8 + lo
}

// palaceIndex 全 64 卦 → 八宫归属(初始化时按规律生成)。
var palaceIndex map[int]palaceEntry

func init() {
	palaceIndex = make(map[int]palaceEntry, 64)
	for p := 0; p < 8; p++ {
		t := meihua.TrigramByNum(p + 1)
		pure := [6]bool{t.Lines[0], t.Lines[1], t.Lines[2], t.Lines[0], t.Lines[1], t.Lines[2]}
		for seq, lines := range palaceHexagrams(pure) {
			palaceIndex[hexKey(lines)] = palaceEntry{palace: p, seq: seq, shi: shiPositions[seq]}
		}
	}
}

// ── 装卦 ─────────────────────────────────────────────────────

// Yao 一爻(自下而上第 Pos 爻)。
type Yao struct {
	Pos      int    `json:"pos"`  // 1-6
	Yang     bool   `json:"yang"` // 阳爻
	Moving   bool   `json:"moving"`
	Stem     string `json:"stem"`   // 纳甲天干
	Branch   string `json:"branch"` // 纳甲地支
	Element  string `json:"element"`
	LiuQin   string `json:"liuQin"`  // 六亲
	LiuShen  string `json:"liuShen"` // 六神(按日干)
	IsShi    bool   `json:"isShi"`   // 世
	IsYing   bool   `json:"isYing"`  // 应
	BianYao  *Yao   `json:"bianYao,omitempty"` // 动爻之变(仅动爻有,变卦对应爻)
}

// Result 六爻装卦结果。
type Result struct {
	Question  string `json:"question,omitempty"`
	LunarText string `json:"lunarText"` // 日干支定六神:「甲子日」
	DayStem   string `json:"dayStem"`
	DayBranch string `json:"dayBranch"`
	MonthJian string `json:"monthJian"` // 月建地支
	RiJian    string `json:"riJian"`    // 日辰地支(同 DayBranch,断卦口径字段)

	BenName    string `json:"benName"`
	BianName   string `json:"bianName,omitempty"` // 有动爻才有变卦
	Palace     string `json:"palace"`             // 本卦所属宫,如「乾宫」
	PalaceSeq  string `json:"palaceSeq"`          // 纯卦/一世…游魂/归魂
	Yaos       [6]Yao `json:"yaos"`
	MovingNums []int  `json:"movingNums"` // 动爻位置列表(可为空=静卦)

	// Tosses 摇卦原始记录(每爻背面数 0-3;报数起卦为空)
	Tosses []int `json:"tosses,omitempty"`
}

var seqNames = []string{"八纯卦", "一世卦", "二世卦", "三世卦", "四世卦", "五世卦", "游魂卦", "归魂卦"}

// liuQin 宫五行 → 爻五行 的六亲。
func liuQin(palaceEl, yaoEl int) string {
	switch {
	case palaceEl == yaoEl:
		return "兄弟"
	case (palaceEl+1)%5 == yaoEl: // 我生
		return "子孙"
	case (palaceEl+2)%5 == yaoEl: // 我克
		return "妻财"
	case (yaoEl+2)%5 == palaceEl: // 克我
		return "官鬼"
	default: // 生我
		return "父母"
	}
}

// assemble 由六爻阴阳+动爻+日干支装出完整卦。
func assemble(lines [6]bool, moving []int, dayStem, dayBranch int, monthJian rune) (*Result, error) {
	key := hexKey(lines)
	entry, ok := palaceIndex[key]
	if !ok {
		return nil, fmt.Errorf("装卦失败:卦象无宫位归属")
	}
	palaceEl := trigramElement[entry.palace]
	lower := meihua.TrigramByLines([3]bool{lines[0], lines[1], lines[2]})
	upper := meihua.TrigramByLines([3]bool{lines[3], lines[4], lines[5]})

	movingSet := map[int]bool{}
	for _, m := range moving {
		if m < 1 || m > 6 {
			return nil, fmt.Errorf("动爻位置非法: %d", m)
		}
		movingSet[m] = true
	}

	// 变卦
	bianLines := lines
	for m := range movingSet {
		bianLines[m-1] = !bianLines[m-1]
	}

	r := &Result{
		BenName:   meihua.HexagramNameByNums(upper.Num, lower.Num),
		Palace:    meihua.TrigramByNum(entry.palace + 1).Name + "宫",
		PalaceSeq: seqNames[entry.seq],
		DayStem:   string(stems[dayStem]),
		DayBranch: string(branches[dayBranch]),
		RiJian:    string(branches[dayBranch]),
		MonthJian: string(monthJian),
		// 静卦时也须输出 [] 而非 null(JSON 列表契约)
		MovingNums: []int{},
	}
	if len(movingSet) > 0 {
		bl := meihua.TrigramByLines([3]bool{bianLines[0], bianLines[1], bianLines[2]})
		bu := meihua.TrigramByLines([3]bool{bianLines[3], bianLines[4], bianLines[5]})
		r.BianName = meihua.HexagramNameByNums(bu.Num, bl.Num)
	}

	// 变卦纳甲按变卦自身的内外卦纳
	bianLower := meihua.TrigramByLines([3]bool{bianLines[0], bianLines[1], bianLines[2]})
	bianUpper := meihua.TrigramByLines([3]bool{bianLines[3], bianLines[4], bianLines[5]})

	shenStart := liuShenStart(dayStem)
	for i := 0; i < 6; i++ {
		pos := i + 1
		var stem rune
		var branch rune
		if i < 3 {
			na := najia[lower.Num-1]
			stem, branch = na.innerStem, na.inner[i]
		} else {
			na := najia[upper.Num-1]
			stem, branch = na.outerStem, na.outer[i-3]
		}
		bIdx := 0
		for j, br := range branches {
			if br == branch {
				bIdx = j
			}
		}
		// 应爻 = 世 ±3(规范化到 1-6)
		ying := entry.shi + 3
		if ying > 6 {
			ying -= 6
		}
		yao := Yao{
			Pos:     pos,
			Yang:    lines[i],
			Moving:  movingSet[pos],
			Stem:    string(stem),
			Branch:  string(branch),
			Element: elementNames[branchElement[bIdx]],
			LiuQin:  liuQin(palaceEl, branchElement[bIdx]),
			LiuShen: liuShen[(shenStart+i)%6],
			IsShi:   pos == entry.shi,
			IsYing:  pos == ying,
		}

		// 动爻:装变卦对应爻(六亲仍以本宫五行论)
		if yao.Moving {
			var bs rune
			var bb rune
			if i < 3 {
				na := najia[bianLower.Num-1]
				bs, bb = na.innerStem, na.inner[i]
			} else {
				na := najia[bianUpper.Num-1]
				bs, bb = na.outerStem, na.outer[i-3]
			}
			bbIdx := 0
			for j, br := range branches {
				if br == bb {
					bbIdx = j
				}
			}
			yao.BianYao = &Yao{
				Pos: pos, Yang: bianLines[i],
				Stem: string(bs), Branch: string(bb),
				Element: elementNames[branchElement[bbIdx]],
				LiuQin:  liuQin(palaceEl, branchElement[bbIdx]),
			}
		}
		r.Yaos[i] = yao
		if yao.Moving {
			r.MovingNums = append(r.MovingNums, pos)
		}
	}
	return r, nil
}

// ── 起卦入口 ─────────────────────────────────────────────────

// tossToYao 背面数 → (阳爻, 动)。1背=少阳、2背=少阴、3背=老阳动、0背=老阴动。
func tossToYao(backs int) (yang, moving bool, err error) {
	switch backs {
	case 1:
		return true, false, nil
	case 2:
		return false, false, nil
	case 3:
		return true, true, nil
	case 0:
		return false, true, nil
	default:
		return false, false, fmt.Errorf("每爻背面数须为 0-3,得到 %d", backs)
	}
}

// dayGanZhi 当前时刻 → 农历日干支索引与月建。
func dayGanZhi(t time.Time) (dayStem, dayBranch int, monthJian rune, lunarText string, err error) {
	if t.Year() < 1902 || t.Year() > 2098 {
		return 0, 0, ' ', "", fmt.Errorf("时间超出支持范围(1902-2098)")
	}
	lunar := calendar.NewSolarFromDate(t).GetLunar()
	dgz := []rune(lunar.GetDayInGanZhi())
	if len(dgz) != 2 {
		return 0, 0, ' ', "", fmt.Errorf("日干支解析失败")
	}
	for i, s := range stems {
		if s == dgz[0] {
			dayStem = i
		}
	}
	for i, b := range branches {
		if b == dgz[1] {
			dayBranch = i
		}
	}
	mgz := []rune(lunar.GetMonthInGanZhi())
	monthJian = mgz[len(mgz)-1]
	lunarText = fmt.Sprintf("%s月%s日(%s日)", lunar.GetMonthInChinese(), lunar.GetDayInChinese(), lunar.GetDayInGanZhi())
	return dayStem, dayBranch, monthJian, lunarText, nil
}

// ByTosses 报爻起卦:六爻背面数(自下而上)。
func ByTosses(tosses []int, at time.Time, question string) (*Result, error) {
	if len(tosses) != 6 {
		return nil, fmt.Errorf("六爻需六次摇卦记录")
	}
	var lines [6]bool
	var moving []int
	for i, backs := range tosses {
		yang, mv, err := tossToYao(backs)
		if err != nil {
			return nil, err
		}
		lines[i] = yang
		if mv {
			moving = append(moving, i+1)
		}
	}
	ds, db, mj, lt, err := dayGanZhi(at)
	if err != nil {
		return nil, err
	}
	r, err := assemble(lines, moving, ds, db, mj)
	if err != nil {
		return nil, err
	}
	r.Question = question
	r.LunarText = lt
	r.Tosses = append([]int(nil), tosses...)
	return r, nil
}

// Shake 服务端摇卦:crypto/rand 模拟三枚铜钱六掷。
func Shake(at time.Time, question string) (*Result, error) {
	tosses := make([]int, 6)
	buf := make([]byte, 18) // 6 爻 × 3 枚
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	for i := 0; i < 6; i++ {
		backs := 0
		for c := 0; c < 3; c++ {
			if buf[i*3+c]%2 == 0 {
				backs++
			}
		}
		tosses[i] = backs
	}
	return ByTosses(tosses, at, question)
}
