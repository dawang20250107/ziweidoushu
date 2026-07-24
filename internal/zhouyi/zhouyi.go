// Package zhouyi 《周易》公版经文(卦辞/爻辞/用九用六):六爻与梅花断卦的经文层。
//
// 数据来源:Wikisource《周易》64 子页(维基文库整理之通行本经文,公有领域),
// 抓取解析管线见 research/zhouyi-jingwen.md;抓取时逐卦校验:爻题序与上下卦
// 爻画一致、页面自带「X下Y上」组成与金文卦序表互证、卦辞连文体例
// (否之匪人/履虎尾/同人于野/艮其背)与多段卦辞(坤)均按通行本保留。
package zhouyi

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed zhouyi.json
var raw []byte

// Gua 一卦经文。YaoCi 自下而上(index 0 = 初爻),文本带爻题前缀(「初九:潜龙勿用。」)。
type Gua struct {
	Name  string   `json:"name"`
	Upper int      `json:"upper"` // 上卦先天数(乾1兑2离3震4巽5坎6艮7坤8)
	Lower int      `json:"lower"`
	GuaCi string   `json:"guaCi"`
	YaoCi []string `json:"yaoCi"`
	Yong  string   `json:"yong,omitempty"` // 乾用九/坤用六
}

var byTrigram map[[2]int]*Gua

func init() {
	var list []*Gua
	if err := json.Unmarshal(raw, &list); err != nil {
		panic(fmt.Sprintf("zhouyi.json 解析失败: %v", err))
	}
	if len(list) != 64 {
		panic(fmt.Sprintf("zhouyi.json 应含 64 卦,得 %d", len(list)))
	}
	byTrigram = make(map[[2]int]*Gua, 64)
	for _, g := range list {
		byTrigram[[2]int{g.Upper, g.Lower}] = g
	}
}

// ByTrigrams 按上/下卦先天数取经文;未知组合返回 nil(参数越界)。
func ByTrigrams(upper, lower int) *Gua {
	return byTrigram[[2]int{upper, lower}]
}
