# iztro 2.5.8 安星算法精确规格（Go 移植基准）

> 本文档逐行核对 iztro v2.5.8 编译产物（`lib/`）写成，所有查表完整抄录，所有公式精确到取模与偏移常数。
> 源码根目录：`node_modules/iztro/lib/`（版本已由 `package.json` 确认为 `2.5.8`）。

---

## 0. 基础约定（移植前必读）

### 0.1 宫位索引体系

- 全库宫位数组长度 12，**下标 0 = 寅宫**，顺时针递增：

| 宫位索引 | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 地支 | 寅 | 卯 | 辰 | 巳 | 午 | 未 | 申 | 酉 | 戌 | 亥 | 子 | 丑 |

- `EARTHLY_BRANCHES`（`data/constants.js`）是**原始地支序**：`子0 丑1 寅2 卯3 辰4 巳5 午6 未7 申8 酉9 戌10 亥11`。
- `HEAVENLY_STEMS`（`data/constants.js`）：`甲0 乙1 丙2 丁3 戊4 己5 庚6 辛7 壬8 癸9`。
- 换算函数 `fixEarthlyBranchIndex(地支)`（`utils/index.js`）：
  `宫位索引 = fixIndex(EARTHLY_BRANCHES.indexOf(地支) - EARTHLY_BRANCHES.indexOf(寅))`，即 **原始地支序 − 2（mod 12）**。
- 归一化函数 `fixIndex(index, max=12)`（`utils/index.js`）：递归把 index 收敛到 `[0, max-1]`（负数 +max，超界 −max，`-0` 归 0）。**注意 JS 的 `%` 对负数返回负值，源码所有可能为负的地方都套了 `fixIndex`，Go 里可以实现 `fixIndex(i) = ((i % 12) + 12) % 12`。**

### 0.2 时辰索引 timeIndex

`0=早子时(00:00~01:00)，1=丑时，2=寅时 … 11=亥时，12=晚子时(23:00~00:00)`（`data/constants.js` 的 `CHINESE_TIME`/`TIME_RANGE`，`utils/index.js` 的 `timeToIndex`）。

**timeIndex 使用规则（源码逐处核对）：**

| 用途 | 函数 | 用法 |
|---|---|---|
| 文昌文曲 | `getChangQuIndex` | `fixIndex(timeIndex)`（12→0，晚子时按子时支算） |
| 地空地劫 | `getKongJieIndex` | `fixIndex(timeIndex)` |
| 火星铃星 | `getHuoLingIndex` | `fixIndex(timeIndex)` |
| 台辅封诰 | `getTimelyStarIndex` | `fixIndex(timeIndex)` |
| 农历日索引 | `fixLunarDayIndex` | 原始 timeIndex，判断 `timeIndex >= 12` |
| 闰月折算 | `fixLunarMonthIndex` | 原始 timeIndex，判断 `timeIndex !== 12` |
| 紫微借日 | `getStartIndex` | 原始 timeIndex，判断 `timeIndex === 12` |
| 命宫身宫 | `getSoulAndBody` | 通过 lunar-lite 求时支后取 `EARTHLY_BRANCHES.indexOf(时支)`（晚子时时支为子，等效 12→0） |

### 0.3 农历月折算 `fixLunarMonthIndex`（utils/index.js）

```
monthIndex = fixIndex( lunarMonth + 1 - EARTHLY_BRANCHES.indexOf(寅=2) + (needToAdd ? 1 : 0) )
           = fixIndex( lunarMonth - 1 + (needToAdd ? 1 : 0) )
needToAdd  = isLeap && fixLeap && lunarDay > 15 && timeIndex !== 12
```

- 返回值是 **0 基**月索引（正月=0 … 十二月=11），同时也正是"正月建寅"后该月对应的宫位索引。
- 闰月处理：仅当 `fixLeap=true` 且当月是闰月且农历日 >15 且**不是晚子时**，才把月份 +1（闰月前 15 天按本月、之后按下月）。
- **凡下文提到"月索引 monthIndex"，一律指本函数的返回值（即 fixLeap 折算后的月），不是原始农历月。**

### 0.4 农历日索引 `fixLunarDayIndex`（utils/index.js）

```
dayIndex = (timeIndex >= 12) ? lunarDay : lunarDay - 1
```

初一为 1 的 `lunarDay` 转为 0 基偏移；晚子时（>=12）视作加一天所以不减一。用于三台八座恩光天贵。

### 0.5 全局配置默认值（astro/astro.js）

| 配置 | 默认值 | 含义 |
|---|---|---|
| `yearDivide` | `'normal'` | 年干支分界：normal=正月初一，exact=立春（`data/types/astro.d.ts`） |
| `horoscopeDivide` | `'normal'` | 运限/流耀年分界，同上 |
| `dayDivide` | `'forward'` | 晚子时归属：forward=算次日，current=算当天 |
| `algorithm` | `'default'` | `'default' | 'zhongzhou'`，影响天伤天使、截路空亡/截空、劫杀、大耗、龙德 |
| `mutagens` / `brightness` | 空 | 可覆盖四化表 / 亮度表（`getTargetMutagens`、`getBrightness`） |

