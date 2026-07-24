package daliuren

// 大六壬确定性断语层:据课体 + 三传(初/中/末)+ 三传对日干生克机械推演,不走 LLM。
//
// 大六壬断法:课体定事之格局(贼克主克战、比用主比助、涉害主曲折、遥克主远扰、
// 昴星主意外、伏吟主静滞、返吟主反复);三传述事之来由(初)、经过(中)、归宿(末),
// 以三传对日干(本身)之生克定吉凶——传生干为得助、克干为受制、干克传为得利、
// 干生传为耗力。天将/年命层俟贵人全表另加,此为课体+三传的确定性骨架。

import (
	"fmt"
	"strings"
)

// Judgment 大六壬断语。
type Judgment struct {
	Conclusion string   `json:"conclusion"` // 综断一句
	Level      string   `json:"level"`      // good/caution/neutral
	KeTypeText string   `json:"keTypeText"` // 课体含义
	SanChuan   []string `json:"sanChuan"`   // 三传逐传解
	Points     []string `json:"points"`     // 逐条断语
}

// stemElement 天干五行(木0火1土2金3水4)。
var stemElement = []int{0, 0, 1, 1, 2, 2, 3, 3, 4, 4}

// elementNames 五行名(木0火1土2金3水4),与 branchElement 同序。
var elementNames = []string{"木", "火", "土", "金", "水"}

var keTypeMeaning = map[string]string{
	"贼克": "贼克课——下贼上为始入、上克下为元首,事有克战、吉凶分明;来意直、应期速,当机立断、以刚制之则决",
	"比用": "比用课——多克取比,择与日干同类者用之;事赖同侪比助、须择亲熟之路,循类相求、和同则济",
	"涉害": "涉害课——涉历深浅以定发用,事必经磨、过程曲折;深者得力迟成、浅者易举早应,耐历炼方见归着",
	"遥克": "遥克课——四课无克、隔位遥相干,事自远方外界而来,其扰间接、力道轻缓;虚声多于实祸,察其来路即可安",
	"昴星": "昴星课——无克取昴,虎视之象、事出非常;阳仰阴俯,防意外惊扰、暗中窥伺,宜静守勿妄动",
	"别责": "别责课——三课无克、别取一神为用,正路不通而旁求门径;事有缺憾不全,须借他力斜出,曲折乃成",
	"八专": "八专课——干支同宫、上下一体,公私相混、二人同心;同心则事济,亦防内外不分、狎昵不明之弊",
	"伏吟": "伏吟课——天地盘重合,伏而不动、事主静止,忧疑滞塞;守旧待时,动在冲开之期",
	"返吟": "返吟课——天地盘对冲,返复不定、动荡不安;事多反复去而复来,宜防中途生变,一动不如一静",
}

func branchIdx(s string) int {
	r := []rune(s)
	if len(r) == 0 {
		return -1
	}
	for i, b := range branches {
		if b == r[0] {
			return i
		}
	}
	return -1
}

func stemIdx(s string) int {
	r := []rune(s)
	if len(r) == 0 {
		return -1
	}
	for i, b := range stems {
		if b == r[0] {
			return i
		}
	}
	return -1
}

// wxRel a 相对 b 的五行关系(木0火1土2金3水4,+1 为生)。
func wxRel(aEl, bEl int) string {
	switch {
	case aEl == bEl:
		return "比和"
	case (aEl+1)%5 == bEl:
		return "生" // a 生 b
	case (aEl+2)%5 == bEl:
		return "克" // a 克 b
	case (bEl+1)%5 == aEl:
		return "被生" // b 生 a
	default:
		return "被克" // b 克 a
	}
}

// chuanRelText 一传对日干关系的断语与吉凶增减。
func chuanRelText(rel string) (string, int) {
	switch rel {
	case "生":
		return "来生日干、有情得助", 2
	case "被克":
		return "为日干所克、可掌握(克者为财利)", 1
	case "比和":
		return "与日干比和、同类相扶", 1
	case "被生":
		return "日干生之、我耗力付出", -1
	default: // 克
		return "克身、事有阻力压制", -2
	}
}

// Judge 由起课机械推演大六壬断语(确定性)。
func (r *Result) Judge() *Judgment {
	j := &Judgment{}
	j.KeTypeText = keTypeMeaning[r.KeType]
	if j.KeTypeText == "" {
		j.KeTypeText = r.KeType + "课"
	}
	j.Points = append(j.Points, "课体:"+j.KeTypeText)

	gi := stemIdx(r.DayStem)
	ganEl := -1
	if gi >= 0 {
		ganEl = stemElement[gi]
	}

	score := 0
	frames := []string{"初传(发端·事之来由)", "中传(经过·事之转折)", "末传(结局·事之归宿)"}
	for i := 0; i < 3 && i < len(r.Chuan); i++ {
		ci := branchIdx(r.Chuan[i])
		if ci < 0 || ganEl < 0 {
			continue
		}
		rel := wxRel(branchElement[ci], ganEl)
		txt, d := chuanRelText(rel)
		// 末传主结局,权重加重;初传次之。
		w := 1
		if i == 2 {
			w = 2
		}
		score += d * w
		line := fmt.Sprintf("%s:%s(%s),%s", frames[i], r.Chuan[i], elementNames[branchElement[ci]], txt)
		// 乘将附断:凶将小减、贵人青龙小增(吉凶系于天将,大全义)
		if jg := r.ChuanJiang[i]; jg != "" {
			line += ";乘" + jg + "——" + jiangNote[jg]
			switch jg {
			case "贵人", "青龙", "六合", "太常", "太阴":
				score++
			case "白虎", "玄武", "螣蛇", "天空", "勾陈":
				score--
			}
		}
		j.SanChuan = append(j.SanChuan, line)
	}
	j.Points = append(j.Points, j.SanChuan...)

	// 课体动荡类微调。
	switch r.KeType {
	case "伏吟", "返吟", "昴星":
		score--
	}

	switch {
	case score >= 3:
		j.Level = "good"
	case score <= -2:
		j.Level = "caution"
	default:
		j.Level = "neutral"
	}
	switch j.Level {
	case "good":
		j.Conclusion = "综断:三传归宿有情、生扶日干,课体不悖,事体大致顺遂、可为。"
	case "caution":
		j.Conclusion = "综断:末传克身或课体动荡反复,事多阻滞,宜守宜慎、勿轻进。"
	default:
		j.Conclusion = "综断:吉凶相参、课体平常,事在人为,宜审时度势、择机而动。"
	}
	return j
}

// Text 断语纯文本(供无 LLM 时直接呈现)。
func (j *Judgment) Text() string {
	var b strings.Builder
	b.WriteString(j.Conclusion + "\n")
	for _, p := range j.Points {
		b.WriteString("· " + p + "\n")
	}
	return b.String()
}
