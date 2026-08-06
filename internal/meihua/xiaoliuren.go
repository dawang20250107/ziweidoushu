// 小六壬(倪师《天纪》课堂教法):以农历月、日、时辰掐指轮数六位,
// 测当下一事之吉凶缓急。六位顺序与断语依倪师课堂记录。
package meihua

import (
	"fmt"
	"strings"
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
	// Sections 分节深断(掐指路径/落宫详断/应期),免费层呈现纵深。
	Sections []JudgeSection `json:"sections,omitempty"`
}

// xlrDeep 六位深断:宫义 + 五行方位 + 事象细断(倪师课堂义引申,原创行文)。
var xlrDeep = map[string]string{
	"大安": "大安属木,青龙之位,主静而有守。事体安稳、身不动时吉;宜守成、宜正路,谋望在东方或属木之人事上尤顺。断曰:大安事事昌,求财在坤方——安中自有生机,约一周之期见分晓。",
	"留连": "留连属水,玄武之位,主暗昧不明、拖延反复。事被无形之手勾住,急切难成;此时不宜强推,宜查暗节、防阴私,再候时机。断曰:留连事难成,求谋日未明——纵成亦迟,防人从中作梗。",
	"速喜": "速喜属火,朱雀之位,主快、主信。喜讯立至、立竿见影,午前占得其力尤显;有音信文书之应,南方之人事可借力。断曰:速喜喜来临,求财向南行——当下即是应期,趁热打铁。",
	"赤口": "赤口属金,白虎之位,主口舌是非、第三者介入。防争执、防小人、防言语生衅;涉契约官非者尤须谨慎,西方之人事不利。断曰:赤口主口舌,是非要提防——话到嘴边留半句,退一步即是解法。",
	"小吉": "小吉属木,六合之位,主和合、贵人。事有人帮、和气而成,约两周之期渐次见果;宜借中人之力、走和谈之路。断曰:小吉最吉昌,路上好商量——小吉非大吉,得寸即安,勿贪全功。",
	"空亡": "空亡属土,勾陈之位,主落空、无果。谋事如探空囊,音信杳然;此时不宜再投入,宜收手观望、另起炉灶。断曰:空亡事不祥,阴人多乖张——空非终局,过此一节、换时换法再图。",
}

// buildXlrSections 组装小六壬分节深断。
func buildXlrSections(res *XiaoLiuRenResult) {
	// ── 掐指路径:月→日→时三步的完整叙述 ──
	var lj strings.Builder
	lj.WriteString(fmt.Sprintf("倪师课堂掐指法:以农历%s起算——从「大安」起正月顺数至月,落【%s】;自月位起初一数至日,落【%s】;自日位起子时数至时辰,终落【%s】。三步为经过,终位为断。",
		res.LunarText, res.Steps[0], res.Steps[1], res.Steps[2]))
	res.Sections = append(res.Sections, JudgeSection{Key: "qiazhi", Title: "掐指路径", Text: lj.String()})

	// ── 落宫详断 ──
	if deep := xlrDeep[res.Result.Name]; deep != "" {
		res.Sections = append(res.Sections, JudgeSection{Key: "luogong", Title: "落宫详断", Text: deep})
	}

	// ── 途中之象:前两步落位对事体过程的提示 ──
	var tz strings.Builder
	tz.WriteString(fmt.Sprintf("途中之象:月位落【%s】(%s),是事体的大背景;日位落【%s】(%s),是近期的推移。终位吉而途中凶者,先难后易;途中吉而终位凶者,虚好看、防反复。",
		res.Path[0].Name, res.Path[0].Luck, res.Path[1].Name, res.Path[1].Luck))
	res.Sections = append(res.Sections, JudgeSection{Key: "tuzhong", Title: "途中之象", Text: tz.String()})
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

	res := XiaoLiuRenResult{
		Question: question,
		LunarText: fmt.Sprintf("%s月%s日%s时",
			lunar.GetMonthInChinese(), lunar.GetDayInChinese(), lunar.GetTimeZhi()),
		Steps:  [3]string{liuRenPositions[p1].Name, liuRenPositions[p2].Name, liuRenPositions[p3].Name},
		Result: liuRenPositions[p3],
		Path:   []LiuRenPos{liuRenPositions[p1], liuRenPositions[p2], liuRenPositions[p3]},
	}
	buildXlrSections(&res)
	return res, nil
}
