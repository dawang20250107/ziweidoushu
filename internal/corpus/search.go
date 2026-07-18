package corpus

import (
	"sort"
	"strings"
)

// searchIndex 中文二元(bigram)倒排索引。
//
// 索引构建一次后只读,查询无锁。对短查询(单字)退化为一元索引;
// 命中候选段落后回原文做子串校验,保证零误报。
type searchIndex struct {
	// unigram/bigram token → 段落全局序号列表(升序去重)
	unigram map[rune][]int32
	bigram  map[string][]int32
	// 段落全局序号 → 位置
	paras []paraRef
}

type paraRef struct {
	bookIdx    int32
	chapterIdx int32
	paraIdx    int32
}

func buildIndex(books []Book) *searchIndex {
	idx := &searchIndex{
		unigram: make(map[rune][]int32),
		bigram:  make(map[string][]int32),
	}
	for bi := range books {
		for ci := range books[bi].Chapters {
			for pi := range books[bi].Chapters[ci].Paragraphs {
				id := int32(len(idx.paras))
				idx.paras = append(idx.paras, paraRef{int32(bi), int32(ci), int32(pi)})
				text := normalizeQuery(books[bi].Chapters[ci].Paragraphs[pi].Text)
				runes := []rune(text)
				seenUni := map[rune]bool{}
				seenBi := map[string]bool{}
				for i, r := range runes {
					if !seenUni[r] {
						seenUni[r] = true
						idx.unigram[r] = append(idx.unigram[r], id)
					}
					if i+1 < len(runes) {
						bg := string(runes[i : i+2])
						if !seenBi[bg] {
							seenBi[bg] = true
							idx.bigram[bg] = append(idx.bigram[bg], id)
						}
					}
				}
			}
		}
	}
	return idx
}

func normalizeQuery(s string) string { return strings.ToLower(s) }

// candidates 返回可能包含 query 的段落全局序号(交集)。
func (idx *searchIndex) candidates(query string) []int32 {
	runes := []rune(query)
	if len(runes) == 0 {
		return nil
	}
	if len(runes) == 1 {
		return idx.unigram[runes[0]]
	}
	var lists [][]int32
	for i := 0; i+1 < len(runes); i++ {
		list, ok := idx.bigram[string(runes[i:i+2])]
		if !ok {
			return nil
		}
		lists = append(lists, list)
	}
	// 从最短列表开始求交集
	sort.Slice(lists, func(a, b int) bool { return len(lists[a]) < len(lists[b]) })
	result := lists[0]
	for _, l := range lists[1:] {
		result = intersect(result, l)
		if len(result) == 0 {
			return nil
		}
	}
	return result
}

func intersect(a, b []int32) []int32 {
	out := make([]int32, 0, min32(len(a), len(b)))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			out = append(out, a[i])
			i++
			j++
		case a[i] < b[j]:
			i++
		default:
			j++
		}
	}
	return out
}

func min32(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Search 公开全文检索(研究语料不露出)。返回按原文顺序的命中,limit 上限截断。
func (s *Store) Search(query string, limit int) []SearchHit {
	return s.search(query, limit, false)
}

// SearchAll 全域检索(含研究语料)。仅供 AI 解读引用等内部路径使用,
// 严禁直接暴露给公开接口。
func (s *Store) SearchAll(query string, limit int) []SearchHit {
	return s.search(query, limit, true)
}

func (s *Store) search(query string, limit int, includeResearch bool) []SearchHit {
	q := strings.TrimSpace(query)
	if q == "" {
		return []SearchHit{}
	}
	if limit <= 0 || limit > 200 {
		limit = 30
	}
	v := s.view()
	nq := normalizeQuery(q)
	hits := make([]SearchHit, 0, limit)
	for _, id := range v.index.candidates(nq) {
		ref := v.index.paras[id]
		book := &v.books[ref.bookIdx]
		if book.Research && !includeResearch {
			continue
		}
		chapter := &book.Chapters[ref.chapterIdx]
		para := &chapter.Paragraphs[ref.paraIdx]
		pos := strings.Index(normalizeQuery(para.Text), nq)
		if pos < 0 {
			continue // bigram 候选误报,原文校验剔除
		}
		hits = append(hits, SearchHit{
			BookSlug:     book.Slug,
			BookTitle:    book.Title,
			ChapterTitle: chapter.Title,
			ChapterIdx:   int(ref.chapterIdx),
			ParagraphID:  para.ID,
			Snippet:      buildSnippet(para.Text, pos, len(nq)),
			Text:         para.Text,
		})
		if len(hits) >= limit {
			break
		}
	}
	return hits
}

// buildSnippet 提取命中上下文(前后各 40 字),命中词加 <mark> 并转义 HTML。
func buildSnippet(text string, bytePos, byteLen int) string {
	runes := []rune(text)
	// 字节位置 → rune 位置
	runePos := len([]rune(text[:bytePos]))
	runeLen := len([]rune(text[bytePos : bytePos+byteLen]))

	start := runePos - 40
	if start < 0 {
		start = 0
	}
	end := runePos + runeLen + 40
	if end > len(runes) {
		end = len(runes)
	}
	var sb strings.Builder
	if start > 0 {
		sb.WriteString("…")
	}
	sb.WriteString(escapeHTML(string(runes[start:runePos])))
	sb.WriteString("<mark>")
	sb.WriteString(escapeHTML(string(runes[runePos : runePos+runeLen])))
	sb.WriteString("</mark>")
	sb.WriteString(escapeHTML(string(runes[runePos+runeLen : end])))
	if end < len(runes) {
		sb.WriteString("…")
	}
	return sb.String()
}

var htmlEscaper = strings.NewReplacer(
	"&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#039;",
)

func escapeHTML(s string) string { return htmlEscaper.Replace(s) }
