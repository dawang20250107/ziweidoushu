/**
 * 黄金基准数据生成器
 * 用 iztro 2.5.8 (与线上口径一致) 生成完整命盘 JSON,供 Go 引擎逐字段校验。
 * 同时记录 lunar-javascript 的农历转换结果做交叉验证。
 */
const { astro } = require('iztro');
const { Solar } = require('lunar-javascript');
const fs = require('fs');

function lunarOf(y, m, d) {
  const solar = Solar.fromYmd(y, m, d);
  const lunar = solar.getLunar();
  return {
    lunarYear: lunar.getYear(),
    lunarMonth: lunar.getMonth(), // negative = leap
    lunarDay: lunar.getDay(),
    yearGan: lunar.getYearGan(),
    yearZhi: lunar.getYearZhi(),
  };
}

function genCase(y, m, d, hour, gender) {
  const astrolabe = astro.bySolar(`${y}-${m}-${d}`, hour, gender === 'male' ? '男' : '女', true, 'zh-CN');
  const palaces = astrolabe.palaces.map((p) => ({
    index: p.index,
    name: p.name,
    isBodyPalace: p.isBodyPalace,
    isOriginalPalace: p.isOriginalPalace,
    heavenlyStem: p.heavenlyStem,
    earthlyBranch: p.earthlyBranch,
    majorStars: p.majorStars.map((s) => ({ name: s.name, brightness: s.brightness || '', mutagen: s.mutagen || '' })),
    minorStars: p.minorStars.map((s) => ({ name: s.name, type: s.type, brightness: s.brightness || '', mutagen: s.mutagen || '' })),
    adjectiveStars: p.adjectiveStars.map((s) => ({ name: s.name })),
    changsheng12: p.changsheng12,
    boshi12: p.boshi12,
    decadal: { range: p.decadal.range, heavenlyStem: p.decadal.heavenlyStem, earthlyBranch: p.decadal.earthlyBranch },
    ages: p.ages,
  }));
  return {
    input: { year: y, month: m, day: d, hour, gender },
    solarDate: astrolabe.solarDate,
    lunarDate: astrolabe.lunarDate,
    chineseDate: astrolabe.chineseDate,
    time: astrolabe.time,
    sign: astrolabe.sign,
    zodiac: astrolabe.zodiac,
    earthlyBranchOfSoulPalace: astrolabe.earthlyBranchOfSoulPalace,
    earthlyBranchOfBodyPalace: astrolabe.earthlyBranchOfBodyPalace,
    soul: astrolabe.soul,
    body: astrolabe.body,
    fiveElementsClass: astrolabe.fiveElementsClass,
    lunarJS: lunarOf(y, m, d),
    palaces,
  };
}

const cases = [];
const daysInMonth = (y, m) => new Date(y, m, 0).getDate();

// 1) 系统化扫描:1924-2044 每隔 3 年,每年取 4 个月 x 3 天,时辰/性别轮转
let k = 0;
for (let y = 1924; y <= 2044; y += 3) {
  for (const m of [1, 2, 5, 8, 11]) {
    for (const d of [1, 9, 15, 28]) {
      const dd = Math.min(d, daysInMonth(y, m));
      const hour = k % 12;
      const gender = k % 2 === 0 ? 'male' : 'female';
      cases.push([y, m, dd, hour, gender]);
      k++;
    }
  }
}

// 2) 闰月覆盖:扫描 1900-2050,凡 lunar-javascript 判定为闰月的日期抽样纳入
for (let y = 1900; y <= 2050; y++) {
  for (let m = 1; m <= 12; m++) {
    for (const d of [3, 12, 21, 27]) {
      if (d > daysInMonth(y, m)) continue;
      const l = lunarOf(y, m, d);
      if (l.lunarMonth < 0) {
        cases.push([y, m, d, (y + d) % 12, (y + m + d) % 2 === 0 ? 'male' : 'female']);
      }
    }
  }
}

