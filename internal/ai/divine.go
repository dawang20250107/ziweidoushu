// AI 解卦:梅花易数卦象 → 结构化提示词 → 流式解读。
// 与命盘解读共用供应商通道/信号量限流;引用语料(含研究语料)佐证断辞。
package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/dawang20250107/ziweidoushu/internal/corpus"
	"github.com/dawang20250107/ziweidoushu/internal/meihua"
)

const divineSystemPrompt = `你是精研邵康节《梅花易数》与倪海厦《天纪》卦象体系的解卦人。
依据给出的卦象(本卦/互卦/变卦/动爻/体用生克)围绕求测之事解卦:
先断吉凶大势(以体用生克为纲),再以本卦断当下、互卦断过程、变卦断结果,
结合卦象类象落到求测的具体事上,给出可操作的建议与应期倾向。
言之有据:引用给出的古籍参考时注明出处;不确定处直言不确定。
行文用简体中文,克制、清晰、不故弄玄虚;结尾提醒占卜为传统文化参考。`

// BuildDivinePrompt 由卦象 + 问题 + 语料引文构建解卦请求。
func BuildDivinePrompt(r *meihua.Result, store *corpus.Store) Request {
	var sb strings.Builder
	sb.WriteString("## 求测之事\n\n")
	if r.Question != "" {
		sb.WriteString(r.Question + "\n\n")
	} else {
		sb.WriteString("(未具体说明,断大势)\n\n")
	}

	sb.WriteString("## 卦象\n\n")
	if r.Method == "time" {
		sb.WriteString(fmt.Sprintf("- 起卦:时间起卦,农历 %s\n", r.LunarText))
	} else {
		sb.WriteString(fmt.Sprintf("- 起卦:数字起卦 %v\n", r.Numbers))
	}
	sb.WriteString(fmt.Sprintf("- 本卦:%s(上%s%s·下%s%s)\n",
		r.Ben.Name, r.Ben.Upper.Name, r.Ben.Upper.Nature, r.Ben.Lower.Name, r.Ben.Lower.Nature))
	sb.WriteString(fmt.Sprintf("- 动爻:第 %d 爻\n", r.Moving))
	sb.WriteString(fmt.Sprintf("- 互卦:%s\n", r.Hu.Name))
	sb.WriteString(fmt.Sprintf("- 变卦:%s\n", r.Bian.Name))
	sb.WriteString(fmt.Sprintf("- 体用:体=%s(%s),用=%s(%s),%s → %s\n\n",
		r.TiTrigram.Name, r.TiTrigram.Element, r.YongTrigram.Name, r.YongTrigram.Element,
		r.Relation, r.Verdict))

	// 语料引文(含研究语料;仅内部引用)。
	// 每书限引 1 条:同一典籍在检索中易霸榜,分散引用面让断辞更立体。
	if store != nil {
		var cites []string
		seenPara := map[string]bool{}
		seenBook := map[string]int{}
		for _, q := range []string{r.Ben.Name, r.Bian.Name, r.Hu.Name, r.TiTrigram.Name + "卦"} {
			for _, hit := range store.SearchAll(q, 6) {
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
			sb.WriteString("## 古籍参考\n\n")
			sb.WriteString(strings.Join(cites, "\n"))
			sb.WriteString("\n")
		}
	}

	return Request{
		System:   divineSystemPrompt,
		Messages: []Message{{Role: "user", Content: sb.String()}},
	}
}

// Divine 解卦(流式)。无供应商时返回 ErrNoProvider(provider.go 定义),
// 由调用方决定降级策略(付费解卦不走规则降级)。
func (it *Interpreter) Divine(ctx context.Context, r *meihua.Result, onDelta func(string) error) (Result, error) {
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
	req := BuildDivinePrompt(r, it.store)
	req.MaxTokens = it.maxTokens
	text, err := it.provider.Stream(ctx, req, onDelta)
	if err != nil {
		return Result{}, err
	}
	return Result{Text: text, Provider: it.ProviderName()}, nil
}
