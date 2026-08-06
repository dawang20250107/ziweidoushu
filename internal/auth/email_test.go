package auth

import "testing"

// TestValidateEmail 邮箱归一与校验。
func TestValidateEmail(t *testing.T) {
	e, err := ValidateEmail("  User@Example.COM ")
	if err != nil || e != "user@example.com" {
		t.Fatalf("应归一小写: %q %v", e, err)
	}
	for _, bad := range []string{"", "a", "a@b", "a b@c.com", "@x.com", "a@"} {
		if _, err := ValidateEmail(bad); err == nil {
			t.Errorf("ValidateEmail(%q) 应拒绝", bad)
		}
	}
}

// TestValidatePassword 密码策略:8-72 位且含字母与数字。
func TestValidatePassword(t *testing.T) {
	if err := ValidatePassword("abc12345"); err != nil {
		t.Errorf("合规密码被拒: %v", err)
	}
	for _, bad := range []string{"", "short1", "onlyletters", "12345678", "abcdefgh"} {
		if err := ValidatePassword(bad); err == nil {
			t.Errorf("ValidatePassword(%q) 应拒绝", bad)
		}
	}
}

// TestDeviceHash 设备指纹稳定且归一空白。
func TestDeviceHash(t *testing.T) {
	if DeviceHash("Mozilla/5.0 X") != DeviceHash("  Mozilla/5.0 X  ") {
		t.Error("首尾空白应归一")
	}
	if DeviceHash("a") == DeviceHash("b") {
		t.Error("不同 UA 不应同哈希")
	}
}

// TestConstantTimeEqual 管理令牌恒时比较语义。
func TestConstantTimeEqual(t *testing.T) {
	if !ConstantTimeEqual("tok", "tok") || ConstantTimeEqual("tok", "tak") || ConstantTimeEqual("tok", "tok2") {
		t.Error("恒时比较语义错误")
	}
}
