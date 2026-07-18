# iztro 2.5.8 排盘主流程移植规格（Go 移植用）

> 依据源码：`node_modules/iztro/lib/`（编译后 JS，v2.5.8）+ 依赖 `lunar-lite@0.2.8` + `lunar-typescript`（6tail lunar 库）。
> 固定调用口径：`astro.bySolar('YYYY-M-D', timeIndex(0~12), '男'|'女', true /*fixLeap*/, 'zh-CN')`，**全局配置均为默认值**。
> 本文所有公式均已用 fixtures 中真实 iztro 运行结果逐条验证。

## 0. 全局约定与默认配置

### 0.1 默认配置（`astro/astro.js` 顶部模块级变量）

| 配置项 | 默认值 | 含义（影响点） |
|---|---|---|
| `_yearDivide` | `'normal'` | 年干支以**农历正月初一**分界（非立春） |
| `_horoscopeDivide` | `'normal'` | 月干支按**初一**起（五虎遁+农历月），非节气 |
| `_ageDivide` | `'normal'` | 小限只按年份（主流程 ages 数组不受其影响） |
| `_dayDivide` | `'forward'` | 晚子时（timeIndex=12）**算第二天**（不把 12 改成 0） |
| `_algorithm` | `'default'` | 全书派安星；命主按**命宫地支**取（zhongzhou 派才按年支） |

`bySolar` 入口（astro.js `bySolar`）：仅当 `dayDivide === 'current'` 时把 `timeIndex >= 12` 改为 0；默认 `'forward'` 下 **timeIndex 原样保留为 12**。

### 0.2 基础常量（data/constants.js）

```
HEAVENLY_STEMS  = [甲,乙,丙,丁,戊,己,庚,辛,壬,癸]           // 索引 0~9
EARTHLY_BRANCHES = [子,丑,寅,卯,辰,巳,午,未,申,酉,戌,亥]     // 索引 0~11
```

**宫位索引体系**：十二宫数组下标 `i∈[0,11]`，`i=0` 固定为**寅宫**，按地支顺序递增：
`i=0寅, 1卯, 2辰, 3巳, 4午, 5未, 6申, 7酉, 8戌, 9亥, 10子, 11丑`。
宫位 i 的地支 = `EARTHLY_BRANCHES[fixIndex(2 + i)]`（astro.js bySolar 循环内）。

### 0.3 fixIndex（utils/index.js `fixIndex`）

```go
// max 默认 12；对天干用 max=10
func fixIndex(index, max int) int { return ((index % max) + max) % max }
```
JS 原实现为递归加减 max 并处理 -0，语义等价于上式。

`fixEarthlyBranchIndex(支)`（utils/index.js）= `fixIndex(支索引 - 2)`，即"地支→宫位下标"。

### 0.4 时辰索引 timeIndex 与干支计算用的"标称时刻"

`lunar-lite/lib/ganzhi.js getHeavenlyStemAndEarthlyBranchBySolarDate`：

```
hour = max(timeIndex*2 - 1, 0), minute = 30
// timeIndex: 0→00:30(早子) 1→01:30(丑) 2→03:30(寅) ... k→(2k-1):30 ... 11→21:30(亥) 12→23:30(晚子)
```
时支索引 = `timeIndex % 12`（12 即晚子时→子=0）。时辰表（data/constants.js `CHINESE_TIME`/`TIME_RANGE`，zh-CN 译名见 i18n common.json）：

| timeIndex | 名称 | 区间 | 时支 |
|---|---|---|---|
| 0 | 早子时 | 00:00~01:00 | 子 |
| 1 | 丑时 | 01:00~03:00 | 丑 |
| 2 | 寅时 | 03:00~05:00 | 寅 |
| 3 | 卯时 | 05:00~07:00 | 卯 |
| 4 | 辰时 | 07:00~09:00 | 辰 |
| 5 | 巳时 | 09:00~11:00 | 巳 |
| 6 | 午时 | 11:00~13:00 | 午 |
| 7 | 未时 | 13:00~15:00 | 未 |
| 8 | 申时 | 15:00~17:00 | 申 |
| 9 | 酉时 | 17:00~19:00 | 酉 |
| 10 | 戌时 | 19:00~21:00 | 戌 |
| 11 | 亥时 | 21:00~23:00 | 亥 |
| 12 | 晚子时 | 23:00~00:00 | 子 |

