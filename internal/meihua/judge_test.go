package meihua

import (
	"strings"
	"testing"
)

// 观梅占(《梅花易数》卷一):辰年十二月十七日申时。
// 年5+月12+日17=34→兑上;+时9=43→离下;43%6=1 动初爻。
// 泽火革之咸,互天风姤。兑金为体、离火克之(用克体),互巽木生起离火
// (克体之势盛),幸互乾比和、变艮生金——书断:有女子折花伤股,伤而不危。
// 引擎应同向:用克体 + 体党有救 → neutral(凶中有救、小损不致大败)。
func TestGuanMeiJudge(t *testing.T) {
	r := derive(2, 3, 1)
	if r.Ben.Name != "泽火革" || r.Bian.Name != "泽山咸" || r.Hu.Name != "天风姤" {
		t.Fatalf("卦象: 本%s 变%s 互%s, want 泽火革/泽山咸/天风姤", r.Ben.Name, r.Bian.Name, r.Hu.Name)
	}
	if r.TiTrigram.Name != "兑" || r.YongTrigram.Name != "离" || r.Relation != RelYongKeTi {
		t.Fatalf("体用: 体%s 用%s %s, want 体兑 用离 用克体", r.TiTrigram.Name, r.YongTrigram.Name, r.Relation)
	}
	j := r.Judge(12, "")
	// 十二月土旺,土生金 → 体气「相」
	if j.TiQi != "相" {
		t.Errorf("体气: got %s want 相(十二月土旺生金)", j.TiQi)
	}
	// 用克体本凶,但乾比和/兑比和/艮生体救之(巽生用记不利)→ 平,凶中有救
	if j.Level != "neutral" {
		t.Errorf("观梅占应平(凶中有救),got %s(score %d)", j.Level, j.Score)
	}
	if !strings.Contains(j.Conclusion, "救") {
		t.Errorf("结论应含凶中有救之义: %s", j.Conclusion)
	}
	// 巽木生起克体之离火,应记入克伐之势
	joined := strings.Join(j.Points, " ")
	if !strings.Contains(joined, "巽(生起克体之离)") {
		t.Errorf("体党用党应记「巽生起离」助攻: %s", joined)
	}
}

// 牡丹占(《梅花易数》卷一):巳年三月十六日卯时。
// 年6+月3+日16=25→乾上;+时4=29→巽下;29%6=5 动五爻。
// 天风姤之鼎,互见重乾。巽木为体、乾金克之,互变克体之卦多而生意薄——
// 书断:牡丹为马践毁。引擎应同向:caution(克体众、宜守避)。
func TestMuDanJudge(t *testing.T) {
	r := derive(1, 5, 5)
	if r.Ben.Name != "天风姤" || r.Bian.Name != "火风鼎" {
		t.Fatalf("卦象: 本%s 变%s, want 天风姤/火风鼎", r.Ben.Name, r.Bian.Name)
	}
	if r.Hu.Upper.Name != "乾" || r.Hu.Lower.Name != "乾" {
		t.Fatalf("互卦应见重乾, got %s/%s", r.Hu.Upper.Name, r.Hu.Lower.Name)
	}
	if r.TiTrigram.Name != "巽" || r.Relation != RelYongKeTi {
		t.Fatalf("体用: 体%s %s, want 体巽 用克体", r.TiTrigram.Name, r.Relation)
	}
	j := r.Judge(3, "")
	// 三月土旺,木克土 → 体气「囚」
	if j.TiQi != "囚" {
		t.Errorf("体气: got %s want 囚(三月土旺,木克当令)", j.TiQi)
	}
	if j.Level != "caution" {
		t.Errorf("牡丹占应慎(克体众而生意薄),got %s(score %d)", j.Level, j.Score)
	}
}

// 事类专断与应期:求财问题触发求财口径;应期含生体之气与卦数。
func TestTopicAndYingQi(t *testing.T) {
	r := derive(2, 3, 1) // 体兑金,用克体
	j := r.Judge(12, "今年投资生意的财运如何")
	if j.Topic != "求财" {
		t.Errorf("事类: got %s want 求财", j.Topic)
	}
	if !strings.Contains(j.TopicNote, "财") {
		t.Errorf("求财专断缺失: %s", j.TopicNote)
	}
	if !strings.Contains(j.YingQi, "卦数 6") { // 兑2+离3+动1
		t.Errorf("应期应含卦数 6: %s", j.YingQi)
	}
	// 未识别事类 → 谋事通断
	if got := r.Judge(12, "随便看看").Topic; got != "谋事" {
		t.Errorf("默认事类: got %s want 谋事", got)
	}
}

// 万物类象表:八卦齐备,方位正确抽验。
func TestLore(t *testing.T) {
	for _, name := range []string{"乾", "兑", "离", "震", "巽", "坎", "艮", "坤"} {
		if _, ok := LoreOf(name); !ok {
			t.Errorf("类象缺 %s", name)
		}
	}
	qian, _ := LoreOf("乾")
	if qian.Fangwei != "西北" {
		t.Errorf("乾方位: got %s want 西北", qian.Fangwei)
	}
}
