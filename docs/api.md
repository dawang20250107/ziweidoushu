# API 文档

统一响应包裹:

```json
{ "ok": true,  "data": { ... } }
{ "ok": false, "error": { "code": "...", "message": "..." } }
```

- Base URL:`http://<host>:8080`
- 所有接口无鉴权(管理接口除外),按单 IP 令牌桶限流(默认 20 rps / 突发 40)
- 时辰索引口径:`0`=早子时(00:00-01:00),`1`=丑时 … `11`=亥时,`12`=晚子时(23:00-00:00)
- 地支索引口径:`0`=子 `1`=丑 … `11`=亥;天干索引:`0`=甲 … `9`=癸

---

## 排盘

### POST /api/v1/chart

生成完整命盘 + 格局判定。

请求体:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| year / month / day | int | ✓ | 公历生日(1900-2100) |
| hour | int | ✓ | 时辰索引 0-12 |
| gender | string | ✓ | `male` / `female` |
| name | string | | 姓名 |
| longitude | float | | 出生地东经度数 |
| province / city | string | | 出生省市(自动查内置经度表) |
| trueSolarTime | bool | | 真太阳时校正(需 longitude 或 province+city) |
| referenceYear | int | | 「当前年龄/当前大限」参考年,默认系统当前年 |
| withPatterns | bool | | 是否返回格局,默认 true |

```bash
curl -s localhost:8080/api/v1/chart -d '{
  "year":1990,"month":6,"day":15,"hour":5,"gender":"male"
}'
```

响应 `data.chart` 关键字段:

- `lunarInfo` 农历(年/月/日/年干支/闰月)、`fourPillars` 四柱、`lunarDateText`
- `mingGongBranch` / `shenGongBranch` 命身宫地支索引;`mingZhu` / `shenZhu` 命主身主
- `wuxingJu` / `wuxingJuName` 五行局;`ziweiPos` 紫微位置
- `palaces[12]` 十二宫(按地支索引排序):宫名、宫干、星曜(名称/类型/庙旺利陷/生年四化)、大限起讫、长生十二神、博士十二神、小限、空宫借星
- `daXians[12]` 大限序列;`currentDaXianIndex` 当前大限
- `data.patterns[]` 格局:名称、吉凶等级(excellent/good/neutral/caution)、描述、必须/加分/破格条件、古籍出处

### GET /api/v1/calendar/lunar-year?year=1993

农历某年逐月表(供表单农历模式渲染月/日选项;lunar-go 口径,1900-2100)。

响应 `data`: `{ year, months: [{month(1-12), leap, days(29|30)}...] }`,闰月按年内实际位置插入(如 1993 年三月后随闰三月,共 13 项)。

### POST /api/v1/calendar/lunar-to-solar

农历生日 → 公历。请求 `{year, month, leap, day}`(闰月以 `leap:true` 标记);
响应 `data: {year, month, day}`。不存在的闰月/超出当月天数/区间外年份均 400(`bad_lunar_date`)。
换算在提交前完成,排盘/档案/问星下游一律公历。

### POST /api/v1/horoscope

运限叠加:大限(含童限)/ 小限 / 流年 / 流月 / 流日 / 流时。

请求体 = 排盘字段 + `target: {year, month, day, hour}`(目标公历日期与时辰索引 0-12)。

响应 `data.horoscope`:

- `nominalAge` 虚岁(自然年口径,支持 1-120)
- 六层 scope(`decadal/age/yearly/monthly/daily/hourly`),每层含:
  - `palaceBranch` 该层命宫地支索引;`palaceNames[12]` 以该层命宫重排的宫名(下标=地支索引)
  - `stem`/`branch` 该层干支;`mutagen[4]` 该层四化 [禄,权,科,忌]
  - `stars` 流曜(运/流/月/日/时 + 魁钺昌曲禄羊陀马鸾喜,流年另含年解),下标=地支索引
- `suiqian12` / `jiangqian12` 流年岁前/将前十二神

> 体系口径:倪师《天纪》只认**生年四化**(已标注在本命盘星曜上)与**流年四化**;
> 大限/流月等层的 `mutagen` 为飞星派研究字段,展示层自行取舍。

```bash
curl -s localhost:8080/api/v1/horoscope -d '{
  "year":1990,"month":6,"day":15,"hour":5,"gender":"male",
  "target":{"year":2026,"month":7,"day":16,"hour":6}
}'
```