---

## 1. 公历→农历后的月份索引（fixLunarMonthIndex）

依据：`utils/index.js fixLunarMonthIndex`；农历转换 `lunar-lite/lib/convertor.js solar2lunar`（内部 lunar-typescript `Solar.getLunar()`，闰月时 `lunar.getMonth()` 为负数）。

```go
// 输出：0 基月份索引（0=正月建寅，直接作为"从寅宫顺数"的偏移）
func fixLunarMonthIndex(solarDate string, timeIndex int, fixLeap bool) int {
    lunarMonth, lunarDay, isLeap := solar2lunar(solarDate) // lunarMonth 1~12（闰月取绝对值）
    needToAdd := isLeap && fixLeap && lunarDay > 15 && timeIndex != 12
    return fixIndex(lunarMonth + 1 - 2 + b2i(needToAdd), 12) // = fixIndex(lunarMonth-1+add)
}
```

边界规则（**精确口径**）：
- 闰月折算界限是 **农历日 > 15**（即十六及以后折入下月；**十五当天仍算本月**）。
- 折算需同时满足：`isLeap && fixLeap==true && lunarDay>15 && timeIndex != 12`。
  **注意 `timeIndex == 12`（晚子时）被显式排除**：晚子时出生即使在闰月下半月也**不**折入下月（源码条件 `timeIndex !== 12`，已实测：2023-4-6=闰二月十六，卯时命宫丑、晚子时命宫卯——晚子时按二月算）。
- timeIndex 对月份**无进位作用**：晚子时不会把月份 +1（当日为月末时也不会跨月），它只出现在上面这个排除条件里。
- `fixLeap==false` 时闰月一律按本月（=所闰之月）计算。
- 返回值语义：`0=正月(寅), 1=二月(卯), ..., 11=十二月(丑)`。

注意区分：**月干支（chineseDate.monthly）里的闰月折算是另一段代码**（lunar-lite ganzhi.js，条件只有 `isLeap && lunarDay>15`，**没有** timeIndex!=12 排除，也不受 fixLeap 参数控制），见 §11。

## 2. 命宫 / 身宫定位（palace.js getSoulAndBody）

```go
monthIndex   := fixLunarMonthIndex(solarDate, timeIndex, fixLeap) // 0=寅
timeBranch   := timeIndex % 12                                    // 子=0..亥=11（12→0）
soulIndex    := fixIndex(monthIndex - timeBranch, 12)             // 命宫：顺数至生月，逆数生时
bodyIndex    := fixIndex(monthIndex + timeBranch, 12)             // 身宫：顺数至生月，顺数生时
```
- `timeBranch` 源码取的是 `hourly[1]`（干支计算结果的时支）在 `EARTHLY_BRANCHES` 的索引，按 §0.4 的标称时刻等价于 `timeIndex % 12`。
- 命宫地支 = `EARTHLY_BRANCHES[fixIndex(soulIndex + 2)]`；身宫地支 = `EARTHLY_BRANCHES[fixIndex(bodyIndex + 2)]`（astro.js bySolar）。
- 命宫天干见 §3（`heavenlyStemOfSoul = HEAVENLY_STEMS[fixIndex(idx(五虎遁起干) + soulIndex, 10)]`）。

实测例：1990-6-15（农历五月廿三）卯时(3)：monthIndex=4，soulIndex=1(卯)，bodyIndex=7(酉)。✓

## 3. 十二宫天干（五虎遁）

依据：`data/constants.js TIGER_RULE`、`palace.js getSoulAndBody`、`astro.js bySolar` 循环。

五虎遁表（年干→寅宫天干）：

