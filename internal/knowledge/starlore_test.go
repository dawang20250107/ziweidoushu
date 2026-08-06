package knowledge_test

import (
	"strings"
	"testing"

	"github.com/dawang20250107/ziweidoushu/data"
	"github.com/dawang20250107/ziweidoushu/internal/knowledge"
	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// TestStarLoreComplete 星曜档案完备性:真实排盘 + 运限叠加收集盘上出现的
// 全部星名(主星/辅煞/杂曜/补曜/流曜)与四大十二神名目,逐一断言有档——
// 引擎新增任何星曜而档案缺失时,此测试即失败。
func TestStarLoreComplete(t *testing.T) {
	kb, err := knowledge.Load(data.FS)
	if err != nil {
		t.Fatal(err)
	}
	if len(kb.StarLore) < 60 {
		t.Fatalf("星曜档案应覆盖全量星曜,仅 %d 条", len(kb.StarLore))
	}

	// 多组生辰扩大覆盖(不同年支/时辰安星各异)
	births := []ziwei.BirthInfo{
		{Year: 1993, Month: 11, Day: 20, Hour: 4, Gender: "female"},
		{Year: 1994, Month: 2, Day: 24, Hour: 7, Gender: "male"},
		{Year: 2000, Month: 6, Day: 15, Hour: 0, Gender: "male"},
	}
	flowPrefix := "运流月日时"
	for _, b := range births {
		c, err := ziwei.Generate(b, ziwei.Options{})
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range c.Palaces {
			for _, s := range p.Stars {
				if _, ok := kb.StarLore[s.Name]; !ok {
					t.Errorf("盘面星曜 %s 无档案", s.Name)
				}
			}
			for _, s := range p.ExtraStars {
				if _, ok := kb.StarLore[s.Name]; !ok {
					t.Errorf("补充杂曜 %s 无档案", s.Name)
				}
			}
			if p.Changsheng12 != "" {
				if _, ok := kb.StarCycles["changsheng12"][p.Changsheng12]; !ok {
					t.Errorf("长生十二神 %s 无义", p.Changsheng12)
				}
			}
			if p.Boshi12 != "" {
				if _, ok := kb.StarCycles["boshi12"][p.Boshi12]; !ok {
					t.Errorf("博士十二神 %s 无义", p.Boshi12)
				}
			}
		}
		// 运限层:流曜按去前缀字义查档;岁前/将前十二神逐名有义
		h, err := ziwei.GenerateHoroscope(c, 2026, 8, 6, 5)
		if err != nil {
			t.Fatal(err)
		}
		for _, sc := range []ziwei.HoroscopeScope{h.Decadal, h.Yearly, h.Monthly, h.Daily, h.Hourly} {
			for _, cell := range sc.Stars {
				for _, s := range cell {
					if _, ok := kb.StarLore[s.Name]; ok {
						continue // 不带前缀的流层星曜(如流年「年解」)直接入档
					}
					r := []rune(s.Name)
					if len(r) != 2 || !strings.ContainsRune(flowPrefix, r[0]) {
						t.Errorf("流曜命名异常: %s", s.Name)
						continue
					}
					if _, ok := kb.StarFlow[string(r[1])]; !ok {
						t.Errorf("流曜 %s 无义(键 %s)", s.Name, string(r[1]))
					}
				}
			}
		}
		for _, name := range h.Suiqian12 {
			if _, ok := kb.StarCycles["suiqian12"][name]; !ok {
				t.Errorf("岁前十二神 %s 无义", name)
			}
		}
		for _, name := range h.Jiangqian12 {
			if _, ok := kb.StarCycles["jiangqian12"][name]; !ok {
				t.Errorf("将前十二神 %s 无义", name)
			}
		}
	}

	// 档案自身质量:每条有主司与义理,主星层五行/化气齐备
	for name, l := range kb.StarLore {
		if l.Si == "" || len([]rune(l.Gist)) < 18 {
			t.Errorf("星曜 %s 档案过薄: si=%q gist=%q", name, l.Si, l.Gist)
		}
	}
	for _, major := range kb.StarOrder {
		l, ok := kb.StarLore[major]
		if !ok || l.Element == "" || l.Hua == "" {
			t.Errorf("主星 %s 应有五行与化气", major)
		}
	}
}