### GET /api/v1/famous
### GET /api/v1/famous/{id}/chart

名人命盘示例库与一键排盘。

---

## 四化

### GET /api/v1/sihua/liunian?year=2026

流年四化(年干 → 禄权科忌四星)。

### GET /api/v1/sihua/liuyue?year=2026&month=3

流月四化(五虎遁推月干;month 为农历月 1-12)。

---

## 古籍查阅

### GET /api/v1/books

全部书目(书名/朝代/作者/简介/章节数/段落数)与语料统计。

### GET /api/v1/books/{slug}

整本书(含全部章节与段落)。内置:`gusuifu` 骨髓赋、`quanji` 紫微斗数全集、`quanshu` 紫微斗数全书。

### GET /api/v1/books/{slug}/chapters/{idx}

单章内容(0 起)。

### GET /api/v1/search?q=紫微&limit=30

全文检索(中文二元索引 + 原文校验,毫秒级)。命中片段以 `<mark>` 高亮(已转义)。

### POST /api/v1/admin/corpus/reload

热加载 `CORPUS_EXTERNAL_DIR` 外部古籍目录,无需重启。需 `Authorization: Bearer <ADMIN_TOKEN>`。
外部古籍 JSON 格式见 [docs/corpus-schema.md](./corpus-schema.md)。

---

## 倪海厦知识库

### GET /api/v1/nihai/{section}

`tianji`(天纪:课程模块/64卦/堪舆/24集课程表/语录)、`renji`(人纪:针灸经验/透针/汉唐方/经方)、`diji`(地纪)、`bio`(倪师完整传记)。

### GET /api/v1/knowledge/stars

十四主星速览(关键词/五行/吉凶)+ 拼音 slug,及全量星曜档案层:
`lore` 逐星档案(主星/辅煞/杂曜共 69 曜,五行/化气/主司/义理,义引归纳
原创行文)、`cycles` 四大十二神逐名义(changsheng12/boshi12/suiqian12/
jiangqian12 各 12 名)、`flow` 流曜十义(按去前缀字键:魁钺昌曲禄羊陀马鸾喜)。
完备性由测试钉住:真实排盘+运限收集盘上全部星名,逐一断言有档。

### GET /api/v1/knowledge/topics

13 个解读主题(命格总览/感情/事业/财运/健康…)与对应宫位。

### GET /api/v1/knowledge/heming

合盘知识库(十四主星在夫妻宫断语、四化影响、方法论、评分标准)。

### GET /api/v1/cities

中国省市经度表(真太阳时校正用)。

---

## AI 解读

### POST /api/v1/ai/interpret

请求体 = 排盘字段 + 以下扩展:

| 字段 | 类型 | 说明 |
|------|------|------|
| topic | string | 解读主题:overview/personality/love/career/wealth/health/… |
| question | string | 自由提问(≤500 字) |
| stream | bool | true = SSE 流式(或用 `Accept: text/event-stream`) |

- **未配置 AI Key 时自动降级**:返回知识库规则版解读(`degraded: true`),服务始终可用。
- SSE 事件:`delta`(增量文本)→ `done`(供应商/长度)或 `error`。
- 上游并发满载时返回 `429 ai_busy`,带 `Retry-After`。

```bash
# 流式
curl -N localhost:8080/api/v1/ai/interpret -d '{
  "year":1990,"month":6,"day":15,"hour":5,"gender":"male",
  "topic":"career","stream":true
}'
```

### POST /api/v1/heming

合盘分析。请求体:`{"a": {排盘字段}, "b": {排盘字段}, "withAI": false}`。
返回双方命盘 + 夫妻宫主星断语(空宫借对宫)+ 倪师合盘方法论;`withAI: true` 时附加 AI 综合分析。

---

## 用户体系(需配置 DATABASE_URL + JWT_SECRET,否则统一 503)

**内测主通道为邮箱**:注册(邮箱验证码+自设密码+邀请码)→ 平时密码登录 →
忘记走邮箱验证码重置。`AUTH_INVITE_REQUIRED=1`(默认)开启邀请码闸门;
`AUTH_DEVICE_STRICT=1` 时陌生设备密码登录须邮箱验证码升级(环境检测,
设备底账 login_devices)。邮件通道:配置 `SMTP_HOST/PORT/USERNAME/PASSWORD/FROM`
即真发(465 隐式 TLS / 587 STARTTLS),未配置走 dev 通道(验证码写日志,
`SMS_DEV_ECHO_CODE=1` 时回显)。手机号短信通道保留,公测接云厂商后并存。

