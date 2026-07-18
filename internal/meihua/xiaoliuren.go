// 小六壬(倪师《天纪》课堂教法):以农历月、日、时辰掐指轮数六位,
// 测当下一事之吉凶缓急。六位顺序与断语依倪师课堂记录。
package meihua

import (
	"fmt"
	"time"

	"github.com/6tail/lunar-go/calendar"
)

// LiuRenPos 小六壬六位。
type LiuRenPos struct {
	Name    string `json:"name"`
	Luck    string `json:"luck"`    // 吉/凶/平
	Meaning string `json:"meaning"` // 倪师课堂断语
}

// 六位顺序:大安→留连→速喜→赤口→小吉→空亡(倪师口述序)。
var liuRenPositions = []LiuRenPos{
	{"大安", "吉", "事体安稳,约一个礼拜左右看到想要的结果。"},
	{"留连", "平", "事情停止不前、拖延反复,此行不利。"},
	{"速喜", "吉", "立竿见影,马上看到所要的结果。"},
	{"赤口", "凶", "有第三者介入,是非口舌、人为干扰。"},
	{"小吉", "吉", "吉,约两个礼拜后看到所要的结果。"},
	{"空亡", "凶", "凶,谋事落空,不宜进行。"},
}

// XiaoLiuRenResult 小六壬占算结果。
type XiaoLiuRenResult struct {
	Question  string      `json:"question,omitempty"`
	LunarText string      `json:"lunarText"` // 起算农历「正月初一子时」
	Steps     [3]string   `json:"steps"`     // 月/日/时三步落位名
	Result    LiuRenPos   `json:"result"`
	Path      []LiuRenPos `json:"path"` // 三步完整落位(展示掐指过程)
}

// XiaoLiuRen 以当下(或指定)时刻起小六壬:
// 从大安起正月数至月,从月位起初一数至日,从日位起子时数至时辰。
func XiaoLiuRen(t time.Time, question string) (XiaoLiuRenResult, error) {
	if t.Year() < 1902 || t.Year() > 2098 {
		return XiaoLiuRenResult{}, fmt.Errorf("时间超出支持范围(1902-2098)")
	}
	solar := calendar.NewSolarFromDate(t)
	lunar := solar.GetLunar()

	month := lunar.GetMonth()
	if month < 0 {
		month = -month // 闰月按当月数
	}
	day := lunar.GetDay()
	hourN := branchNum(lunar.GetTimeZhi()) // 子=1

	p1 := (month - 1) % 6
	p2 := (p1 + day - 1) % 6
	p3 := (p2 + hourN - 1) % 6

	return XiaoLiuRenResult{
		Question: question,
		LunarText: fmt.Sprintf("%s月%s日%s时",
			lunar.GetMonthInChinese(), lunar.GetDayInChinese(), lunar.GetTimeZhi()),
		Steps:  [3]string{liuRenPositions[p1].Name, liuRenPositions[p2].Name, liuRenPositions[p3].Name},
		Result: liuRenPositions[p3],
		Path:   []LiuRenPos{liuRenPositions[p1], liuRenPositions[p2], liuRenPositions[p3]},
	}, nil
}
