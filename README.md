# 紫微斗数 · AI+ 排盘与古籍查阅平台

基于**倪海厦《天纪》体系**的紫微斗数平台后端:完整排盘引擎、格局判定、古籍全文检索、
倪海厦三纪知识库、AI 命盘解读。**Go 语言实现,纯 API 服务,API-first 设计**,
供 Web / 微信小程序等多端接入。

> v2.0 全面重构:原 Next.js/TypeScript 全栈实现已下线,语言栈整体切换为 Go。
> 排盘不再依赖第三方库——安星算法完整自研实现,并以 iztro 2.5.8 的
> 1566 个黄金基准用例逐字段回归验证(见「正确性保障」)。

## 能力总览

| 模块 | 说明 |
|------|------|
| **排盘引擎** | 完整安星:命身宫、五虎遁、五行局、十四主星、十四辅星、38 杂曜、庙旺利陷、生年四化、大限小限、长生/博士十二神、四柱、真太阳时校正、空宫借对宫 |
| **格局引擎** | 41 个判定器、70+ 经典格局(君臣庆会/紫府同宫/杀破狼/阳梁昌禄…),必须/加分/破格三层条件,古籍出处可考 |
| **古籍查阅** | 骨髓赋、紫微斗数全集、紫微斗数全书;中文二元索引全文检索;**外部目录热加载**(后期古籍资料、倪师著作按 JSON 规范投放即可,无需改代码) |
| **倪海厦知识库** | 天纪(课程/64 卦/堪舆/语录)、人纪(针灸经验 120/透针 31/汉唐方 97/经方 25)、地纪、倪师传记;十四主星速览;合盘夫妻宫断语 |
| **AI 解读** | Anthropic 原生 + OpenAI 兼容协议(DeepSeek/通义/智谱等),SSE 流式;prompt 融合命盘+格局+知识库+古籍引文(RAG);**无 Key 自动降级**为知识库规则版解读 |
| **合盘** | 双人命盘 + 夫妻宫主星断语 + 倪师合盘方法论,可选 AI 综合分析 |

## 体系口径

倪海厦《天纪》三合派正统:**生年四化永远固定不动**,不使用飞星派的宫干自化、
大限四化、来因宫作为排盘输出(相关工具函数保留于 `internal/ziwei/sihua.go` 仅供研究)。
命宫为本、三方四正为用;空宫借对宫主星论。

## 快速开始

```bash
# 本地运行(Go 1.22+)
make run                       # 或: go run ./cmd/server
curl localhost:8080/api/v1/meta

# 排盘
curl -s localhost:8080/api/v1/chart -d '{
  "year":1990,"month":6,"day":15,"hour":5,"gender":"male"
}'

# 古籍检索
curl -s 'localhost:8080/api/v1/search?q=紫微'

# AI 解读(SSE 流式;未配 Key 时返回知识库规则版)
curl -N localhost:8080/api/v1/ai/interpret -d '{
  "year":1990,"month":6,"day":15,"hour":5,"gender":"male",
  "topic":"career","stream":true
}'
```

Docker:

```bash
docker compose up -d
```

配置全部走环境变量,见 [.env.example](./.env.example);完整接口文档见 [docs/api.md](./docs/api.md)。

## 高并发设计

- **无状态服务**:缓存/限流均为进程内加速,不影响正确性,水平扩容加负载均衡即可;
- **排盘纯计算**:单次排盘微秒级、零 IO,叠加 16 路分片 LRU 缓存;
- **语料只读快照**:检索基于不可变快照 + 二元倒排索引,读路径零锁竞争,热加载不阻塞读;
- **AI 上游治理**:信号量并发上限、独立超时、满载快速失败(429 + Retry-After),流式回传不占内存;
- **防护**:单 IP 令牌桶限流(32 分片)、请求体 1MiB 上限、panic 兜底、优雅停机;
- **可观测**:结构化 JSON 日志(含 request-id)、Prometheus 指标、健康/就绪探针。

渠道口径(2026-08 评估拍板):**不做小程序**(占卜类内容在微信/抖音/QQ 属
成文禁止,详见评估记录);微信生态走「公众号 + H5」(即本 web 直接承载),
规模化阶段付费与 AI 解读迁独立 App(App Store 中国区)。所有能力均为标准
HTTP/JSON + SSE,App/H5 直连同一套接口;CORS 与限流可按环境变量收紧。

## 正确性保障

排盘这类算法「重构后结果不能变」是硬约束,本仓库用双重黄金基准锁死口径:

1. **安星基准**:iztro 2.5.8(线上旧版所用引擎)对 1566 个生辰(覆盖 1900-2100
   闰月首尾、农历年界、历法分歧年份、全时辰)输出的完整命盘,与 Go 引擎
   **逐宫逐星逐字段**比对(星曜名称与顺序、庙旺利陷、四化、宫干、大限小限、
   长生/博士十二神、四柱、农历文本、命主身主、五行局)——全部一致;
2. **格局基准**:TS 原版 `patterns.ts` 对同一批输入的判定输出,与 Go 移植版
   比对(格局数量、顺序、名称、等级、描述文案、三层条件、古籍出处)——全部一致;
3. 历法层采用 lunar-go,与 iztro 底层的 lunar-typescript 同源同作者,
   基准中另附 lunar-javascript 转换结果交叉验证。

```bash
make test-race   # 全部测试(含两份黄金基准回归)
make bench       # 排盘引擎微基准
```

基准数据的再生成方法见 [tools/goldgen/README.md](./tools/goldgen/README.md);
iztro 安星算法的移植规格文档见 [docs/algorithm/](./docs/algorithm/)。

## 目录结构

```
cmd/server/          服务入口
internal/
  ziwei/             排盘引擎(安星/四化/格局/运限/黄金基准测试)
  corpus/            古籍语料库(存储/检索/外部导入)
  knowledge/         倪海厦三纪知识库与业务知识
  ai/                AI 解读层(多供应商/SSE/降级)
  httpapi/           HTTP API(路由/中间件/限流/缓存/指标)
  config/            环境变量配置
data/                内嵌数据资产(古籍/三纪/合盘/城市/名人,go:embed)
web/                 Web 前端(Next.js 15 + Tailwind 4,观星台设计系统)
migrations/          PostgreSQL Schema(用户/档案/订单/订阅,P1 用户系统用)
docs/                API 文档 / 古籍投放规范 / 算法规格 / 架构与设计系统
tools/goldgen/       黄金基准生成脚本(Node,存档用)
```

Web 前端本地开发:

```bash
make run                # 终端 1:Go API(:8080)
cd web && npm install && npm run dev   # 终端 2:前端(:3000,/api 已代理到 8080)
```

## 古籍资料扩容(预留)

后期投放古籍资料与倪海厦著作:整理为 [docs/corpus-schema.md](./docs/corpus-schema.md)
规范的 JSON,放入 `CORPUS_EXTERNAL_DIR` 目录,调用
`POST /api/v1/admin/corpus/reload` 即时生效——自动进入全文检索与 AI 引文源。

## 协议

- 代码:[MIT License](./LICENSE)
- 古籍原文(骨髓赋、紫微斗数全集/全书):Public Domain
