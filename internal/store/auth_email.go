package store

import (
	"context"
	"crypto/rand"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ── 邮箱验证码(与短信验证码同构)─────────────────────────────

// InsertEmailCode 落库一条邮箱验证码(哈希存储,审计 + 校验)。
func (s *Store) InsertEmailCode(ctx context.Context, email, codeHash, purpose, clientIP string, ttl time.Duration) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO email_codes (email, code_hash, purpose, expires_at, client_ip)
		VALUES ($1, $2, $3, now() + $4::interval, NULLIF($5,'')::inet)`, email, codeHash, purpose, ttl, clientIP)
	return err
}

// CountRecentEmail 频控计数:email 维度与 IP 维度。
func (s *Store) CountRecentEmail(ctx context.Context, email, clientIP string, window time.Duration) (byEmail, byIP int, err error) {
	err = s.pool.QueryRow(ctx, `SELECT
		count(*) FILTER (WHERE email = $1),
		count(*) FILTER (WHERE client_ip = NULLIF($2,'')::inet)
		FROM email_codes WHERE created_at > now() - $3::interval`, email, clientIP, window).
		Scan(&byEmail, &byIP)
	return
}

// LastEmailAt 该邮箱最近一次发送时间(60s 间隔限制)。
func (s *Store) LastEmailAt(ctx context.Context, email string) (time.Time, error) {
	var t time.Time
	err := s.pool.QueryRow(ctx, `SELECT coalesce(max(created_at), 'epoch'::timestamptz)
		FROM email_codes WHERE email = $1`, email).Scan(&t)
	return t, err
}

// ConsumeEmailCode 校验并消费验证码:未过期、未消费、哈希与用途匹配。
func (s *Store) ConsumeEmailCode(ctx context.Context, email, codeHash, purpose string) (bool, error) {
	tag, err := s.pool.Exec(ctx, `UPDATE email_codes SET consumed_at = now()
		WHERE id = (
			SELECT id FROM email_codes
			WHERE email = $1 AND code_hash = $2 AND purpose = $3
			  AND consumed_at IS NULL AND expires_at > now()
			ORDER BY created_at DESC LIMIT 1
		)`, email, codeHash, purpose)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

// FailEmailAttempt 校验失败防爆破:10 分钟内该邮箱发码超 5 条时作废未消费码。
func (s *Store) FailEmailAttempt(ctx context.Context, email string) error {
	_, err := s.pool.Exec(ctx, `UPDATE email_codes SET consumed_at = now()
		WHERE email = $1 AND consumed_at IS NULL AND expires_at > now()
		  AND (SELECT count(*) FROM email_codes c2
		       WHERE c2.email = $1 AND c2.created_at > now() - interval '10 minutes') > 5`, email)
	return err
}

// ── 邮箱用户 ─────────────────────────────────────────────────

// ErrEmailTaken 邮箱已注册。
var ErrEmailTaken = errors.New("该邮箱已注册")

// FindUserByEmail 邮箱查用户(附密码哈希供登录校验)。
func (s *Store) FindUserByEmail(ctx context.Context, email string) (*User, string, error) {
	var u User
	var hash string
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.nickname, u.avatar_url, u.tier, u.tier_expires_at, u.status, u.session_ver, u.password_hash
		FROM user_identities i JOIN users u ON u.id = i.user_id
		WHERE i.provider = 'email' AND i.identifier = $1`, email).
		Scan(&u.ID, &u.Nickname, &u.AvatarURL, &u.Tier, &u.TierExpiresAt, &u.Status, &u.SessionVer, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", ErrNotFound
	}
	if err != nil {
		return nil, "", err
	}
	return &u, hash, nil
}