年干支来源：主星/辅星/杂曜入口（`majorStar.js`、`minorStar.js`、`adjectiveStar.js`）均用 `getHeavenlyStemAndEarthlyBranchBySolarDate(solarDate, timeIndex, { year: yearDivide })` 的 `yearly = [年干, 年支]`；**但 `getYearlyStarIndex`（location.js）内部另行用 `{ year: horoscopeDivide }` 重新取年干支**（源码注释："流耀应该用立春为界，但为了满足不同流派的需求允许配置"）。默认配置下两者结果一致。

### 0.6 命宫身宫与五行局（astro/palace.js，紫微定位的前置依赖）

`getSoulAndBody(param)`：

```
monthIndex = fixLunarMonthIndex(solarDate, timeIndex, fixLeap)   // fixLeap 折算后
timeBranch = 时支的原始地支序（晚子时=子=0）
soulIndex（命宫） = fixIndex(monthIndex - timeBranch)
bodyIndex（身宫） = fixIndex(monthIndex + timeBranch)
```

命宫天干（五虎遁 `TIGER_RULE`，data/constants.js）：寅宫天干 = `TIGER_RULE[年干]`（甲己→丙，乙庚→戊，丙辛→庚，丁壬→壬，戊癸→甲）；
`命宫天干 = HEAVENLY_STEMS[ fixIndex(indexOf(寅宫天干) + soulIndex, 10) ]`；`命宫地支 = EARTHLY_BRANCHES[ fixIndex(soulIndex + 2) ]`。
若传入 `from`（指定干支起盘），命宫直接取 `fixEarthlyBranchIndex(from.earthlyBranch)`，身宫 = `fixIndex(bodyOffset[timeIndex] + soulIndex)`，`bodyOffset = [0,2,4,6,8,10,0,2,4,6,8,10,0]`。

`getFiveElementsClass(命宫天干, 命宫地支)`（astro/palace.js）：

```
天干数 = floor(HEAVENLY_STEMS.indexOf(干) / 2) + 1        // 甲乙1 丙丁2 戊己3 庚辛4 壬癸5
地支数 = floor(fixIndex(EARTHLY_BRANCHES.indexOf(支), 6) / 2) + 1  // 子丑1 寅卯2 辰巳3 午未1 申酉2 戌亥3
index = 天干数 + 地支数; while (index > 5) index -= 5
五行局 = ['木三局','金四局','水二局','火六局','土五局'][index - 1]
```

`FiveElementsClass` 数值（data/constants.js）：水二局=2，木三局=3，金四局=4，土五局=5，火六局=6。

---

## 1. 紫微星定位（`getStartIndex`，star/location.js）

输入：`solarDate, timeIndex, fixLeap, from?`。

1. 取命宫干支（`getSoulAndBody`，见 0.6）；若传入 `from.heavenlyStem/earthlyBranch` 则用传入干支替代命宫干支。
2. 五行局数值 `f = FiveElementsClass[getFiveElementsClass(干, 支)]`（2/3/4/5/6）。
3. 农历日 `lunarDay = solar2lunar(solarDate).lunarDay`（初一=1）。
4. **晚子时加日**：`_day = (timeIndex === 12 && dayDivide !== 'current') ? lunarDay + 1 : lunarDay`。
5. **跨月借日**：`maxDays = getTotalDaysOfLunarMonth(solarDate)`（当月农历总天数）；若 `_day > maxDays` 则 `_day -= maxDays`。
6. **虚日（借日凑整除）循环**：

```
offset = -1
do {
    offset++
    divisor  = _day + offset
    quotient = floor(divisor / f)
    remainder = divisor % f
} while (remainder != 0)
```

即 offset 是让 `_day + offset` 恰能被局数整除的最小非负补数。

7. 定位：

```
quotient %= 12
ziweiIndex = quotient - 1                 // 商从寅起数，减一转 0 基
if (offset % 2 == 0)  ziweiIndex += offset   // 补数为偶数：顺时针进 offset 格
else                  ziweiIndex -= offset   // 补数为奇数：逆时针退 offset 格
ziweiIndex = fixIndex(ziweiIndex)
```

（注意：源码第 64-70 行的两条注释文字互相矛盾且与代码相反，**以代码为准**：偶数加、奇数减，与文件头口诀示例一致——27日木三局→戌；13日火六局补 5（奇）逆回五宫→亥；6日土五局补 4（偶）顺行 4 格→未。）

8. 天府：`tianfuIndex = fixIndex(12 - ziweiIndex)`（紫微与天府关于寅申轴对称；紫微在寅 0 时 12−0=12→fixIndex→0，天府同在寅）。

---

## 2. 十四主星（`getMajorStar`，star/majorStar.js）

年干支用 `yearDivide` 配置取得（用于四化）。两组偏移数组（下标 i 即偏移量，空串表示该偏移无星）：

### 2.1 紫微系六星 —— 自紫微**逆时针**（索引相减）

`ziweiGroup = ['ziweiMaj','tianjiMaj','','taiyangMaj','wuquMaj','tiantongMaj','','','lianzhenMaj']`

| 星 | 安放位置 |
|---|---|
| 紫微 | `fixIndex(ziweiIndex - 0)` |
| 天机 | `fixIndex(ziweiIndex - 1)` |
| 太阳 | `fixIndex(ziweiIndex - 3)` |
| 武曲 | `fixIndex(ziweiIndex - 4)` |
| 天同 | `fixIndex(ziweiIndex - 5)` |
| 廉贞 | `fixIndex(ziweiIndex - 8)` |

