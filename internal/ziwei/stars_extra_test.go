package ziwei

import "testing"

// 年支系补充杂曜安星回归:大耗(阳支顺七/阴支顺五)、龙德(年支+7)、
// 劫煞(三合绝地)。锚点为《全书》表 + 文墨天机实盘核对(子年:大耗未/龙德未/劫煞巳)。
func TestPlaceExtraStars(t *testing.T) {
	cases := []struct {
		yearBranch int
		daHao      string
		longDe     string
		jieSha     string
	}{
		{0, "未", "未", "巳"}, // 子年(文墨实盘核对)
		{6, "丑", "丑", "亥"}, // 午年(寅午戌劫煞亥)
		{9, "寅", "辰", "寅"}, // 酉年(巳酉丑劫煞寅)
		{3, "申", "戌", "申"}, // 卯年(亥卯未劫煞申)
		{4, "亥", "亥", "巳"}, // 辰年(申子辰劫煞巳)
	}
	for _, c := range cases {
		palaces := make([]Palace, 12)
		for i := range palaces {
			palaces[i].Branch = i
		}
		placeExtraStars(palaces, c.yearBranch)
		find := func(name string) string {
			for _, p := range palaces {
				for _, s := range p.ExtraStars {
					if s.Name == name {
						return Branches[p.Branch]
					}
				}
			}
			return "无"
		}
		if got := find("大耗"); got != c.daHao {
			t.Errorf("%s年大耗: got %s want %s", Branches[c.yearBranch], got, c.daHao)
		}
		if got := find("龙德"); got != c.longDe {
			t.Errorf("%s年龙德: got %s want %s", Branches[c.yearBranch], got, c.longDe)
		}
		if got := find("劫煞"); got != c.jieSha {
			t.Errorf("%s年劫煞: got %s want %s", Branches[c.yearBranch], got, c.jieSha)
		}
	}
}