// CreateUserByEmail 邮箱注册建户(验证码校验后调用):写身份与密码哈希。
// 邮箱已存在返回 ErrEmailTaken。
func (s *Store) CreateUserByEmail(ctx context.Context, email, passwordHash, nickname string) (*User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.WithoutCancel(ctx)) //nolint:errcheck

	var u User
	err = tx.QueryRow(ctx, `INSERT INTO users (nickname, password_hash, password_set_at) VALUES ($1, $2, now())
		RETURNING id, nickname, avatar_url, tier, tier_expires_at, status, session_ver`, nickname, passwordHash).
		Scan(&u.ID, &u.Nickname, &u.AvatarURL, &u.Tier, &u.TierExpiresAt, &u.Status, &u.SessionVer)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO user_identities (user_id, provider, identifier)
		VALUES ($1, 'email', $2)`, u.ID, email); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &u, nil
}

// SetPassword 更新密码哈希(重置/修改共用);调用层负责随后全端下线。
func (s *Store) SetPassword(ctx context.Context, userID, passwordHash string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE users SET password_hash = $2, password_set_at = now(), updated_at = now()
		WHERE id = $1`, userID, passwordHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrNotFound
	}
	return nil
}

// PasswordHash 取用户密码哈希(改密时校验旧密码)。
func (s *Store) PasswordHash(ctx context.Context, userID string) (string, error) {
	var h string
	err := s.pool.QueryRow(ctx, `SELECT password_hash FROM users WHERE id = $1`, userID).Scan(&h)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return h, err
}

// ── 邀请码(内测闸门)────────────────────────────────────────

// ErrInviteInvalid 邀请码无效/已用尽/过期。
var ErrInviteInvalid = errors.New("邀请码无效或已用尽")

const inviteAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 去易混字符

// CreateInvites 铸造 n 个邀请码(每码 maxUses 次,expires 零值不过期)。
func (s *Store) CreateInvites(ctx context.Context, n, maxUses int, note string, expires *time.Time) ([]string, error) {
	codes := make([]string, 0, n)
	for i := 0; i < n; i++ {
		buf := make([]byte, 8)
		if _, err := rand.Read(buf); err != nil {
			return nil, err
		}
		for j := range buf {
			buf[j] = inviteAlphabet[int(buf[j])%len(inviteAlphabet)]
		}
		code := "ZW-" + string(buf)
		if _, err := s.pool.Exec(ctx, `INSERT INTO invite_codes (code, max_uses, note, expires_at)
			VALUES ($1, $2, $3, $4) ON CONFLICT (code) DO NOTHING`, code, maxUses, note, expires); err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	return codes, nil
}

// ConsumeInvite 原子消费一次邀请码并登记使用人;无效/用尽/过期返回 ErrInviteInvalid。
func (s *Store) ConsumeInvite(ctx context.Context, code, userID string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE invite_codes SET used_count = used_count + 1
		WHERE code = $1 AND used_count < max_uses
		  AND (expires_at IS NULL OR expires_at > now())`, code)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrInviteInvalid
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO invite_uses (code, user_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING`, code, userID)
	return err
}

// CheckInvite 预检邀请码可用性(注册第一步即时反馈;消费仍以 ConsumeInvite 原子为准)。
func (s *Store) CheckInvite(ctx context.Context, code string) error {
	var ok bool
	err := s.pool.QueryRow(ctx, `SELECT used_count < max_uses AND (expires_at IS NULL OR expires_at > now())
		FROM invite_codes WHERE code = $1`, code).Scan(&ok)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && !ok) {
		return ErrInviteInvalid
	}
	return err
}

// ── 登录设备底账(环境检测)──────────────────────────────────

// TouchDevice 登记/刷新设备指纹,返回此前是否见过(known)。
func (s *Store) TouchDevice(ctx context.Context, userID, deviceHash string) (known bool, err error) {
	err = s.pool.QueryRow(ctx, `INSERT INTO login_devices (user_id, device_hash)
		VALUES ($1, $2)
		ON CONFLICT (user_id, device_hash) DO UPDATE SET last_seen = now()
		RETURNING (xmax <> 0)`, userID, deviceHash).Scan(&known)
	return
}

// KnownDevice 只读探测设备是否见过(登录前置判断,不落新记录)。
func (s *Store) KnownDevice(ctx context.Context, userID, deviceHash string) (bool, error) {
	var one int
	err := s.pool.QueryRow(ctx, `SELECT 1 FROM login_devices WHERE user_id = $1 AND device_hash = $2`,
		userID, deviceHash).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}
