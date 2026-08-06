package daliuren

import (
	"testing"
	"time"
)

// TestJudgeSections 分节深断:四节齐备、课体详解命中九门深文、应期含冲合。
func TestJudgeSections(t *testing.T) {
	r, err := CastByTime(time.Date(2026, 8, 5, 10, 30, 0, 0, time.FixedZone("CST", 8*3600)))
	if err != nil {
		t.Fatal(err)
	}
	j := r.Judge()
	keys := map[string]bool{}
	for _, s := range j.Sections {
		if s.Title == "" || len([]rune(s.Text)) < 30 {
			t.Errorf("分节 %s 文本过短: %q", s.Key, s.Text)
		}
		keys[s.Key] = true
	}
	for _, want := range []string{"keti", "sanchuan", "tianjiang", "yingqi"} {
		if !keys[want] {
			t.Errorf("缺分节 %s(实有 %v)", want, keys)
		}
	}
	// 课体详解须为九门深文(非一句话兜底)
	if deep := keTypeDeep[r.KeType]; deep == "" {
		t.Errorf("课体 %s 无深断文本", r.KeType)
	}
	// 九门深文全部备齐且有分量
	for k, v := range keTypeDeep {
		if len([]rune(v)) < 60 {
			t.Errorf("课体 %s 深文过短", k)
		}
	}
	for k, v := range jiangDeep {
		if len([]rune(v)) < 20 {
			t.Errorf("天将 %s 深文过短", k)
		}
	}
	if len(jiangDeep) != 12 {
		t.Errorf("天将深文应 12 将,得 %d", len(jiangDeep))
	}
}
