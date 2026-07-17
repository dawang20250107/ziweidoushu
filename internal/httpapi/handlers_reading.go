package httpapi

import (
	"errors"
	"net/http"

	"github.com/dawang20250107/ziweidoushu/internal/store"
)

// ── 古籍阅读:跨端续读进度 + 书签(全部 requireAuth)──────────────

// maxExcerptRunes 书签摘录上限(字符数,非字节)。
const maxExcerptRunes = 200

// truncateRunes 按字符(rune)截断,避免中文被拦腰截断为半个字节。
func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) > max {
		return string(r[:max])
	}
	return s
}

// handleGetReading GET /api/v1/me/reading → {progress:[...]}(最近读的在前)。
func (s *Server) handleGetReading(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	progress, err := s.store.ListReadingProgress(r.Context(), claims.Sub)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "reading_failed", err.Error())
		return
	}
	if progress == nil {
		progress = []store.ReadingProgress{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"progress": progress})
}

// handlePutReading PUT /api/v1/me/reading/{slug} → {ok:true}(upsert;chapterIdx<0 拒 400)。
func (s *Server) handlePutReading(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	slug := r.PathValue("slug")
	if slug == "" {
		writeError(w, http.StatusBadRequest, "invalid_slug", "书籍标识缺失")
		return
	}
	var req struct {
		ChapterIdx  int    `json:"chapterIdx"`
		ParagraphID string `json:"paragraphId"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.ChapterIdx < 0 {
		writeError(w, http.StatusBadRequest, "invalid_chapter", "章节索引非法")
		return
	}
	if err := s.store.UpsertReadingProgress(r.Context(), claims.Sub, slug, req.ChapterIdx, req.ParagraphID); err != nil {
		writeError(w, http.StatusInternalServerError, "reading_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// handleListBookmarks GET /api/v1/me/bookmarks?book={slug?} → {bookmarks:[...]}。
func (s *Server) handleListBookmarks(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	book := r.URL.Query().Get("book")
	bookmarks, err := s.store.ListBookmarks(r.Context(), claims.Sub, book)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "bookmarks_failed", err.Error())
		return
	}
	if bookmarks == nil {
		bookmarks = []store.Bookmark{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"bookmarks": bookmarks})
}

// handleAddBookmark POST /api/v1/me/bookmarks → {bookmark}(excerpt 截断 ≤200 字符;同段幂等)。
func (s *Server) handleAddBookmark(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	var req struct {
		BookSlug    string `json:"bookSlug"`
		ChapterIdx  int    `json:"chapterIdx"`
		ParagraphID string `json:"paragraphId"`
		Excerpt     string `json:"excerpt"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.BookSlug == "" {
		writeError(w, http.StatusBadRequest, "invalid_book", "书籍标识缺失")
		return
	}
	if req.ChapterIdx < 0 {
		writeError(w, http.StatusBadRequest, "invalid_chapter", "章节索引非法")
		return
	}
	excerpt := truncateRunes(req.Excerpt, maxExcerptRunes)
	bm, err := s.store.AddBookmark(r.Context(), claims.Sub, req.BookSlug, req.ChapterIdx, req.ParagraphID, excerpt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "bookmark_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"bookmark": bm})
}

// handleDeleteBookmark DELETE /api/v1/me/bookmarks/{id} → {deleted:true}(越权/不存在均 404)。
func (s *Server) handleDeleteBookmark(w http.ResponseWriter, r *http.Request) {
	claims := currentClaims(r.Context())
	err := s.store.DeleteBookmark(r.Context(), claims.Sub, r.PathValue("id"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "书签不存在")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "bookmark_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}
