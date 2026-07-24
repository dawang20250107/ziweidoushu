package liuyao

// 六爻确定性断语层:据用神旺衰 / 元忌力量 / 动变 / 世应机械推演吉凶,不走 LLM。
//
// 宗《增删卜易》《卜筮正宗》断卦次第:先定用神,次看用神旺衰空破,再论元神生扶、
// 忌神克制,合动变(回头生克/进退/墓绝)与世应生克定成败,末推应期(冲空填实/值日)。
// 只据本包已算出的可验证字段推演,是与 AI 解卦互补的确定性骨架。

import (
	"fmt"
	"strings"
)

// Judgment 六爻断语。
type Judgment struct {
	Conclusion string   `json:"conclusion"` // 综断一句
	Level      string   `json:"level"`      // good/caution/neutral
	YongShen   string   `json:"yongShen"`   // 用神状态摘要
	YingQi     string   `json:"yingQi"`     // 应期提示
	Points     []string `json:"points"`     // 逐条断语
}

var yaoNumCN = [6]string{"初", "二", "三", "四", "五", "上"}

func numCN(pos int) string {
	if pos >= 1 && pos <= 6 {
		return yaoNumCN[pos-1]
	}
	return fmt.Sprintf("%d", pos)
}

func elIndex(name string) int {
	for i, n := range elementNames {
		if n == name {
			return i
		}
	}
	return -1
}

// wuxingRel a 相对 b 的关系(五行序 木0火1土2金3水4,+1 为生)。
func wuxingRel(a, b int) string {
	switch {
	case a == b:
		return "比和"
	case (a+1)%5 == b:
		return "生" // a 生 b
	case (a+2)%5 == b:
		return "克" // a 克 b
	case (b+1)%5 == a:
		return "被生" // b 生 a
	default:
		return "被克" // b 克 a
	}
}

type yaoAssess struct {
	text   string
	delta  int
	strong bool
}

// assessYong 评一爻旺衰空破(月令 + 日辰 + 破空)。
func assessYong(y Yao) yaoAssess {
	var parts []string
	d := 0
	switch y.MonthState {
	case "旺":
		parts, d = append(parts, "月建旺"), d+2
	case "相":
		parts, d = append(parts, "得月令相"), d+1
	case "囚":
		parts, d = append(parts, "月令囚"), d-1
	case "死":
		parts, d = append(parts, "月令死"), d-2
	}
	switch y.DayRelation {
	case "临":
		parts, d = append(parts, "临日辰"), d+2
	case "扶":
		parts, d = append(parts, "日辰扶"), d+1
	case "生":
		parts, d = append(parts, "日辰生"), d+1
	case "合":
		parts = append(parts, "日辰合起")
	case "冲":
		parts, d = append(parts, "日辰冲"), d-1
	case "克":
		parts, d = append(parts, "日辰克"), d-2
	case "泄":
		parts, d = append(parts, "日辰泄气"), d-1
	case "耗":
		parts, d = append(parts, "日辰耗"), d-1
	}
	if y.YuePo {
		parts, d = append(parts, "月破"), d-2
	}
	if y.RiPo {
		parts, d = append(parts, "日破"), d-2
	}
	if y.XunKong {
		parts, d = append(parts, "旬空"), d-1
	}
	if y.AnDong {
		parts = append(parts, "暗动")
	}
	return yaoAssess{text: strings.Join(parts, "、"), delta: d, strong: d >= 2}
}

// bianEffect 动爻之变对用神力量的影响。
func bianEffect(rel string) (int, string) {
	switch rel {
	case "化进神":
		return 1, "化进神,渐入佳境、事有进展"
	case "化退神":
		return -1, "化退神,渐退、事难维持"
	case "回头生":
		return 2, "得变爻回头相生,力增、吉"
	case "回头克":
		return -2, "遭变爻回头克,力损、凶,须防"
	case "化长生":
		return 1, "化长生,生机渐旺"
	case "化墓":
		return -2, "化墓,事入墓、迟滞难展"
	case "化绝":
		return -2, "化绝,力尽、事恐难成"
	case "化空":
		return -1, "化空,一时落空、待出空"
	case "化合":
		return 0, "化合,事有牵绊、待冲开"
	case "伏吟":
		return -1, "伏吟,呻吟迟滞、进退两难"
	case "反吟":
		return -1, "反吟,反复不定、事多变卦"
	}
	return 0, ""
}

// aggVerdict 聚合力量评估:任一有力则有力,否则见无力即无力。
func aggVerdict(notes []PowerNote) string {
	res := ""
	for _, n := range notes {
		if n.Verdict == "有力" {
			return "有力"
		}
		if n.Verdict == "无力" {
			res = "无力"
		}
	}
	return res
}

func (r *Result) shiYing() (shi, ying *Yao) {
	for i := range r.Yaos {
		if r.Yaos[i].IsShi {
			shi = &r.Yaos[i]
		}
		if r.Yaos[i].IsYing {
			ying = &r.Yaos[i]
		}
	}
	return
}

