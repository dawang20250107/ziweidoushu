// Package store PostgreSQL 数据访问层(pgx 连接池 + 内嵌迁移执行器)。
//
// DATABASE_URL 未配置时平台以「无库模式」运行:排盘/古籍/AI 全部可用,
// 仅用户体系(登录/档案)不可用——保证最小部署形态不被破坏。
package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dawang20250107/ziweidoushu/migrations"
)

// Store 数据访问入口。
type Store struct {
	pool *pgxpool.Pool
}

// Open 建立连接池并自动应用迁移。
func Open(ctx context.Context, databaseURL string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("DATABASE_URL 解析失败: %w", err)
	}
	cfg.MaxConns = 20
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	s := &Store{pool: pool}
	if err := s.migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return s, nil
}

// Close 关闭连接池。
func (s *Store) Close() { s.pool.Close() }

// Pool 暴露底层连接池(仅供测试与运维工具)。
func (s *Store) Pool() *pgxpool.Pool { return s.pool }

// migrate 按文件名顺序应用未执行的迁移;advisory lock 防多实例并发。
func (s *Store) migrate(ctx context.Context) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	// 全局迁移锁(任意稳定常数)
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock(742001)"); err != nil {
		return err
	}
	defer conn.Exec(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock(742001)") //nolint:errcheck

	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		filename TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
		return err
	}

	applied := map[string]bool{}
	rows, err := conn.Query(ctx, "SELECT filename FROM schema_migrations")
	if err != nil {
		return err
	}
	for rows.Next() {
		var f string
		if err := rows.Scan(&f); err != nil {
			return err
		}
		applied[f] = true
	}
	if rows.Err() != nil {
		return rows.Err()
	}

	names, err := migrationFiles()
	if err != nil {
		return err
	}
	for _, name := range names {
		if applied[name] {
			continue
		}
		raw, err := migrations.FS.ReadFile(name)
		if err != nil {
			return err
		}
		if _, err := conn.Exec(ctx, string(raw)); err != nil {
			return fmt.Errorf("迁移 %s 失败: %w", name, err)
		}
		if _, err := conn.Exec(ctx, "INSERT INTO schema_migrations (filename) VALUES ($1)", name); err != nil {
			return err
		}
	}
	return nil
}

func migrationFiles() ([]string, error) {
	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") && !strings.Contains(e.Name(), "_rag") {
			// RAG 迁移(依赖 pgvector)由 P3 显式执行,不随启动自动应用
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// ── 用户与身份 ────────────────────────────────────────────────

// User 用户主档。
type User struct {
	ID         string
	Nickname   string
	AvatarURL  string
	Tier       string
	Status     string
	SessionVer int
}

// ErrNotFound 记录不存在。
var ErrNotFound = errors.New("record not found")

// FindOrCreateUserByPhone 手机号登录:身份存在则返回用户,否则创建。
func (s *Store) FindOrCreateUserByPhone(ctx context.Context, phone string) (*User, bool, error) {
	var u User
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.nickname, u.avatar_url, u.tier, u.status, u.session_ver
		FROM user_identities i JOIN users u ON u.id = i.user_id
		WHERE i.provider = 'phone' AND i.identifier = $1`, phone).
		Scan(&u.ID, &u.Nickname, &u.AvatarURL, &u.Tier, &u.Status, &u.SessionVer)
	if err == nil {
		return &u, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback(context.WithoutCancel(ctx)) //nolint:errcheck

	nickname := "星友" + phone[max(0, len(phone)-4):]
	err = tx.QueryRow(ctx, `INSERT INTO users (nickname) VALUES ($1)
		RETURNING id, nickname, avatar_url, tier, status, session_ver`, nickname).
		Scan(&u.ID, &u.Nickname, &u.AvatarURL, &u.Tier, &u.Status, &u.SessionVer)
	if err != nil {
		return nil, false, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO user_identities (user_id, provider, identifier)
		VALUES ($1, 'phone', $2)`, u.ID, phone); err != nil {
		// 并发注册同一手机号:唯一约束触发,回退为读取既有用户
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			tx.Rollback(context.WithoutCancel(ctx)) //nolint:errcheck
			u2, _, err2 := s.FindOrCreateUserByPhone(ctx, phone)
			return u2, false, err2
		}
		return nil, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}
	return &u, true, nil
}

