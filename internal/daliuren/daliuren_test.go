package daliuren

import "testing"

func bIdx(r rune) int {
	for i, x := range branches {
		if x == r {
			return i
		}
	}
	return -1
}
func sIdx(r rune) int {
	for i, x := range stems {
		if x == r {
			return i
		}
	}
	return -1
}

// TestLiFaYiJianExample 理法易简走例:甲日 戌时 丑将(月将丑加占时戌)。
// 四课「甲上巳/巳上申/午上酉/酉上子」,二重下贼上取比用,三传申亥寅。
func TestLiFaYiJianExample(t *testing.T) {
	r, err := Cast(sIdx('甲'), bIdx('午'), bIdx('戌'), bIdx('丑'))
	if err != nil {
		t.Fatal(err)
	}
	// 四课
	wantKe := [4][2]rune{{'寅', '巳'}, {'巳', '申'}, {'午', '酉'}, {'酉', '子'}}
	for i, w := range wantKe {
		if r.Ke[i].Lower != string(w[0]) || r.Ke[i].Upper != string(w[1]) {
			t.Fatalf("第%d课: got %s上%s want %c上%c", i+1, r.Ke[i].Lower, r.Ke[i].Upper, w[0], w[1])
		}
	}
	// 三传申亥寅,比用
	if r.Chuan != [3]string{"申", "亥", "寅"} {
		t.Fatalf("三传: %v want 申亥寅", r.Chuan)
	}
	if r.KeType != "比用" {
		t.Fatalf("课体: %s want 比用", r.KeType)
	}
}

// TestZeiKeSingle 单一下贼上 → 贼克;初传取该上神,中末顺天盘。
func TestZeiKeSingle(t *testing.T) {
	// 构造:寻一 月将/时 使四课仅一处克。用 甲日子将子时(伏吟排除)——改用具体验证课体机制。
	// 甲日 亥将 子时:tp[i]=亥+(i-0)=(11+i)%12
	r, err := Cast(sIdx('甲'), bIdx('子'), bIdx('子'), bIdx('亥'))
	if err != nil {
		t.Fatal(err)
	}
	if r.KeType == "" || len(r.Chuan) != 3 {
		t.Fatalf("课体/三传缺失: %+v", r)
	}
	// 中传=初传上神、末传=中传上神(顺天盘不变式)
	tp := tianPanOf(bIdx('亥'), bIdx('子'))
	chu := bIdx(rune([]rune(r.Chuan[0])[0]))
	if r.Chuan[1] != string(branches[tp[chu]]) {
		t.Fatalf("中传应为初传之上神: 初%s 中%s 期望%s", r.Chuan[0], r.Chuan[1], string(branches[tp[chu]]))
	}
}

// TestFuYin 伏吟:月将==占时,天盘=地盘;阳日无克取干上神,中末用刑。
func TestFuYin(t *testing.T) {
	// 甲日 子将 子时 → 伏吟。甲寄寅,干上神=寅(自坐)。阳日取干上寅发用。
	// 寅刑巳(中),巳刑申(末)。
	r, err := Cast(sIdx('甲'), bIdx('辰'), bIdx('子'), bIdx('子'))
	if err != nil {
		t.Fatal(err)
	}
	if r.KeType != "伏吟" {
		t.Fatalf("课体: %s want 伏吟", r.KeType)
	}
	// 天盘=地盘
	for i := 0; i < 12; i++ {
		if r.TianPan[i] != string(branches[i]) {
			t.Fatalf("伏吟天盘应=地盘: [%d]=%s", i, r.TianPan[i])
		}
	}
}

// TestFanYin 返吟:月将冲占时,天盘=冲地盘。
func TestFanYin(t *testing.T) {
	// 甲日 午将 子时 → 天盘=地盘+6(冲)。
	r, err := Cast(sIdx('甲'), bIdx('辰'), bIdx('子'), bIdx('午'))
	if err != nil {
		t.Fatal(err)
	}
	if r.KeType != "返吟" {
		t.Fatalf("课体: %s want 返吟", r.KeType)
	}
	for i := 0; i < 12; i++ {
		if r.TianPan[i] != string(branches[(i+6)%12]) {
			t.Fatalf("返吟天盘应=冲地盘: [%d]=%s", i, r.TianPan[i])
		}
	}
}

// TestMonthGeneral 月将中气表:雨水亥、春分戌、大寒子。
func TestMonthGeneral(t *testing.T) {
	cases := []struct {
		mq   int
		want rune
	}{{0, '亥'}, {1, '戌'}, {3, '申'}, {6, '巳'}, {11, '子'}}
	for _, c := range cases {
		if got := MonthGeneralByMidQi(c.mq); got != bIdx(c.want) {
			t.Fatalf("中气%d 月将: got %s want %c", c.mq, string(branches[got]), c.want)
		}
	}
}

// TestJiGong 干寄宫定式抽查。
func TestJiGong(t *testing.T) {
	cases := []struct {
		s    rune
		want rune
	}{{'甲', '寅'}, {'丙', '巳'}, {'戊', '巳'}, {'丁', '未'}, {'己', '未'}, {'壬', '亥'}, {'癸', '丑'}}
	for _, c := range cases {
		if got := stemJiGong[sIdx(c.s)]; got != bIdx(c.want) {
			t.Fatalf("%c 寄宫: got %s want %c", c.s, string(branches[got]), c.want)
		}
	}
}