| 年干 | 甲 | 乙 | 丙 | 丁 | 戊 | 己 | 庚 | 辛 | 壬 | 癸 |
|---|---|---|---|---|---|---|---|---|---|---|
| 寅宫干 | 丙 | 戊 | 庚 | 壬 | 甲 | 丙 | 戊 | 庚 | 壬 | 甲 |

```go
startStem := TIGER_RULE[年干]                       // 年干取自 §11 年柱（yearDivide=normal：正月初一分界）
palaceStem[i] := HEAVENLY_STEMS[fixIndex(idx(startStem) + i, 10)]  // i=0(寅)..11(丑)
```
源码写法是 `HEAVENLY_STEMS[fixIndex(idx(命宫干) - soulIndex + i, 10)]`（astro.js L182），与上式恒等。

**子丑两宫的处理**：不做"甲寅乙卯…甲子乙丑"重起，就是继续 mod 10 顺推——
`i=10(子宫)` 干 = `startStem + 10 ≡ startStem`（与寅宫同干）；`i=11(丑宫)` 干 = `startStem + 1`（与卯宫同干）。
例：庚午年 → 戊寅、己卯、庚辰、辛巳、壬午、癸未、甲申、乙酉、丙戌、丁亥、**戊子、己丑**。✓（实测）

**来因宫（isOriginalPalace）**：宫干 == 年干 且 宫支 ∉ {子, 丑}（astro.js L188）。

## 4. 宫名排布

依据：`data/constants.js PALACES`、`palace.js getPalaceNames`。

`PALACES` 顺序（kot 键 → zh-CN，译文出自 i18n/locales/zh-CN/palace.js）：

| k | PALACES[k] | zh-CN |
|---|---|---|
| 0 | soulPalace | 命宫 |
| 1 | parentsPalace | 父母 |
| 2 | spiritPalace | 福德 |
| 3 | propertyPalace | 田宅 |
| 4 | careerPalace | 官禄 |
| 5 | friendsPalace | 仆役 |
| 6 | surfacePalace | 迁移 |
| 7 | healthPalace | 疾厄 |
| 8 | wealthPalace | 财帛 |
| 9 | childrenPalace | 子女 |
| 10 | spousePalace | 夫妻 |
| 11 | siblingsPalace | 兄弟 |

```go
// names[i] 为宫位 i（0=寅）的名字
names[i] = PALACES[fixIndex(i - soulIndex, 12)]
// 等价：names[fixIndex(soulIndex + k)] = PALACES[k]
```
即从命宫开始，沿宫位索引**递增方向**（寅→卯→辰…，盘面逆时针）依次安：命宫、父母、福德、田宅、官禄、仆役、迁移、疾厄、财帛、子女、夫妻、兄弟。
（身宫不是独立宫名，只是 `isBodyPalace = (bodyIndex == i)` 标志；zh-CN "身宫"= `bodyPalace`。）

## 5. 五行局（palace.js getFiveElementsClass）

由**命宫干支**计算（非查 60 纳音表，用口诀公式，两者结果一致）：

```go
stemNum   := idx10(命宫干)/2 + 1                    // 甲乙1 丙丁2 戊己3 庚辛4 壬癸5
branchNum := (idx12(命宫支) % 6)/2 + 1              // 子丑午未1 寅卯申酉2 辰巳戌亥3
sum := stemNum + branchNum
for sum > 5 { sum -= 5 }
// sum: 1→木三局(3) 2→金四局(4) 3→水二局(2) 4→火六局(6) 5→土五局(5)
```
`FiveElementsClass` 枚举值（data/constants.js）：water2nd=2, wood3rd=3, metal4th=4, earth5th=5, fire6th=6。

完整 60 干支→局对照（由上式生成，已与 iztro `getFiveElementsClass` 逐一比对）：

