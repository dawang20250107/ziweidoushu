package auth

import (
	"strings"
	"testing"
	"time"
)

const testSecret = "0123456789abcdef0123456789abcdef" // 32 字节

func TestJWTSignVerify(t *testing.T) {
	j, err := NewJWT(testSecret, "", 30*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	token, err := j.Sign("user-1", "pro", 3, now)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := j.Verify(token, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("校验失败: %v", err)
	}
	if claims.Sub != "user-1" || claims.Tier != "pro" || claims.SV != 3 {
		t.Fatalf("载荷不符: %+v", claims)
	}
}

func TestJWTExpiry(t *testing.T) {
	j, _ := NewJWT(testSecret, "", time.Minute)
	now := time.Now()
	token, _ := j.Sign("u", "free", 1, now)
	if _, err := j.Verify(token, now.Add(2*time.Minute)); err != ErrTokenExpired {
		t.Fatalf("过期令牌应报 ErrTokenExpired, got %v", err)
	}
}

func TestJWTTamper(t *testing.T) {
	j, _ := NewJWT(testSecret, "", time.Minute)
	token, _ := j.Sign("u", "free", 1, time.Now())
	parts := strings.Split(token, ".")

	// 篡改载荷
	tampered := parts[0] + "." + parts[1][:len(parts[1])-2] + "xx" + "." + parts[2]
	if _, err := j.Verify(tampered, time.Now()); err == nil {
		t.Fatal("篡改载荷应校验失败")
	}
	// 篡改签名
	tampered = parts[0] + "." + parts[1] + "." + parts[2][:len(parts[2])-2] + "xx"
	if _, err := j.Verify(tampered, time.Now()); err == nil {
		t.Fatal("篡改签名应校验失败")
	}
	// 错误密钥
	j2, _ := NewJWT("another-secret-another-secret-32", "", time.Minute)
	if _, err := j2.Verify(token, time.Now()); err == nil {
		t.Fatal("异密钥应校验失败")
	}
	// alg 替换(none 攻击):header 非固定值即拒绝
	fake := "eyJhbGciOiJub25lIn0" + "." + parts[1] + "." + parts[2]
	if _, err := j.Verify(fake, time.Now()); err == nil {
		t.Fatal("alg 替换应校验失败")
	}
}

func TestJWTKeyRotation(t *testing.T) {
	old, _ := NewJWT(testSecret, "", time.Minute)
	token, _ := old.Sign("u", "free", 1, time.Now())

	rotated, _ := NewJWT("new-secret-new-secret-new-secr32", testSecret, time.Minute)
	if _, err := rotated.Verify(token, time.Now()); err != nil {
		t.Fatalf("轮换期应兼容旧密钥: %v", err)
	}
	dropped, _ := NewJWT("new-secret-new-secret-new-secr32", "", time.Minute)
	if _, err := dropped.Verify(token, time.Now()); err == nil {
		t.Fatal("旧密钥移除后应拒绝")
	}
}

func TestShortSecretRejected(t *testing.T) {
	if _, err := NewJWT("short", "", time.Minute); err == nil {
		t.Fatal("短密钥应被拒绝")
	}
}

func TestRandomCode(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		code, err := RandomCode(6)
		if err != nil || len(code) != 6 {
			t.Fatalf("验证码生成异常: %q %v", code, err)
		}
		for _, r := range code {
			if r < '0' || r > '9' {
				t.Fatalf("验证码含非数字: %q", code)
			}
		}
		seen[code] = true
	}
	if len(seen) < 40 {
		t.Fatalf("验证码随机性可疑: %d/50 唯一", len(seen))
	}
}

func TestValidatePhone(t *testing.T) {
	for _, ok := range []string{"13800138000", "19912345678"} {
		if err := ValidatePhone(ok); err != nil {
			t.Errorf("%s 应合法: %v", ok, err)
		}
	}
	for _, bad := range []string{"", "2380013800", "138001380", "1380013800a", "138001380001"} {
		if err := ValidatePhone(bad); err == nil {
			t.Errorf("%q 应非法", bad)
		}
	}
}

func TestRefreshTokenHash(t *testing.T) {
	tok, hash, err := NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	if HashToken(tok) != hash {
		t.Fatal("哈希不一致")
	}
	tok2, _, _ := NewRefreshToken()
	if tok == tok2 {
		t.Fatal("刷新令牌不应重复")
	}
}
