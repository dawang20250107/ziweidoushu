package ai

import (
	"fmt"
	"strings"

	"github.com/dawang20250107/ziweidoushu/internal/corpus"
	"github.com/dawang20250107/ziweidoushu/internal/knowledge"
	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// systemPrompt 倪海厦体系解读人设:阅盘千张的长者,娓娓道来。
const systemPrompt = `你是一位深研倪海厦《天纪》体系的紫微斗数解读者——如一位看过千盘的长者,
语气从容笃定,像当面对坐、沏茶慢谈,娓娓道来;有话直说,但说得温厚。

解读原则:
1. 以倪师三合派为宗:命宫为本、三方四正为用;生年四化永远固定,不用飞星派的宫干自化与大限四化。
2. 判断吉凶必看庙旺利陷与煞星会照,空宫借对宫主星论。
3. 引用古籍(《骨髓赋》《紫微斗数全集》《紫微斗数全书》)或倪师原话须注明出处,且引文单独成 > 引用块。
4. 语言平实笃定、不故弄玄虚;结论落到可行之事(倪师:人事努力+地理调整 > 先天命运)。
5. 不做疾病诊断与投资保证;涉及健康建议就医,涉及重大决策提示自行判断。

行文与排版(务必遵守):
- 以连贯短段落叙述为主,一段三四句、说透一层意思;不满篇列表,唯并列的宜忌清单可用 -。
- 用 ### 小标题分节,标题三到六字(如「命之根基」「财从何来」「行运之节」),全篇三到五节。
- 每节先下判断、再讲缘由、后给做法;关键结论以 **加粗** 点睛,一节至多一两处。
- 术语随手用白话点破(如「化忌——即此处易有牵绊执念」),不堆术语。
- 忌 AI 腔:不用「首先/其次/综上所述」,不用表情符号,不复述提问,不写客套结尾。
- 篇末以「### 叮嘱」收束:两三句长者式嘱咐,落到本月内可做的一两件实事。
- 输出为简体中文 Markdown。`

// BuildInterpretPrompt 由命盘 + 格局 + 知识库 + 古籍引文构建解读请求。
func BuildInterpretPrompt(
	chart *ziwei.Chart,
	patterns []ziwei.Pattern,
	kb *knowledge.Base,
	store *corpus.Store,
	topic string,
	question string,
) Request {
	var sb strings.Builder
	sb.WriteString("## 命盘数据\n\n")
	sb.WriteString(ChartSummary(chart))

	// 结构化断语骨架:逐宫「主星×庙旺×四化×煞吉」的确定性判定,供 LLM 贴盘发挥,
	// 避免脱离本盘写通用套话(这是不同命盘断语雷同的根因)。
	if rd := buildReading(chart, patterns); rd != nil {
		sb.WriteString("\n## 逐宫判定骨架(须据此贴盘,不得写通用套话)\n\n")
		sb.WriteString("命格总论:" + rd.Overview + "\n")
		for _, s := range rd.Sections {
			sb.WriteString(fmt.Sprintf("- 【%s·%s】%s星曜[%s]:%s\n",
				s.Title, s.Palace, levelWord(s.Level), strings.Join(s.Stars, " "), s.Text))
		}
	}

	if len(patterns) > 0 {
		sb.WriteString("\n## 已识别格局\n\n")
		for _, p := range patterns {
			sb.WriteString(fmt.Sprintf("- 【%s】(%s)%s", p.Name, levelLabel(p.Level), p.Description))
			if p.Source != "" {
				sb.WriteString(fmt.Sprintf("(出处:%s)", p.Source))
			}
			sb.WriteString("\n")
		}
	}

	// 命宫主星知识(倪海厦体系速览)
	ming := chart.MingGong()
	mainStars := ming.MajorStarNames()
	if len(mainStars) == 0 {
		mainStars = ming.BorrowedStars
	}
	if kb != nil && len(mainStars) > 0 {
		sb.WriteString("\n## 主星参考(倪海厦体系)\n\n")
		for _, name := range mainStars {
			if d, ok := kb.StarDesc[name]; ok {
				sb.WriteString(fmt.Sprintf("- %s:%s|五行属%s|%s\n", name, d.Keywords, d.Element, d.Nature))
			}
		}
	}

	// 古籍引文(RAG:按命宫主星检索原文;含研究语料——仅内部引用,不对外露出全文)。
	// 每星每书限引 1 条,分散引用面防单书霸榜。
	if store != nil && len(mainStars) > 0 {
		var cites []string
		seen := map[string]bool{}
		for _, name := range mainStars {
			seenBook := map[string]int{}
			for _, hit := range store.SearchAll(name, 8) {
				if seen[hit.ParagraphID] || seenBook[hit.BookSlug] >= 1 {
					continue
				}
				seen[hit.ParagraphID] = true
				seenBook[hit.BookSlug]++
				cites = append(cites, fmt.Sprintf("- 《%s·%s》:%s", hit.BookTitle, hit.ChapterTitle, truncateRunes(hit.Text, 80)))
				if len(seenBook) >= 3 {
					break
				}
			}
		}
		if len(cites) > 0 {
			sb.WriteString("\n## 古籍原文参考\n\n")
			sb.WriteString(strings.Join(cites, "\n"))
			sb.WriteString("\n")
		}
	}

	// 择时参考:该主题的宜忌年月(确定性推算,供 LLM 给出具体择时建议)。
	if brief := TimingBriefForTopic(chart, topic); brief != "" {
		sb.WriteString("\n## 择时参考(据本盘确定性推算,须结合命格给出宜忌年月)\n\n")
		sb.WriteString(brief)
	}

	// 解读主题
	sb.WriteString("\n## 解读要求\n\n")
	if label, ok := topicLabel(kb, topic); ok {
		sb.WriteString(fmt.Sprintf("请围绕「%s」主题,", label))
		if palace, ok2 := topicPalace(kb, topic); ok2 {
			sb.WriteString(fmt.Sprintf("以【%s】及其三方四正为核心,", palace))
		}
		sb.WriteString("结合命盘数据与格局给出深入解读。\n")
	} else {
		sb.WriteString("请给出命格总览:性格特质、格局高低、事业财运方向、感情婚姻、健康注意点,以及当前大限的运势重点。\n")
	}
	if question != "" {
		sb.WriteString(fmt.Sprintf("\n命主的具体问题:%s\n", question))
	}

	return Request{
		System:   systemPrompt,
		Messages: []Message{{Role: "user", Content: sb.String()}},
	}
}

// ChartSummary 命盘的紧凑文本表示(供 prompt 与降级解读共用)。
func ChartSummary(c *ziwei.Chart) string {
	var sb strings.Builder
	gender := "男"
	if c.BirthInfo.Gender == ziwei.Female {
		gender = "女"
	}
	sb.WriteString(fmt.Sprintf("- 性别:%s命\n", gender))
	sb.WriteString(fmt.Sprintf("- 公历:%d-%d-%d %s\n", c.BirthInfo.Year, c.BirthInfo.Month, c.BirthInfo.Day, c.TimeName))
	sb.WriteString(fmt.Sprintf("- 农历:%s(四柱:%s %s %s %s)\n", c.LunarDateText,
		c.FourPillars.Year, c.FourPillars.Month, c.FourPillars.Day, c.FourPillars.Hour))
	sb.WriteString(fmt.Sprintf("- 五行局:%s|命主:%s|身主:%s|生肖:%s\n", c.WuxingJuName, c.MingZhu, c.ShenZhu, c.Zodiac))
	if c.SiZhu != nil && c.SiZhu.GeJu != nil {
		sb.WriteString(fmt.Sprintf("- 四柱视角:日主%s%s,月令%s(%s;%s)\n",
			c.SiZhu.DayMaster, c.SiZhu.DayMasterElement, c.SiZhu.GeJu.Name, c.SiZhu.GeJu.Basis, c.SiZhu.GeJu.Source))
	}
	if c.SiZhu != nil && len(c.SiZhu.ShenSha) > 0 {
		parts := make([]string, 0, len(c.SiZhu.ShenSha))
		for _, s := range c.SiZhu.ShenSha {
			parts = append(parts, fmt.Sprintf("%s(%s)", s.Name, strings.Join(s.Pillars, "")))
		}
		sb.WriteString("- 四柱神煞:" + strings.Join(parts, "、") + "(参照维度,轻重以格局旺衰为主)\n")
	}
	if c.SiZhu != nil && c.SiZhu.DaYun != nil {
		dy := c.SiZhu.DaYun
		for _, d := range dy.List {
			if d.IsCurrent {
				sb.WriteString(fmt.Sprintf("- 当前大运:%s(%s,%d 岁起,%s局)\n",
					d.GanZhi, d.StemShiShen, d.StartAge, d.NaYin))
				break
			}
		}
		for _, l := range dy.CurrentLiuNian {
			if l.IsCurrent {
				sb.WriteString(fmt.Sprintf("- 流年:%s(%s)\n", l.GanZhi, l.StemShiShen))
				break
			}
		}
	}
	sb.WriteString(fmt.Sprintf("- 命宫:%s宫|身宫:%s宫\n", ziwei.Branches[c.MingGongBranch], ziwei.Branches[c.ShenGongBranch]))
	if c.CurrentDaXianIndex >= 0 && c.CurrentDaXianIndex < len(c.DaXians) {
		dx := c.DaXians[c.CurrentDaXianIndex]
		sb.WriteString(fmt.Sprintf("- 当前大限:%d-%d 岁,行【%s】(%s宫)\n", dx.StartAge, dx.EndAge, dx.PalaceName, ziwei.Branches[dx.PalaceBranch]))
	}
	sb.WriteString("\n十二宫概览(宫名|地支|星曜[亮度/四化]):\n")
	// 从命宫起顺时针罗列(命宫→父母→福德→…,地支索引逐位 +1)
	for i := 0; i < 12; i++ {
		p := c.PalaceByBranch(c.MingGongBranch + i)
		if p == nil {
			continue
		}
		var stars []string
		for _, s := range p.Stars {
			tag := s.Name
			if s.Brightness != "" {
				tag += "(" + s.Brightness + ")"
			}
			if s.SiHua != "" {
				tag += "化" + string(s.SiHua)
			}
			stars = append(stars, tag)
		}
		line := fmt.Sprintf("- %s|%s", p.Name, ziwei.Branches[p.Branch])
		if p.IsShenGong {
			line += "(身宫)"
		}
		if len(stars) > 0 {
			line += "|" + strings.Join(stars, " ")
		}
		if p.IsEmpty && len(p.BorrowedStars) > 0 {
			line += fmt.Sprintf("|空宫借对宫【%s】:%s", p.BorrowedFromName, strings.Join(p.BorrowedStars, " "))
		}
		sb.WriteString(line + "\n")
	}
	return sb.String()
}

func levelLabel(level string) string {
	switch level {
	case "excellent":
		return "上格"
	case "good":
		return "吉格"
	case "caution":
		return "凶格"
	default:
		return "中性"
	}
}

func topicLabel(kb *knowledge.Base, topic string) (string, bool) {
	if kb == nil || topic == "" {
		return "", false
	}
	l, ok := kb.Topics.Label[topic]
	return l, ok
}

func topicPalace(kb *knowledge.Base, topic string) (string, bool) {
	if kb == nil || topic == "" {
		return "", false
	}
	p, ok := kb.Topics.PalaceName[topic]
	return p, ok
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
