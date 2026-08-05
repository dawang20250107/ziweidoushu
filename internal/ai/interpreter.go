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
	Text     string   `json:"text"`
	Provider string   `json:"provider"`
	Degraded bool     `json:"degraded"`          // true = 规则化降级(未走 LLM)
	Reading  *Reading `json:"reading,omitempty"` // 结构化多维断语骨架(始终随盘生成)
}

// Interpret 解读命盘。onDelta 非 nil 时流式回调增量文本(降级模式一次性回调全文)。
func (it *Interpreter) Interpret(
	ctx context.Context,
	chart *ziwei.Chart,
	patterns []ziwei.Pattern,
	topic, question string,
	onDelta func(string) error,
) (Result, error) {
	reading := buildReading(chart, patterns)
	if it.provider == nil {
		text := it.fallbackText(chart, patterns, topic, question)
		if onDelta != nil {
			if err := onDelta(text); err != nil {
				return Result{}, err
			}
		}
		return Result{Text: text, Provider: it.ProviderName(), Degraded: true, Reading: reading}, nil
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
	return Result{Text: text, Provider: it.ProviderName(), Reading: reading}, nil
}

// ErrBusy 上游并发已满。
var ErrBusy = fmt.Errorf("AI 解读并发已满,请稍后重试")

// fallbackText 规则化解读:不依赖 LLM,由命盘组合多维断语(逐宫随星曜实配而变)。
func (it *Interpreter) fallbackText(chart *ziwei.Chart, patterns []ziwei.Pattern, topic, question string) string {
	var sb strings.Builder
	sb.WriteString("# 命盘多维断语(规则版)\n\n")
	sb.WriteString("> 以下由命盘各宫「主星 × 庙旺 × 四化 × 煞吉」组合而成,随盘而异;配置 AI 模型后另有深度行文解读。\n\n")

	rd := it.BuildReading(chart, patterns)
	sb.WriteString(renderReadingMarkdown(rd, topic))

	// 格局细节(条件/破格)
	if len(patterns) > 0 {
		sb.WriteString("\n## 格局详解\n\n")
		for _, p := range patterns {
			sb.WriteString(fmt.Sprintf("### %s(%s)\n\n%s\n\n", p.Name, levelLabel(p.Level), p.Description))
			if p.Conditions != nil {
				if len(p.Conditions.Required) > 0 {
					sb.WriteString(fmt.Sprintf("- 成格条件:%s\n", strings.Join(p.Conditions.Required, ";")))
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

	// 古籍原文参考(按命宫主星检索)
	if it.store != nil {
		ming := chart.MingGong()
		mainStars := ming.MajorStarNames()
		if len(mainStars) == 0 {
			mainStars = ming.BorrowedStars
		}
		var cites []string
		for _, name := range mainStars {
			for _, hit := range it.store.Search(name, 1) {
				cites = append(cites, fmt.Sprintf("- 《%s·%s》:%s", hit.BookTitle, hit.ChapterTitle, truncateRunes(hit.Text, 60)))
			}
		}
		if len(cites) > 0 {
			sb.WriteString("\n## 古籍参考\n\n" + strings.Join(cites, "\n") + "\n")
		}
	}

	if question != "" {
		sb.WriteString(fmt.Sprintf("\n## 关于你的问题\n\n「%s」——规则版按上表定盘作答;配置 AI 模型可获针对性深度解读。\n", question))
	}

	if it.kb != nil && len(it.kb.TianjiQuotes) > 0 {
		quote := it.kb.TianjiQuotes[(chart.MingGongBranch*7+chart.LunarInfo.LunarDay)%len(it.kb.TianjiQuotes)]
		sb.WriteString(fmt.Sprintf("\n---\n\n> 倪师语录:%s\n", quote.Text))
	}
	return sb.String()
}

// renderReadingMarkdown 把结构化多维断语渲染为 Markdown(topic 非空时把对应维度提前)。
func renderReadingMarkdown(rd *Reading, topic string) string {
	if rd == nil {
		return ""
	}
	levelMark := map[string]string{"good": "◎ 吉", "caution": "△ 慎", "neutral": "○ 平"}
	var sb strings.Builder
	if rd.Overview != "" {
		sb.WriteString("## 命格总论\n\n" + rd.Overview + "\n")
	}
	for _, s := range rd.Sections {
		sb.WriteString(fmt.Sprintf("\n## %s  _%s_\n\n", s.Title, levelMark[s.Level]))
		if len(s.Stars) > 0 {
			sb.WriteString("星曜:" + strings.Join(s.Stars, "、") + "\n\n")
		}
		sb.WriteString(s.Text + "\n")
	}
	return sb.String()
}
