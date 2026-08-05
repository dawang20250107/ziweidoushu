package liuyao

// 补强层测试:六冲六合卦性、伏神之法、三合局、应期具体地支、断语新层接线。

import (
	"strings"
	"testing"
)

// 乾为天:八纯卦皆六冲(子午/寅申/辰戌)。
func TestGuaXingZhiLiuChong(t *testing.T) {
	r, err := AssembleForResearch([6]bool{true, true, true, true, true, true}, nil, 0, 0, '午')
	if err != nil {
		t.Fatal(err)
	}
	if r.BenXingZhi != "六冲" {
		t.Fatalf("乾为天应为六冲卦,得 %q", r.BenXingZhi)
	}
	if r.BianXingZhi != "" {
		t.Fatalf("静卦不应有变卦卦性,得 %q", r.BianXingZhi)
	}
}

// 地天泰(乾下坤上):子丑/寅亥/辰酉三对相合,六合卦。
func TestGuaXingZhiLiuHe(t *testing.T) {
	r, err := AssembleForResearch([6]bool{true, true, true, false, false, false}, nil, 0, 0, '午')
	if err != nil {
		t.Fatal(err)
	}
	if r.BenXingZhi != "六合" {
		t.Fatalf("地天泰应为六合卦,得 %q", r.BenXingZhi)
	}
}

// 天风姤占财:妻财(寅卯木)不上卦,伏神取乾宫首卦寅木于二爻,
// 飞神亥水生伏——飞来生伏能出;出伏应期含值伏神(寅)与冲飞神(巳)。
func TestFuShen(t *testing.T) {
	// 姤 = 巽下乾上;甲戌日(旬空申酉,不涉飞伏),月建午
	r, err := AssembleForResearch([6]bool{false, true, true, true, true, true}, nil, 0, 10, '午')
	if err != nil {
		t.Fatal(err)
	}
	if r.BenName != "天风姤" {
		t.Fatalf("卦名: %s", r.BenName)
	}
	r.Question = "求财"
	r.applyYongShen()
	r.applyFuShen()
	if r.YongShen != "妻财" || len(r.YongShenPos) != 0 {
		t.Fatalf("姤卦占财应用神妻财不上卦: %s %v", r.YongShen, r.YongShenPos)
	}
	fs := r.FuShen
	if fs == nil {
		t.Fatal("应取出伏神")
	}
	if fs.Branch != "寅" || fs.Pos != 2 || fs.Fei != "亥" {
		t.Fatalf("伏神应为寅伏二爻亥下: %+v", fs)
	}
	if !fs.CanOut || !strings.Contains(fs.Note, "飞来生伏") {
		t.Fatalf("亥水生寅木应判飞来生伏能出: %+v", fs)
	}
	if !strings.Contains(fs.ChuFuRi, "寅") || !strings.Contains(fs.ChuFuRi, "巳") {
		t.Fatalf("出伏应期应含值寅/冲巳: %s", fs.ChuFuRi)
	}
	// Judge 接线:不上卦断语应走伏神实断而非断头路
	j := r.Judge()
	if j.Level != "neutral" || !strings.Contains(strings.Join(j.Points, ""), "伏于二爻亥之下") {
		t.Fatalf("伏神断语接线失败: %+v", j)
	}
}

// 三合局:乾卦三/五爻动(辰申),日建子补一字 → 申子辰水局;日建丑则不成。
func TestSanHeJu(t *testing.T) {
	r, err := AssembleForResearch([6]bool{true, true, true, true, true, true}, []int{3, 5}, 0, 0, '未')
	if err != nil {
		t.Fatal(err)
	}
	el, name, ok := r.sanHeJu()
	if !ok || el != 4 || name != "申子辰水局" {
		t.Fatalf("申辰动+子日应成水局: %v %s %v", el, name, ok)
	}
	r2, _ := AssembleForResearch([6]bool{true, true, true, true, true, true}, []int{3, 5}, 0, 1, '未')
	if _, _, ok := r2.sanHeJu(); ok {
		t.Fatal("缺子字且日月不补,不应成局")
	}
}

// 应期落具体地支:旬空得冲空/填实之日,静旺得值/冲之日。
func TestYingQiConcrete(t *testing.T) {
	got := yingQi(Yao{Branch: "子", Element: "水", XunKong: true})
	if !strings.Contains(got, "午日") || !strings.Contains(got, "子日") {
		t.Fatalf("子空应期应含冲午/值子: %s", got)
	}
	got = yingQi(Yao{Branch: "酉", Element: "金", MonthState: "旺"})
	if !strings.Contains(got, "酉") || !strings.Contains(got, "卯") {
		t.Fatalf("酉旺静应期应含值酉/冲卯: %s", got)
	}
	got = yingQi(Yao{Branch: "午", Element: "火", DayStage: "墓"})
	if !strings.Contains(got, "辰日") { // 火墓戌,冲戌者辰
		t.Fatalf("火入墓应期应含冲开戌墓之辰日: %s", got)
	}
}

// 断语新层接线冒烟:六冲卦性、六神、爻位诸点应入 Points。
func TestJudgeEnrichWiring(t *testing.T) {
	// 乾为天静卦,甲子日(六神起青龙),月建午,自占(世爻用神)
	r, err := AssembleForResearch([6]bool{true, true, true, true, true, true}, nil, 0, 0, '午')
	if err != nil {
		t.Fatal(err)
	}
	r.Question = ""
	r.applyYongShen()
	r.applyFuShen()
	r.applyPower()
	j := r.Judge()
	all := strings.Join(j.Points, "\n")
	if !strings.Contains(all, "六冲") {
		t.Fatalf("六冲卦性未入断语: %s", all)
	}
	if !strings.Contains(all, "爻位:") {
		t.Fatalf("爻位象义未入断语: %s", all)
	}
	if !strings.Contains(all, "用神临") {
		t.Fatalf("六神事象未入断语: %s", all)
	}
}