### POST /api/v1/auth/email/send-code

`{email, purpose: "register"|"reset"|"login"}` → `{sent: true, devCode?}`。
频控与短信同刻度(60s/次、1h≤5、24h≤10、同 IP 24h≤20)。register 用途
且邮箱已注册时 400 `email_taken`。

### POST /api/v1/auth/email/register

`{email, code, password, invite?}` → `{tokens, user, created: true}`。
密码 8-72 位须含字母与数字(bcrypt 存储);邀请码闸门开启时 invite 必填,
无效/用尽/过期 400 `invite_invalid`;验证码错误 400 `code_invalid`。

### POST /api/v1/auth/email/login

`{email, password, code?}` → `{tokens, user}`;凭证错误统一 401
`login_failed`(防枚举,恒时比对)。strict 模式陌生设备且密码正确时返回
200 `{needVerify: true}`,客户端发 login 用途验证码后带 `code` 重试。

### POST /api/v1/auth/password/reset

`{email, code, newPassword}` → `{reset: true}`;成功后全端下线。

### POST /api/v1/auth/password/change(需鉴权)

`{oldPassword, newPassword}` → `{changed: true}`;其余端下线。

### POST /api/v1/admin/invites(Bearer ADMIN_TOKEN)

`{count: 1-200, maxUses: 1-10000, note?, expiresInDays?}` → `{codes: ["ZW-XXXXXXXX", ...]}`。
内测运营铸码;`used_count/invite_uses` 落库可审计。

### POST /api/v1/auth/sms/send

`{phone}` → `{sent: true}`。频控:同号 60s/次、1h≤5、24h≤10;同 IP 24h≤20。
dev 短信通道 + `SMS_DEV_ECHO_CODE=1` 时附 `devCode`(仅本地)。

### POST /api/v1/auth/sms/verify

`{phone, code}` → `{tokens: {access, refresh, expiresAt}, user, created}`。
未注册手机号自动建号。验证码 5 分钟有效、一次性、错误话术统一防遍历。

### POST /api/v1/auth/refresh

`{refresh}` → 新令牌对(旋转:旧 refresh 即刻失效)。
**复用检测**:已撤销的 refresh 被再次使用时,撤销该用户全部令牌并全端下线。

### POST /api/v1/auth/logout(需鉴权)

`{refresh?, all?}`:撤销当前设备令牌;`all=true` 时全端下线(会话版本 +1,
所有已签发 access 在 10 秒内失效)。

### GET /api/v1/me(需鉴权)

`Authorization: Bearer <access>` → 用户信息(id/昵称/头像/会员层级/到期时间)。

---

## 变现:订阅 + 按次付费(需用户体系)

商品双轨:`kind=subscription`(会员时长)与 `kind=credits`(次卡,如深度报告)。
订单状态机:`created → paying → paid → fulfilled`;超时 `closed`、退款 `refunded`。
金额一律服务端取自商品表,不信任客户端。

### GET /api/v1/products

上架商品列表 → `{products: [{id, title, kind, tier?, durationDays?, creditType?,
creditAmount?, priceCents, originalPriceCents?}], devPayEnabled}`。

### POST /api/v1/orders(需鉴权)

`{productId}` → `{order, devPayEnabled}`。订单 2 小时未支付自动关闭。
微信/支付宝渠道接入后在此返回 prepay 参数。

### GET /api/v1/orders(需鉴权)/ GET /api/v1/orders/{id}(需鉴权)

订单列表(近 50 条)/ 单个订单。

### POST /api/v1/orders/{id}/dev-pay(需鉴权,仅 PAY_DEV_ENABLED=1)

dev 支付渠道:模拟渠道回调,标记支付成功并立即履约(订阅顺延发放 /
次数入账)。重复调用幂等。生产环境禁开。

### GET /api/v1/me/entitlements(需鉴权)

`{tier, entitlements: [{tier, startsAt, endsAt}], credits: {deep_report: n}}`。
订阅续费顺延:新时段起点 = max(now, 当前同层级最晚到期)。

