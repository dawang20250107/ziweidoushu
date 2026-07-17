// Package corpus 古籍语料库:内存存储、全文检索、外部导入。
//
// 设计目标:
//   - 读多写少、高并发:数据加载后不可变,读路径无锁竞争(替换指针 + RWMutex);
//   - 后期扩容:倪海厦著作、更多古籍以 JSON 文件投放到外部目录即可被导入,
//     无需改代码、无需数据库;
//   - 检索:中文二元(bigram)倒排索引 + 原文校验,毫秒级全文搜索。
package corpus

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Paragraph 古籍段落。
type Paragraph struct {
	ID          string `json:"id"`
	Idx         int    `json:"idx"`
	Text        string `json:"text"`
	Translation string `json:"translation,omitempty"`
	NiNote      string `json:"niNote,omitempty"` // 倪师注解
}

// Chapter 章节。
type Chapter struct {
	Title      string      `json:"title"`
	Subtitle   string      `json:"subtitle,omitempty"`
	Paragraphs []Paragraph `json:"paragraphs"`
}

// Book 一部古籍。
type Book struct {
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	Dynasty   string    `json:"dynasty"`
	Author    string    `json:"author"`
	Intro     string    `json:"intro"`
	WordCount int       `json:"wordCount"`
	Chapters  []Chapter `json:"chapters"`
	// Source 数据来源:embedded(随二进制打包)| external(外部目录导入)。
	Source string `json:"source,omitempty"`
	// Research 研究语料:不进书架、不进公开检索,仅供 AI 解读引用(内部研究)。
	Research bool `json:"research,omitempty"`
}

// BookMeta 书目摘要(列表接口用,不携带全文)。
type BookMeta struct {
	Title      string `json:"title"`
	Slug       string `json:"slug"`
	Dynasty    string `json:"dynasty"`
	Author     string `json:"author"`
	Intro      string `json:"intro"`
	WordCount  int    `json:"wordCount"`
	Chapters   int    `json:"chapters"`
	Paragraphs int    `json:"paragraphs"`
	Source     string `json:"source,omitempty"`
}

// SearchHit 检索命中。
type SearchHit struct {
	BookSlug     string `json:"bookSlug"`
	BookTitle    string `json:"bookTitle"`
	ChapterTitle string `json:"chapterTitle"`
	ChapterIdx   int    `json:"chapterIdx"`
	ParagraphID  string `json:"paragraphId"`
	// Snippet 高亮片段,命中词以 <mark></mark> 包裹(文本已做 HTML 转义)。
	Snippet string `json:"snippet"`
	Text    string `json:"text"`
}

// snapshot 一次完整加载的不可变视图。读路径只解引用,不加锁遍历可变结构。
type snapshot struct {
	books  []Book
	bySlug map[string]int
	index  *searchIndex
}

// Store 语料库。并发安全:所有读操作基于原子快照。
type Store struct {
	mu   sync.RWMutex
	snap *snapshot
}

// NewStore 从内置文件系统加载全部古籍。
// embeddedFS 传入 data.FS,dir 为古籍子目录(classics)。
func NewStore(embeddedFS fs.FS, dir string) (*Store, error) {
	books, err := loadBooksFromFS(embeddedFS, dir, "embedded")
	if err != nil {
		return nil, err
	}
	s := &Store{}
	s.replace(books)
	return s, nil
}

// LoadExternalDir 从外部目录导入古籍 JSON(为后期投放古籍资料/倪师著作准备)。
// 目录内每个 *.json 须符合 Book schema;slug 冲突时外部数据覆盖内置数据。
// 可重复调用实现热加载,调用期间读请求不受影响。
func (s *Store) LoadExternalDir(dir string) (int, error) {
	if dir == "" {
		return 0, nil
	}
	info, err := os.Stat(dir)
	if err != nil {
		return 0, fmt.Errorf("外部语料目录不可用: %w", err)
	}
	if !info.IsDir() {
		return 0, fmt.Errorf("外部语料路径不是目录: %s", dir)
	}
	external, err := loadBooksFromFS(os.DirFS(dir), ".", "external")
	if err != nil {
		return 0, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	merged := make([]Book, 0, len(s.snap.books)+len(external))
	overridden := map[string]bool{}
	for _, b := range external {
		overridden[b.Slug] = true
	}
	for _, b := range s.snap.books {
		if b.Source == "embedded" && !overridden[b.Slug] {
			merged = append(merged, b)
		}
	}
	merged = append(merged, external...)
	s.replaceLocked(merged)
	return len(external), nil
}

func loadBooksFromFS(fsys fs.FS, dir, source string) ([]Book, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("读取语料目录失败: %w", err)
	}
	var books []Book
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		raw, err := fs.ReadFile(fsys, filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("读取 %s 失败: %w", e.Name(), err)
		}
		var b Book
		if err := json.Unmarshal(raw, &b); err != nil {
			return nil, fmt.Errorf("解析 %s 失败: %w", e.Name(), err)
		}
		if b.Slug == "" || b.Title == "" || len(b.Chapters) == 0 {
			return nil, fmt.Errorf("%s 缺少必要字段(title/slug/chapters)", e.Name())
		}
		b.Source = source
		books = append(books, b)
	}
	sort.Slice(books, func(i, j int) bool { return books[i].Slug < books[j].Slug })
	return books, nil
}

func (s *Store) replace(books []Book) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.replaceLocked(books)
}

func (s *Store) replaceLocked(books []Book) {
	bySlug := make(map[string]int, len(books))
	for i := range books {
		bySlug[books[i].Slug] = i
	}
	s.snap = &snapshot{books: books, bySlug: bySlug, index: buildIndex(books)}
}

func (s *Store) view() *snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snap
}

// Books 公开书目摘要(研究语料不露出)。
func (s *Store) Books() []BookMeta {
	v := s.view()
	out := make([]BookMeta, 0, len(v.books))
	for i := range v.books {
		b := &v.books[i]
		if b.Research {
			continue
		}
		paras := 0
		for _, c := range b.Chapters {
			paras += len(c.Paragraphs)
		}
		out = append(out, BookMeta{
			Title: b.Title, Slug: b.Slug, Dynasty: b.Dynasty, Author: b.Author,
			Intro: b.Intro, WordCount: b.WordCount, Chapters: len(b.Chapters),
			Paragraphs: paras, Source: b.Source,
		})
	}
	return out
}

// Book 按 slug 取整本书,不存在返回 nil。
func (s *Store) Book(slug string) *Book {
	v := s.view()
	if i, ok := v.bySlug[slug]; ok {
		return &v.books[i]
	}
	return nil
}

// Chapter 取某书某章,越界返回 nil。
func (s *Store) Chapter(slug string, idx int) (*Book, *Chapter) {
	b := s.Book(slug)
	if b == nil || idx < 0 || idx >= len(b.Chapters) {
		return nil, nil
	}
	return b, &b.Chapters[idx]
}

// Stats 语料统计。
func (s *Store) Stats() map[string]int {
	v := s.view()
	chapters, paragraphs, chars := 0, 0, 0
	for i := range v.books {
		for _, c := range v.books[i].Chapters {
			chapters++
			paragraphs += len(c.Paragraphs)
			for _, p := range c.Paragraphs {
				chars += len([]rune(p.Text))
			}
		}
	}
	research := 0
	for i := range v.books {
		if v.books[i].Research {
			research++
		}
	}
	return map[string]int{
		"books":         len(v.books) - research,
		"researchBooks": research,
		"chapters":      chapters,
		"paragraphs":    paragraphs,
		"characters":    chars,
	}
}
