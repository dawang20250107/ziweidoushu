// AI 六爻解卦:装卦结构 → 提示词 → 流式解读(引六爻正典语料)。
package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/dawang20250107/ziweidoushu/internal/liuyao"
)

const liuYaoSystemPrompt = `你是精研火珠林法的六爻解卦人,宗《增删卜易》《卜筮正宗》断法。
依据装卦(本卦/变卦/世应/六亲/六神/动爻/日月建)围绕求测之事解卦:
先取用神(按所测事项取对应六亲,说明取用理由),看用神旺衰、动爻生克冲合、
世应关系,断成败应期;引用给出的古籍参考须注明出处。
装卦已按经文标注客观事实层:各爻对月建旺衰(旺相休囚死)、日辰作用
(临/冲/合/扶/生/克/泄/耗)、日辰四态(长生/帝旺/墓/绝,野鹤口径)、
月破、旬空、暗动、日破,动爻另标动变作用(化进退神/伏吟反吟/化长生
墓绝合/回头生克)——以上以标注为准,不必自行推算,径直据此论生扶
克害;应期(冲空实空/墓库冲开/破待填合等)结合古籍参考推断。
行文简体中文,条理清晰、不故弄玄虚;不确定处直言;结尾提醒占卜为传统文化参考。`

// yaoLine 单爻描述行。
func yaoLine(y liuyao.Yao) string {
	mark := ""
	if y.IsShi {
		mark = " 世"
	}
	if y.IsYing {
		mark = " 应"
	}
	shape := "▅▅▅▅▅"
	if !y.Yang {
		shape = "▅▅ ▅▅"
	}
	s := fmt.Sprintf("%s %s %s%s(%s·%s,日%s)%s", y.LiuShen, y.LiuQin, y.Stem, y.Branch, y.Element, y.MonthState, y.DayRelation, mark)
	if y.DayStage != "" {
		s += " 日辰" + y.DayStage
	}
	if y.YuePo {
		s += " 月破"
	}
	if y.XunKong {
		s += " 旬空"
	}
	if y.AnDong {
		s += " 暗动"
	}
	if y.RiPo {
		s += " 日破"
	}
	if y.Moving {
		s += " ×动"
		if y.BianYao != nil {
			s += fmt.Sprintf(" → 变 %s %s%s(%s)", y.BianYao.LiuQin, y.BianYao.Stem, y.BianYao.Branch, y.BianYao.Element)
		}
		if y.BianRelation != "" {
			s += " " + y.BianRelation
		}
	}
	return fmt.Sprintf("第%d爻 %s  %s", y.Pos, shape, s)
}

// DivineLiuYao 六爻解卦(流式)。无供应商返回 ErrNoProvider(付费不降级)。
func (it *Interpreter) DivineLiuYao(ctx context.Context, r *liuyao.Result, onDelta func(string) error) (Result, error) {
	if it.provider == nil {
		return Result{}, ErrNoProvider
	}
	select {
	case it.sem <- struct{}{}:
		defer func() { <-it.sem }()
	case <-ctx.Done():
		return Result{}, ctx.Err()
	default:
		return Result{}, ErrBusy
	}

	var sb strings.Builder
	sb.WriteString("## 求测之事\n\n")
	if r.Question != "" {
		sb.WriteString(r.Question + "\n\n")
	} else {
		sb.WriteString("(未具体说明,断大势)\n\n")
	}
	sb.WriteString("## 装卦\n\n")
	sb.WriteString(fmt.Sprintf("- 起卦:%s,月建%s、日辰%s%s\n", r.LunarText, r.MonthJian, r.DayStem, r.DayBranch))
	sb.WriteString(fmt.Sprintf("- 本卦:%s(%s·%s)", r.BenName, r.Palace, r.PalaceSeq))
	if r.BianName != "" {
		sb.WriteString(fmt.Sprintf(",变卦:%s", r.BianName))
	}
	sb.WriteString("\n\n自上而下:\n")
	for i := 5; i >= 0; i-- {
		sb.WriteString(yaoLine(r.Yaos[i]) + "\n")
	}
	if r.YongShen != "" {
		sb.WriteString("\n用神建议:" + r.YongShen)
		if len(r.YongShenPos) > 0 {
			pos := make([]string, len(r.YongShenPos))
			for i, p := range r.YongShenPos {
				pos[i] = fmt.Sprintf("%d", p)
			}
			sb.WriteString("(第 " + strings.Join(pos, "、") + " 爻)")
		} else if r.YongShen != "世爻" {
			sb.WriteString("(用神不上卦,须论伏神)")
		}
		sb.WriteString("——" + r.YongShenBasis + "。事类识别或有出入,若与所测事理不符,以你按经义取用为准。\n")
	}

	// 语料引文:每书限一条,分散引用面
	if it.store != nil {
		var cites []string
		seenPara := map[string]bool{}
		seenBook := map[string]int{}
		queries := []string{r.BenName, "用神", "世应"}
		if r.BianName != "" {
			queries = append(queries, r.BianName)
		}
		for _, q := range queries {
			for _, hit := range it.store.SearchAll(q, 6) {
				if seenPara[hit.ParagraphID] || seenBook[hit.BookSlug] >= 1 {
					continue
				}
				seenPara[hit.ParagraphID] = true
				seenBook[hit.BookSlug]++
				cites = append(cites, fmt.Sprintf("- 《%s·%s》:%s", hit.BookTitle, hit.ChapterTitle, truncateRunes(hit.Text, 80)))
				if len(cites) >= 6 {
					break
				}
			}
		}
		if len(cites) > 0 {
			sb.WriteString("\n## 古籍参考\n\n")
			sb.WriteString(strings.Join(cites, "\n"))
			sb.WriteString("\n")
		}
	}

	req := Request{
		System:    liuYaoSystemPrompt,
		Messages:  []Message{{Role: "user", Content: sb.String()}},
		MaxTokens: it.maxTokens,
	}
	text, err := it.provider.Stream(ctx, req, onDelta)
	if err != nil {
		return Result{}, err
	}
	return Result{Text: text, Provider: it.ProviderName()}, nil
}