### 2.2 天府系八星 —— 自天府**顺时针**（索引相加）

`tianfuGroup = ['tianfuMaj','taiyinMaj','tanlangMaj','jumenMaj','tianxiangMaj','tianliangMaj','qishaMaj','','','','pojunMaj']`

| 星 | 安放位置 |
|---|---|
| 天府 | `fixIndex(tianfuIndex + 0)` |
| 太阴 | `fixIndex(tianfuIndex + 1)` |
| 贪狼 | `fixIndex(tianfuIndex + 2)` |
| 巨门 | `fixIndex(tianfuIndex + 3)` |
| 天相 | `fixIndex(tianfuIndex + 4)` |
| 天梁 | `fixIndex(tianfuIndex + 5)` |
| 七杀 | `fixIndex(tianfuIndex + 6)` |
| 破军 | `fixIndex(tianfuIndex + 10)` |

每颗主星以 `{name, type:'major', scope:'origin', brightness:getBrightness(名, 宫位索引), mutagen:getMutagen(名, 年干)}` 入宫。

---

## 3. 十四辅星 minorStars（`getMinorStar`，star/minorStar.js）

入口取 `yearly`（`yearDivide` 配置）、`monthIndex = fixLunarMonthIndex(...)`（**fixLeap 折算后**）。

### 3.1 左辅右弼（按月，`getZuoYouIndex`，location.js）

调用方式：`getZuoYouIndex(monthIndex + 1)`（参数是 1 基月份）。

```
zuoIndex = fixIndex( fixEarthlyBranchIndex(辰=2) + (lunarMonth - 1) )   // 辰上起正月，顺行
youIndex = fixIndex( fixEarthlyBranchIndex(戌=8) - (lunarMonth - 1) )   // 戌上起正月，逆行
```

即左辅 = `fixIndex(2 + monthIndex)`，右弼 = `fixIndex(8 - monthIndex)`。**用折算月，不是原始农历月。**

### 3.2 文昌文曲（按时支，`getChangQuIndex`，location.js）

```
changIndex = fixIndex( 8 - fixIndex(timeIndex) )   // 戌上起子时，逆数至生时 → 文昌
quIndex    = fixIndex( 2 + fixIndex(timeIndex) )   // 辰上起子时，顺数至生时 → 文曲
```

**以时起，不以月起**；晚子时 `fixIndex(12)=0` 按子时算。

### 3.3 天魁天钺（年干表，`getKuiYueIndex`，location.js）

| 年干 | 天魁 | 天钺 |
|---|---|---|
| 甲、戊、庚 | 丑(11) | 未(5) |
| 乙、己 | 子(10) | 申(6) |
| 辛 | 午(4) | 寅(0) |
| 丙、丁 | 亥(9) | 酉(7) |
| 壬、癸 | 卯(1) | 巳(3) |

### 3.4 禄存、擎羊、陀罗（年干表，`getLuYangTuoMaIndex`，location.js）

| 年干 | 禄存 |
|---|---|
| 甲 | 寅(0) |
| 乙 | 卯(1) |
| 丙、戊 | 巳(3) |
| 丁、己 | 午(4) |
| 庚 | 申(6) |
| 辛 | 酉(7) |
| 壬 | 亥(9) |
| 癸 | 子(10) |

```
yangIndex（擎羊） = fixIndex(luIndex + 1)   // 禄前
tuoIndex（陀罗） = fixIndex(luIndex - 1)   // 禄后
```

### 3.5 天马（年支三合表，同函数 `getLuYangTuoMaIndex`）

| 年支 | 天马 |
|---|---|
| 寅、午、戌 | 申(6) |
| 申、子、辰 | 寅(0) |
| 巳、酉、丑 | 亥(9) |
| 亥、卯、未 | 巳(3) |

### 3.6 火星铃星（年支三合定起点 + 时辰顺数，`getHuoLingIndex`，location.js）

| 年支 | 火星起点 | 铃星起点 |
|---|---|---|
| 寅、午、戌 | 丑(11) | 卯(1) |
| 申、子、辰 | 寅(0) | 戌(8) |
| 巳、酉、丑 | 卯(1) | 戌(8) |
| 亥、卯、未 | 酉(7) | 戌(8) |

```
huoIndex  = fixIndex( 起点 + fixIndex(timeIndex) )
lingIndex = fixIndex( 起点 + fixIndex(timeIndex) )
```

**iztro 不区分男女、不逆行**：一律从起点顺数至生时（部分流派的"阳男阴女顺、阴男阳女逆"未实现）。

### 3.7 地空地劫（按时支，`getKongJieIndex`，location.js）

```
kongIndex（地空） = fixIndex( 9 - fixIndex(timeIndex) )   // 亥上起子时逆行
jieIndex（地劫） = fixIndex( 9 + fixIndex(timeIndex) )   // 亥上起子时顺行
```

---

## 4. 杂曜 adjectiveStars（`getAdjectiveStar`，star/adjectiveStar.js）

