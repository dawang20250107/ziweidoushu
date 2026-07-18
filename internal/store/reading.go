package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// ── 古籍阅读:跨端续读进度 + 书签 ────────────────────────────────

// ReadingProgress 单本书的阅读进度(跨端续读的最小事实)。
type ReadingProgress struct {
	BookSlug    string    `json:"bookSlug"`
	ChapterIdx  int       `json:"chapterIdx"`
	ParagraphID string    `json:"paragraphId"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Bookmark 阅读书签(定位到某书某章某段 + 摘录)。
type Bookmark struct {
	ID          string    `json:"id"`
	BookSlug    string    `json:"bookSlug"`
	ChapterIdx  int       `json:"chapterIdx"`
	ParagraphID string    `json:"paragraphId"`
	Excerpt     string    `json:"excerpt"`
	CreatedAt   time.Time `json:"createdAt"`
}

// UpsertReadingProgress 写入/覆盖某书进度(同书二次写入取后值)。
// reading_progress 无 updated_at 触发器,故此处显式刷新时间戳。
func (s *Store) UpsertReadingProgress(ctx context.Context, userID, bookSlug string, chapterIdx int, paragraphID string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO reading_progress (user_id, book_slug, chapter_idx, paragraph_id, updated_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (user_id, book_slug) DO UPDATE
		SET chapter_idx  = EXCLUDED.chapter_idx,
		    paragraph_id = EXCLUDED.paragraph_id,
		    updated_at   = now()`,
		userID, bookSlug, chapterIdx, paragraphID)
	return err
}

// ListReadingProgress 用户全部书籍进度(最近读的在前)。
func (s *Store) ListReadingProgress(ctx context.Context, userID string) ([]ReadingProgress, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT book_slug, chapter_idx, paragraph_id, updated_at
		FROM reading_progress WHERE user_id = $1
		ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ReadingProgress
	for rows.Next() {
		var p ReadingProgress
		if err := rows.Scan(&p.BookSlug, &p.ChapterIdx, &p.ParagraphID, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// AddBookmark 新增书签;命中唯一键(同段重复)时幂等返回既有书签,不改摘录。
func (s *Store) AddBookmark(ctx context.Context, userID, bookSlug string, chapterIdx int, paragraphID, excerpt string) (*Bookmark, error) {
	var b Bookmark
	// ON CONFLICT DO UPDATE(写回自身列)以便命中既有行时也能 RETURNING;
	// 摘录取既有值,保证「再次加书签」不覆盖首次内容,行为幂等。
	err := s.pool.QueryRow(ctx, `
		INSERT INTO reading_bookmarks (user_id, book_slug, chapter_idx, paragraph_id, excerpt)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, book_slug, chapter_idx, paragraph_id) DO UPDATE
		SET excerpt = reading_bookmarks.excerpt
		RETURNING id, book_slug, chapter_idx, paragraph_id, excerpt, created_at`,
		userID, bookSlug, chapterIdx, paragraphID, excerpt).
		Scan(&b.ID, &b.BookSlug, &b.ChapterIdx, &b.ParagraphID, &b.Excerpt, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// ListBookmarks 用户书签列表;bookSlug 非空时按书过滤。均按创建时间倒序。
func (s *Store) ListBookmarks(ctx context.Context, userID, bookSlug string) ([]Bookmark, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, book_slug, chapter_idx, paragraph_id, excerpt, created_at
		FROM reading_bookmarks
		WHERE user_id = $1 AND ($2 = '' OR book_slug = $2)
		ORDER BY created_at DESC`, userID, bookSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Bookmark
	for rows.Next() {
		var b Bookmark
		if err := rows.Scan(&b.ID, &b.BookSlug, &b.ChapterIdx, &b.ParagraphID, &b.Excerpt, &b.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// DeleteBookmark 删除书签;校验归属:非本人或不存在(含非法 UUID)均返回 ErrNotFound。
func (s *Store) DeleteBookmark(ctx context.Context, userID, id string) error {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM reading_bookmarks WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		// 非法 UUID 文本(22P02)视为不存在,避免越权探测返回 500
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "22P02" {
			return ErrNotFound
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
