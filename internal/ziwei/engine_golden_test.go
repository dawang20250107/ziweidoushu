package ziwei

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

// 黄金基准:tools/goldgen/gen.js 用 iztro 2.5.8 生成的 1500+ 用例,
// 覆盖 1900-2100 闰月、年界、边界年份、全时辰。逐字段比对。

type goldenStar struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Brightness string `json:"brightness"`
	Mutagen    string `json:"mutagen"`
}

type goldenPalace struct {
	Index          int          `json:"index"`
	Name           string       `json:"name"`
	IsBodyPalace   bool         `json:"isBodyPalace"`
	HeavenlyStem   string       `json:"heavenlyStem"`
	EarthlyBranch  string       `json:"earthlyBranch"`
	MajorStars     []goldenStar `json:"majorStars"`
	MinorStars     []goldenStar `json:"minorStars"`
	AdjectiveStars []goldenStar `json:"adjectiveStars"`
	Changsheng12   string       `json:"changsheng12"`
	Boshi12        string       `json:"boshi12"`
	Decadal        struct {
		Range []int `json:"range"`
	} `json:"decadal"`
	Ages []int `json:"ages"`
}

type goldenCase struct {
	Input struct {
		Year   int    `json:"year"`
		Month  int    `json:"month"`
		Day    int    `json:"day"`
		Hour   int    `json:"hour"`
		Gender string `json:"gender"`
	} `json:"input"`
	LunarDate                 string `json:"lunarDate"`
	ChineseDate               string `json:"chineseDate"`
	Time                      string `json:"time"`
	Sign                      string `json:"sign"`
	Zodiac                    string `json:"zodiac"`
	EarthlyBranchOfSoulPalace string `json:"earthlyBranchOfSoulPalace"`
	EarthlyBranchOfBodyPalace string `json:"earthlyBranchOfBodyPalace"`
	Soul                      string `json:"soul"`
	Body                      string `json:"body"`
	FiveElementsClass         string `json:"fiveElementsClass"`
	LunarJS                   struct {
		LunarYear  int    `json:"lunarYear"`
		LunarMonth int    `json:"lunarMonth"`
		LunarDay   int    `json:"lunarDay"`
		YearGan    string `json:"yearGan"`
		YearZhi    string `json:"yearZhi"`
	} `json:"lunarJS"`
	Palaces []goldenPalace `json:"palaces"`
}