依赖四个索引组：`getYearlyStarIndex`（**内部年干支用 horoscopeDivide**）、`getMonthlyStarIndex`、`getDailyStarIndex`、`getTimelyStarIndex`（均在 location.js），外加 `getLuanXiIndex`（红鸾天喜，用入口的 yearDivide 年支）与 `getYearly12().suiqian12`（龙德，中州派，decorativeStar.js，timeIndex=0、horoscopeDivide）。

下文约定：`YB = EARTHLY_BRANCHES.indexOf(年支)`（原始序，子=0）、`HS = HEAVENLY_STEMS.indexOf(年干)`（甲=0）、`monthIndex` 为 0 基折算月、`soulIndex/bodyIndex` 为命宫/身宫索引。**逐星列表（一颗不漏，共 38 颗 + 派别分支）：**

### 4.1 年支系（`getYearlyStarIndex` / `getLuanXiIndex` / `getHuagaiXianchiIndex` / `getGuGuaIndex` / `getJieshaAdjIndex` / `getDahaoIndex` / `getNianjieIndex` / `getTianshiTianshangIndex`）

| 星 | 公式 / 完整查表（宫位索引） | 源码函数 |
|---|---|---|
| 红鸾 | `fixIndex(1 - YB)`。查表：子→卯(1) 丑→寅(0) 寅→丑(11) 卯→子(10) 辰→亥(9) 巳→戌(8) 午→酉(7) 未→申(6) 申→未(5) 酉→午(4) 戌→巳(3) 亥→辰(2) | `getLuanXiIndex` |
| 天喜 | `fixIndex(红鸾 + 6)`（对宫） | `getLuanXiIndex` |
| 华盖 | 申子辰→辰(2)；寅午戌→戌(8)；巳酉丑→丑(11)；亥卯未→未(5) | `getHuagaiXianchiIndex` |
| 咸池 | 申子辰→酉(7)；寅午戌→卯(1)；巳酉丑→午(4)；亥卯未→子(10) | `getHuagaiXianchiIndex` |
| 孤辰 | 寅卯辰→巳(3)；巳午未→申(6)；申酉戌→亥(9)；亥子丑→寅(0) | `getGuGuaIndex` |
| 寡宿 | 寅卯辰→丑(11)；巳午未→辰(2)；申酉戌→未(5)；亥子丑→戌(8) | `getGuGuaIndex` |
| 破碎 | `['巳','丑','酉'][YB % 3]` → 子午卯酉→巳(3)；丑辰未戌→丑(11)；寅巳申亥→酉(7) | `getYearlyStarIndex`（posui） |
| 蜚廉 | 查表（按年支子→亥）：`['申','酉','戌','巳','午','未','寅','卯','辰','亥','子','丑'][YB]`，即 子→申(6) 丑→酉(7) 寅→戌(8) 卯→巳(3) 辰→午(4) 巳→未(5) 午→寅(0) 未→卯(1) 申→辰(2) 酉→亥(9) 戌→子(10) 亥→丑(11) | `getYearlyStarIndex`（feilian） |
| 龙池 | `fixIndex(2 + YB)`（辰宫起子顺数至年支） | `getYearlyStarIndex`（longchi） |
| 凤阁 | `fixIndex(8 - YB)`（戌宫起子逆数至年支） | `getYearlyStarIndex`（fengge） |
| 天哭 | `fixIndex(4 - YB)`（午宫起子逆行） | `getYearlyStarIndex`（tianku） |
| 天虚 | `fixIndex(4 + YB)`（午宫起子顺行） | `getYearlyStarIndex`（tianxu） |
| 天德 | `fixIndex(7 + YB)`（酉宫起子顺行） | `getYearlyStarIndex`（tiande） |
| 月德 | `fixIndex(3 + YB)`（巳宫起子顺行） | `getYearlyStarIndex`（yuede） |
| 天才 | `fixIndex(soulIndex + YB)`（命宫起子顺数至年支） | `getYearlyStarIndex`（tiancai） |
| 天寿 | `fixIndex(bodyIndex + YB)`（身宫起子顺数至年支） | `getYearlyStarIndex`（tianshou） |
| 天空 | `fixIndex(fixEarthlyBranchIndex(年支) + 1)`（年支宫位的下一宫） | `getYearlyStarIndex`（tiankong） |
| 年解 | 查表（按年支子→亥）：`['戌','酉','申','未','午','巳','辰','卯','寅','丑','子','亥'][YB]`，即 子→戌(8) 丑→酉(7) 寅→申(6) 卯→未(5) 辰→午(4) 巳→巳(3) 午→辰(2) 未→卯(1) 申→寅(0) 酉→丑(11) 戌→子(10) 亥→亥(9)（戌上起子逆数至年支） | `getNianjieIndex` |
| 天伤 | `fixIndex(5 + soulIndex)`（仆役/交友宫，`PALACES.indexOf('friendsPalace')=5`）；**中州派且阴男/阳女时与天使互换** | `getTianshiTianshangIndex` |
| 天使 | `fixIndex(7 + soulIndex)`（疾厄宫，`PALACES.indexOf('healthPalace')=7`）；互换规则同上。判定：`sameYinyang = (YB % 2) === indexOf(gender in ['male','female'])`（年支阳=0 配男=0），`algorithm==='zhongzhou' && !sameYinyang` 时 swap | `getTianshiTianshangIndex` |
| 劫杀（中州派限定） | 申子辰→3(巳)；亥卯未→6(申)；寅午戌→9(亥)；巳酉丑→0(寅)（函数直接返回宫位索引常数） | `getJieshaAdjIndex` |
| 大耗（中州派限定，杂曜版） | 先按年支查对映支：子→未 丑→午 寅→酉 卯→申 辰→亥 巳→戌 午→丑 未→子 申→卯 酉→寅 戌→巳 亥→辰；宫位 = `fixIndex(EARTHLY_BRANCHES.indexOf(对映支) - 2)`。即：子→未(5) 丑→午(4) 寅→酉(7) 卯→申(6) 辰→亥(9) 巳→戌(8) 午→丑(11) 未→子(10) 申→卯(1) 酉→寅(0) 戌→巳(3) 亥→辰(2) | `getDahaoIndex` |

