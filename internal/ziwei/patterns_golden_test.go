package ziwei

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

// 格局差分基准:对与安星黄金基准相同的 1566 个输入,
// 运行 TS 原版 patterns.ts(detectPatterns + getMingGongSummary)生成的结果,
// 与 Go 移植版逐例逐字段比对。

type goldenPatternCondition struct {
	Required []string `json:"required"`
	Bonus    []string `json:"bonus"`
	Breaking []string `json:"breaking"`
}

type goldenPattern struct {
	Name        string                  `json:"name"`
	Level       string                  `json:"level"`
	Description string                  `json:"description"`
	Palaces     []string                `json:"palaces"`
	Conditions  *goldenPatternCondition `json:"conditions"`
	Source      string                  `json:"source"`
}

type goldenPatternCase struct {
	Input struct {
		Year   int    `json:"year"`
		Month  int    `json:"month"`
		Day    int    `json:"day"`
		Hour   int    `json:"hour"`
		Gender string `json:"gender"`
	} `json:"input"`
	Patterns []goldenPattern `json:"patterns"`
	Summary  struct {
		Stars    []string `json:"stars"`
		Keywords []string `json:"keywords"`
		Nature   string   `json:"nature"`
	} `json:"summary"`
}

func TestPatternsGoldenParity(t *testing.T) {
	f, err := os.Open("testdata/patterns_golden.json.gz")
	if err != nil {
		t.Fatalf("打开格局基准失败: %v", err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	var cases []goldenPatternCase
	if err := json.NewDecoder(zr).Decode(&cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 1000 {
		t.Fatalf("格局基准用例数异常: %d", len(cases))
	}

	failures := 0
	fail := func(label, format string, args ...any) {
		failures++
		if failures <= 30 {
			t.Errorf("[%s] %s", label, fmt.Sprintf(format, args...))
		}
	}

	for _, g := range cases {
		label := fmt.Sprintf("%d-%d-%d/h%d/%s", g.Input.Year, g.Input.Month, g.Input.Day, g.Input.Hour, g.Input.Gender)
		chart, err := Generate(BirthInfo{
			Year: g.Input.Year, Month: g.Input.Month, Day: g.Input.Day,
			Hour: g.Input.Hour, Gender: Gender(g.Input.Gender),
		}, Options{ReferenceYear: 2026})
		if err != nil {
			fail(label, "Generate 报错: %v", err)
			continue
		}
		got := DetectPatterns(chart)

		if len(got) != len(g.Patterns) {
			gotNames := make([]string, len(got))
			for i := range got {
				gotNames[i] = got[i].Name
			}
			wantNames := make([]string, len(g.Patterns))
			for i := range g.Patterns {
				wantNames[i] = g.Patterns[i].Name
			}
			fail(label, "格局数量: got %v want %v", gotNames, wantNames)
			continue
		}
		for i, want := range g.Patterns {
			gp := got[i]
			if gp.Name != want.Name || gp.Level != want.Level || gp.Description != want.Description || gp.Source != want.Source {
				fail(label, "格局[%d]: got {%s %s %q %q} want {%s %s %q %q}",
					i, gp.Name, gp.Level, gp.Description, gp.Source,
					want.Name, want.Level, want.Description, want.Source)
				continue
			}
			if !equalStrings(gp.Palaces, want.Palaces) {
				fail(label, "格局[%s]宫位: got %v want %v", gp.Name, gp.Palaces, want.Palaces)
			}
			wantCond := want.Conditions
			gotCond := gp.Conditions
			switch {
			case wantCond == nil && gotCond != nil:
				fail(label, "格局[%s]条件: TS 无而 Go 有 %+v", gp.Name, gotCond)
			case wantCond != nil && gotCond == nil:
				fail(label, "格局[%s]条件: TS 有 %+v 而 Go 无", gp.Name, wantCond)
			case wantCond != nil && gotCond != nil:
				if !equalStrings(gotCond.Required, wantCond.Required) ||
					!equalStrings(gotCond.Bonus, wantCond.Bonus) ||
					!equalStrings(gotCond.Breaking, wantCond.Breaking) {
					fail(label, "格局[%s]条件: got %+v want %+v", gp.Name, gotCond, wantCond)
				}
			}
		}

		// 命宫摘要
		sum := GetMingGongSummary(chart)
		if !equalStrings(sum.Stars, g.Summary.Stars) ||
			!equalStrings(sum.Keywords, g.Summary.Keywords) ||
			sum.Nature != g.Summary.Nature {
			fail(label, "命宫摘要: got %+v want %+v", sum, g.Summary)
		}
	}
	if failures > 0 {
		t.Fatalf("格局差分比对失败共 %d 处(仅显示前 30 处)", failures)
	}
}

// equalStrings 空切片与 nil 等价。
func equalStrings(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	return reflect.DeepEqual(a, b)
}