func loadGolden(t *testing.T) []goldenCase {
	t.Helper()
	f, err := os.Open("testdata/iztro_golden.json.gz")
	if err != nil {
		t.Fatalf("打开基准数据失败: %v", err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("解压基准数据失败: %v", err)
	}
	var cases []goldenCase
	if err := json.NewDecoder(zr).Decode(&cases); err != nil {
		t.Fatalf("解析基准数据失败: %v", err)
	}
	if len(cases) < 1000 {
		t.Fatalf("基准用例数异常: %d", len(cases))
	}
	return cases
}

func caseLabel(g goldenCase) string {
	return fmt.Sprintf("%d-%d-%d/h%d/%s", g.Input.Year, g.Input.Month, g.Input.Day, g.Input.Hour, g.Input.Gender)
}

func TestGoldenParity(t *testing.T) {
	cases := loadGolden(t)
	failures := 0
	fail := func(g goldenCase, format string, args ...any) {
		failures++
		if failures <= 40 {
			t.Errorf("[%s] %s", caseLabel(g), fmt.Sprintf(format, args...))
		}
	}

	for _, g := range cases {
		chart, err := Generate(BirthInfo{
			Year: g.Input.Year, Month: g.Input.Month, Day: g.Input.Day,
			Hour: g.Input.Hour, Gender: Gender(g.Input.Gender),
		}, Options{ReferenceYear: 2026})
		if err != nil {
			fail(g, "Generate 报错: %v", err)
			continue
		}

		// ── 顶层字段 ──
		if got := Branches[chart.MingGongBranch]; got != g.EarthlyBranchOfSoulPalace {
			fail(g, "命宫地支: got %s want %s", got, g.EarthlyBranchOfSoulPalace)
		}
		if got := Branches[chart.ShenGongBranch]; got != g.EarthlyBranchOfBodyPalace {
			fail(g, "身宫地支: got %s want %s", got, g.EarthlyBranchOfBodyPalace)
		}
		if chart.WuxingJuName != g.FiveElementsClass {
			fail(g, "五行局: got %s want %s", chart.WuxingJuName, g.FiveElementsClass)
		}
		if chart.MingZhu != g.Soul {
			fail(g, "命主: got %s want %s", chart.MingZhu, g.Soul)
		}
		if chart.ShenZhu != g.Body {
			fail(g, "身主: got %s want %s", chart.ShenZhu, g.Body)
		}
		if chart.Zodiac != g.Zodiac {
			fail(g, "生肖: got %s want %s", chart.Zodiac, g.Zodiac)
		}
		if chart.Sign != g.Sign {
			fail(g, "星座: got %s want %s", chart.Sign, g.Sign)
		}
		if chart.TimeName != g.Time {
			fail(g, "时辰名: got %s want %s", chart.TimeName, g.Time)
		}
		wantPillars := strings.Split(g.ChineseDate, " ")
		gotPillars := []string{chart.FourPillars.Year, chart.FourPillars.Month, chart.FourPillars.Day, chart.FourPillars.Hour}
		if len(wantPillars) == 4 {
			for i := range wantPillars {
				if gotPillars[i] != wantPillars[i] {
					fail(g, "四柱[%d]: got %s want %s (完整: got %v want %v)", i, gotPillars[i], wantPillars[i], gotPillars, wantPillars)
				}
			}
		}
		if chart.LunarDateText != g.LunarDate {
			fail(g, "农历文本: got %s want %s", chart.LunarDateText, g.LunarDate)
		}
		// 农历数值(与 lunar-javascript 交叉验证)
		if chart.LunarInfo.LunarYear != g.LunarJS.LunarYear ||
			chart.LunarInfo.LunarMonth != abs(g.LunarJS.LunarMonth) ||
			chart.LunarInfo.LunarDay != g.LunarJS.LunarDay ||
			chart.LunarInfo.IsLeapMonth != (g.LunarJS.LunarMonth < 0) {
			fail(g, "农历数值: got %+v want %+v", chart.LunarInfo, g.LunarJS)
		}
		if Stems[chart.LunarInfo.YearStem] != g.LunarJS.YearGan || Branches[chart.LunarInfo.YearBranch] != g.LunarJS.YearZhi {
			fail(g, "年柱干支: got %s%s want %s%s", Stems[chart.LunarInfo.YearStem], Branches[chart.LunarInfo.YearBranch], g.LunarJS.YearGan, g.LunarJS.YearZhi)
		}

		// ── 逐宫比对 ──
		for _, gp := range g.Palaces {
			branch := fix12(gp.Index + 2)
			p := chart.PalaceByBranch(branch)
			if p == nil {
				fail(g, "宫位缺失 branch=%d", branch)
				continue
			}
			if p.Name != gp.Name {
				fail(g, "宫[%s]名: got %s want %s", Branches[branch], p.Name, gp.Name)
			}
			if Stems[p.Stem] != gp.HeavenlyStem {
				fail(g, "宫[%s]干: got %s want %s", Branches[branch], Stems[p.Stem], gp.HeavenlyStem)
			}
			if Branches[p.Branch] != gp.EarthlyBranch {
				fail(g, "宫[%s]支: got %s want %s", Branches[branch], Branches[p.Branch], gp.EarthlyBranch)
			}
			if p.IsShenGong != gp.IsBodyPalace {
				fail(g, "宫[%s]身宫标记: got %v want %v", Branches[branch], p.IsShenGong, gp.IsBodyPalace)
			}
			if p.Changsheng12 != gp.Changsheng12 {
				fail(g, "宫[%s]长生12: got %s want %s", Branches[branch], p.Changsheng12, gp.Changsheng12)
			}
			if p.Boshi12 != gp.Boshi12 {
				fail(g, "宫[%s]博士12: got %s want %s", Branches[branch], p.Boshi12, gp.Boshi12)
			}
			if len(gp.Decadal.Range) == 2 && (p.DaXianStart != gp.Decadal.Range[0] || p.DaXianEnd != gp.Decadal.Range[1]) {
				fail(g, "宫[%s]大限: got [%d,%d] want %v", Branches[branch], p.DaXianStart, p.DaXianEnd, gp.Decadal.Range)
			}
			if len(p.Ages) != len(gp.Ages) {
				fail(g, "宫[%s]小限长度: got %d want %d", Branches[branch], len(p.Ages), len(gp.Ages))
			} else {
				for i := range p.Ages {
					if p.Ages[i] != gp.Ages[i] {
						fail(g, "宫[%s]小限[%d]: got %d want %d", Branches[branch], i, p.Ages[i], gp.Ages[i])
						break
					}
				}
			}

			// 星曜:引擎按 主星→辅星→杂曜 顺序追加,与 iztro 三组拼接一致。
			var wantStars []goldenStar
			wantStars = append(wantStars, gp.MajorStars...)
			wantStars = append(wantStars, gp.MinorStars...)
			wantStars = append(wantStars, gp.AdjectiveStars...)
			if len(p.Stars) != len(wantStars) {
				gotNames := make([]string, len(p.Stars))
				for i, s := range p.Stars {
					gotNames[i] = s.Name
				}
				wantNames := make([]string, len(wantStars))
				for i, s := range wantStars {
					wantNames[i] = s.Name
				}
				fail(g, "宫[%s]星曜数量: got %v want %v", Branches[branch], gotNames, wantNames)
				continue
			}
			nMajor := len(gp.MajorStars)
			nMinor := len(gp.MinorStars)
			for i, want := range wantStars {
				got := p.Stars[i]
				if got.Name != want.Name {
					fail(g, "宫[%s]星[%d]名: got %s want %s", Branches[branch], i, got.Name, want.Name)
					continue
				}
				if string(got.SiHua) != want.Mutagen {
					fail(g, "宫[%s]星[%s]四化: got %q want %q", Branches[branch], got.Name, got.SiHua, want.Mutagen)
				}
				switch {
				case i < nMajor: // 主星:比对亮度与类型
					if got.Type != StarMajor {
						fail(g, "宫[%s]星[%s]类型: got %s want major", Branches[branch], got.Name, got.Type)
					}
					if got.Brightness != want.Brightness {
						fail(g, "宫[%s]星[%s]亮度: got %q want %q", Branches[branch], got.Name, got.Brightness, want.Brightness)
					}
				case i < nMajor+nMinor: // 辅星:比对亮度与业务类型映射
					if got.Brightness != want.Brightness {
						fail(g, "宫[%s]星[%s]亮度: got %q want %q", Branches[branch], got.Name, got.Brightness, want.Brightness)
					}
					wantType := classifyStar(want.Name, want.Type)
					if got.Type != wantType {
						fail(g, "宫[%s]星[%s]类型: got %s want %s(iztro=%s)", Branches[branch], got.Name, got.Type, wantType, want.Type)
					}
				default: // 杂曜:类型恒为 minor
					if got.Type != StarMinor {
						fail(g, "宫[%s]星[%s]类型: got %s want minor", Branches[branch], got.Name, got.Type)
					}
				}
			}
		}
	}
	if failures > 0 {
		t.Fatalf("黄金基准比对失败共 %d 处(仅显示前 40 处)", failures)
	}
}
