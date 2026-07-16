# 黄金基准生成器

本目录脚本用于再生成 `internal/ziwei/testdata/` 下的两份回归基准:

| 基准文件 | 生成脚本 | 内容 |
|----------|----------|------|
| `iztro_golden.json.gz` | `gen.js` | iztro 2.5.8 完整命盘输出,1566 用例(闰月/年界/边界年份/全时辰覆盖) |
| `patterns_golden.json.gz` | `genpatterns.mts` | TS 原版 patterns.ts 的格局判定输出(同一批输入) |

Go 引擎的正确性由这两份基准的逐字段比对保障
(`engine_golden_test.go` / `patterns_golden_test.go`)。

## 再生成方法

```bash
# 1) 安装依赖(临时,node_modules 已 gitignore)
npm install iztro@2.5.8 lunar-javascript@1.7.3 tsx --no-save

# 2) 安星基准
node tools/goldgen/gen.js   # 产出 fixtures.json
gzip -c fixtures.json > internal/ziwei/testdata/iztro_golden.json.gz

# 3) 格局基准 —— 依赖旧 TS 源码(lib/ziwei/*.ts),已在 Go 重构时删除,
#    需检出历史提交获取:
git worktree add /tmp/ts-legacy 88194a4
#    再将 genpatterns.mts 中的 REPO 指向 /tmp/ts-legacy 运行:
npx tsx tools/goldgen/genpatterns.mts
```

> `dump-data.mts` 为一次性数据迁移脚本(TS 常量 → data/*.json),
> 同样依赖历史提交中的 lib/ 目录,仅作存档。

## 口径说明

- iztro 调用参数:`astro.bySolar(date, timeIndex, gender, fixLeap=true, 'zh-CN')`,
  全部默认配置(《紫微斗数全书》安星法、年界正月初一);
- 基准中同时记录 lunar-javascript 的农历转换结果,与 Go 侧 lunar-go 交叉验证
  (两者同源同作者,1900-2100 全区间一致)。
