// Package auth 鉴权核心:JWT(HS256,纯标准库)、刷新令牌旋转、会话版本。
// 设计契约见 docs/architecture/auth.md。
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Claims 访问令牌载荷。
type Claims struct {
	Sub  string `json:"sub"`  // user id
	Tier string `json:"tier"` // free/pro/master(缓存,权益判定以数据库为准)
	SV   int    `json:"sv"`   // 会话版本
	IAT  int64  `json:"iat"`
	EXP  int64  `json:"exp"`
	JTI  string `json:"jti"`
}

var (
	// ErrTokenInvalid 签名/格式非法。
	ErrTokenInvalid = errors.New("令牌无效")
	// ErrTokenExpired 已过期。
	ErrTokenExpired = errors.New("令牌已过期")
)

// JWT HS256 签发与校验器,支持双密钥轮换(新签发用主密钥,校验兼容旧密钥)。
type JWT struct {
	secret     []byte
	prevSecret []byte
	ttl        time.Duration
}

// NewJWT 构建签发器。prevSecret 可为空。
func NewJWT(secret, prevSecret string, ttl time.Duration) (*JWT, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT_SECRET 长度需 ≥ 32 字节")
	}
	j := &JWT{secret: []byte(secret), ttl: ttl}
	if prevSecret != "" {
		j.prevSecret = []byte(prevSecret)
	}
	return j, nil
}

var b64 = base64.RawURLEncoding

const jwtHeader = `{"alg":"HS256","typ":"JWT"}`

// Sign 签发访问令牌。
func (j *JWT) Sign(userID, tier string, sessionVer int, now time.Time) (string, error) {
	jti, err := randomHex(8)
	if err != nil {
		return "", err
	}
	claims := Claims{
		Sub: userID, Tier: tier, SV: sessionVer,
		IAT: now.Unix(), EXP: now.Add(j.ttl).Unix(), JTI: jti,
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	signingInput := b64.EncodeToString([]byte(jwtHeader)) + "." + b64.EncodeToString(payload)
	sig := hmacSign([]byte(signingInput), j.secret)
	return signingInput + "." + b64.EncodeToString(sig), nil
}

// Verify 校验令牌并返回载荷(不校验会话版本——由调用层比对数据库)。
func (j *JWT) Verify(token string, now time.Time) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrTokenInvalid
	}
	signingInput := parts[0] + "." + parts[1]
	sig, err := b64.DecodeString(parts[2])
	if err != nil {
		return nil, ErrTokenInvalid
	}
	if !hmacValid([]byte(signingInput), sig, j.secret) {
		if j.prevSecret == nil || !hmacValid([]byte(signingInput), sig, j.prevSecret) {
			return nil, ErrTokenInvalid
		}
	}
	// 校验 header(防算法替换)
	headerRaw, err := b64.DecodeString(parts[0])
	if err != nil || string(headerRaw) != jwtHeader {
		return nil, ErrTokenInvalid
	}
	payloadRaw, err := b64.DecodeString(parts[1])
	if err != nil {
		return nil, ErrTokenInvalid
	}
	var claims Claims
	if err := json.Unmarshal(payloadRaw, &claims); err != nil {
		return nil, ErrTokenInvalid
	}
	if claims.Sub == "" {
		return nil, ErrTokenInvalid
	}
	if now.Unix() >= claims.EXP {
		return nil, ErrTokenExpired
	}
	return &claims, nil
}

func hmacSign(data, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

func hmacValid(data, sig, key []byte) bool {
	expected := hmacSign(data, key)
	return subtle.ConstantTimeCompare(expected, sig) == 1
}

// NewRefreshToken 生成 256bit 随机刷新令牌(返回明文与 SHA-256 哈希)。
func NewRefreshToken() (token, hash string, err error) {
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return "", "", err
	}
	token = b64.EncodeToString(raw)
	return token, HashToken(token), nil
}

// HashToken 刷新令牌哈希(库中不存明文)。
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomHex(n int) (string, error) {
	raw := make([]byte, n)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

// RandomCode 生成 n 位数字验证码。
func RandomCode(n int) (string, error) {
	const digits = "0123456789"
	raw := make([]byte, n)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	var sb strings.Builder
	for _, ch := range raw {
		sb.WriteByte(digits[int(ch)%10])
	}
	return sb.String(), nil
}

// ValidatePhone 中国大陆手机号(11 位,1 开头)。
func ValidatePhone(phone string) error {
	if len(phone) != 11 || phone[0] != '1' {
		return fmt.Errorf("手机号格式不正确")
	}
	for _, r := range phone {
		if r < '0' || r > '9' {
			return fmt.Errorf("手机号格式不正确")
		}
	}
	return nil
}