| 干支 | 局 | 干支 | 局 | 干支 | 局 | 干支 | 局 | 干支 | 局 | 干支 | 局 |
|---|---|---|---|---|---|---|---|---|---|---|---|
| 甲子 | 金四局 | 乙丑 | 金四局 | 丙寅 | 火六局 | 丁卯 | 火六局 | 戊辰 | 木三局 | 己巳 | 木三局 |
| 庚午 | 土五局 | 辛未 | 土五局 | 壬申 | 金四局 | 癸酉 | 金四局 | 甲戌 | 火六局 | 乙亥 | 火六局 |
| 丙子 | 水二局 | 丁丑 | 水二局 | 戊寅 | 土五局 | 己卯 | 土五局 | 庚辰 | 金四局 | 辛巳 | 金四局 |
| 壬午 | 木三局 | 癸未 | 木三局 | 甲申 | 水二局 | 乙酉 | 水二局 | 丙戌 | 土五局 | 丁亥 | 土五局 |
| 戊子 | 火六局 | 己丑 | 火六局 | 庚寅 | 木三局 | 辛卯 | 木三局 | 壬辰 | 水二局 | 癸巳 | 水二局 |
| 甲午 | 金四局 | 乙未 | 金四局 | 丙申 | 火六局 | 丁酉 | 火六局 | 戊戌 | 木三局 | 己亥 | 木三局 |
| 庚子 | 土五局 | 辛丑 | 土五局 | 壬寅 | 金四局 | 癸卯 | 金四局 | 甲辰 | 火六局 | 乙巳 | 火六局 |
| 丙午 | 水二局 | 丁未 | 水二局 | 戊申 | 土五局 | 己酉 | 土五局 | 庚戌 | 金四局 | 辛亥 | 金四局 |
| 壬子 | 木三局 | 癸丑 | 木三局 | 甲寅 | 水二局 | 乙卯 | 水二局 | 丙辰 | 土五局 | 丁巳 | 土五局 |
| 戊午 | 火六局 | 己未 | 火六局 | 庚申 | 木三局 | 辛酉 | 木三局 | 壬戌 | 水二局 | 癸亥 | 水二局 |

## 6. 大限（palace.js getHoroscope，decadals 部分）

- 阴阳判定：`GENDER = {male:'阳', female:'阴'}`（data/constants.js）；年支阴阳（data/earthlyBranches.js `yinYang`）：**阳支=子寅辰午申戌，阴支=丑卯巳未酉亥**。
- **顺逆**：`GENDER[性别] == 年支阴阳` → 顺行（阳男、阴女）；否则逆行（阴男、阳女）。年支取自 `{year: yearDivide}` 配置（默认初一分界）。

```go
start0 := FiveElementsClassValue // 水2 木3 金4 土5 火6（起运虚岁）
for i := 0; i < 12; i++ {
    var idx int
    if 顺行 { idx = fixIndex(soulIndex + i) } else { idx = fixIndex(soulIndex - i) }
    start := start0 + 10*i
    decadals[idx] = Decadal{
        Range:         [2]int{start, start + 9},           // 闭区间，虚岁
        HeavenlyStem:  HEAVENLY_STEMS[fixIndex(idx(TIGER_RULE[年干]) + idx, 10)], // 即该宫自身宫干
        EarthlyBranch: EARTHLY_BRANCHES[fixIndex(2 + idx)],                      // 即该宫自身宫支
    }
}
```
即：命宫起第一限 `[局数, 局数+9]`，每宫 +10 岁；`range` 数组=每宫 `[起, 止]` 两元素。
实测：土五局阳男，寅宫为命宫后第 12 宫（逆推 i=11）→ `[115,124]`。✓

## 7. 小限（palace.js getHoroscope，ages 部分 + utils/index.js getAgeIndex）

起宫（按**年支三合**，getAgeIndex）：

| 年支 | 起宫（地支） | 宫位索引 |
|---|---|---|
| 寅、午、戌 | 辰 | 2 |
| 申、子、辰 | 戌 | 8 |
| 巳、酉、丑 | 未 | 5 |
| 亥、卯、未 | 丑 | 11 |

