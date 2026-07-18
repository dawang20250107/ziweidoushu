package liuyao

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/6tail/lunar-go/calendar"
	"github.com/dawang20250107/ziweidoushu/internal/meihua"
)

// 《卜筮正宗》八宫六十四卦定表(每宫八卦:纯/一世…五世/游魂/归魂)。
// 程序按变爻规律生成宫属,与此定表逐卦对照——任何纳甲/宫属规则偏差都会在此爆炸。
var bagongTable = map[string][8]string{
	"乾": {"乾为天", "天风姤", "天山遁", "天地否", "风地观", "山地剥", "火地晋", "火天大有"},
	"坎": {"坎为水", "水泽节", "水雷屯", "水火既济", "泽火革", "雷火丰", "地火明夷", "地水师"},
	"艮": {"艮为山", "山火贲", "山天大畜", "山泽损", "火泽睽", "天泽履", "风泽中孚", "风山渐"},
	"震": {"震为雷", "雷地豫", "雷水解", "雷风恒", "地风升", "水风井", "泽风大过", "泽雷随"},
	"巽": {"巽为风", "风天小畜", "风火家人", "风雷益", "天雷无妄", "火雷噬嗑", "山雷颐", "山风蛊"},
	"离": {"离为火", "火山旅", "火风鼎", "火水未济", "山水蒙", "风水涣", "天水讼", "天火同人"},
	"坤": {"坤为地", "地雷复", "地泽临", "地天泰", "雷天大壮", "泽天夬", "水天需", "水地比"},
	"兑": {"兑为泽", "泽水困", "泽地萃", "泽山咸", "水山蹇", "地山谦", "雷山小过", "雷泽归妹"},
}

// TestBaGongTable 八宫归属:程序生成 vs 卜筮正宗定表,64 卦全对照。
func TestBaGongTable(t *testing.T) {
	count := 0
	for p := 0; p < 8; p++ {
		tr := meihua.TrigramByNum(p + 1)
		pure := [6]bool{tr.Lines[0], tr.Lines[1], tr.Lines[2], tr.Lines[0], tr.Lines[1], tr.Lines[2]}
		want := bagongTable[tr.Name]
		for seq, lines := range palaceHexagrams(pure) {
			lo := meihua.TrigramByLines([3]bool{lines[0], lines[1], lines[2]})
			hi := meihua.TrigramByLines([3]bool{lines[3], lines[4], lines[5]})
			name := meihua.HexagramNameByNums(hi.Num, lo.Num)
			if name != want[seq] {
				t.Fatalf("%s宫第 %d 卦: got %s want %s", tr.Name, seq, name, want[seq])
			}
			entry := palaceIndex[hexKey(lines)]
			if entry.palace != p || entry.seq != seq {
				t.Fatalf("%s 宫属登记错误: %+v", name, entry)
			}
			count++
		}
	}
	if count != 64 {
		t.Fatalf("八宫总卦数: %d", count)
	}
}

