package httpapi

import (
	"crypto/subtle"
	"net/http"
	"strconv"
)

func (s *Server) handleBooks(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"books": s.corpus.Books(),
		"stats": s.corpus.Stats(),
	})
}

func (s *Server) handleBook(w http.ResponseWriter, r *http.Request) {
	book := s.corpus.Book(r.PathValue("slug"))
	if book == nil {
		writeError(w, http.StatusNotFound, "not_found", "未收录该古籍")
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func (s *Server) handleChapter(w http.ResponseWriter, r *http.Request) {
	idx, err := strconv.Atoi(r.PathValue("idx"))
	if err != nil || idx < 0 {
		writeError(w, http.StatusBadRequest, "bad_index", "章节序号非法")
		return
	}
	book, chapter := s.corpus.Chapter(r.PathValue("slug"), idx)
	if chapter == nil {
		writeError(w, http.StatusNotFound, "not_found", "古籍或章节不存在")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"book":    map[string]string{"title": book.Title, "slug": book.Slug},
		"index":   idx,
		"total":   len(book.Chapters),
		"chapter": chapter,
	})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, http.StatusBadRequest, "bad_query", "缺少查询参数 q")
		return
	}
	limit := 30
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	hits := s.corpus.Search(q, limit)
	writeJSON(w, http.StatusOK, map[string]any{"query": q, "count": len(hits), "hits": hits})
}

// handleCorpusReload 重新加载外部古籍目录(投放新资料后热更新,无需重启)。
// 需配置 ADMIN_TOKEN 并以 Authorization: Bearer <token> 调用。
func (s *Server) handleCorpusReload(w http.ResponseWriter, r *http.Request) {
	token := s.cfg.AdminToken
	if token == "" {
		writeError(w, http.StatusForbidden, "disabled", "未配置 ADMIN_TOKEN,管理接口已禁用")
		return
	}
	auth := r.Header.Get("Authorization")
	if len(auth) < 8 || subtle.ConstantTimeCompare([]byte(auth), []byte("Bearer "+token)) != 1 {
		writeError(w, http.StatusUnauthorized, "unauthorized", "凭证无效")
		return
	}
	if s.cfg.CorpusExternalDir == "" {
		writeError(w, http.StatusBadRequest, "no_external_dir", "未配置 CORPUS_EXTERNAL_DIR")
		return
	}
	n, err := s.corpus.LoadExternalDir(s.cfg.CorpusExternalDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "reload_failed", err.Error())
		return
	}
	s.logger.Info("外部语料热加载完成", "count", n, "dir", s.cfg.CorpusExternalDir)
	writeJSON(w, http.StatusOK, map[string]any{"loaded": n, "stats": s.corpus.Stats()})
}
