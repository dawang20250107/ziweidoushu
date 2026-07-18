// researchgen:研究语料文本 → corpus Book JSON(research:true)。
//
// 输入:research/corpus/*.txt + manifest.json(来源/格式/字符数)
// 输出:research/books-json/r-<hash8>.json,可经 CORPUS_EXTERNAL_DIR 热加载。
// 研究语料不进书架、不进公开检索,仅 AI 解读引用(见 internal/corpus Research 标志)。
//
// 用法:go run ./tools/researchgen [输入目录] [输出目录]
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type paragraph struct {
	ID   string `json:"id"`
	Idx  int    `json:"idx"`
	Text string `json:"text"`
}

type chapter struct {
	Title      string      `json:"title"`
	Paragraphs []paragraph `json:"paragraphs"`
}

type book struct {
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	Dynasty   string    `json:"dynasty"`
	Author    string    `json:"author"`
	Intro     string    `json:"intro"`
	WordCount int       `json:"wordCount"`
	Chapters  []chapter `json:"chapters"`
	Research  bool      `json:"research"`
}

type manifestDoc struct {
	File   string `json:"file"`
	Source string `json:"source"`
	Format string `json:"format"`
	Chars  int    `json:"chars"`
	TopDir string `json:"topdir"`
}

const (
	maxParaRunes  = 420 // 段落切分上限(RAG 引用粒度)
	parasPerChap  = 40  // 每章段落数
	minParaRunes  = 8   // 短于此的独立行并入下一段
)

var sentenceEnd = regexp.MustCompile(`[。!?!?;;]`)

// splitParas 文本 → 段落序列:按空行切,超长段在句界二分。
func splitParas(text string) []string {
	var out []string
	for _, blk := range regexp.MustCompile(`\n\s*\n`).Split(text, -1) {
		blk = strings.TrimSpace(strings.ReplaceAll(blk, "\n", ""))
		if blk == "" {
			continue
		}
		runes := []rune(blk)
		for len(runes) > maxParaRunes {
			cut := maxParaRunes
			// 在 [maxParaRunes/2, maxParaRunes] 内找最近的句界
			seg := string(runes[maxParaRunes/2 : maxParaRunes])
			if loc := sentenceEnd.FindAllStringIndex(seg, -1); len(loc) > 0 {
				last := loc[len(loc)-1]
				cut = maxParaRunes/2 + len([]rune(seg[:last[1]]))
			}
			out = append(out, string(runes[:cut]))
			runes = runes[cut:]
		}
		if len(runes) > 0 {
			out = append(out, string(runes))
		}
	}
	// 极短行并入后段(标题残片等)
	var merged []string
	for i := 0; i < len(out); i++ {
		if len([]rune(out[i])) < minParaRunes && i+1 < len(out) {
			out[i+1] = out[i] + " " + out[i+1]
			continue
		}
		merged = append(merged, out[i])
	}
	return merged
}

func buildBook(id, title, topdir, source, text string) book {
	paras := splitParas(text)
	var chapters []chapter
	for i := 0; i < len(paras); i += parasPerChap {
		end := min(i+parasPerChap, len(paras))
		ch := chapter{Title: fmt.Sprintf("第 %d 部分", len(chapters)+1)}
		for j, p := range paras[i:end] {
			ch.Paragraphs = append(ch.Paragraphs, paragraph{
				ID: fmt.Sprintf("%s-%d-%d", id, len(chapters)+1, j+1), Idx: j, Text: p,
			})
		}
		chapters = append(chapters, ch)
	}
	chars := 0
	for _, p := range paras {
		chars += len([]rune(p))
	}
	return book{
		Title: title, Slug: "r-" + id, Dynasty: "研究语料",
		Author: topdir, Intro: "内部研究语料,来源:" + source,
		WordCount: chars, Chapters: chapters, Research: true,
	}
}

func main() {
	inDir, outDir := "research/corpus", "research/books-json"
	if len(os.Args) > 2 {
		inDir, outDir = os.Args[1], os.Args[2]
	}
	raw, err := os.ReadFile(filepath.Join(inDir, "manifest.json"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "读取 manifest 失败:", err)
		os.Exit(1)
	}
	var manifest struct {
		Docs []manifestDoc `json:"docs"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		fmt.Fprintln(os.Stderr, "解析 manifest 失败:", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	totalBooks, totalChars := 0, 0
	for _, d := range manifest.Docs {
		text, err := os.ReadFile(filepath.Join(inDir, d.File))
		if err != nil {
			fmt.Fprintf(os.Stderr, "跳过 %s: %v\n", d.File, err)
			continue
		}
		id := strings.SplitN(d.File, "-", 2)[0]
		title := strings.TrimSuffix(strings.SplitN(d.File, "-", 2)[1], ".txt")
		title = strings.ReplaceAll(title, "_", " ")
		b := buildBook(id, title, d.TopDir, d.Source, string(text))
		if len(b.Chapters) == 0 {
			continue
		}
		out, _ := json.MarshalIndent(b, "", " ")
		if err := os.WriteFile(filepath.Join(outDir, b.Slug+".json"), out, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		totalBooks++
		totalChars += b.WordCount
	}
	fmt.Printf("生成 %d 部研究语料书,共 %d 字符 → %s\n", totalBooks, totalChars, outDir)
}
