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

func containsPos(ps []int, p int) bool {
	for _, x := range ps {
		if x == p {
			return true
		}
	}
	return false
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

// yingQi 应期推断,落到具体地支(增删卜易应期章义:静应值冲、动应值合、
// 空应冲空填实、破应合破填实、墓应冲开墓库、绝应逢生)。
func yingQi(y Yao) string {
	b := branchIndexOf(y.Branch)
	if b < 0 {
		return "应期待详"
	}
	el := branchElement[b]
	zhi := string(branches[b])
	chong := string(branches[(b+6)%12])
	he := string(branches[liuHe[b]])
	switch {
	case y.XunKong:
		return fmt.Sprintf("用神旬空,应在冲空(%s日)或出旬填实(%s日)之期", chong, zhi)
	case y.YuePo:
		return fmt.Sprintf("用神月破,应在合破(%s日)、出月或填实(%s日)之期", he, zhi)
	case y.DayRelation == "合":
		return fmt.Sprintf("用神逢合绊,应在冲开之日(%s日)", chong)
	case y.DayStage == "墓" || y.BianRelation == "化墓":
		return fmt.Sprintf("用神入墓,应在冲开墓库之期(%s日)", string(branches[(muBranch[el]+6)%12]))
	case y.DayStage == "绝" || y.BianRelation == "化绝":
		return fmt.Sprintf("用神逢绝,应在绝处逢生之期(%s日)", string(branches[csBranch[el]]))
	case y.MonthState == "旺" || y.MonthState == "相":
		if y.Moving {
			return fmt.Sprintf("用神旺而发动,应期近——动应值日(%s)或合日(%s)", zhi, he)
		}
		return fmt.Sprintf("用神旺相安静,应在值日(%s)或冲动之日(%s)", zhi, chong)
	default:
		return fmt.Sprintf("用神休囚,应期较缓,待生扶之月令或值日(%s)", zhi)
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
		j.YongShen = fmt.Sprintf("用神【%s】不上卦", name)
		if fs := r.FuShen; fs != nil {
			// 伏神之法:本宫首卦纳甲取伏,论出伏与应期(卜筮正宗·飞伏章同源)
			j.Points = append(j.Points, fmt.Sprintf(
				"用神【%s】不上卦,取伏神:%s%s伏于%s爻%s之下——%s。",
				name, fs.Branch, fs.Element, numCN(fs.Pos), fs.Fei, fs.Note))
			if fs.CanOut {
				j.Level = "neutral"
				j.Conclusion = "综断:用神虽不上卦,伏而有气能出——事非无望,待出伏之期自见端倪。"
				j.YingQi = "用神伏藏,应在出伏之期——" + fs.ChuFuRi
			} else {
				j.Level = "caution"
				j.Conclusion = "综断:用神不上卦,伏神受制难出——所问之事一时无绪,宜待时或另择日再占。"
				j.YingQi = "伏神难出,纵论亦在" + fs.ChuFuRi + ",到期无验则非其时"
			}
			j.Points = append(j.Points, "应期:"+j.YingQi)
		} else {
			j.Level = "caution"
			j.Conclusion = "用神不上卦:所问之事一时无绪、或非其时,宜另择时再占。"
			j.Points = []string{j.Conclusion}
		}
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

	// 六神事象(只作事象修饰,不入吉凶分)+ 爻位象义与内外远近
	if note := liuShenNote[firstYong.LiuShen]; note != "" {
		j.Points = append(j.Points, fmt.Sprintf("用神临%s——%s。", firstYong.LiuShen, note))
	}
	if p := firstYong.Pos; p >= 1 && p <= 6 {
		nearFar := "居内卦,事近己身、应期较速"
		if p >= 4 {
			nearFar = "居外卦,事在远方外界、应期较缓"
		}
		j.Points = append(j.Points, fmt.Sprintf("爻位:用神在%s,%s。", yaoPosNote[p-1], nearFar))
	}

	// 动爻对用神:合绊(应待冲开)/冲动(旺衰分冲起冲散)
	yongB := branchIndexOf(firstYong.Branch)
	yongStrongState := firstYong.MonthState == "旺" || firstYong.MonthState == "相"
	for _, z := range r.Yaos {
		if !(z.Moving || z.AnDong) || containsPos(r.YongShenPos, z.Pos) {
			continue
		}
		zb := branchIndexOf(z.Branch)
		if zb < 0 || yongB < 0 {
			continue
		}
		if liuHe[zb] == yongB {
			j.Points = append(j.Points, fmt.Sprintf(
				"%s爻%s发动与用神成合——用神被合绊,事有牵制,应在冲开之日(%s日)。",
				numCN(z.Pos), z.Branch, string(branches[(zb+6)%12])))
		} else if (zb+6)%12 == yongB {
			if yongStrongState {
				j.Points = append(j.Points, fmt.Sprintf(
					"%s爻%s发动冲用神——用神旺相,冲则愈起,事得激发。", numCN(z.Pos), z.Branch))
			} else {
				score--
				j.Points = append(j.Points, fmt.Sprintf(
					"%s爻%s发动冲用神——用神休囚,恐被冲散,防事中道而废。", numCN(z.Pos), z.Branch))
			}
		}
	}

	// 三合局:动爻会局(缺一字日月补),局五行对用神论生克
	if juEl, juName, ok := r.sanHeJu(); ok {
		if ye := elIndex(firstYong.Element); ye >= 0 {
			switch wuxingRel(juEl, ye) {
			case "生":
				score += 2
				j.Points = append(j.Points, fmt.Sprintf("动爻会成%s生用神——众力相扶,其吉倍增。", juName))
			case "克":
				score -= 2
				j.Points = append(j.Points, fmt.Sprintf("动爻会成%s克用神——合党为忌,其力甚大,须防。", juName))
			case "比和":
				score++
				j.Points = append(j.Points, fmt.Sprintf("动爻会成%s与用神比和——同气连枝,用神得势。", juName))
			case "被生":
				j.Points = append(j.Points, fmt.Sprintf("动爻会成%s,用神生局泄气——耗力于外,略嫌分神。", juName))
			}
		}
	}

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
		// 世空/应空专断(增删卜易:月日冲合皆为填实,应期在填实之日)
		if shi.XunKong {
			j.Points = append(j.Points, fmt.Sprintf(
				"世爻旬空——己心未定、进退未决,应在填实之日(值%s日)。", shi.Branch))
		}
		if ying.XunKong {
			score--
			j.Points = append(j.Points, fmt.Sprintf(
				"应爻旬空——对方心虚不实,谋望难凭,待填实(值%s日)再看。", ying.Branch))
		}
	}

	// 卦性:六冲主散、六合主成;冲合互变则先后有别(增删卜易·冲合章)
	switch {
	case r.BenXingZhi == "六冲" && r.BianXingZhi == "六合":
		score++
		j.Points = append(j.Points, "六冲变六合——先散后聚、先难后成,久必有终。")
	case r.BenXingZhi == "六合" && r.BianXingZhi == "六冲":
		score--
		j.Points = append(j.Points, "六合变六冲——先合后散、始易终难,防有始无终。")
	case r.BenXingZhi == "六冲":
		score--
		j.Points = append(j.Points, "本卦六冲——事难聚而易散,谋事难成;惟近病逢冲即愈、忧散讼解反吉。")
	case r.BenXingZhi == "六合":
		score++
		j.Points = append(j.Points, "本卦六合——事有成就、合而能久;惟近病逢合缠绵,占散忧者嫌其绊。")
	case r.BianXingZhi == "六冲":
		j.Points = append(j.Points, "变卦六冲——事纵有成,成后防散。")
	}

	// 游魂/归魂卦义(占行人出行久事尤验)
	switch r.PalaceSeq {
	case "游魂卦":
		j.Points = append(j.Points, "游魂卦——心神不定、事多漂泊反复,占行人主在外未归。")
	case "归魂卦":
		j.Points = append(j.Points, "归魂卦——事归本位、有回归安顿之象,占行人主归,出行不宜远动。")
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
