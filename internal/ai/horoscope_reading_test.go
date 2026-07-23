package ai

import (
	"strings"
	"testing"

	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// TestHoroscopeReading 运限逐层断语:五层齐备、各层贴本命、流年含四化飞宫。
func TestHoroscopeReading(t *testing.T) {
	c, err := ziwei.Generate(ziwei.BirthInfo{Year: 1990, Month: 6, Day: 15, Hour: 6, Gender: ziwei.Male}, ziwei.Options{ReferenceYear: 2026})
	if err != nil {
		t.Fatal(err)
	}
	h, err := ziwei.GenerateHoroscope(c, 2026, 8, 20, 6)
	if err != nil {
		t.Fatal(err)
	}
	hr := buildHoroscopeReading(c, h)
	if hr == nil || len(hr.Sections) != 5 {
		t.Fatalf("运限断语应为 5 层(大限/流年/流月/流日/流时),得 %d", len(hr.Sections))
	}
	byKey := map[string]ReadingSection{}
	for _, s := range hr.Sections {
		byKey[s.Key] = s
	}
	for _, k := range []string{"decadal", "yearly", "monthly", "daily", "hourly"} {
		s, ok := byKey[k]
		if !ok {
			t.Errorf("缺运限层 %s", k)
			continue
		}
		if s.Text == "" || !strings.Contains(s.Text, "命宫落于本命") {
			t.Errorf("%s 层断语应含「命宫落于本命」,实际:%s", k, s.Text)
		}
	}
	// 流年层须含流年四化飞入本命宫位(倪师认可的动态层)。
	if !strings.Contains(byKey["yearly"].Text, "流年四化") || !strings.Contains(byKey["yearly"].Text, "飞入本命") {
		t.Errorf("流年层应含四化飞宫,实际:%s", byKey["yearly"].Text)
	}
	// 流月/流日/流时须声明四化属飞星派、不作定断。
	for _, k := range []string{"monthly", "daily", "hourly"} {
		if !strings.Contains(byKey[k].Text, "飞星派") {
			t.Errorf("%s 层应声明宫干四化属飞星派仅供参考,实际:%s", k, byKey[k].Text)
		}
	}

	// 不同目标日期,流日断语应不同(下钻到日的意义所在)。
	h2, err := ziwei.GenerateHoroscope(c, 2026, 8, 25, 6)
	if err != nil {
		t.Fatal(err)
	}
	hr2 := buildHoroscopeReading(c, h2)
	var d1, d2 string
	for _, s := range hr.Sections {
		if s.Key == "daily" {
			d1 = s.Text
		}
	}
	for _, s := range hr2.Sections {
		if s.Key == "daily" {
			d2 = s.Text
		}
	}
	if d1 == d2 {
		t.Error("不同日期的流日断语不应雷同")
	}
}
