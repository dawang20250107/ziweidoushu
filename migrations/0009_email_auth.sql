-- 0009 内测鉴权:邮箱注册登录(密码)+ 邮箱验证码 + 邀请码 + 登录设备
--
-- 口径:
--   1. 邮箱为内测期主凭证(user_identities.provider='email',identifier=邮箱小写);
--      手机号通道保留,公测接短信服务商后并存。
--   2. 密码只在注册(邮箱验证码验证后)自设;登录用密码;忘记/修改走邮箱验证码。
--   3. 邀请码内测闸门:注册须携带有效邀请码(服务端 AUTH_INVITE_REQUIRED 控制)。
--   4. login_devices 环境检测底账:陌生设备登录可要求验证码升级(strict 模式)。

ALTER TABLE user_identities DROP CONSTRAINT user_identities_provider_check;
ALTER TABLE user_identities ADD CONSTRAINT user_identities_provider_check
    CHECK (provider IN ('phone', 'email', 'wechat', 'apple', 'google'));

ALTER TABLE users ADD COLUMN password_hash TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN password_set_at TIMESTAMPTZ;

-- 邮箱验证码(与 sms_codes 同构;purpose: register / reset / login)
CREATE TABLE email_codes (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email         TEXT        NOT NULL,
    code_hash     TEXT        NOT NULL,
    purpose       TEXT        NOT NULL CHECK (purpose IN ('register', 'reset', 'login')),
    expires_at    TIMESTAMPTZ NOT NULL,
    consumed_at   TIMESTAMPTZ,
    client_ip     INET,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_email_codes_lookup ON email_codes(email, created_at DESC);

-- 邀请码(内测闸门):一码可多用(max_uses),过期即废
CREATE TABLE invite_codes (
    code        TEXT        PRIMARY KEY,
    max_uses    INT         NOT NULL DEFAULT 1,
    used_count  INT         NOT NULL DEFAULT 0,
    note        TEXT        NOT NULL DEFAULT '',
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE invite_uses (
    code      TEXT        NOT NULL REFERENCES invite_codes(code) ON DELETE CASCADE,
    user_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    used_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (code, user_id)
);

-- 登录设备底账(环境检测):device_hash = sha256(UA 归一)
CREATE TABLE login_devices (
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_hash TEXT        NOT NULL,
    first_seen  TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, device_hash)
);
