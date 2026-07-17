package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/dawang20250107/ziweidoushu/internal/corpus"
	"github.com/dawang20250107/ziweidoushu/internal/knowledge"
	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// Interpreter 命盘解读器:LLM 优先,无供应商时规则化降级。
// 并发安全;上游调用受信号量限流,避免把上游打挂或撑爆本服务内存。
type Interpreter struct {
	provider  Provider // 可为 nil(降级模式)
	kb        *knowledge.Base
	store     *corpus.Store
	sem       chan struct{}
	maxTokens int
}

// NewInterpreter 构建解读器。provider 可为 nil。
func NewInterpreter(p Provider, kb *knowledge.Base, store *corpus.Store, cfg Config) *Interpreter {
	n := cfg.MaxConcurrency
	if n <= 0 {
		n = 8
	}
	return &Interpreter{
		provider:  p,
		kb:        kb,
		store:     store,
		sem:       make(chan struct{}, n),
		maxTokens: cfg.MaxTokens,
	}
}

// HasProvider 是否配置了真实 LLM 供应商(false = 只有规则化降级)。
func (it *Interpreter) HasProvider() bool { return it.provider != nil }

// ProviderName 当前供应商标识,降级模式返回 "fallback/rule-based"。
func (it *Interpreter) ProviderName() string {
	if it.provider == nil {
		return "fallback/rule-based"
	}
	return it.provider.Name()
}

// Result 解读结果。
type Result struct {
	Text     string `json:"text"`
	Provider string `json:"provider"`
	Degraded bool   `json:"degraded"` // true = 规则化降级(未走 LLM)
}

// Interpret 解读命盘。onDelta 非 nil 时流式回调增量文本(降级模式一次性回调全文)。
func (it *Interpreter) Interpret(
	ctx context.Context,
	chart *ziwei.Chart,
	patterns []ziwei.Pattern,
	topic, question string,
	onDelta func(string) error,
) (Result, error) {
	if it.provider == nil {
		text := it.fallbackText(chart, patterns, topic, question)
		if onDelta != nil {
			if err := onDelta(text); err != nil {
				return Result{}, err
			}
		}
		return Result{Text: text, Provider: it.ProviderName(), Degraded: true}, nil
	}

	// 上游并发限流:满载时快速失败,把压力信号还给调用端(可重试),
	// 而不是让请求堆积拖垮整个服务。
	select {
	case it.sem <- struct{}{}:
		defer func() { <-it.sem }()
	case <-ctx.Done():
		return Result{}, ctx.Err()
	default:
		return Result{}, ErrBusy
	}

	req := BuildInterpretPrompt(chart, patterns, it.kb, it.store, topic, question)
	req.MaxTokens = it.maxTokens
	text, err := it.provider.Stream(ctx, req, onDelta)
	if err != nil {
		return Result{}, err
	}
	return Result{Text: text, Provider: it.ProviderName()}, nil
}

// ErrBusy 上游并发已满。
var ErrBusy = fmt.Errorf("AI 解读并发已满,请稍后重试")