// 3) 农历年界与月首月末边界(春节前后、除夕、初一、月底 29/30)
const boundaries = [
  [2000, 2, 4], [2000, 2, 5], [2000, 2, 6], // 2000 春节 2/5
  [1984, 2, 1], [1984, 2, 2], [1984, 2, 3], // 1984 春节 2/2
  [2024, 2, 9], [2024, 2, 10], [2024, 2, 11],
  [1900, 1, 30], [1900, 1, 31], [1900, 2, 1],
  [2033, 1, 30], [2033, 1, 31], // 2033 闰十一月问题年
  [2033, 12, 21], [2033, 12, 22], [2034, 1, 1],
  [1985, 2, 19], [1985, 2, 20],
  [2023, 3, 21], [2023, 3, 22], [2023, 4, 19], [2023, 4, 20], // 闰二月首尾
  [2020, 5, 22], [2020, 5, 23], [2020, 6, 20], [2020, 6, 21], // 闰四月首尾
  [1995, 8, 25], [1995, 8, 26], [1995, 9, 24], [1995, 9, 25], // 闰八月首尾
  [2057, 1, 1], [2089, 1, 1], // 已知各历法库易分歧年份
];
for (const [y, m, d] of boundaries) {
  for (const hour of [0, 5, 11]) {
    cases.push([y, m, d, hour, 'male']);
    cases.push([y, m, d, hour, 'female']);
  }
}

// 4) 全时辰覆盖:同一天 12 个时辰全排
for (let hour = 0; hour < 12; hour++) {
  cases.push([1990, 6, 15, hour, 'male']);
  cases.push([1990, 6, 15, hour, 'female']);
  cases.push([1964, 9, 10, hour, 'male']);
}

// 5) 农历大月 30 日覆盖(紫微安星依赖农历日 1-30)
for (let y = 1950; y <= 2030; y += 5) {
  for (let m = 1; m <= 12; m += 2) {
    for (const d of [5, 17, 26]) {
      cases.push([y, m, d, (y + m + d) % 12, d % 2 === 0 ? 'male' : 'female']);
    }
  }
}

// ── P47 扩容分层:补齐抽样空洞(月×时 137/144、农历日 24/30、年界系统性缺失) ──
const { Lunar } = require('lunar-javascript');

// 6) 月×时穷举:命宫定位 = f(月,时) 的 144 格全覆盖 × 4 个年代(甲子分散)
for (const y of [1937, 1968, 1999, 2041]) {
  for (let m = 1; m <= 12; m++) {
    for (let hour = 0; hour < 12; hour++) {
      cases.push([y, m, 16, hour, (m + hour) % 2 === 0 ? 'male' : 'female']);
    }
  }
}

// 7) 整农历年逐日扫描:农历 1-30 日全覆盖 + 闰月整月连续覆盖
//    1993(闰三月)与 2033(闰十一月难题年):自春节前 3 天起连扫 400 天
for (const startY of [1993, 2033]) {
  let cur = Lunar.fromYmd(startY, 1, 1).getSolar().next(-3);
  for (let i = 0; i < 400; i++) {
    cases.push([cur.getYear(), cur.getMonth(), cur.getDay(), i % 12, i % 2 === 0 ? 'male' : 'female']);
    cur = cur.next(1);
  }
}

// 8) 年界穷举:1900-2089 每年除夕 + 正月初一(生年干支翻转点,逐年不漏)
for (let y = 1900; y <= 2089; y++) {
  const spring = Lunar.fromYmd(y, 1, 1).getSolar();
  const eve = spring.next(-1);
  cases.push([eve.getYear(), eve.getMonth(), eve.getDay(), y % 12, y % 2 === 0 ? 'male' : 'female']);
  cases.push([spring.getYear(), spring.getMonth(), spring.getDay(), (y + 6) % 12, y % 2 === 0 ? 'female' : 'male']);
}

// 9) 确定性随机加密(LCG 可复现):全区间均匀撒点,加密组合路径
let seed = 20260805;
const rnd = () => (seed = (seed * 48271) % 2147483647) / 2147483647;
for (let i = 0; i < 3600; i++) {
  const y = 1900 + Math.floor(rnd() * 190);
  const m = 1 + Math.floor(rnd() * 12);
  const d = 1 + Math.floor(rnd() * daysInMonth(y, m));
  cases.push([y, m, d, Math.floor(rnd() * 12), rnd() < 0.5 ? 'male' : 'female']);
}

const seen = new Set();
const out = [];
let failed = 0;
for (const [y, m, d, h, g] of cases) {
  const key = `${y}-${m}-${d}-${h}-${g}`;
  if (seen.has(key)) continue;
  seen.add(key);
  try {
    out.push(genCase(y, m, d, h, g));
  } catch (e) {
    failed++;
    console.error(`FAIL ${key}: ${e.message}`);
  }
}

fs.writeFileSync('fixtures.json', JSON.stringify(out));
console.log(`generated ${out.length} cases, ${failed} failed`);