### POST /api/v1/ai/report(需鉴权,消耗 1 次 deep_report)

`{排盘字段, topic?}` → `{report, topic, remainingCredits}`。
原子扣次 → 生成深度报告;AI 失败自动退还次数。未配置 LLM 时返回 503
且不扣次数(付费报告不走规则化降级)。次数不足返回 402 `no_credits`。

---

## 命盘档案库(需鉴权)

档案 = 生辰输入 + 排盘快照(引擎版本随存;版本升级后读取时自动按最新口径重排)。
免费层最多 3 份,pro/master 不限。

### POST /api/v1/profiles

`{label, relation?, isDefault?, 排盘字段}` → `{profile}`。生辰不能排盘则 400;
超出免费层上限返回 403 `profile_limit`。

### GET /api/v1/profiles

`{profiles: [不含快照的轻量列表], limit}`(limit=0 表示不限)。

### GET /api/v1/profiles/{id}

单个档案,含 `chartSnapshot`(命盘 + 格局)。

### DELETE /api/v1/profiles/{id}

软删除。`POST /api/v1/profiles/{id}/default` 设为默认档案。

---

## 古籍阅读:跨端续读 + 书签(需鉴权)

### GET /api/v1/me/reading

`{progress: [{bookSlug, chapterIdx, paragraphId, updatedAt}]}`(按最近阅读排序)。

### PUT /api/v1/me/reading/{slug}

`{chapterIdx, paragraphId}` → `{ok: true}`。同书 upsert;前端滚动节流 ≥2s 上报,
离开章节补发最后位置。未登录时前端以 localStorage 本地记忆兜底。

### GET /api/v1/me/bookmarks?book={slug?}

书签列表(可按书过滤)。`POST /api/v1/me/bookmarks`
`{bookSlug, chapterIdx, paragraphId, excerpt}` → `{bookmark}`(摘录截断 ≤200 字,
同段重复添加幂等);`DELETE /api/v1/me/bookmarks/{id}` → `{deleted: true}`(仅本人)。

## 占卜:梅花易数 + 六爻纳甲 + 小六壬

起卦免费;AI 深度解卦按次付费(credit_type=divination,商品 divine_3/divine_10)。

问辞上限均为 200 字(按字符计)。

**时区口径:北京时间(UTC+8)。** `castAt` 为 unix 秒(绝对时刻,本身无时区);
服务端推农历/时辰/日干支前一律先归一到东八区墙钟,与部署机器时区无关
(容器常为 UTC,回归测试钉住)。时辰为占法最小时间粒度(两小时一辰),
秒与毫秒不参与推演;同一时辰的正时课(六壬正时/梅花与小六壬无问辞)人人相同,
求同刻各异用问辞字数(梅花/小六壬)、报数(六壬活时)或摇卦随机(六爻)。

### POST /api/v1/divination/meihua

新增 `method: "zi"` 测字起卦:`ziText` 传一或二个汉字,按 Unihan 简体笔画起数
(一字:字画起上卦、加时辰配下卦并取动爻;二字:两仪平分,总画加时取动爻);
响应带 `ziText`/`ziStrokes` 溯源。六爻起卦可传 `yongShen`(世爻/妻财/官鬼/父母/子孙/兄弟)
显式定用神,优先于问辞推断,响应 `yongShenOverride` 回传快照;大六壬可传 `birthYear`
(1900-2100)加断年命上神,响应带 `nianMing`。AI 解卦回传同一卦时须原样回传上述字段。

`{method: "time"|"number", numbers?, castAt?, question?}` → `{result, castAt}`。
`time` 法:**有问辞按字数起数**(声音占义:问辞字数起上卦、加时辰数配下卦并
取动爻——众人同刻问辞各异,卦自不同;`result.castBasis` 溯源起数依据),
无问辞回退年月日时法(观梅体);数字起卦支持两数/三数式。result 含本卦/
互卦/变卦(各带 `guaCi` 周易卦辞)/动爻/体用五行生克与吉凶倾向。
`castAt` 须在近 24 小时内(防伪造历史卦)。正确性由邵康节观梅占黄金测试钉住。

### POST /api/v1/divination/liuyao