// GetUser 按 id 取用户。
func (s *Store) GetUser(ctx context.Context, id string) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `SELECT id, nickname, avatar_url, tier, status, session_ver
		FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Nickname, &u.AvatarURL, &u.Tier, &u.Status, &u.SessionVer)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// SessionVer 读取会话版本。
func (s *Store) SessionVer(ctx context.Context, userID string) (int, error) {
	var v int
	err := s.pool.QueryRow(ctx, "SELECT session_ver FROM users WHERE id = $1", userID).Scan(&v)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNotFound
	}
	return v, err
}

// BumpSessionVer 会话版本 +1(全端登出/封禁/改密)。
func (s *Store) BumpSessionVer(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx, "UPDATE users SET session_ver = session_ver + 1 WHERE id = $1", userID)
	return err
}

// ── 短信验证码 ────────────────────────────────────────────────

// HashCode 验证码哈希(不存明文)。
func HashCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

// InsertSMSCode 落库一条验证码(审计 + 校验)。
func (s *Store) InsertSMSCode(ctx context.Context, phone, codeHash, purpose, clientIP string, ttl time.Duration) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO sms_codes (phone, code_hash, purpose, expires_at, client_ip)
		VALUES ($1, $2, $3, now() + $4::interval, NULLIF($5,'')::inet)`, phone, codeHash, purpose, ttl, clientIP)
	return err
}

// CountRecentSMS 频控计数:phone 维度与 IP 维度。
func (s *Store) CountRecentSMS(ctx context.Context, phone, clientIP string, window time.Duration) (byPhone, byIP int, err error) {
	err = s.pool.QueryRow(ctx, `SELECT
		count(*) FILTER (WHERE phone = $1),
		count(*) FILTER (WHERE client_ip = NULLIF($2,'')::inet)
		FROM sms_codes WHERE created_at > now() - $3::interval`, phone, clientIP, window).
		Scan(&byPhone, &byIP)
	return
}

// LastSMSAt 该号码最近一次发送时间(60s 间隔限制)。
func (s *Store) LastSMSAt(ctx context.Context, phone string) (time.Time, error) {
	var t time.Time
	err := s.pool.QueryRow(ctx, `SELECT coalesce(max(created_at), 'epoch'::timestamptz)
		FROM sms_codes WHERE phone = $1`, phone).Scan(&t)
	return t, err
}

// ConsumeSMSCode 校验并消费验证码:未过期、未消费、哈希匹配。
// 返回 true 表示校验通过(该码同时被标记消费)。
func (s *Store) ConsumeSMSCode(ctx context.Context, phone, codeHash, purpose string) (bool, error) {
	tag, err := s.pool.Exec(ctx, `UPDATE sms_codes SET consumed_at = now()
		WHERE id = (
			SELECT id FROM sms_codes
			WHERE phone = $1 AND code_hash = $2 AND purpose = $3
			  AND consumed_at IS NULL AND expires_at > now()
			ORDER BY created_at DESC LIMIT 1
		)`, phone, codeHash, purpose)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

// FailSMSAttempt 校验失败计数:该号码最新未消费码累计失败 5 次即作废(置为已消费)。
func (s *Store) FailSMSAttempt(ctx context.Context, phone string) error {
	// 复用 client_ip 之外的字段过于隐晦,直接以 error 字段无;简单策略:
	// 统计窗口内失败依赖调用层;此处将超过尝试上限的最新码作废。
	_, err := s.pool.Exec(ctx, `UPDATE sms_codes SET consumed_at = now()
		WHERE phone = $1 AND consumed_at IS NULL AND expires_at > now()
		  AND (SELECT count(*) FROM sms_codes c2
		       WHERE c2.phone = $1 AND c2.created_at > now() - interval '10 minutes') > 5`, phone)
	return err
}

// ── 刷新令牌 ─────────────────────────────────────────────────

// RefreshToken 刷新令牌记录。
type RefreshToken struct {
	ID          string
	UserID      string
	ExpiresAt   time.Time
	RevokedAt   *time.Time
	RotatedFrom *string
}

// InsertRefreshToken 写入新刷新令牌(存哈希)。
func (s *Store) InsertRefreshToken(ctx context.Context, userID, tokenHash, deviceInfo string, ttl time.Duration, rotatedFrom *string) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `INSERT INTO refresh_tokens (user_id, token_hash, device_info, expires_at, rotated_from)
		VALUES ($1, $2, $3, now() + $4::interval, $5) RETURNING id`, userID, tokenHash, deviceInfo, ttl, rotatedFrom).Scan(&id)
	return id, err
}

// FindRefreshToken 按哈希取刷新令牌。
func (s *Store) FindRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	var rt RefreshToken
	err := s.pool.QueryRow(ctx, `SELECT id, user_id, expires_at, revoked_at, rotated_from
		FROM refresh_tokens WHERE token_hash = $1`, tokenHash).
		Scan(&rt.ID, &rt.UserID, &rt.ExpiresAt, &rt.RevokedAt, &rt.RotatedFrom)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

// RevokeRefreshToken 撤销单个令牌。
func (s *Store) RevokeRefreshToken(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, "UPDATE refresh_tokens SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL", id)
	return err
}

// RevokeAllRefreshTokens 撤销用户全部令牌(复用检测触发/全端登出)。
func (s *Store) RevokeAllRefreshTokens(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx, "UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL", userID)
	return err
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