顺逆：**男顺女逆**（只看性别，不看年支阴阳：源码 `kot(gender) === 'male' ? ageIdx+i : ageIdx-i`）。

```go
ageIdx := getAgeIndex(年支)
for i := 0; i < 12; i++ {
    age := make([]int, 10)
    for j := 0; j < 10; j++ { age[j] = 12*j + i + 1 }   // 1岁起，每轮+12，共10个
    var idx int
    if 男 { idx = fixIndex(ageIdx + i) } else { idx = fixIndex(ageIdx - i) }
    ages[idx] = age
}
```
即起宫放 `[1,13,25,...,109]`，下一宫（男顺/女逆）放 `[2,14,26,...,110]`，以此类推；每宫恰好 10 个虚岁数。
实测：庚午年男，起辰(2)，寅宫(0)=i10 → `[11,23,35,47,59,71,83,95,107,119]`。✓

## 8. 命主 / 身主（data/earthlyBranches.js soul/body 字段）

- **命主**：默认派别按**命宫地支**查 `soul`（astro.js L210-212；仅 `algorithm==='zhongzhou'` 时按年支）。
- **身主**：一律按**生年年支**查 `body`。

| 地支 | 命主 soul | 身主 body |
|---|---|---|
| 子 | 贪狼 | 火星 |
| 丑 | 巨门 | 天相 |
| 寅 | 禄存 | 天梁 |
| 卯 | 文曲 | 天同 |
| 辰 | 廉贞 | 文昌 |
| 巳 | 武曲 | 天机 |
| 午 | 破军 | 火星 |
| 未 | 武曲 | 天相 |
| 申 | 廉贞 | 天梁 |
| 酉 | 文曲 | 天同 |
| 戌 | 禄存 | 文昌 |
| 亥 | 巨门 | 天机 |

## 9. 长生十二神（star/decorativeStar.js getchangsheng12 / getChangesheng12StartIndex）

长生起宫（按**命宫五行局**）：

| 五行局 | 长生所在地支 | 宫位索引 |
|---|---|---|
| 水二局 | 申 | 6 |
| 木三局 | 亥 | 9 |
| 金四局 | 巳 | 3 |
| 土五局 | 申 | 6 |
| 火六局 | 寅 | 0 |

顺逆：与大限相同——**阳男阴女顺行，阴男阳女逆行**（`GENDER[性别] == 年支阴阳` 判断，年支取 timeIndex=0、yearDivide 配置下的年支，默认口径下与命盘年支一致）。

十二神固定顺序（zh-CN 译名，i18n/locales/zh-CN/star.js）：
`长生、沐浴、冠带、临官、帝旺、衰、病、死、墓、绝、胎、养`

```go
startIdx := 长生起宫[五行局]
for i, name := range [12]string{长生,...,养} {
    var idx int
    if 顺行 { idx = fixIndex(startIdx + i) } else { idx = fixIndex(startIdx - i) }
    changsheng12[idx] = name
}
```

## 10. 博士十二神（star/decorativeStar.js getBoShi12 + star/location.js getLuYangTuoMaIndex）

起宫 = **禄存**所在宫（按年干）：

| 年干 | 甲 | 乙 | 丙 | 丁 | 戊 | 己 | 庚 | 辛 | 壬 | 癸 |
|---|---|---|---|---|---|---|---|---|---|---|
| 禄存地支 | 寅 | 卯 | 巳 | 午 | 巳 | 午 | 申 | 酉 | 亥 | 子 |
| 宫位索引 | 0 | 1 | 3 | 4 | 3 | 4 | 6 | 7 | 9 | 10 |

顺逆：**阳男阴女顺行，阴男阳女逆行**（同 §9 规则，年支阴阳 vs 性别阴阳）。

十二神固定顺序（zh-CN）：
`博士、力士、青龙、小耗、将军、奏书、飞廉、喜神、病符、大耗、伏兵、官府`

