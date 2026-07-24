// AI 解课:大六壬起课结构 → 提示词 → 流式解读(引六壬语料)。
package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/dawang20250107/ziweidoushu/internal/daliuren"
)

const daLiuRenSystemPrompt = `你是精研大六壬的解课人,宗《六壬大全》《大六壬指南》课经毕法之义。
依据起课(天地盘/四课/三传/课体/十二天将)围绕求测之事解课:
以课体定格局、三传定来龙去脉(初传事之发端、中传经过、末传归宿),
天将定吉凶色彩(凡壬课吉凶系于天将),结合月将占时论旺衰应期。
已给出确定性断语骨架(课体+三传生克+乘将机械推演)——须据此贴课发挥、
不得与骨架吉凶相悖空谈;引用古籍参考须注明出处;不确定处直言。
行文如一位断课多年的长者当面讲解:短段落娓娓道来,先断后释再嘱;
用 ### 小标题分节(如「课象大势」「三传始末」「天将所示」「应期」「叮嘱」),
引文单独成 > 引用块,关键断语以 **加粗** 点睛;不用「首先/其次」腔调、
不用表情符号。简体中文;结尾「### 叮嘱」提醒占卜为传统文化参考并落一件实事。`

// DivineDaLiuRen 大六壬解课(流式)。无供应商返回 ErrNoProvider(付费不降级)。
func (it *Interpreter) DivineDaLiuRen(ctx context.Context, r *daliuren.Result, question string, onDelta func(string) error) (Result, error) {
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
	if question != "" {
		sb.WriteString(question + "\n\n")
	} else {
		sb.WriteString("(未具体说明,断大势)\n\n")
	}
	gui := "夜贵"
	if r.GuiIsDay {
		gui = "昼贵"
	}
	sb.WriteString("## 起课\n\n")
	sb.WriteString(fmt.Sprintf("- %s%s日 %s时占,月将%s,用%s,课体【%s】\n",
		r.DayStem, r.DayBranch, r.HourBranch, r.MonthGen, gui, r.KeType))
	branches := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	var tp []string
	for i := 0; i < 12; i++ {
		tp = append(tp, fmt.Sprintf("%s上%s乘%s", branches[i], r.TianPan[i], r.TianJiang[i]))
	}
	sb.WriteString("- 天地盘:" + strings.Join(tp, "、") + "\n")
	var ke []string
	for i, k := range r.Ke {
		ke = append(ke, fmt.Sprintf("%d课%s/%s", i+1, k.Upper, k.Lower))
	}
	sb.WriteString("- 四课:" + strings.Join(ke, " ") + "\n")
	sb.WriteString(fmt.Sprintf("- 三传:初%s乘%s、中%s乘%s、末%s乘%s\n\n",
		r.Chuan[0], r.ChuanJiang[0], r.Chuan[1], r.ChuanJiang[1], r.Chuan[2], r.ChuanJiang[2]))

	if j := r.Judgment; j != nil {
		sb.WriteString("## 断语骨架(确定性推演,须据此贴课、不得脱课空谈)\n\n")
		sb.WriteString(j.Conclusion + "\n- " + j.KeTypeText + "\n")
		for _, p := range j.Points {
			sb.WriteString("- " + p + "\n")
		}
		sb.WriteString("\n")
	}

	// 语料引文:每书限一条(现有六壬语料:精校本六壬大全/课经集/指南注解等)
	if it.store != nil {
		var cites []string
		seenPara := map[string]bool{}
		seenBook := map[string]int{}
		for _, q := range []string{r.KeType + "课", "三传", "天将", "月将"} {
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
			sb.WriteString("## 古籍参考\n\n" + strings.Join(cites, "\n") + "\n")
		}
	}

	req := Request{
		System:    daLiuRenSystemPrompt,
		Messages:  []Message{{Role: "user", Content: sb.String()}},
		MaxTokens: it.maxTokens,
	}
	text, err := it.provider.Stream(ctx, req, onDelta)
	if err != nil {
		return Result{}, err
	}
	return Result{Text: text, Provider: it.ProviderName()}, nil
}
