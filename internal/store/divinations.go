package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ── 卦档(占卜记录存档)──────────────────────────────────────
// 登录用户起卦自动存档;AI 解卦成功后回填 reading。
// 每用户保留最近 divinationRecordCap 条,插入时截断旧档。

const divinationRecordCap = 200

// DivinationRecord 一次占卜的存档。
type DivinationRecord struct {
	ID              string          `json:"id"`
	Kind            string          `json:"kind"` // meihua | liuyao
	Question        string          `json:"question"`
	Summary         string          `json:"summary"` // 卦名摘要,如「地天泰 → 山风蛊」
	Payload         json.RawMessage `json:"payload,omitempty"`
	Reading         string          `json:"reading,omitempty"`
	ReadingProvider string          `json:"readingProvider,omitempty"`
	HasReading      bool            `json:"hasReading"`
	CastAt          time.Time       `json:"castAt"`
	CreatedAt       time.Time       `json:"createdAt"`
}

// SaveDivination 存档一次起卦,返回记录 ID;超出上限时截断最旧记录。
func (s *Store) SaveDivination(ctx context.Context, userID, kind, question, summary string,
	payload json.RawMessage, castAt time.Time) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `INSERT INTO divination_records
		(user_id, kind, question, summary, payload, cast_at)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		userID, kind, question, summary, payload, castAt).Scan(&id)
	if err != nil {
		return "", err
	}
	_, err = s.pool.Exec(ctx, `DELETE FROM divination_records WHERE user_id = $1 AND id IN (
		SELECT id FROM divination_records WHERE user_id = $1
		ORDER BY created_at DESC OFFSET $2)`, userID, divinationRecordCap)
	return id, err
}

// AttachDivinationReading 回填 AI 解卦文本(仅本人记录)。
func (s *Store) AttachDivinationReading(ctx context.Context, userID, recordID, reading, provider string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE divination_records
		SET reading = $3, reading_provider = $4
		WHERE id = $1 AND user_id = $2`, recordID, userID, reading, provider)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListDivinations 卦档列表(轻量:不含卦象与解卦全文)。kind 为空表示全部占法。
func (s *Store) ListDivinations(ctx context.Context, userID, kind string, limit, offset int) ([]DivinationRecord, int, error) {
	where := `WHERE user_id = $1`
	args := []any{userID}
	if kind != "" {
		where += ` AND kind = $2`
		args = append(args, kind)
	}
	var total int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM divination_records `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	listSQL := fmt.Sprintf(`SELECT id, kind, question, summary,
		reading IS NOT NULL AND reading <> '', cast_at, created_at
		FROM divination_records %s
		ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2)
	rows, err := s.pool.Query(ctx, listSQL, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []DivinationRecord{}
	for rows.Next() {
		var r DivinationRecord
		if err := rows.Scan(&r.ID, &r.Kind, &r.Question, &r.Summary, &r.HasReading, &r.CastAt, &r.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, r)
	}
	return out, total, rows.Err()
}

// GetDivination 单条卦档(含卦象与解卦全文)。
func (s *Store) GetDivination(ctx context.Context, userID, recordID string) (*DivinationRecord, error) {
	var r DivinationRecord
	var reading, provider *string
	err := s.pool.QueryRow(ctx, `SELECT id, kind, question, summary, payload,
		reading, reading_provider, cast_at, created_at
		FROM divination_records WHERE id = $1 AND user_id = $2`, recordID, userID).
		Scan(&r.ID, &r.Kind, &r.Question, &r.Summary, &r.Payload, &reading, &provider, &r.CastAt, &r.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if reading != nil {
		r.Reading = *reading
		r.HasReading = r.Reading != ""
	}
	if provider != nil {
		r.ReadingProvider = *provider
	}
	return &r, nil
}

// DeleteDivination 删除卦档(仅本人)。
func (s *Store) DeleteDivination(ctx context.Context, userID, recordID string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM divination_records
		WHERE id = $1 AND user_id = $2`, recordID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
