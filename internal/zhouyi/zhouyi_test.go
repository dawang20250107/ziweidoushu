package zhouyi

import (
	"strings"
	"testing"
)

// 先天八卦爻画(自下而上),与 meihua 同序:乾1…坤8。
var trigLines = [9][3]bool{
	{}, {true, true, true}, {true, true, false}, {true, false, true}, {true, false, false},
	{false, true, true}, {false, true, false}, {false, false, true}, {false, false, false},
}

// TestCoverage 64 组合齐备,每卦 6 爻辞非空、爻题与卦画逐位一致。
func TestCoverage(t *testing.T) {
	yang := []string{"初九", "九二", "九三", "九四", "九五", "上九"}
	yin := []string{"初六", "六二", "六三", "六四", "六五", "上六"}
	n := 0
	for up := 1; up <= 8; up++ {
		for lo := 1; lo <= 8; lo++ {
			g := ByTrigrams(up, lo)
			if g == nil {
				t.Fatalf("缺卦: 上%d下%d", up, lo)
			}
			n++
			if g.GuaCi == "" || len(g.YaoCi) != 6 {
				t.Fatalf("%s: 卦辞/爻辞不全", g.Name)
			}
			lines := append(append([]bool{}, trigLines[lo][:]...), trigLines[up][:]...)
			for i, yc := range g.YaoCi {
				want := yin[i]
				if lines[i] {
					want = yang[i]
				}
				if !strings.HasPrefix(yc, want) {
					t.Fatalf("%s 第%d爻: %q 应以 %s 起", g.Name, i+1, yc, want)
				}
			}
		}
	}
	if n != 64 {
		t.Fatalf("卦数 %d", n)
	}
}

// TestSpotChecks 通行本抽查:乾坤用九用六、卦名连文卦辞、多段卦辞、多句爻辞。
func TestSpotChecks(t *testing.T) {
	qian := ByTrigrams(1, 1)
	if qian.Name != "乾" || qian.GuaCi != "元亨。利贞。" || !strings.HasPrefix(qian.Yong, "用九") {
		t.Fatalf("乾: %+v", qian)
	}
	kun := ByTrigrams(8, 8)
	if !strings.HasPrefix(kun.Yong, "用六") || !strings.Contains(kun.GuaCi, "西南得朋") {
		t.Fatalf("坤: %+v", kun)
	}
	// 卦名连文体例
	if pi := ByTrigrams(1, 8); !strings.HasPrefix(pi.GuaCi, "否之匪人") {
		t.Fatalf("否卦辞: %s", pi.GuaCi)
	}
	if gen := ByTrigrams(7, 7); !strings.HasPrefix(gen.GuaCi, "艮其背") {
		t.Fatalf("艮卦辞: %s", gen.GuaCi)
	}
	// 多句爻辞不截断(明夷初九三句)
	mingyi := ByTrigrams(8, 3)
	if !strings.Contains(mingyi.YaoCi[0], "三日不食") {
		t.Fatalf("明夷初九: %s", mingyi.YaoCi[0])
	}
	// 既济/未济方位互证
	if ByTrigrams(6, 3).Name != "既济" || ByTrigrams(3, 6).Name != "未济" {
		t.Fatal("既济/未济组成错位")
	}
}
