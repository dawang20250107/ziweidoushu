package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// ── 命盘档案(数据锁定核心)──────────────────────────────────

// Profile 命盘档案。
type Profile struct {
	ID            string          `json:"id"`
	Label         string          `json:"label"`
	Relation      string          `json:"relation"`
	BirthInput    json.RawMessage `json:"birthInput"`
	ChartSnapshot json.RawMessage `json:"chartSnapshot,omitempty"`
	EngineVersion string          `json:"engineVersion"`
	IsDefault     bool            `json:"isDefault"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

// ErrProfileLimit 免费层档案数量已达上限。
var ErrProfileLimit = errors.New("档案数量已达当前会员层级上限")

// CountProfiles 用户档案数(未删除)。
func (s *Store) CountProfiles(ctx context.Context, userID string) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT count(*) FROM chart_profiles
		WHERE user_id = $1 AND deleted_at IS NULL`, userID).Scan(&n)
	return n, err
}

// CreateProfile 新建档案。isDefault=true 时先取消旧默认(部分唯一索引约束)。
func (s *Store) CreateProfile(ctx context.Context, userID, label, relation string,
	birthInput, snapshot json.RawMessage, engineVersion string, isDefault bool) (*Profile, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.WithoutCancel(ctx)) //nolint:errcheck

	if isDefault {
		if _, err := tx.Exec(ctx, `UPDATE chart_profiles SET is_default = false
			WHERE user_id = $1 AND is_default AND deleted_at IS NULL`, userID); err != nil {
			return nil, err
		}
	}
	var p Profile
	err = tx.QueryRow(ctx, `INSERT INTO chart_profiles
		(user_id, label, relation, birth_input, chart_snapshot, engine_version, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, label, relation, birth_input, engine_version, is_default, created_at, updated_at`,
		userID, label, relation, birthInput, snapshot, engineVersion, isDefault).
		Scan(&p.ID, &p.Label, &p.Relation, &p.BirthInput, &p.EngineVersion, &p.IsDefault, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &p, nil
}

// ListProfiles 档案列表(不含快照,轻量)。
func (s *Store) ListProfiles(ctx context.Context, userID string) ([]Profile, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, label, relation, birth_input, engine_version, is_default, created_at, updated_at
		FROM chart_profiles WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY is_default DESC, created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Profile
	for rows.Next() {
		var p Profile
		if err := rows.Scan(&p.ID, &p.Label, &p.Relation, &p.BirthInput, &p.EngineVersion, &p.IsDefault, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// GetProfile 单个档案(含命盘快照)。
func (s *Store) GetProfile(ctx context.Context, userID, profileID string) (*Profile, error) {
	var p Profile
	err := s.pool.QueryRow(ctx, `SELECT id, label, relation, birth_input, chart_snapshot, engine_version, is_default, created_at, updated_at
		FROM chart_profiles WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, profileID, userID).
		Scan(&p.ID, &p.Label, &p.Relation, &p.BirthInput, &p.ChartSnapshot, &p.EngineVersion, &p.IsDefault, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// DeleteProfile 软删除。
func (s *Store) DeleteProfile(ctx context.Context, userID, profileID string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE chart_profiles SET deleted_at = now(), is_default = false
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, profileID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// SetDefaultProfile 设为默认档案。
func (s *Store) SetDefaultProfile(ctx context.Context, userID, profileID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.WithoutCancel(ctx)) //nolint:errcheck
	if _, err := tx.Exec(ctx, `UPDATE chart_profiles SET is_default = false
		WHERE user_id = $1 AND is_default AND deleted_at IS NULL`, userID); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `UPDATE chart_profiles SET is_default = true
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, profileID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return tx.Commit(ctx)
}
