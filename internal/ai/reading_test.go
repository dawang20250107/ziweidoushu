package ai

import (
	"strings"
	"testing"

	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// TestReadingDifferentiation 核心验收:不同命盘的多维断语必须显著不同,
// 且覆盖十二宫多维度(解决「不同人断语雷同」的问题)。
func TestReadingDifferentiation(t *testing.T) {
	a, err := ziwei.Generate(ziwei.BirthInfo{Year: 1990, Month: 6, Day: 15, Hour: 6, Gender: ziwei.Male}, ziwei.Options{ReferenceYear: 2024})
	if err != nil {
		t.Fatal(err)
	}
	b, err := ziwei.Generate(ziwei.BirthInfo{Year: 1985, Month: 11, Day: 20, Hour: 0, Gender: ziwei.Female}, ziwei.Options{ReferenceYear: 2024})
	if err != nil {
		t.Fatal(err)
	}
	ra := buildReading(a, ziwei.DetectPatterns(a))
	rb := buildReading(b, ziwei.DetectPatterns(b))

	// 多维:至少 12 个维度(十二宫,大限视年龄)。
	if len(ra.Sections) < 12 {
		t.Fatalf("维度数应≥12,得 %d", len(ra.Sections))
	}
	// 总论必须不同。
	if ra.Overview == rb.Overview {
		t.Errorf("两盘命格总论不应雷同")
	}
	// 逐宫比对:同名维度的断语大多应不同。
	byKeyB := map[string]string{}
	for _, s := range rb.Sections {
		byKeyB[s.Key] = s.Text
	}
	same, total := 0, 0
	for _, s := range ra.Sections {
		if tb, ok := byKeyB[s.Key]; ok {
			total++
			if tb == s.Text {
				same++
			}
		}
	}
	if total == 0 {
		t.Fatal("无可比对维度")
	}
	if same*100/total > 25 { // 允许极少数空宫巧合雷同,但绝大多数须不同
		t.Errorf("两盘断语雷同度过高:%d/%d 维度相同", same, total)
	}
}

// TestPairTraitOf 双主星组合定名与查断:与星序无关,须归一化到规范名。
func TestPairTraitOf(t *testing.T) {
	cases := []struct{ a, b, wantName string }{
		{"紫微", "破军", "紫微破军"},
		{"破军", "紫微", "紫微破军"}, // 反序须归一
		{"武曲", "贪狼", "武曲贪狼"},
		{"贪狼", "武曲", "武曲贪狼"},
		{"廉贞", "七杀", "廉贞七杀"},
		{"太阴", "太阳", "太阳太阴"},
	}
	for _, c := range cases {
		name, trait := pairTraitOf([]string{c.a, c.b})
		if name != c.wantName || trait == "" {
			t.Errorf("pairTraitOf(%s,%s)=%q(trait空=%v),期望规范名 %s 且有断语",
				c.a, c.b, name, trait == "", c.wantName)
		}
	}
	// 非组合(单星/三星/未知)须返回空。
	if n, _ := pairTraitOf([]string{"紫微"}); n != "" {
		t.Errorf("单主星不应有组合断语,得 %s", n)
	}
	if n, _ := pairTraitOf([]string{"紫微", "天府", "天相"}); n != "" {
		t.Errorf("三主星不应有双星组合断语,得 %s", n)
	}
	// 24 组组合断语键须全部可查且规范有序、无空文。
	for key, trait := range starPairTrait {
		if trait == "" {
			t.Errorf("组合 %s 断语为空", key)
		}
	}
	if len(starPairTrait) != 24 {
		t.Errorf("双主星组合应为 24 组,得 %d", len(starPairTrait))
	}
}

// TestReadingChartGrounded 断语须引用本盘实配星曜(非通用套话)。
func TestReadingChartGrounded(t *testing.T) {
	c, err := ziwei.Generate(ziwei.BirthInfo{Year: 1990, Month: 6, Day: 15, Hour: 6, Gender: ziwei.Male}, ziwei.Options{ReferenceYear: 2024})
	if err != nil {
		t.Fatal(err)
	}
	rd := buildReading(c, ziwei.DetectPatterns(c))
	// 命宫维度的断语应含命宫的某颗主星名(或借星)。
	ming := c.MingGong()
	stars := ming.MajorStarNames()
	if len(stars) == 0 {
		stars = ming.BorrowedStars
	}
	if len(stars) == 0 {
		return // 极少数命宫空且无借星,跳过
	}
	var mingText string
	for _, s := range rd.Sections {
		if s.Key == "ming" {
			mingText = s.Text
		}
	}
	hit := false
	for _, name := range stars {
		if strings.Contains(mingText, name) {
			hit = true
		}
	}
	if !hit {
		t.Errorf("命宫断语应引用命宫主星%v,实际:%s", stars, mingText)
	}
	// 每个维度都应有 level 且文本非空。
	for _, s := range rd.Sections {
		if s.Text == "" || s.Level == "" {
			t.Errorf("维度 %s 断语或吉凶档缺失", s.Key)
		}
	}
}