### 4.2 年干系（`getYearlyStarIndex` 内查表，下标 = HS，甲→癸）

| 星 | 完整查表（甲 乙 丙 丁 戊 己 庚 辛 壬 癸） | 源码变量 |
|---|---|---|
| 天厨 | 巳(3) 午(4) 子(10) 巳(3) 午(4) 申(6) 寅(0) 午(4) 酉(7) 亥(9) | `tianchuIndex`，表 `['si','woo','zi','si','woo','shen','yin','woo','you','hai']` |
| 天官 | 未(5) 辰(2) 巳(3) 寅(0) 卯(1) 酉(7) 亥(9) 酉(7) 戌(8) 午(4) | `tianguanIndex`，表 `['wei','chen','si','yin','mao','you','hai','you','xu','woo']` |
| 天福 | 酉(7) 申(6) 子(10) 亥(9) 卯(1) 寅(0) 午(4) 巳(3) 午(4) 巳(3) | `tianfuIndex`，表 `['you','shen','zi','hai','mao','yin','woo','si','woo','si']` |
| 截路（非中州派） | `['申','午','辰','寅','子'][HS % 5]` → 甲己→申(6)；乙庚→午(4)；丙辛→辰(2)；丁壬→寅(0)；戊癸→子(10) | `jieluIndex` |
| 空亡（非中州派） | `['酉','未','巳','卯','丑'][HS % 5]` → 甲己→酉(7)；乙庚→未(5)；丙辛→巳(3)；丁壬→卯(1)；戊癸→丑(11) | `kongwangIndex` |
| 截空（中州派限定） | `年支阳（YB%2==0）? 截路索引 : 空亡索引`（"生年阳干在阳宫，阴干在阴宫"——实际按**年支**阴阳取） | `jiekongIndex` |
| 旬空 | `xunkong = fixIndex( fixEarthlyBranchIndex(年支) + 9 - HS + 1 )`（9 = `HEAVENLY_STEMS.indexOf(癸)`）；然后若 `(YB % 2) !== (xunkong % 2)` 则 `xunkong = fixIndex(xunkong + 1)`（年支阴阳与宫位阴阳对齐） | `xunkongIndex` |

### 4.3 月系（`getMonthlyStarIndex`，monthIndex 为 0 基折算月，正月=0）

| 星 | 公式 / 完整查表 | 源码变量 |
|---|---|---|
| 解神（月解） | `['申','戌','子','寅','辰','午'][floor(monthIndex / 2)]` → 正二月→申(6)；三四月→戌(8)；五六月→子(10)；七八月→寅(0)；九十月→辰(2)；十一十二月→午(4) | `jieshenIndex`（返回名 `yuejieIndex`） |
| 天姚 | `fixIndex(11 + monthIndex)`（丑宫起正月顺行） | `tianyaoIndex` |
| 天刑 | `fixIndex(7 + monthIndex)`（酉宫起正月顺行） | `tianxingIndex` |
| 阴煞 | `['寅','子','戌','申','午','辰'][monthIndex % 6]` → 正/七月→寅(0)；二/八月→子(10)；三/九月→戌(8)；四/十月→申(6)；五/十一月→午(4)；六/十二月→辰(2) | `yinshaIndex` |
| 天月 | 查表（正月→十二月）：`['戌','巳','辰','寅','未','卯','亥','未','寅','午','戌','寅'][monthIndex]`，即 正→戌(8) 二→巳(3) 三→辰(2) 四→寅(0) 五→未(5) 六→卯(1) 七→亥(9) 八→未(5) 九→寅(0) 十→午(4) 十一→戌(8) 十二→寅(0) | `tianyueIndex` |
| 天巫 | `['巳','申','寅','亥'][monthIndex % 4]` → 正五九月→巳(3)；二六十月→申(6)；三七十一月→寅(0)；四八十二月→亥(9) | `tianwuIndex` |

### 4.4 日系（`getDailyStarIndex`，依赖左辅右弼/文昌文曲位置 + 农历日）

```
monthIndex = fixLunarMonthIndex(solarDate, timeIndex, fixLeap)          // 折算月
{zuoIndex, youIndex} = getZuoYouIndex(monthIndex + 1)
{changIndex, quIndex} = getChangQuIndex(timeIndex)
dayIndex = fixLunarDayIndex(lunarDay, timeIndex)                        // 晚子时不减一

santaiIndex（三台）  = fixIndex( (zuoIndex  + dayIndex) % 12 )          // 左辅起初一顺行至生日
bazuoIndex（八座）   = fixIndex( (youIndex  - dayIndex) % 12 )          // 右弼起初一逆行至生日
enguangIndex（恩光） = fixIndex( ((changIndex + dayIndex) % 12) - 1 )   // 文昌起初一顺行至生日退一步
tianguiIndex（天贵） = fixIndex( ((quIndex   + dayIndex) % 12) - 1 )    // 文曲起初一顺行至生日退一步
```