`{method: "shake"|"tosses", tosses?, castAt?, question?}` → `{result, castAt}`。
`shake` 为服务端 crypto/rand 模拟三枚铜钱六掷;`tosses` 为报爻起卦,六爻背面数
自下而上各 0-3(1背少阳/2背少阴/3背老阳动/0背老阴动)。result 为完整装卦:
本卦/变卦名、八宫宫属与世次(纯卦~归魂)、逐爻纳甲干支五行/六亲/六神/世应/
动变爻,及月建日辰与摇卦原始记录 `tosses`(回传同一卦的凭据)。静卦
`movingNums` 恒为 `[]`。正确性由《卜筮正宗》八宫六十四卦定表逐卦对照钉住。

### POST /api/v1/divination/daliuren

`{question?, castAt?, baoShu?}` → `{result, castAt, recordId?}`。
`baoShu` 三态:**缺省=正时起课**(以当下时辰——同一时辰之课人人相同,
古以年命分断);**0=活时代摇**(服务端 crypto 取 1-12);**>0=活时报数**
(自子顺数报数所至之支为占时,日干支/月将仍按实时)。result 为完整起课:
天地盘/四课/三传(带旬遁干 `chuanDunGan`)/课体/十二天将/旬空 `xunKong`/
确定性断语,活时课带 `baoShu` 与 `hourNote`(「活时·报数7」)溯源——
复现同一课的凭据为 `castAt+baoShu`。正确性由《六壬断案》全书 205 例
批量回归钉住。

### POST /api/v1/divination/xiaoliuren

小六壬快占(倪师《天纪》课堂教法):`{question?, castAt?}` →
掐指落位(月/日/时)与断语(大安/留连/速喜/赤口/小吉/空亡)。
带 `question` 时以问辞字数自时位再入一数(问数第四跳,响应 `qNum`/
`steps`×4 溯源)——同刻问事各异,落宫自不同;无问辞为传统正时三跳。

### POST /api/v1/ai/divine(需鉴权,消耗 1 次 divination)

`{kind?: "meihua"|"liuyao"|"daliuren", method?, numbers?, tosses?, castAt,
baoShu?, question}`(question 必填,kind 缺省 meihua)→
`{result, reading, remainingCredits}`。
服务端按 kind 重推卦象(不信任客户端),并强制「同一卦契约」:
`castAt` 一律必传(缺失 400 `bad_cast_time`——卦象/旺衰随时辰走,缺省
落到当下即另一卦);六爻须回传起卦返回的六掷 `tosses`(缺失 400
`bad_tosses`);大六壬活时课须回传 `baoShu`,且 baoShu≤0 返回 400
`bad_baoshu`——付费解课不代摇。校验通过后引语料 RAG 解卦;
AI 失败自动退还;未配置 LLM 返回 503 不扣次;次数不足 402 `no_credits`。
注:梅花 time 法卦象依赖问辞字数,回传时问辞须与起卦时一致方为同一卦。

### 卦档(登录后自动存档)

登录用户起卦(梅花/六爻/小六壬)时服务端自动存档并在起卦响应中附
`recordId`;AI 解卦请求携带 `recordId` 则解卦文本回填该档(匿名起卦后
才登录的场景,解卦时自动补建一条带解卦的档)。每用户保留最近 200 条,
超出截断最旧。

- `GET /api/v1/me/divinations?limit&offset&kind` → `{records, total}`
  (轻量列表:kind/question/summary/hasReading/castAt,无卦象与解卦全文;
  `kind` 可选 meihua|liuyao|daliuren|xiaoliuren,缺省全部)
- `GET /api/v1/me/divinations/{id}` → `{record}`(含 payload 卦象与 reading)
- `DELETE /api/v1/me/divinations/{id}` → `{deleted: true}`(仅本人)

## 研究语料(内部)

`Book.research: true` 的语料(`research/books-json/`,经 `CORPUS_EXTERNAL_DIR`
加载)对外完全不可见:不出现在 `/books`、公开 `/search` 不命中、按 slug 直接
访问返回 404;仅 AI 解读内部检索(SearchAll)可引用其内容片段。

## 运维

| 接口 | 说明 |
|------|------|
| GET /healthz | 存活探针 |
| GET /readyz | 就绪探针(数据加载完成) |
| GET /metrics | Prometheus 文本指标(请求量/延迟/缓存命中/AI 调用) |
| GET /api/v1/meta | 版本、引擎口径、AI 供应商、语料统计 |
