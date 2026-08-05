package ai

import (
	"strings"
	"testing"

	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// 形神表:十四主星齐备;紫微坐命盘整链形神句落文。
func TestXingShen(t *testing.T) {
	for _, s := range []string{"紫微", "天机", "太阳", "武曲", "天同", "廉贞", "天府", "太阴", "贪狼", "巨门", "天相", "天梁", "七杀", "破军"} {
		if xingShenNote[s] == "" {
			t.Errorf("形神缺 %s", s)
		}
	}
	c, err := ziwei.Generate(ziwei.BirthInfo{Year: 1996, Month: 7, Day: 14, Hour: 0, Gender: ziwei.Male}, ziwei.Options{ReferenceYear: 2026})
	if err != nil {
		t.Fatal(err)
	}
	rd := buildReading(c, nil)
	if rd == nil || !strings.Contains(rd.Overview, "形神多见") || !strings.Contains(rd.Overview, "天庭饱满") {
		t.Errorf("命格总论应含紫微形神速写: %s", rd.Overview)
	}
}