（JS 的 `%` 可产生负值，`fixIndex` 负责归一；Go 移植时直接用 `((x % 12) + 12) % 12`。）

### 4.5 时系（`getTimelyStarIndex`）

```
taifuIndex（台辅） = fixIndex( 4 + fixIndex(timeIndex) )   // 午宫起子时顺行
fenggaoIndex（封诰） = fixIndex( 0 + fixIndex(timeIndex) )  // 寅宫起子时顺行
```

### 4.6 龙德（中州派限定，adjectiveStar.js + decorativeStar.js `getYearly12`）

`getYearly12(solarDate)`（timeIndex 固定 0，年分界用 `horoscopeDivide`）从**年支对应宫位**起"岁建"顺行安岁前 12 神；中州派顺序为 `岁建 晦气 丧门 贯索 官符 小耗 岁破 龙德 白虎 天德 吊客 病符`（非中州派把"岁破"换成"大耗"，其余相同）。龙德排第 8（偏移 7），故：

```
龙德索引 = suiqian12.indexOf('龙德') = fixIndex( fixEarthlyBranchIndex(年支) + 7 )
```

---

## 5. 星耀归属与 type 字段（`StarType`，data/types/general.d.ts）

`type StarType = 'major' | 'soft' | 'tough' | 'adjective' | 'flower' | 'helper' | 'lucun' | 'tianma'`；本命星 `scope` 一律 `'origin'`。

### 5.1 majorStars（majorStar.js）

紫微 天机 太阳 武曲 天同 廉贞 天府 太阴 贪狼 巨门 天相 天梁 七杀 破军 —— 全部 `type:'major'`，带 `brightness` 与 `mutagen` 字段。

### 5.2 minorStars（minorStar.js，共 14 颗）

| type | 星 | 附带字段 |
|---|---|---|
| `soft` | 左辅、右弼、文昌、文曲 | brightness + mutagen |
| `soft` | 天魁、天钺 | 仅 brightness（无 mutagen 字段） |
| `lucun` | 禄存 | 仅 brightness |
| `tianma` | 天马 | 仅 brightness |
| `tough` | 地空、地劫、火星、铃星、擎羊、陀罗 | 仅 brightness |

### 5.3 adjectiveStars（adjectiveStar.js，均无 brightness/mutagen 字段）

| type | 星 |
|---|---|
| `flower` | 红鸾、天喜、天姚、咸池 |
| `helper` | 解神（月解）、年解 |
| `adjective` | 三台、八座、恩光、天贵、龙池、凤阁、天才、天寿、台辅、封诰、天巫、华盖、天官、天福、天厨、天月、天德、月德、天空、旬空、孤辰、寡宿、蜚廉、破碎、天刑、阴煞、天哭、天虚、天使、天伤 |
| `adjective`（仅非中州派） | 截路、空亡 |
| `adjective`（仅中州派） | 龙德、截空、劫杀、大耗 |

（流耀 horoscopeStar.js 供参考：运/流魁钺昌曲=`soft`，流禄=`lucun`，流羊流陀=`tough`，流马=`tianma`，流鸾流喜=`flower`，流年年解=`helper`。）

---

## 6. 亮度 brightness

### 6.1 机制（`getBrightness`，utils/index.js）

`getBrightness(星名, 宫位索引)`：优先取 `getConfig().brightness[星]` 覆盖表，否则取 `data/stars.js` 的 `STARS_INFO[星].brightness`；无表则返回空串。返回 `brightness[fixIndex(index)]` 的翻译值。

### 6.2 下标基准（源码证据）

`data/stars.js` 第 9 行注释：**"亮度（bright）, 按照宫位地支排序（从寅开始）"**——即**下标 0 = 寅宫**，与宫位索引一致，`getBrightness` 直接用宫位索引查数组，无任何换算。

### 6.3 有亮度表的星（仅以下 20 颗；键值 miao=庙 wang=旺 de=得 li=利 ping=平 bu=不 xian=陷，zh-CN 翻译见 `i18n/locales/zh-CN/brightness.js`）

下表按宫位索引 0→11（寅 卯 辰 巳 午 未 申 酉 戌 亥 子 丑）完整抄录：

