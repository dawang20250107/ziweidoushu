package daliuren

// 别责/八专走例(《六壬大全·课经集》原例钉死)与旬空/遁干定式。

import "testing"

// 丙辰日卯时辰将:四课三课备无克无遥,刚日丙合辛、辛寄戌、戌上亥发用,
// 中末俱并干上午 → 三传亥午午,别责。
func TestBieZeExample(t *testing.T) {
	r, err := Cast(sIdx('丙'), bIdx('辰'), bIdx('卯'), bIdx('辰'))
	if err != nil {
		t.Fatal(err)
	}
	if r.KeType != "别责" {
		t.Fatalf("课体: %s want 别责", r.KeType)
	}
	if r.Chuan != [3]string{"亥", "午", "午"} {
		t.Fatalf("三传: %v want 亥午午", r.Chuan)
	}
}

// 甲寅日辰时丑将:干支同位无克(不论申金遥克),刚日干上亥顺数三位至丑,
// 中末俱并干上 → 三传丑亥亥,八专。
func TestBaZhuanYangExample(t *testing.T) {
	r, err := Cast(sIdx('甲'), bIdx('寅'), bIdx('辰'), bIdx('丑'))
	if err != nil {
		t.Fatal(err)
	}
	if r.KeType != "八专" {
		t.Fatalf("课体: %s want 八专", r.KeType)
	}
	if r.Chuan != [3]string{"丑", "亥", "亥"} {
		t.Fatalf("三传: %v want 丑亥亥", r.Chuan)
	}
}

// 丁未日丑时辰将:柔日第四课上神丑逆数三位至亥,中末俱并干上戌 →
// 三传亥戌戌,八专(帷薄不修格之走例)。
func TestBaZhuanYinExample(t *testing.T) {
	r, err := Cast(sIdx('丁'), bIdx('未'), bIdx('丑'), bIdx('辰'))
	if err != nil {
		t.Fatal(err)
	}
	if r.KeType != "八专" {
		t.Fatalf("课体: %s want 八专", r.KeType)
	}
	if r.Chuan != [3]string{"亥", "戌", "戌"} {
		t.Fatalf("三传: %v want 亥戌戌", r.Chuan)
	}
}

// 旬空/遁干:己卯日属甲戌旬,空申酉(断案案01 书头「申酉空亡」);
// 三传巳戌卯遁干辛甲己(书载「辛巳/甲戌/己卯」)。
func TestXunKongDunGan(t *testing.T) {
	r, err := Cast(sIdx('己'), bIdx('卯'), bIdx('酉'), bIdx('寅'))
	if err != nil {
		t.Fatal(err)
	}
	if r.XunKong != [2]string{"申", "酉"} {
		t.Fatalf("旬空: %v want 申酉", r.XunKong)
	}
	if r.ChuanDunGan != [3]string{"辛", "甲", "己"} {
		t.Fatalf("遁干: %v want 辛甲己(辛巳/甲戌/己卯)", r.ChuanDunGan)
	}
	// 传落空亡时遁干为空:甲子旬空戌亥——甲子日戌将子时? 构造含戌之传验证。
	r2, _ := Cast(0, 0, bIdx('子'), bIdx('戌')) // 甲子日子时戌将
	if r2.XunKong != [2]string{"戌", "亥"} {
		t.Fatalf("甲子旬空: %v want 戌亥", r2.XunKong)
	}
	for i, c := range r2.Chuan {
		if (c == "戌" || c == "亥") && r2.ChuanDunGan[i] != "" {
			t.Fatalf("空亡之传不应有遁干: 传%s 遁%s", c, r2.ChuanDunGan[i])
		}
	}
}