func yingQi(y Yao) string {
	switch {
	case y.XunKong:
		return "用神旬空,应在冲空或出空(值日填实)之期"
	case y.YuePo:
		return "用神月破,应在合破、出月或填实之期"
	case y.DayRelation == "合":
		return "用神逢合绊,应在冲开之日"
	case y.MonthState == "旺" || y.MonthState == "相":
		if y.Moving {
			return "用神旺而发动,应期近、值日或合日可期"
		}
		return "用神旺相,应在值日、值旬或合日"
	default:
		return "用神休囚,应期较缓、待生扶之月令或用神值日"
	}
}

func judgeLevel(score int) string {
	switch {
	case score >= 3:
		return "good"
	case score <= -2:
		return "caution"
	default:
		return "neutral"
	}
}

// Judge 由已装之卦机械推演六爻断语(确定性)。
func (r *Result) Judge() *Judgment {
	j := &Judgment{}
	name := r.YongShen
	if name == "" {
		name = "用神"
	}
	if len(r.YongShenPos) == 0 {
		j.Level = "caution"
		j.YongShen = fmt.Sprintf("用神【%s】不上卦", name)
		j.Conclusion = "用神不上卦:所问之事一时无绪、或非其时,宜细察伏神,或另择时再占。"
		j.Points = []string{j.Conclusion}
		return j
	}

	score := 0
	var states []string
	yongStrong := false
	var firstYong Yao
	var bianPoints []string
	for i, pos := range r.YongShenPos {
		if pos < 1 || pos > 6 {
			continue
		}
		y := r.Yaos[pos-1]
		if i == 0 {
			firstYong = y
		}
		a := assessYong(y)
		score += a.delta
		if a.strong {
			yongStrong = true
		}
		s := fmt.Sprintf("%s爻%s(%s%s)", numCN(pos), y.LiuQin, y.Branch, y.Element)
		if a.text != "" {
			s += "——" + a.text
		}
		states = append(states, s)
		if y.Moving && y.BianRelation != "" {
			d, t := bianEffect(y.BianRelation)
			score += d
			if t != "" {
				bianPoints = append(bianPoints, "用神发动:"+t)
			}
		}
	}
	j.YongShen = fmt.Sprintf("用神【%s】:%s", name, strings.Join(states, ";"))
	j.Points = append(j.Points, j.YongShen)
	j.Points = append(j.Points, bianPoints...)

	switch aggVerdict(r.YuanShenPower) {
	case "有力":
		score += 2
		j.Points = append(j.Points, fmt.Sprintf("元神【%s】有力,生扶用神有源。", r.YuanShen))
	case "无力":
		j.Points = append(j.Points, fmt.Sprintf("元神【%s】无力,生扶乏力、用神少援。", r.YuanShen))
	}
	switch aggVerdict(r.JiShenPower) {
	case "有力":
		score -= 2
		j.Points = append(j.Points, fmt.Sprintf("忌神【%s】有力,克伤用神,主阻、宜防。", r.JiShen))
	case "无力":
		score++
		j.Points = append(j.Points, fmt.Sprintf("忌神【%s】无力,难伤用神,无大碍。", r.JiShen))
	}

	if shi, ying := r.shiYing(); shi != nil && ying != nil {
		si, yi := elIndex(shi.Element), elIndex(ying.Element)
		if si >= 0 && yi >= 0 {
			switch wuxingRel(yi, si) { // 应 相对 世
			case "生":
				score++
				j.Points = append(j.Points, "应生世,他来就我、对方有意,利求谋、求人、求财。")
			case "克":
				score--
				j.Points = append(j.Points, "应克世,对方压制、阻力较大,谋事费力宜缓。")
			case "被生":
				j.Points = append(j.Points, "世生应,我施于人、主动付出,较耗神。")
			case "被克":
				score++
				j.Points = append(j.Points, "世克应,谋事在我、能操主动,多可制彼成事。")
			case "比和":
				score++
				j.Points = append(j.Points, "世应比和,两情相协、事多顺遂。")
			}
		}
	}

	j.YingQi = yingQi(firstYong)
	j.Points = append(j.Points, "应期:"+j.YingQi)

	j.Level = judgeLevel(score)
	switch j.Level {
	case "good":
		j.Conclusion = "综断:用神得地、生扶有力,所问之事大体可成,吉。"
		if yongStrong {
			j.Conclusion = "综断:用神旺相得力、元神相生,所问之事成算颇高,吉。"
		}
	case "caution":
		j.Conclusion = "综断:用神受制或逢空破、忌神当权,事多阻滞,宜谨慎、防其不成。"
	default:
		j.Conclusion = "综断:吉凶参半、用神平平,成事多在人为,宜积极促成、把握应期。"
	}
	return j
}