| 星 | 寅 | 卯 | 辰 | 巳 | 午 | 未 | 申 | 酉 | 戌 | 亥 | 子 | 丑 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 紫微 | 旺 | 旺 | 得 | 旺 | 庙 | 庙 | 旺 | 旺 | 得 | 旺 | 平 | 庙 |
| 天机 | 得 | 旺 | 利 | 平 | 庙 | 陷 | 得 | 旺 | 利 | 平 | 庙 | 陷 |
| 太阳 | 旺 | 庙 | 旺 | 旺 | 旺 | 得 | 得 | 陷 | 不 | 陷 | 陷 | 不 |
| 武曲 | 得 | 利 | 庙 | 平 | 旺 | 庙 | 得 | 利 | 庙 | 平 | 旺 | 庙 |
| 天同 | 利 | 平 | 平 | 庙 | 陷 | 不 | 旺 | 平 | 平 | 庙 | 旺 | 不 |
| 廉贞 | 庙 | 平 | 利 | 陷 | 平 | 利 | 庙 | 平 | 利 | 陷 | 平 | 利 |
| 天府 | 庙 | 得 | 庙 | 得 | 旺 | 庙 | 得 | 旺 | 庙 | 得 | 庙 | 庙 |
| 太阴 | 旺 | 陷 | 陷 | 陷 | 不 | 不 | 利 | 不 | 旺 | 庙 | 庙 | 庙 |
| 贪狼 | 平 | 利 | 庙 | 陷 | 旺 | 庙 | 平 | 利 | 庙 | 陷 | 旺 | 庙 |
| 巨门 | 庙 | 庙 | 陷 | 旺 | 旺 | 不 | 庙 | 庙 | 陷 | 旺 | 旺 | 不 |
| 天相 | 庙 | 陷 | 得 | 得 | 庙 | 得 | 庙 | 陷 | 得 | 得 | 庙 | 庙 |
| 天梁 | 庙 | 庙 | 庙 | 陷 | 庙 | 旺 | 陷 | 得 | 庙 | 陷 | 庙 | 旺 |
| 七杀 | 庙 | 旺 | 庙 | 平 | 旺 | 庙 | 庙 | 庙 | 庙 | 平 | 旺 | 庙 |
| 破军 | 得 | 陷 | 旺 | 平 | 庙 | 旺 | 得 | 陷 | 旺 | 平 | 庙 | 旺 |
| 文昌 | 陷 | 利 | 得 | 庙 | 陷 | 利 | 得 | 庙 | 陷 | 利 | 得 | 庙 |
| 文曲 | 平 | 旺 | 得 | 庙 | 陷 | 旺 | 得 | 庙 | 陷 | 旺 | 得 | 庙 |
| 火星 | 庙 | 利 | 陷 | 得 | 庙 | 利 | 陷 | 得 | 庙 | 利 | 陷 | 得 |
| 铃星 | 庙 | 利 | 陷 | 得 | 庙 | 利 | 陷 | 得 | 庙 | 利 | 陷 | 得 |
| 擎羊 | （空） | 陷 | 庙 | （空） | 陷 | 庙 | （空） | 陷 | 庙 | （空） | 陷 | 庙 |
| 陀罗 | 陷 | （空） | 庙 | 陷 | （空） | 庙 | 陷 | （空） | 庙 | 陷 | （空） | 庙 |

- 擎羊/陀罗表中的空串保留为空（源码即空串 `''`）。
- **左辅、右弼、天魁、天钺、禄存、天马、地空、地劫在 `STARS_INFO` 中无条目**：minorStar.js 虽调用了 `getBrightness`，但返回空串，故这 8 颗辅星亮度恒为 `''`。所有杂曜不调用 `getBrightness`。

---

## 7. 四化 mutagen

### 7.1 十天干四化完整表（`data/heavenlyStems.js`，顺序【禄，权，科，忌】；`MUTAGEN = ['sihuaLu','sihuaQuan','sihuaKe','sihuaJi']` → 禄/权/科/忌，data/stars.js + i18n/locales/zh-CN/mutagen.js）

| 年干 | 化禄 | 化权 | 化科 | 化忌 |
|---|---|---|---|---|
| 甲 | 廉贞 | 破军 | 武曲 | 太阳 |
| 乙 | 天机 | 天梁 | 紫微 | 太阴 |
| 丙 | 天同 | 天机 | 文昌 | 廉贞 |
| 丁 | 太阴 | 天同 | 天机 | 巨门 |
| 戊 | 贪狼 | 太阴 | 右弼 | 天机 |
| 己 | 武曲 | 贪狼 | 天梁 | 文曲 |
| 庚 | 太阳 | 武曲 | 太阴 | 天同 |
| 辛 | 巨门 | 太阳 | 文曲 | 文昌 |
| 壬 | 天梁 | 紫微 | 左辅 | 武曲 |
| 癸 | 破军 | 巨门 | 太阴 | 贪狼 |

### 7.2 挂载方式（`getMutagen` / `getTargetMutagens`，utils/index.js）

- `getMutagen(星名, 年干)`：取该年干的四化数组 `target`（优先 `getConfig().mutagens[年干]` 覆盖，否则 `heavenlyStems[年干].mutagen`），返回 `MUTAGEN[target.indexOf(星)]` 的翻译（'禄'/'权'/'科'/'忌'）；星不在数组中时 `indexOf=-1`，结果为 `undefined`（即无四化）。
- 只有 **14 主星（majorStar.js）与 左辅、右弼、文昌、文曲（minorStar.js）** 在安星时调用 `getMutagen` 并写入 `mutagen` 字段；其余星（含天魁天钺禄存天马火铃羊陀空劫及全部杂曜）不带该字段。
- 所用年干：majorStar/minorStar 的 `yearly[0]`，按 `yearDivide` 配置取得（默认 normal=正月初一分界）。

---

## 8. 月份/时辰使用细节汇总（易错点清单）