```go
luIndex := 禄存宫[年干]
for i, name := range [12]string{博士,...,官府} {
    idx := fixIndex(顺行 ? luIndex+i : luIndex-i)
    boshi12[idx] = name
}
```

## 11. chineseDate 四柱（lunar-lite/lib/ganzhi.js + lunar-typescript/dist/lib/Lunar.mjs）

历法数据来源：**lunar-typescript（6tail lunar 库 TS 版）**，农历/节气由寿星天文历法（ShouXingUtil）实时计算，支持范围远超 1900~2100。Go 移植建议直接用同作者的 **github.com/6tail/lunar-go**（API 同名：`GetYearGan/GetDayGanExact/GetTimeGan` 等，算法一致），或以黄金基准数据校验自研实现。

计算前先构造标称时刻：`Solar.fromYmdHms(y, m, d, max(timeIndex*2-1,0), 30, 0)`（见 §0.4）。

### 11.1 年柱（yearDivide='normal'，正月初一分界）

```go
lunarYear := solar2lunar(date).lunarYear     // 农历年号（初一分界）
ganIdx := ((lunarYear - 4) % 10 + 10) % 10   // 甲=0
zhiIdx := ((lunarYear - 4) % 12 + 12) % 12   // 子=0
```
（`Lunar._computeYear`：`offset = year - 4`。若配置为 `'exact'` 则用 `getYearGanByLiChun`：仅比较**日期**（不含时刻）与当年立春日，立春前减一年——默认不启用。）

### 11.2 月柱（horoscopeDivide='normal'，按初一 + 五虎遁）

`lunar-lite ganzhi.js calculateMonthlyGanZhi` 的 normal 分支：

```go
fixLeap := 0
if isLeap && lunarDay > 15 { fixLeap = 1 }   // 注意：此处无 timeIndex!=12 排除，也不受 bySolar 的 fixLeap 参数控制
m := lunarMonth - 1 + fixLeap                // 0=正月
monthGan := HEAVENLY_STEMS[(idx(FIVE_TIGER[idx(年干)]) + m) % 10]  // FIVE_TIGER 同 §3 五虎遁表
monthZhi := MONTHLY_EARTHLY_BRANCHES[m]      // [寅,卯,辰,巳,午,未,申,酉,戌,亥,子,丑]，无回绕（闰十二月在支持范围内不存在）
```
实测：闰二月十六（2023-4-6）任何时辰月柱均为丙辰（按三月）；闰二月十五为乙卯（按二月）。✓

### 11.3 日柱（关键：晚子时进位）

`Lunar._computeDay`：

```go
jd := julianDay(y, m, d, 12, 0, 0)           // 当日正午的儒略日（公式见下）
offset := int(math.Floor(jd)) - 11
dayGanIdx := ((offset % 10) + 10) % 10       // 甲=0
dayZhiIdx := ((offset % 12) + 12) % 12       // 子=0
// Exact 修正：标称时刻在 [23:00, 23:59] 内（即 timeIndex==12 晚子时→23:30）时日柱进一位：
if timeIndex == 12 /* 且 dayDivide != 'current' 的默认口径 */ {
    dayGanIdx = (dayGanIdx + 1) % 10
    dayZhiIdx = (dayZhiIdx + 1) % 12
}
// 早子时（timeIndex==0 → 00:30）不进位，用当日日柱。
```
儒略日公式（`Solar.getJulianDay`，标准格里高利/儒略历转换）：

```go
d := day + ((sec/60 + min)/60 + hour) / 24
n, g := 0, y*372 + m*31 + int(math.Floor(d)) >= 588829
if m <= 2 { m += 12; y-- }
if g { a := y / 100; n = 2 - a + a/4 }
jd = math.Floor(365.25*float64(y+4716)) + math.Floor(30.6001*float64(m+1)) + d + float64(n) - 1524.5
```
实测：1990-6-15 日柱辛亥；timeIndex=12 时为壬子（进位）。✓

### 11.4 时柱（五鼠遁，基于**进位后的**日干）