// mustAssemble 以固定日干支装卦(甲子日,月建寅)。
func mustAssemble(t *testing.T, lines [6]bool, moving []int) *Result {
	t.Helper()
	r, err := assemble(lines, moving, 0, 0, '寅')
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// TestQianNaJiaLiuQin 乾为天:纳甲/六亲/世应全爻对照(卜筮正宗定式)。
func TestQianNaJiaLiuQin(t *testing.T) {
	all := [6]bool{true, true, true, true, true, true}
	r := mustAssemble(t, all, nil)
	if r.BenName != "乾为天" || r.Palace != "乾宫" || r.PalaceSeq != "八纯卦" {
		t.Fatalf("卦名/宫属: %+v", r)
	}
	want := []struct {
		stem, branch, liuQin string
	}{
		{"甲", "子", "子孙"}, {"甲", "寅", "妻财"}, {"甲", "辰", "父母"},
		{"壬", "午", "官鬼"}, {"壬", "申", "兄弟"}, {"壬", "戌", "父母"},
	}
	for i, w := range want {
		y := r.Yaos[i]
		if y.Stem != w.stem || y.Branch != w.branch || y.LiuQin != w.liuQin {
			t.Fatalf("乾为天第 %d 爻: got %s%s %s want %s%s %s",
				i+1, y.Stem, y.Branch, y.LiuQin, w.stem, w.branch, w.liuQin)
		}
	}
	if !r.Yaos[5].IsShi || !r.Yaos[2].IsYing {
		t.Fatalf("纯卦世应: 世6应3, got 世%v 应%v", r.Yaos[5].IsShi, r.Yaos[2].IsYing)
	}
	// 甲日六神:初爻青龙起
	if r.Yaos[0].LiuShen != "青龙" || r.Yaos[5].LiuShen != "玄武" {
		t.Fatalf("甲日六神: %s...%s", r.Yaos[0].LiuShen, r.Yaos[5].LiuShen)
	}
}

// TestDiShuiShi 地水师(坎宫归魂):内坎外坤纳甲、世3应6。
func TestDiShuiShi(t *testing.T) {
	kan := meihua.TrigramByNum(6).Lines
	kun := meihua.TrigramByNum(8).Lines
	lines := [6]bool{kan[0], kan[1], kan[2], kun[0], kun[1], kun[2]}
	r := mustAssemble(t, lines, nil)
	if r.BenName != "地水师" || r.Palace != "坎宫" || r.PalaceSeq != "归魂卦" {
		t.Fatalf("%+v", r)
	}
	wantNa := []string{"戊寅", "戊辰", "戊午", "癸丑", "癸亥", "癸酉"}
	for i, w := range wantNa {
		got := r.Yaos[i].Stem + r.Yaos[i].Branch
		if got != w {
			t.Fatalf("师卦第 %d 爻纳甲: got %s want %s", i+1, got, w)
		}
	}
	if !r.Yaos[2].IsShi || !r.Yaos[5].IsYing {
		t.Fatal("归魂卦应为世3应6")
	}
}

// TestYouHunShiYing 游魂卦世4应1(火地晋)。
func TestYouHunShiYing(t *testing.T) {
	kun := meihua.TrigramByNum(8).Lines
	li := meihua.TrigramByNum(3).Lines
	lines := [6]bool{kun[0], kun[1], kun[2], li[0], li[1], li[2]}
	r := mustAssemble(t, lines, nil)
	if r.BenName != "火地晋" || r.PalaceSeq != "游魂卦" {
		t.Fatalf("%+v", r.BenName)
	}
	if !r.Yaos[3].IsShi || !r.Yaos[0].IsYing {
		t.Fatal("游魂卦应为世4应1")
	}
}

// TestMovingBianGua 乾为天动初爻 → 变天风姤,变爻纳巽内卦辛丑。
func TestMovingBianGua(t *testing.T) {
	all := [6]bool{true, true, true, true, true, true}
	r := mustAssemble(t, all, []int{1})
	if r.BianName != "天风姤" {
		t.Fatalf("变卦: got %s want 天风姤", r.BianName)
	}
	b := r.Yaos[0].BianYao
	if b == nil || b.Stem+b.Branch != "辛丑" {
		t.Fatalf("变爻纳甲: %+v", b)
	}
	if len(r.MovingNums) != 1 || r.MovingNums[0] != 1 {
		t.Fatalf("动爻: %v", r.MovingNums)
	}
}

// TestLiuShenByDay 戊日初爻起勾陈。
func TestLiuShenByDay(t *testing.T) {
	all := [6]bool{true, true, true, true, true, true}
	r, err := assemble(all, nil, 4, 4, '寅') // 戊辰日
	if err != nil {
		t.Fatal(err)
	}
	if r.Yaos[0].LiuShen != "勾陈" || r.Yaos[1].LiuShen != "腾蛇" {
		t.Fatalf("戊日六神: %s %s", r.Yaos[0].LiuShen, r.Yaos[1].LiuShen)
	}
}

// TestByTosses 摇卦映射:3背=老阳动、0背=老阴动、1背=少阳、2背=少阴。
func TestByTosses(t *testing.T) {
	// 农历一个确定日子(2024 正月初一)
	lunar := calendar.NewLunar(2024, 1, 1, 10, 0, 0)
	solar := lunar.GetSolar()
	at := time.Date(solar.GetYear(), time.Month(solar.GetMonth()), solar.GetDay(), 10, 0, 0, 0, time.Local)

	r, err := ByTosses([]int{3, 1, 1, 2, 2, 0}, at, "测试")
	if err != nil {
		t.Fatal(err)
	}
	// 爻象:阳阳阳 阴阴阴 = 内乾外坤 → 地天泰;动爻 1、6
	if r.BenName != "地天泰" {
		t.Fatalf("本卦: got %s want 地天泰", r.BenName)
	}
	if len(r.MovingNums) != 2 || r.MovingNums[0] != 1 || r.MovingNums[1] != 6 {
		t.Fatalf("动爻: %v", r.MovingNums)
	}
	if !r.Yaos[0].Moving || r.Yaos[0].BianYao == nil {
		t.Fatal("初爻应为动爻且有变爻")
	}
	if _, err := ByTosses([]int{4, 1, 1, 1, 1, 1}, at, ""); err == nil {
		t.Fatal("非法背面数应拒绝")
	}
}

// TestWangShuaiAnnotations 旺衰/旬空/月破/日辰作用标注(规则见 research/liuyao-wangshuai.md)。
func TestWangShuaiAnnotations(t *testing.T) {
	all := [6]bool{true, true, true, true, true, true}
	// 甲子日、月建寅,乾为天:爻支自下而上 子寅辰午申戌
	r := mustAssemble(t, all, nil)
	want := []struct {
		state, dayRel             string
		yuePo, kong, anDong, riPo bool
	}{
		{"休", "临", false, false, false, false}, // 子水:水生木令为休;临日辰
		{"旺", "生", false, false, false, false}, // 寅木:当令;日辰子水生之
		{"死", "耗", false, false, false, false}, // 辰土:木令克土;爻克日为耗
		{"相", "冲", false, false, true, false},  // 午火:令生为相;子冲午,旺相静爻=暗动
		{"囚", "泄", true, false, false, false},  // 申金:克令为囚;寅申冲=月破;爻生日为泄
		{"死", "耗", false, true, false, false},  // 戌土:死;甲子旬戌亥空
	}
	for i, w := range want {
		y := r.Yaos[i]
		if y.MonthState != w.state || y.DayRelation != w.dayRel ||
			y.YuePo != w.yuePo || y.XunKong != w.kong || y.AnDong != w.anDong || y.RiPo != w.riPo {
			t.Fatalf("第 %d 爻标注: state=%s rel=%s 破%v 空%v 暗%v 日破%v, want %+v",
				i+1, y.MonthState, y.DayRelation, y.YuePo, y.XunKong, y.AnDong, y.RiPo, w)
		}
	}
	// 增删卜易:「静爻休囚,日辰冲之,曰日破」——子月午火死地,子日冲之
	r2, err := assemble(all, nil, 0, 0, '子')
	if err != nil {
		t.Fatal(err)
	}
	if y := r2.Yaos[3]; !y.RiPo || y.AnDong || y.MonthState != "死" {
		t.Fatalf("子月子日午爻应为日破: %+v", y)
	}
	// 动爻逢日冲不入暗动/日破(另论)
	r3, err := assemble(all, []int{4}, 0, 0, '寅')
	if err != nil {
		t.Fatal(err)
	}
	if y := r3.Yaos[3]; y.AnDong || y.RiPo {
		t.Fatalf("动爻不应标暗动/日破: %+v", y)
	}
	// 六甲旬空诀全表对照
	wantKong := [6][2]int{{10, 11}, {8, 9}, {6, 7}, {4, 5}, {2, 3}, {0, 1}}
	for xun := 0; xun < 6; xun++ {
		a, b := xunKongBranches(0, (xun*10)%12)
		if a != wantKong[xun][0] || b != wantKong[xun][1] {
			t.Fatalf("第 %d 旬空亡: got %d,%d want %v", xun, a, b, wantKong[xun])
		}
	}
}

// TestDongBianAnnotations 生旺墓绝四态与动变作用(规则见 research/liuyao-dongbian.md)。
func TestDongBianAnnotations(t *testing.T) {
	// 日辰四态(野鹤口径:金巳生 酉旺 丑墓 寅绝;土绝巳论生不论绝)
	stageCases := []struct {
		el, db int
		want   string
	}{
		{3, 5, "长生"}, {3, 9, "帝旺"}, {3, 1, "墓"}, {3, 2, "绝"}, // 金
		{0, 11, "长生"}, {0, 7, "墓"}, {0, 8, "绝"}, // 木
		{2, 5, ""}, {2, 8, "长生"}, {2, 4, "墓"}, // 土:绝于巳不标
		{4, 4, "墓"}, {4, 5, "绝"}, // 水
		{1, 6, "帝旺"}, {3, 0, ""}, // 非四态位不标
	}
	for _, c := range stageCases {
		if got := dayStage(c.el, c.db); got != c.want {
			t.Fatalf("dayStage(%d,%d)=%q want %q", c.el, c.db, got, c.want)
		}
	}
	// 动变作用与优先级
	bianCases := []struct {
		ben, bian int
		want      string
	}{
		{2, 3, "化进神"}, {3, 2, "化退神"}, // 寅→卯 / 卯→寅
		{1, 4, "化进神"}, {10, 7, "化退神"}, // 丑→辰 / 戌→未(土)
		{6, 6, "伏吟"},                    // 午→午
		{2, 8, "反吟"}, {5, 11, "反吟"},    // 寅→申 / 巳→亥(冲优先于绝)
		{9, 5, "化长生"},                  // 酉金→巳:论长生不论克
		{6, 10, "化墓"}, {9, 1, "化墓"},    // 午→戌(火墓)/ 酉→丑(金墓,墓优先于回头生)
		{6, 11, "化绝"}, {3, 8, "化绝"},    // 午→亥(绝优先于回头克)/ 卯→申
		{0, 1, "化合"},                    // 子→丑(合优先于回头克)
		{7, 5, "回头生"}, // 未土→巳:土绝巳论生
		{0, 4, "化墓"},  // 子水→辰:辰乃水墓,墓优先于回头克
		{5, 0, "回头克"}, // 巳火→子水
		{8, 0, ""},     // 申金→子:化泄不标
	}
	for _, c := range bianCases {
		if got := bianRelation(c.ben, c.bian); got != c.want {
			t.Fatalf("bianRelation(%d,%d)=%q want %q", c.ben, c.bian, got, c.want)
		}
	}
	// 集成:乾动初爻(甲子日寅月)→ 姤,子化辛丑=化合
	all := [6]bool{true, true, true, true, true, true}
	r := mustAssemble(t, all, []int{1})
	if r.Yaos[0].BianRelation != "化合" {
		t.Fatalf("乾初动子化丑: %q", r.Yaos[0].BianRelation)
	}
	// 甲子日:土爻(辰/戌)帝旺于子(野鹤「土长生于申,旺于子」),申金爻非四态位
	if r.Yaos[5].DayStage != "帝旺" || r.Yaos[2].DayStage != "帝旺" || r.Yaos[4].DayStage != "" {
		t.Fatalf("子日四态: 戌%q 辰%q 申%q", r.Yaos[5].DayStage, r.Yaos[2].DayStage, r.Yaos[4].DayStage)
	}
}

// TestStaticGuaJSONContract 静卦 movingNums 序列化为 [] 而非 null(前端列表契约)。
func TestStaticGuaJSONContract(t *testing.T) {
	all := [6]bool{true, true, true, true, true, true}
	r := mustAssemble(t, all, nil)
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"movingNums":[]`) {
		t.Fatalf("静卦 movingNums 应为 []: %s", b)
	}
}

// TestYongShenSuggest 用神事类映射(人物优先于事类)与装卦定位。
func TestYongShenSuggest(t *testing.T) {
	cases := []struct{ q, want string }{
		{"此番求职能否成", "官鬼"},
		{"投资能否获利", "妻财"},
		{"母亲病情如何", "父母"},
		{"儿子考试能否上榜", "子孙"}, // 所占之人定用神,非事类
		{"明日有雨否", "父母"},
		{"明日天晴否", "子孙"},
		{"老屋翻修可动工否", "父母"},
		{"丈夫升迁如何", "官鬼"},
		{"此事可成否", "世爻"},
	}
	for _, c := range cases {
		if got, _ := SuggestYongShen(c.q); got != c.want {
			t.Fatalf("%q → %s want %s", c.q, got, c.want)
		}
	}
	// 集成:地天泰(坤宫,世3)问财 → 妻财爻位 1、5(甲子水/癸亥水)
	lunar := calendar.NewLunar(2024, 1, 1, 10, 0, 0)
	solar := lunar.GetSolar()
	at := time.Date(solar.GetYear(), time.Month(solar.GetMonth()), solar.GetDay(), 10, 0, 0, 0, time.Local)
	r, err := ByTosses([]int{1, 1, 1, 2, 2, 2}, at, "投资能否获利")
	if err != nil {
		t.Fatal(err)
	}
	if r.YongShen != "妻财" || len(r.YongShenPos) != 2 || r.YongShenPos[0] != 1 || r.YongShenPos[1] != 5 {
		t.Fatalf("泰卦问财用神: %s %v", r.YongShen, r.YongShenPos)
	}
	// 无事类 → 世爻(泰为三世卦,世在 3)
	r2, err := ByTosses([]int{1, 1, 1, 2, 2, 2}, at, "")
	if err != nil {
		t.Fatal(err)
	}
	if r2.YongShen != "世爻" || len(r2.YongShenPos) != 1 || r2.YongShenPos[0] != 3 {
		t.Fatalf("默认世爻用神: %s %v", r2.YongShen, r2.YongShenPos)
	}
}

// TestShake 服务端摇卦:结构合法(可多次)。
func TestShake(t *testing.T) {
	at := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 20; i++ {
		r, err := Shake(at, "")
		if err != nil {
			t.Fatal(err)
		}
		if r.BenName == "" || r.Palace == "" {
			t.Fatalf("摇卦结果不完整: %+v", r)
		}
		shi, ying := 0, 0
		for _, y := range r.Yaos {
			if y.IsShi {
				shi++
			}
			if y.IsYing {
				ying++
			}
		}
		if shi != 1 || ying != 1 {
			t.Fatalf("世应各应恰一: 世%d 应%d", shi, ying)
		}
	}
}