// fallbackText 规则化解读:不依赖 LLM,由命盘 + 格局 + 知识库生成结构化文本。
func (it *Interpreter) fallbackText(chart *ziwei.Chart, patterns []ziwei.Pattern, topic, question string) string {
	var sb strings.Builder
	sb.WriteString("# 命盘解读(知识库规则版)\n\n")
	sb.WriteString("> 当前未配置 AI 模型,以下内容由倪海厦体系知识库规则生成;配置 ANTHROPIC_API_KEY 或 OPENAI_API_KEY 后可获得 AI 深度解读。\n\n")

	sb.WriteString("## 命盘基础\n\n")
	sb.WriteString(ChartSummary(chart))

	// 命宫主星解析
	ming := chart.MingGong()
	mainStars := ming.MajorStarNames()
	borrowed := false
	if len(mainStars) == 0 {
		mainStars = ming.BorrowedStars
		borrowed = true
	}
	if len(mainStars) > 0 && it.kb != nil {
		sb.WriteString("\n## 命宫主星\n\n")
		if borrowed {
			sb.WriteString(fmt.Sprintf("命宫无主星,借对宫【%s】主星论:\n\n", ming.BorrowedFromName))
		}
		for _, name := range mainStars {
			if d, ok := it.kb.StarDesc[name]; ok {
				sb.WriteString(fmt.Sprintf("- **%s**(五行属%s,%s):%s\n", name, d.Element, d.Nature, d.Keywords))
			}
		}
	}

	if len(patterns) > 0 {
		sb.WriteString("\n## 格局判定\n\n")
		for _, p := range patterns {
			sb.WriteString(fmt.Sprintf("### %s(%s)\n\n%s\n\n", p.Name, levelLabel(p.Level), p.Description))
			if p.Conditions != nil {
				if len(p.Conditions.Required) > 0 {
					sb.WriteString(fmt.Sprintf("- 成格条件:%s\n", strings.Join(p.Conditions.Required, ";")))
				}
				if len(p.Conditions.Bonus) > 0 {
					sb.WriteString(fmt.Sprintf("- 加分:%s\n", strings.Join(p.Conditions.Bonus, ";")))
				}
				if len(p.Conditions.Breaking) > 0 {
					sb.WriteString(fmt.Sprintf("- ⚠️ 破格警示:%s\n", strings.Join(p.Conditions.Breaking, ";")))
				}
			}
			if p.Source != "" {
				sb.WriteString(fmt.Sprintf("- 出处:%s\n", p.Source))
			}
			sb.WriteString("\n")
		}
	}

	// 感情主题:引用合盘知识库夫妻宫断语
	if (topic == "love" || topic == "") && it.kb != nil {
		fuqi := chart.PalaceByName("夫妻")
		if fuqi != nil {
			stars := fuqi.MajorStarNames()
			if len(stars) == 0 {
				stars = fuqi.BorrowedStars
			}
			var parts []string
			for _, name := range stars {
				if entry, ok := it.kb.Heming.StarInFuqi[name]; ok {
					parts = append(parts, fmt.Sprintf("- **%s坐夫妻宫**:%s(婚期建议:%s)", name, entry.Summary, entry.Timing))
				}
			}
			if len(parts) > 0 {
				sb.WriteString("\n## 感情婚姻(夫妻宫)\n\n")
				sb.WriteString(strings.Join(parts, "\n"))
				sb.WriteString("\n")
			}
		}
	}

	// 古籍原文参考
	if it.store != nil && len(mainStars) > 0 {
		var cites []string
		for _, name := range mainStars {
			for _, hit := range it.store.Search(name, 1) {
				cites = append(cites, fmt.Sprintf("- 《%s·%s》:%s", hit.BookTitle, hit.ChapterTitle, truncateRunes(hit.Text, 60)))
			}
		}
		if len(cites) > 0 {
			sb.WriteString("\n## 古籍参考\n\n")
			sb.WriteString(strings.Join(cites, "\n"))
			sb.WriteString("\n")
		}
	}

	// 当前大限
	if chart.CurrentDaXianIndex >= 0 && chart.CurrentDaXianIndex < len(chart.DaXians) {
		dx := chart.DaXians[chart.CurrentDaXianIndex]
		sb.WriteString(fmt.Sprintf("\n## 当前大限\n\n%d-%d 岁行【%s】大限。倪师体系下四化固定不动,大限重在观察宫位星曜组合的十年主题切换。\n",
			dx.StartAge, dx.EndAge, dx.PalaceName))
	}

	if question != "" {
		sb.WriteString(fmt.Sprintf("\n## 关于你的问题\n\n「%s」——规则版暂无法针对性作答,建议配置 AI 模型获得深度解读。\n", question))
	}

	if it.kb != nil && len(it.kb.TianjiQuotes) > 0 {
		quote := it.kb.TianjiQuotes[(chart.MingGongBranch*7+chart.LunarInfo.LunarDay)%len(it.kb.TianjiQuotes)]
		sb.WriteString(fmt.Sprintf("\n---\n\n> 倪师语录:%s\n", quote.Text))
	}
	return sb.String()
}
