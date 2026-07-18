/** 格局差分基准:对黄金用例输入跑 TS 原版 generateChart + detectPatterns。 */
import { readFileSync, writeFileSync } from 'node:fs';
import { gzipSync, gunzipSync } from 'node:zlib';

const REPO = '/home/user/ziweidoushu';
const { generateChart } = await import(`${REPO}/lib/ziwei/algorithm.ts`);
const { detectPatterns, getMingGongSummary } = await import(`${REPO}/lib/ziwei/patterns.ts`);

const golden = JSON.parse(
  gunzipSync(readFileSync(`${REPO}/internal/ziwei/testdata/iztro_golden.json.gz`)).toString(),
);

const out: any[] = [];
for (const g of golden) {
  const { year, month, day, hour, gender } = g.input;
  const chart = generateChart({ year, month, day, hour, gender });
  out.push({
    input: g.input,
    patterns: detectPatterns(chart),
    summary: getMingGongSummary(chart),
  });
}
writeFileSync(
  `${REPO}/internal/ziwei/testdata/patterns_golden.json.gz`,
  gzipSync(JSON.stringify(out)),
);
console.log(`patterns golden: ${out.length} cases`);
const withPatterns = out.filter((o) => o.patterns.length > 0).length;
console.log(`cases with >=1 pattern: ${withPatterns}`);
const counts: Record<string, number> = {};
for (const o of out) for (const p of o.patterns) counts[p.name] = (counts[p.name] ?? 0) + 1;
console.log('distinct patterns:', Object.keys(counts).length);
console.log(JSON.stringify(counts, null, 0));
