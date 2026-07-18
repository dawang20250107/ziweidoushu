# 鉴权架构设计(P0 定稿)

> 口径:先国内——手机号+短信验证码为主,微信扫码登录为辅;
> UnionID 打通 Web 与(后期)小程序身份;海外渠道(Apple/Google)以 provider 扩展位预留。

## 1. 身份模型

```
users(1) ──< user_identities(N)     provider ∈ {phone, wechat, apple, google}
```

- **手机号**:`identifier = E.164 手机号`,是国内主身份;
- **微信**:`identifier = unionid`(开放平台维度,Web 扫码与小程序同值),
  `credentials` 存 `{openid_web, openid_mp}`——同一 unionid 在两端登录归并为同一 user;
- 一个 user 可绑多渠道;任一渠道登录得到同一账号。首次登录自动注册。

## 2. 令牌模型(JWT 双 token + 会话版本)

| 令牌 | 形态 | 有效期 | 存储 |
|------|------|--------|------|
| Access Token | JWT(HS256,服务端密钥) | 30 分钟 | 客户端内存 |
| Refresh Token | 随机 256bit,服务端存 SHA-256 | 30 天,旋转式 | 客户端安全存储 / HttpOnly Cookie(Web) |

**JWT claims**:`sub`(user_id)、`tier`、`sv`(会话版本)、`iat/exp`、`jti`。

**会话版本(sv)机制**:Redis 存 `session_ver:{user_id} = N`。
校验时比对 JWT 内 `sv` 与 Redis 值,不等即拒——实现**秒级全端登出/封禁**
(改密、封号、退出所有设备时 `INCR`),无需维护 JWT 黑名单。
Redis 不可用时的降级策略:仅信任 JWT 签名与有效期(可用性优先,窗口 ≤30 分钟)。

**Refresh 旋转与复用检测**:每次刷新签发新 refresh token 并撤销旧值,
`rotated_from` 记录旋转链;若检测到已撤销 token 被再次使用(疑似被盗),
撤销整条链并 `INCR` 会话版本,强制全端重新登录。

## 3. 登录流程

### 3.1 手机号 + 验证码

```
POST /api/v1/auth/sms/send     {phone}
POST /api/v1/auth/sms/verify   {phone, code}  → {access, refresh, user}
```

- 验证码 6 位数字,有效期 5 分钟,单码最多试错 5 次(Redis 计数);
- 发送频控(Redis):同号 60s/次、1h≤5 次、24h≤10 次;同 IP 24h≤20 次;
- 验证码哈希落库(`sms_codes`,审计),明文只进短信通道;
- 防遍历:验证失败统一返回「验证码错误或已过期」。

### 3.2 微信扫码(Web)

```
GET  /api/v1/auth/wechat/url        → {authUrl, state}
GET  /api/v1/auth/wechat/callback   ?code=&state=   → 重定向携带一次性票据
POST /api/v1/auth/wechat/exchange   {ticket}        → {access, refresh, user}
```

- `state` 为一次性随机值(Redis,5 分钟),防 CSRF;
- 后端用 code 换 `access_token + openid + unionid`,按 unionid 归并账号;
- 一次性票据(60s)换 token,避免 token 经 URL 传递。

### 3.3 账号绑定

已登录用户可绑定另一渠道(`purpose=bind` 的验证码 / 扫码);
冲突(目标渠道已属他人)返回明确错误,不做自动合并(合并属高危操作,人工通道)。

## 4. 中间件与权限

```
withAuth(required bool)  → ctx 注入 {userID, tier, sv}
```

- 排盘/古籍读取:匿名可用(获客面),登录后才有档案/对话/进度;
- AI 对话:匿名给极小额度(Redis 按 IP),登录按 tier 配额(`ai_usage_daily`);
- 管理接口:独立 `ADMIN_TOKEN`(后续升级为 admin 角色 + 双因素)。

## 5. 威胁模型与对策

| 威胁 | 对策 |
|------|------|
| 短信轰炸 | 多维频控(号/IP/设备指纹)+ 图形验证码兜底(触发阈值后) |
| 验证码爆破 | 单码 5 次试错即作废;错误响应不区分「不存在/错误」 |
| Token 泄露 | Access 短时效;Refresh 旋转+复用检测;会话版本全端吊销 |
| 重放 | 关键写操作要求幂等键;支付回调验签+事件表幂等 |
| CSRF(Web Cookie 模式) | SameSite=Lax + 自定义头校验;扫码 state 一次性 |
| 水平越权 | 所有资源查询强制 `WHERE user_id = ctx.userID`(sqlc 模板固化) |
| PIPL 合规 | 生辰属个人敏感信息:传输 TLS、手机号打码展示、注销即软删+30 天物理清除 |

## 6. 实施边界(P1)

- Go 侧新增 `internal/auth`(令牌/中间件)、`internal/user`(账号/档案)、
  `internal/store`(pgx + sqlc);
- 密钥管理:`JWT_SECRET` 环境变量注入,支持双密钥轮换(`JWT_SECRET_PREV`);
- 短信通道抽象 `SMSProvider` 接口(阿里云/腾讯云实现二选一,配置切换);
- 本文档为实施契约,接口签名以 `docs/api.md` 增补为准。