```go
timeZhiIdx := timeIndex % 12                 // 子=0（getTimeZhiIndex 对 23:00~00:59 返回 0）
timeGanIdx := (dayGanIdxExact % 5 * 2 + timeZhiIdx) % 10   // dayGanIdxExact = 11.3 的（含晚子时进位）日干
```
`(dayGan % 5) * 2` 即五鼠遁 RAT_RULE：甲己→甲子起，乙庚→丙子，丙辛→戊子，丁壬→庚子，戊癸→壬子。
晚子时：时支=子，时干按**次日**日干遁（实测 1990-6-15 t12：日壬子、时庚子 = 壬日庚子起 ✓）。

### 11.5 字符串格式（utils/index.js translateChineseDate）

zh-CN 下每个干支字译文长度为 1，走单空格分支：

```
"{年干}{年支} {月干}{月支} {日干}{日支} {时干}{时支}"    // 例："庚午 壬午 辛亥 辛卯"
```
（柱内不空格，柱间单空格；多字符语言才用 " - " 分隔，zh-CN 不会触发。）

## 12. zodiac 生肖、sign 星座

### 12.1 生肖（astro.js getZodiacBySolarDate + lunar-lite misc.js getZodiac）

以 **timeIndex=0、yearDivide 配置（默认初一分界）** 的年支查表：

| 年支 | 子 | 丑 | 寅 | 卯 | 辰 | 巳 | 午 | 未 | 申 | 酉 | 戌 | 亥 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 生肖 | 鼠 | 牛 | 虎 | 兔 | 龙 | 蛇 | 马 | 羊 | 猴 | 鸡 | 狗 | 猪 |

即分界为**农历正月初一**（非立春）。zh-CN 输出单字（i18n common.json）。

### 12.2 星座（lunar-lite misc.js getSign → lunar-typescript Solar.getXingZuo）

纯按公历月日，`y = month*100 + day`：

| 区间（含端点） | zh-CN |
|---|---|
| 0321~0419 | 白羊座 |
| 0420~0520 | 金牛座 |
| 0521~0621 | 双子座 |
| 0622~0722 | 巨蟹座 |
| 0723~0822 | 狮子座 |
| 0823~0922 | 处女座 |
| 0923~1023 | 天秤座 |
| 1024~1122 | 天蝎座 |
| 1123~1221 | 射手座 |
| 1222~1231 或 0101~0119 | 摩羯座 |
| 0120~0218 | 水瓶座 |
| 0219~0320 | 双鱼座 |

## 13. lunarDate 中文农历字符串

依据：astro.js `lunarDate.toString(true)` → lunar-typescript `Lunar.toString()`：

```
{年中文}年{[闰]月中文}月{日中文}
```
- **年**：逐位数字转中文，`0~9 → 〇一二三四五六七八九`（如 1990→一九九〇，2023→二〇二三）。
- **月**：`1~12 → 正二三四五六七八九十冬腊`；闰月前缀 `闰`（如 闰二月）。
- **日**：`1~10 → 初一…初十`；`11~19 → 十一…十九`；`20 → 二十`；`21~29 → 廿一…廿九`；`30 → 三十`。
- 完整示例：`一九九〇年五月廿三`、`二〇二三年闰二月十六`。✓（实测）
- 农历月/日取的是**公历当日**对应的农历（不做晚子时进位）。

## 14. fiveElementsClass 中文名

i18n/locales/zh-CN/fiveElementsClass.js，格式 `{五行}{数字}局`，固定 5 个值：

| kot 键 | 数值 | zh-CN |
|---|---|---|
| water2nd | 2 | 水二局 |
| wood3rd | 3 | 木三局 |
| metal4th | 4 | 金四局 |
| earth5th | 5 | 土五局 |
| fire6th | 6 | 火六局 |

---

## 附录 A：紫微星定位（star/location.js getStartIndex，供主星安星参考）

