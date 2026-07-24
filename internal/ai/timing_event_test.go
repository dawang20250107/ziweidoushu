package ai

import (
	"strings"
	"testing"

	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// TestEventTiming 事项择吉:目录齐备、宫映射正确、利年→利月→利日结构完整、未知事项返回空。
func TestEventTiming(t *testing.T) {
	cat := EventCatalog()
	if len(cat) != 7 {
		t.Errorf("事项目录应为 7 项,得 %d", len(cat))
	}
	// 宫映射抽验。
	wantPalace := map[string]string{"marriage": "夫妻", "wealth": "财帛", "career": "官禄", "relocate": "田宅", "travel": "迁移"}
	byKey := map[string]EventCatalogItem{}
	for _, c := range cat {
		byKey[c.Key] = c
	}
	for k, p := range wantPalace {
		if byKey[k].Palace != p {
			t.Errorf("%s 应对应 %s 宫,得 %s", k, p, byKey[k].Palace)
		}
	}

	c, err := ziwei.Generate(ziwei.BirthInfo{Year: 1990, Month: 6, Day: 15, Hour: 6, Gender: ziwei.Male}, ziwei.Options{ReferenceYear: 2026})
	if err != nil {
		t.Fatal(err)
	}
	et := buildEventTiming(c, "wealth")
	if et == nil {
		t.Fatal("求财择吉不应为空")
	}
	if et.Palace != "财帛" || et.Summary == "" || et.Advice == "" {
		t.Errorf("择吉结果字段缺失:%+v", et)
	}
	// 利年须落十年窗口且带干支理由;利日应引用流日行财帛。
	for _, y := range et.Years {
		if y.Year < 2026 || y.Year >= 2036 || y.GanZhi == "" || y.Note == "" {
			t.Errorf("利年 %+v 异常", y)
		}
	}
	for _, d := range et.BestDays {
		if !strings.Contains(d, "财帛") {
			t.Errorf("利日应引用流日行财帛:%s", d)
		}
	}
	// 未知事项返回 nil。
	if buildEventTiming(c, "nonsense") != nil {
		t.Error("未知事项应返回 nil")
	}
}
