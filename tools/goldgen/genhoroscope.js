/**
 * 运限黄金基准生成器:iztro astrolabe.horoscope(targetDate, timeIndex) 输出。
 * 覆盖:童限、大限各段、闰月目标日、晚子时、跨世纪目标。
 */
const { astro } = require('iztro');
const fs = require('fs');
const zlib = require('zlib');

// 从安星基准中抽取出生用例(每 7 个取 1,约 220 个,保持多样性)
const golden = JSON.parse(
  zlib.gunzipSync(fs.readFileSync('/home/user/ziweidoushu/internal/ziwei/testdata/iztro_golden.json.gz')),
);
const births = golden.filter((_, i) => i % 7 === 0).map((g) => g.input);

function serializeStars(stars) {
  return stars.map((cell) => cell.map((s) => ({ name: s.name, type: s.type })));
}

function serializeScope(sc, withStars, withDec) {
  const out = {
    index: sc.index,
    name: sc.name,
    heavenlyStem: sc.heavenlyStem,
    earthlyBranch: sc.earthlyBranch,
    palaceNames: sc.palaceNames,
    mutagen: sc.mutagen,
  };
  if (sc.nominalAge !== undefined) out.nominalAge = sc.nominalAge;
  if (withStars && sc.stars) out.stars = serializeStars(sc.stars);
  if (withDec && sc.yearlyDecStar) {
    out.suiqian12 = sc.yearlyDecStar.suiqian12;
    out.jiangqian12 = sc.yearlyDecStar.jiangqian12;
  }
  return out;
}

const cases = [];
let failed = 0;
for (const b of births) {
  const astrolabe = astro.bySolar(`${b.year}-${b.month}-${b.day}`, b.hour, b.gender === 'male' ? '男' : '女', true, 'zh-CN');
  // 目标日期组:童限期、青年期、闰月日、近期、晚子时
  const targets = [
    { y: b.year + 2, m: 3, d: 9, h: 4 },            // 童限
    { y: b.year + 5, m: 11, d: 27, h: 0 },          // 童限/初运边界
    { y: b.year + 33, m: 7, d: 15, h: 7 },          // 壮年大限
    { y: 2023, m: 3, d: 25, h: 6 },                 // 闰二月内目标日
    { y: 2026, m: 7, d: 16, h: 12 },                // 当前 + 晚子时
    { y: b.year + 61, m: 1, d: 30, h: 9 },          // 老年 + 农历年界附近
  ];
  for (const t of targets) {
    if (t.y < b.year || t.y > 2100) continue;
    if (t.y === b.year) continue;
    try {
      const h = astrolabe.horoscope(`${t.y}-${t.m}-${t.d}`, t.h);
      cases.push({
        birth: b,
        target: t,
        solarDate: h.solarDate,
        lunarDate: h.lunarDate,
        decadal: serializeScope(h.decadal, true, false),
        age: serializeScope(h.age, false, false),
        yearly: serializeScope(h.yearly, true, true),
        monthly: serializeScope(h.monthly, true, false),
        daily: serializeScope(h.daily, true, false),
        hourly: serializeScope(h.hourly, true, false),
      });
    } catch (e) {
      failed++;
      console.error(`FAIL ${JSON.stringify(b)} -> ${JSON.stringify(t)}: ${e.message}`);
    }
  }
}

fs.writeFileSync(
  '/home/user/ziweidoushu/internal/ziwei/testdata/horoscope_golden.json.gz',
  zlib.gzipSync(JSON.stringify(cases)),
);
console.log(`horoscope golden: ${cases.length} cases, ${failed} failed`);
const childhood = cases.filter((c) => c.decadal.name === '童限').length;
console.log(`childhood cases: ${childhood}`);