```go
day := lunarDay                       // 农历日，初一=1
if timeIndex == 12 { day++ }          // 晚子时 +1 天（dayDivide 默认 forward）
if day > 当月农历总天数 { day -= 总天数 }  // 跨月修正（getTotalDaysOfLunarMonth）
offset := 0
for (day+offset) % 局数 != 0 { offset++ }
quotient := ((day + offset) / 局数) % 12
ziweiIndex := quotient - 1
if offset % 2 == 0 { ziweiIndex += offset } else { ziweiIndex -= offset }
ziweiIndex = fixIndex(ziweiIndex)
tianfuIndex := fixIndex(12 - ziweiIndex)
```

## 附录 B：验证样例（真实 iztro 2.5.8 输出）

`bySolar('1990-6-15', 3, '男', true, 'zh-CN')`：
- lunarDate=一九九〇年五月廿三；chineseDate=庚午 壬午 辛亥 辛卯
- 命宫卯、身宫酉；土五局；命主文曲、身主火星；生肖马、星座双子座
- 宫干（寅起）：戊寅 己卯 庚辰 辛巳 壬午 癸未 甲申 乙酉 丙戌 丁亥 戊子 己丑
- 寅宫（兄弟）：大限[115,124]、小限[11,23,...,119]、长生十二神=病、博士十二神=飞廉

边界验证：
- `('1990-6-15', 12, ...)`：chineseDate=庚午 壬午 **壬子 庚子**（日柱进位、时柱按次日日干遁）；命宫与早子时相同（时支均为子）。
- 2023 闰二月：十五(4-5)命宫按二月；十六(4-6)fixLeap=true 按三月、fixLeap=false 按二月、**晚子时按二月**（timeIndex==12 排除规则）。

## 已读源码文件清单

iztro 2.5.8（`fixtures/node_modules/iztro/lib/`）：
- `astro/astro.js`（bySolar 主流程、config 默认值、getZodiacBySolarDate、getSignBySolarDate、命主/身主取用）
- `astro/palace.js`（getSoulAndBody、getFiveElementsClass、getPalaceNames、getHoroscope 大限/小限）
- `utils/index.js`（fixIndex、fixEarthlyBranchIndex、fixLunarMonthIndex、fixLunarDayIndex、timeToIndex、getAgeIndex、translateChineseDate）
- `data/constants.js`（HEAVENLY_STEMS、EARTHLY_BRANCHES、PALACES、GENDER、FiveElementsClass、CHINESE_TIME、TIME_RANGE、TIGER_RULE、RAT_RULE、ZODIAC）
- `data/earthlyBranches.js`（yinYang、soul 命主、body 身主表）
- `star/decorativeStar.js`（getchangsheng12、getChangesheng12StartIndex、getBoShi12、getYearly12）
- `star/location.js`（getStartIndex 紫微定位、getLuYangTuoMaIndex 禄存）
- `star/index.js`（导出确认：decorativeStar 为实际生效模块）
- `i18n/locales/zh-CN/`：`palace.js`、`fiveElementsClass.js`、`earthlyBranch.js`、`heavenlyStem.js`、`gender.js`、`mutagen.js`、`star.js`、`common.json`

lunar-lite 0.2.8（`fixtures/node_modules/lunar-lite/lib/`）：
- `convertor.js`（solar2lunar/lunar2solar）、`ganzhi.js`（getHeavenlyStemAndEarthlyBranchBySolarDate、calculateMonthlyGanZhi）、`misc.js`（getSign/getZodiac/getTotalDaysOfLunarMonth）、`constants.js`（FIVE_TIGER、MONTHLY_EARTHLY_BRANCHES）

lunar-typescript（`fixtures/node_modules/lunar-typescript/dist/lib/`）：
- `Lunar.mjs`（_computeYear/_computeMonth/_computeDay/_computeTime、getYearInChinese/getMonthInChinese/getDayInChinese、toString）
- `LunarUtil.mjs`（getTimeZhiIndex、NUMBER/MONTH/DAY 常量）、`Solar.mjs`（getXingZuo、getJulianDay）、`I18n.mjs`（中文字典值）
