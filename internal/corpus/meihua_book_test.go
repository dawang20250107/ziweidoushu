package corpus

import (
	"strings"
	"testing"
)

// 《梅花易数》全文检索:体用总诀/观梅占/卦名均可命中,供 AI 解卦引文。
func TestMeihuaBookSearch(t *testing.T) {
	s := newTestStore(t)
	for _, q := range []string{"体用总诀", "观梅", "泽火革", "万物赋"} {
		hits := s.SearchAll(q, 5)
		found := false
		for _, h := range hits {
			if h.BookSlug == "meihuayishu" {
				found = true
				if h.BookTitle != "梅花易数" || h.Text == "" {
					t.Errorf("%s 命中字段异常: %+v", q, h)
				}
			}
		}
		if !found {
			t.Errorf("检索 %q 未命中《梅花易数》", q)
		}
	}
	// 观梅占原文细节(书中起卦数演算)确在库中
	hits := s.SearchAll("二雀争枝", 3)
	if len(hits) == 0 || !strings.Contains(hits[0].Text, "辰年") {
		t.Errorf("观梅占原文未入库或不可检索: %+v", hits)
	}
}