1. **折算月 vs 原始农历月**：左辅右弼（3.1）、三台八座恩光天贵的左右起点（4.4）、全部月系杂曜（4.3）、命宫身宫（0.6，进而影响五行局与紫微）都用 `fixLunarMonthIndex` 的**折算月**；**没有任何安星公式直接用原始 lunarMonth**（原始月只作为 `fixLunarMonthIndex` 的输入）。
2. **紫微用的日**：原始 `lunarDay`，加两条修正——晚子时（`timeIndex===12` 且 `dayDivide!=='current'`）+1 天；超过当月农历总天数则减去总天数（跨月借日）。之后再做"虚日 offset"循环（第 1 节）。
3. **三台八座恩光天贵用的日**：`fixLunarDayIndex = timeIndex>=12 ? lunarDay : lunarDay-1`（晚子时等效加一天）。
4. **timeIndex**：所有"顺/逆数至生时"的公式先做 `fixIndex(timeIndex)`（12→0）；只有晚子时判断（`===12`/`>=12`/`!==12`）用原始值。见 0.2 表。
5. **年干支的两个口径**：入口层（major/minor/adjective）用 `yearDivide`；`getYearlyStarIndex` 内部与 `getYearly12`（龙德）用 `horoscopeDivide`。于是**同一张盘里红鸾天喜（入口年支）与其余年系杂曜（horoscopeDivide 年支）理论上可能用不同年支**——默认配置两者一致，Go 移植必须保留这两条独立取值路径。
6. **命宫身宫时支**：`getSoulAndBody` 用 lunar-lite 返回的时支（晚子时=子），等效 `fixIndex(timeIndex)`。
7. **fixIndex 是逻辑取模**：Go 中所有下标运算需用 `((x % 12) + 12) % 12`，尤其是 `bazuoIndex`（右弼减日数）与红鸾（1 − 年支序）这类必然出负数的公式。
8. **中州派开关**（`algorithm==='zhongzhou'`）影响：天伤天使互换条件（4.1）、截路/空亡 → 龙德+截空+劫杀+大耗（4.1/4.2/4.6）。

---

## 9. 附：流昌流曲（`getChangQuIndexByHeavenlyStem`，location.js，供运限星移植参考）

按（大限/流年）天干：

| 天干 | 流昌 | 流曲 |
|---|---|---|
| 甲 | 巳(3) | 酉(7) |
| 乙 | 午(4) | 申(6) |
| 丙、戊 | 申(6) | 午(4) |
| 丁、己 | 酉(7) | 巳(3) |
| 庚 | 亥(9) | 卯(1) |
| 辛 | 子(10) | 寅(0) |
| 壬 | 寅(0) | 子(10) |
| 癸 | 卯(1) | 亥(9) |

流耀（运/流魁钺昌曲禄羊陀马鸾喜、流年年解）复用本文 3.3/3.4/3.5/4.1 的同名函数，见 `star/horoscopeStar.js`。

---

## 附录：本规格所读源码文件清单

均位于 iztro@2.5.8 npm 包的 `lib/` 目录(编译后 JS):

- `star/location.js`（核心：getStartIndex、getLuYangTuoMaIndex、getKuiYueIndex、getZuoYouIndex、getChangQuIndex、getDailyStarIndex、getTimelyStarIndex、getKongJieIndex、getHuoLingIndex、getLuanXiIndex、getHuagaiXianchiIndex、getGuGuaIndex、getJieshaAdjIndex、getDahaoIndex、getYearlyStarIndex、getTianshiTianshangIndex、getNianjieIndex、getMonthlyStarIndex、getChangQuIndexByHeavenlyStem）
- `star/majorStar.js`（getMajorStar）
- `star/minorStar.js`（getMinorStar）
- `star/adjectiveStar.js`（getAdjectiveStar）
- `star/index.js`（initStars 与模块导出）
- `star/star.js`（旧版实现，未被 index.js 导出，仅核对用）
- `star/decorativeStar.js`(getYearly12、getBoShi12、getchangsheng12、getJiangqian12StartIndex、getChangesheng12StartIndex)
- `star/horoscopeStar.js`（getHoroscopeStar，流耀）
- `data/constants.js`（HEAVENLY_STEMS、EARTHLY_BRANCHES、PALACES、GENDER、FiveElementsClass、CHINESE_TIME、TIME_RANGE、TIGER_RULE、RAT_RULE）
- `data/stars.js`（MUTAGEN、STARS_INFO 亮度表）
- `data/heavenlyStems.js`（十天干四化表）
- `data/earthlyBranches.js`（地支阴阳/命主身主）
- `data/types/general.d.ts`（StarType 定义）、`data/types/astro.d.ts`（yearDivide/horoscopeDivide/dayDivide 语义）
- `utils/index.js`（fixIndex、fixEarthlyBranchIndex、fixLunarMonthIndex、fixLunarDayIndex、getBrightness、getMutagen、timeToIndex）
- `astro/palace.js`（getSoulAndBody、getFiveElementsClass）
- `astro/astro.js`（getConfig 默认值：yearDivide/horoscopeDivide='normal'、dayDivide='forward'、algorithm='default'）
- `astro/astro.d.ts`
- `i18n/locales/zh-CN/star.js`、`i18n/locales/zh-CN/brightness.js`、`i18n/locales/zh-CN/mutagen.js`（key→中文名映射）
- `package.json`（版本确认 2.5.8）
