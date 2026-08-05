package ai

import (
	"strings"
	"testing"

	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// 四化坐宫细则表:4 化 × 12 宫 = 48 条齐备,无空条目。
func TestSiHuaGongNoteComplete(t *testing.T) {
	palaces := []string{"命宫", "兄弟", "夫妻", "子女", "财帛", "疾厄", "迁移", "仆役", "官禄", "田宅", "福德", "父母"}
	for _, hua := range []ziwei.SiHua{ziwei.HuaLu, ziwei.HuaQuan, ziwei.HuaKe, ziwei.HuaJi} {
		for _, p := range palaces {
			if sihuaGongClause(hua, p) == "" {
				t.Errorf("化%s入%s 细则缺失", hua, p)
			}
		}
	}
}

// 整链:丙年盘(1996-07-14)生年四化 天同禄疾厄/天机权兄弟/文昌科官禄/廉贞忌官禄,
// 断语骨架四化节应携带对应坐宫专断。
func TestSiHuaGongInReading(t *testing.T) {
	c, err := ziwei.Generate(ziwei.BirthInfo{Year: 1996, Month: 7, Day: 14, Hour: 0, Gender: ziwei.Male}, ziwei.Options{ReferenceYear: 2026})
	if err != nil {
		t.Fatal(err)
	}
	sec := sectionForSihua(c)
	if sec == nil {
		t.Fatal("四化节为空")
	}
	for _, kw := range []string{
		"病有速愈",   // 化禄入疾厄
		"平辈中有强人", // 化权入兄弟
		"不大起大落",  // 化科入官禄
		"磨折方成",   // 化忌入官禄
	} {
		if !strings.Contains(sec.Text, kw) {
			t.Errorf("四化节应含坐宫专断 %q,text=%s", kw, sec.Text)
		}
	}
}
