package ziwei

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// 运限黄金基准:tools/goldgen/genhoroscope.js 以 iztro astrolabe.horoscope()
// 生成 1200+ 用例(含童限、闰月目标日、晚子时、老年大限),逐层逐字段比对。

type goldenHoroscopeScope struct {
	Index         int          `json:"index"` // 宫位索引(寅=0)
	Name          string       `json:"name"`
	HeavenlyStem  string       `json:"heavenlyStem"`
	EarthlyBranch string       `json:"earthlyBranch"`
	PalaceNames   []string     `json:"palaceNames"`
	Mutagen       []string     `json:"mutagen"`
	NominalAge    int          `json:"nominalAge"`
	Stars         [][]struct { // 宫位索引空间
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"stars"`
	Suiqian12   []string `json:"suiqian12"`
	Jiangqian12 []string `json:"jiangqian12"`
}

type goldenHoroscopeCase struct {
	Birth struct {
		Year   int    `json:"year"`
		Month  int    `json:"month"`
		Day    int    `json:"day"`
		Hour   int    `json:"hour"`
		Gender string `json:"gender"`
	} `json:"birth"`
	Target struct {
		Y int `json:"y"`
		M int `json:"m"`
		D int `json:"d"`
		H int `json:"h"`
	} `json:"target"`
	SolarDate string               `json:"solarDate"`
	LunarDate string               `json:"lunarDate"`
	Decadal   goldenHoroscopeScope `json:"decadal"`
	Age       goldenHoroscopeScope `json:"age"`
	Yearly    goldenHoroscopeScope `json:"yearly"`
	Monthly   goldenHoroscopeScope `json:"monthly"`
	Daily     goldenHoroscopeScope `json:"daily"`
	Hourly    goldenHoroscopeScope `json:"hourly"`
}

func TestHoroscopeGoldenParity(t *testing.T) {
	f, err := os.Open("testdata/horoscope_golden.json.gz")
	if err != nil {
		t.Fatalf("打开运限基准失败: %v", err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	var cases []goldenHoroscopeCase
	if err := json.NewDecoder(zr).Decode(&cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 1000 {
		t.Fatalf("运限基准用例数异常: %d", len(cases))
	}

	failures := 0
	fail := func(label, format string, args ...any) {
		failures++
		if failures <= 30 {
			t.Errorf("[%s] %s", label, fmt.Sprintf(format, args...))
		}
	}

	for _, g := range cases {
		label := fmt.Sprintf("%d-%d-%d/h%d/%s→%d-%d-%d/h%d",
			g.Birth.Year, g.Birth.Month, g.Birth.Day, g.Birth.Hour, g.Birth.Gender,
			g.Target.Y, g.Target.M, g.Target.D, g.Target.H)

		chart, err := Generate(BirthInfo{
			Year: g.Birth.Year, Month: g.Birth.Month, Day: g.Birth.Day,
			Hour: g.Birth.Hour, Gender: Gender(g.Birth.Gender),
		}, Options{ReferenceYear: 2026})
		if err != nil {
			fail(label, "Generate 报错: %v", err)
			continue
		}
		h, err := GenerateHoroscope(chart, g.Target.Y, g.Target.M, g.Target.D, g.Target.H)
		// iztro 对虚岁越界(>120)返回退化值(index=-1),Go 侧口径为干净报错。
		if g.Decadal.Index < 0 || g.Age.Index < 0 {
			if err == nil {
				fail(label, "虚岁越界应报错(iztro 返回退化值的用例)")
			}
			continue
		}
		if err != nil {
			fail(label, "GenerateHoroscope 报错: %v", err)
			continue
		}

		if h.TargetSolarDate != g.SolarDate {
			fail(label, "目标公历: got %s want %s", h.TargetSolarDate, g.SolarDate)
		}
		if h.TargetLunarText != g.LunarDate {
			fail(label, "目标农历: got %s want %s", h.TargetLunarText, g.LunarDate)
		}
		if h.NominalAge != g.Age.NominalAge {
			fail(label, "虚岁: got %d want %d", h.NominalAge, g.Age.NominalAge)
		}

		compareScope(t, label, "大限", &h.Decadal, &g.Decadal, fail)
		compareScope(t, label, "小限", &h.Age, &g.Age, fail)
		compareScope(t, label, "流年", &h.Yearly, &g.Yearly, fail)
		compareScope(t, label, "流月", &h.Monthly, &g.Monthly, fail)
		compareScope(t, label, "流日", &h.Daily, &g.Daily, fail)
		compareScope(t, label, "流时", &h.Hourly, &g.Hourly, fail)

		// 岁前/将前十二神
		for i, want := range g.Yearly.Suiqian12 {
			if got := h.Suiqian12[fix12(i+2)]; got != want {
				fail(label, "岁前12[%d]: got %s want %s", i, got, want)
				break
			}
		}
		for i, want := range g.Yearly.Jiangqian12 {
			if got := h.Jiangqian12[fix12(i+2)]; got != want {
				fail(label, "将前12[%d]: got %s want %s", i, got, want)
				break
			}
		}
	}
	if failures > 0 {
		t.Fatalf("运限基准比对失败共 %d 处(仅显示前 30 处)", failures)
	}
}

func compareScope(t *testing.T, label, scopeName string, got *HoroscopeScope, want *goldenHoroscopeScope, fail func(string, string, ...any)) {
	t.Helper()
	if got.Name != want.Name {
		fail(label, "%s 层名: got %s want %s", scopeName, got.Name, want.Name)
	}
	if wantBranch := fix12(want.Index + 2); got.PalaceBranch != wantBranch {
		fail(label, "%s 落宫: got %s want %s", scopeName, Branches[got.PalaceBranch], Branches[wantBranch])
		return
	}
	if got.Stem != want.HeavenlyStem || got.Branch != want.EarthlyBranch {
		fail(label, "%s 干支: got %s%s want %s%s", scopeName, got.Stem, got.Branch, want.HeavenlyStem, want.EarthlyBranch)
	}
	for i, wantName := range want.PalaceNames {
		if got.PalaceNames[fix12(i+2)] != wantName {
			fail(label, "%s 宫名[%d]: got %s want %s", scopeName, i, got.PalaceNames[fix12(i+2)], wantName)
			break
		}
	}
	for i, wantM := range want.Mutagen {
		if got.Mutagen[i] != wantM {
			fail(label, "%s 四化[%d]: got %s want %s", scopeName, i, got.Mutagen[i], wantM)
			break
		}
	}
	if want.Stars != nil {
		for i, wantCell := range want.Stars {
			branch := fix12(i + 2)
			gotCell := got.Stars[branch]
			if len(gotCell) != len(wantCell) {
				gotNames := make([]string, len(gotCell))
				for k := range gotCell {
					gotNames[k] = gotCell[k].Name
				}
				wantNames := make([]string, len(wantCell))
				for k := range wantCell {
					wantNames[k] = wantCell[k].Name
				}
				fail(label, "%s 流曜[%s]: got %v want %v", scopeName, Branches[branch], gotNames, wantNames)
				continue
			}
			for k := range wantCell {
				if gotCell[k].Name != wantCell[k].Name {
					fail(label, "%s 流曜[%s][%d]: got %s want %s", scopeName, Branches[branch], k, gotCell[k].Name, wantCell[k].Name)
				} else if wantType := classifyStar(wantCell[k].Name, wantCell[k].Type); gotCell[k].Type != wantType {
					fail(label, "%s 流曜[%s]类型: got %s want %s", scopeName, gotCell[k].Name, gotCell[k].Type, wantType)
				}
			}
		}
	}
}
