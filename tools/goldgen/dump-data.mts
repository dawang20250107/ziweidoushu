/**
 * 机械化数据迁移:直接执行仓库 TS 模块,把导出常量序列化为 JSON 数据文件。
 * 运行:npx tsx dump.mts  (cwd = 仓库根)
 */
import { writeFileSync, mkdirSync } from 'node:fs';

const REPO = '/home/user/ziweidoushu';

const classics = await import(`${REPO}/lib/classics/index.ts`);
const nihaiTianji = await import(`${REPO}/lib/nihai/tianji.ts`);
const nihaiRenji = await import(`${REPO}/lib/nihai/renji.ts`);
const nihaiDiji = await import(`${REPO}/lib/nihai/diji.ts`);
const nihaiIndex = await import(`${REPO}/lib/nihai/index.ts`);
const heming = await import(`${REPO}/lib/ziwei/heming-knowledge.ts`);
const cities = await import(`${REPO}/lib/ziwei/cities.ts`);
const famous = await import(`${REPO}/lib/ziwei/famous.ts`);
const constants = await import(`${REPO}/lib/ziwei/constants.ts`);
const dbAnalysis = await import(`${REPO}/lib/ziwei/db-analysis.ts`);
const seo = await import(`${REPO}/lib/seo/knowledge.ts`);

function dump(path: string, data: unknown) {
  writeFileSync(path, JSON.stringify(data, null, 1) + '\n');
  console.log('wrote', path);
}

mkdirSync(`${REPO}/data/classics`, { recursive: true });
mkdirSync(`${REPO}/data/nihai`, { recursive: true });
mkdirSync(`${REPO}/data/knowledge`, { recursive: true });

// ── 古籍:一书一文件 ──
for (const book of classics.ALL_BOOKS) {
  dump(`${REPO}/data/classics/${book.slug}.json`, book);
}

// ── 倪海厦三纪 ──
dump(`${REPO}/data/nihai/tianji.json`, {
  modules: nihaiTianji.TIANJI_MODULES,
  hexagrams: nihaiTianji.HEXAGRAMS,
  fengshui: nihaiTianji.FENGSHUI_ENTRIES,
  episodes: nihaiTianji.TIANJI_EPISODES,
  quotes: nihaiTianji.TIANJI_QUOTES,
  stats: nihaiTianji.TIANJI_STATS,
});
dump(`${REPO}/data/nihai/renji.json`, {
  modules: nihaiRenji.RENJI_MODULES,
  acuExperiences: nihaiRenji.ACU_EXPERIENCES,
  transNeedling: nihaiRenji.TRANS_NEEDLING,
  hantangFormulas: nihaiRenji.HANTANG_FORMULAS,
  classicFormulas: nihaiRenji.CLASSIC_FORMULAS,
  stats: nihaiRenji.RENJI_STATS,
});
dump(`${REPO}/data/nihai/diji.json`, {
  modules: nihaiDiji.DIJI_MODULES,
  stats: nihaiDiji.DIJI_STATS,
});
dump(`${REPO}/data/nihai/bio.json`, {
  bio: nihaiIndex.NI_HAIXIA_BIO,
  categories: nihaiIndex.SANJI_CATEGORIES,
});

// ── 合盘知识库 ──
dump(`${REPO}/data/knowledge/heming.json`, {
  starInFuqi: heming.STAR_IN_FUQI_GU,
  sihuaInFuqi: heming.SIHUA_IN_FUQI_GU,
  methodology: heming.HEMING_METHODOLOGY,
  marriageStarsBrief: heming.MARRIAGE_STARS_BRIEF,
  scoreCriteria: heming.HEMING_SCORE_CRITERIA,
});

// ── 主星知识 + 主题标签 ──
dump(`${REPO}/data/knowledge/stars.json`, {
  descriptions: constants.STAR_DESCRIPTIONS,
  slugs: seo.STAR_TO_SLUG,
  order: seo.ALL_STARS,
});
dump(`${REPO}/data/knowledge/topics.json`, {
  palaceName: dbAnalysis.TOPIC_PALACE_NAME,
  label: dbAnalysis.TOPIC_LABEL,
  order: seo.ALL_TOPICS,
});

// ── 城市经度 + 名人盘 ──
dump(`${REPO}/data/cities.json`, { provinces: cities.PROVINCES });
dump(`${REPO}/data/famous.json`, {
  persons: famous.FAMOUS_PERSONS,
  categories: famous.FAMOUS_CATEGORIES,
});

// ── 完整性统计,供人工核对 ──
const stats = {
  books: classics.ALL_BOOKS.map((b: any) => ({
    slug: b.slug,
    chapters: b.chapters.length,
    paragraphs: b.chapters.reduce((s: number, c: any) => s + c.paragraphs.length, 0),
  })),
  totalParagraphs: classics.TOTAL_PARAGRAPHS,
  tianjiModules: nihaiTianji.TIANJI_MODULES.length,
  hexagrams: nihaiTianji.HEXAGRAMS.length,
  fengshui: nihaiTianji.FENGSHUI_ENTRIES.length,
  episodes: nihaiTianji.TIANJI_EPISODES.length,
  quotes: nihaiTianji.TIANJI_QUOTES.length,
  renjiModules: nihaiRenji.RENJI_MODULES.length,
  acuExperiences: nihaiRenji.ACU_EXPERIENCES.length,
  transNeedling: nihaiRenji.TRANS_NEEDLING.length,
  hantangFormulas: nihaiRenji.HANTANG_FORMULAS.length,
  classicFormulas: nihaiRenji.CLASSIC_FORMULAS.length,
  dijiModules: nihaiDiji.DIJI_MODULES.length,
  hemingStars: Object.keys(heming.STAR_IN_FUQI_GU).length,
  provinces: cities.PROVINCES.length,
  cities: cities.PROVINCES.reduce((s: number, p: any) => s + p.cities.length, 0),
  famous: famous.FAMOUS_PERSONS.length,
  starDescriptions: Object.keys(constants.STAR_DESCRIPTIONS).length,
};
console.log(JSON.stringify(stats, null, 2));
