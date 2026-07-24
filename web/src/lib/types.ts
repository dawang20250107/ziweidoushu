/**
 * API 类型契约 —— 与 Go 后端 JSON 逐字段对应。
 * 后端为唯一事实源:internal/ziwei/types.go、horoscope.go、httpapi/*。
 */

export type Gender = "male" | "female";
export type SiHua = "禄" | "权" | "科" | "忌";
export type StarType = "major" | "lucky" | "sha" | "minor";
export type BrightnessLevel = "bright" | "normal" | "dim" | "";

export interface Star {
  name: string;
  type: StarType;
  brightness?: string; // 庙旺得利平不陷
  brightnessLevel?: BrightnessLevel;
  siHua?: SiHua;
}

export interface Palace {
  branch: number; // 地支索引 0=子 … 11=亥
  stem: number; // 天干索引 0=甲 … 9=癸
  name: string; // 命宫/兄弟/夫妻/…/父母(iztro 口径)
  stars: Star[];
  daXianStart: number;
  daXianEnd: number;
  isMingGong: boolean;
  isShenGong: boolean;
  oppositeBranch: number;
  isEmpty: boolean;
  borrowedFromBranch?: number;
  borrowedFromName?: string;
  borrowedStars?: string[];
  changsheng12?: string;
  boshi12?: string;
  ages?: number[];
}

export interface DaXian {
  startAge: number;
  endAge: number;
  palaceBranch: number;
  palaceName: string;
}

export interface FourPillars {
  year: string;
  month: string;
  day: string;
  hour: string;
}

export interface LunarInfo {
  lunarYear: number;
  lunarMonth: number;
  lunarDay: number;
  yearStem: number;
  yearBranch: number;
  isLeapMonth: boolean;
}

export interface BirthInfo {
  year: number;
  month: number;
  day: number;
  hour: number; // 时辰索引 0=早子 … 11=亥 12=晚子
  gender: Gender;
  name?: string;
  longitude?: number;
  trueSolarTime?: boolean; // 是否按真太阳时校正时辰
  province?: string; // 国内出生地(查经度)
  city?: string;
  worldCity?: string; // 国际出生地(查经度 + 时区标准经线)
  utcOffset?: number; // 直传时区偏移(小时);worldCity 已含则无需
}

// WorldCity 世界主要城市:经度 + UTC 偏移,用于国际真太阳时。
export interface WorldCity {
  country: string;
  name: string;
  longitude: number;
  utcOffset: number;
  dst?: boolean; // 该国实行夏令时(数据为标准时)
}

// ProvinceCities 国内省份及其城市经度。
export interface ProvinceCities {
  name: string;
  cities: { name: string; longitude: number }[];
}

// ── 四柱视角(八字附加层)────────────────────────────
export interface HiddenStem {
  stem: string;
  element: string;
  shiShen: string;
}
export interface SiZhuPillar {
  name: string;
  stem: string;
  branch: string;
  stemElement: string;
  branchElement: string;
  stemShiShen: string; // 日柱为「日主」
  hidden: HiddenStem[];
  naYin: string;
  xunKong?: boolean; // 此柱地支落日柱旬空
}
export interface SiZhuGeJu {
  name: string; // 正官格/七杀格/…/建禄格/阳刃格/月劫格/杂气月垣
  basis: string; // 取格依据
  note: string; // 格局大意(子平真诠)
  source: string; // 出处
}
export interface ShenSha {
  name: string; // 神煞名
  pillars: string[]; // 命中柱:年/月/日/时
  basis: string; // 查法:年支三合/年支/日干/日柱旬
}
export interface DaYunEntry {
  index: number;
  ganZhi: string;
  startAge: number; // 虚岁
  startYear: number;
  stemShiShen: string; // 运干十神
  naYin: string;
  xunKong: string;
  isCurrent: boolean;
}
export interface LiuNianEntry {
  year: number;
  age: number;
  ganZhi: string;
  stemShiShen: string;
  naYin: string;
  isCurrent: boolean;
}
export interface DaYunView {
  forward: boolean; // 顺行/逆行
  startAge: number; // 起运虚岁
  startDesc: string; // 「X 年 Y 月后起运」
  list: DaYunEntry[];
  currentLiuNian: LiuNianEntry[];
}
export interface SiZhuView {
  dayMaster: string;
  dayMasterElement: string;
  pillars: SiZhuPillar[];
  elementCount: Record<string, number>;
  geJu?: SiZhuGeJu;
  shenSha?: ShenSha[];
  daYun?: DaYunView;
}

export interface Chart {
  birthInfo: BirthInfo;
  lunarInfo: LunarInfo;
  lunarDateText: string;
  fourPillars: FourPillars;
  siZhu?: SiZhuView;
  timeName: string;
  zodiac: string;
  sign: string;
  mingGongBranch: number;
  shenGongBranch: number;
  mingZhu: string;
  shenZhu: string;
  wuxingJu: number;
  wuxingJuName: string;
  ziweiPos: number;
  palaces: Palace[]; // 按地支索引 0-11
  daXians: DaXian[];
  referenceYear: number;
  currentAge: number;
  currentDaXianIndex: number;
}

export type PatternLevel = "excellent" | "good" | "neutral" | "caution";

export interface PatternCondition {
  required: string[];
  bonus?: string[];
  breaking?: string[];
}

export interface Pattern {
  name: string;
  level: PatternLevel;
  description: string;
  palaces: string[];
  conditions?: PatternCondition;
  source?: string;
}

export interface ReadingSection {
  key: string;
  title: string; // 事业·官禄
  palace: string;
  stars: string[]; // 带庙旺/四化标记
  level: "good" | "caution" | "neutral";
  text: string;
}
export interface Reading {
  overview: string;
  sections: ReadingSection[];
}
export interface ChartResponse {
  chart: Chart;
  patterns?: Pattern[];
  reading?: Reading; // 结构化多维断语(确定性,随盘生成)
}

// ── 运限 ──────────────────────────────────────────────

export interface HoroscopeScope {
  name: string; // 大限/童限/小限/流年/流月/流日/流时
  palaceBranch: number;
  stem: string;
  branch: string;
  palaceNames: string[]; // 下标=地支索引
  mutagen: [string, string, string, string]; // 禄权科忌
  stars?: Star[][]; // 流曜,下标=地支索引
  nominalAge?: number;
}

export interface Horoscope {
  targetSolarDate: string;
  targetLunarText: string;
  nominalAge: number;
  decadal: HoroscopeScope;
  age: HoroscopeScope;
  yearly: HoroscopeScope;
  monthly: HoroscopeScope;
  daily: HoroscopeScope;
  hourly: HoroscopeScope;
  suiqian12: string[];
  jiangqian12: string[];
}

// 运限逐层断语(大限→流年→流月→流日→流时),复用 ReadingSection。
export interface HoroscopeReading {
  target: string;
  sections: ReadingSection[];
}

// ── 事项择吉 ──────────────────────────────────────────
export interface EventCatalogItem {
  key: string;
  label: string;
  palace: string;
  kind: "auspicious" | "avoid";
}
export interface TimingYear {
  year: number;
  ganZhi: string;
  note: string;
}
export interface EventTiming {
  event: string;
  palace: string;
  kind: "auspicious" | "avoid";
  summary: string;
  years: TimingYear[];
  bestMonth?: string;
  bestDays?: string[];
  baseQuality: string; // 佳/中/弱
  baseNote: string;
  advice: string;
}

// ── 合盘与名人 ────────────────────────────────────────

export interface FamousPerson {
  id: string;
  name: string;
  category: string;
  description: string;
  year: number;
  month: number;
  day: number;
  hour: number;
  gender: Gender;
  notable: string;
}

/** 夫妻宫星曜断语(倪师口径)。 */
export interface HemingReading {
  star: string;
  summary: string;
  good: string;
  bad: string;
  spouseTraits: string;
  timing: string;
  niQuote: string;
}

export interface HemingSide {
  chart: Chart;
  patterns: Pattern[];
  fuqiStars: string[];
  /** 夫妻宫空宫时借对宫(官禄宫)主星 */
  fuqiBorrowed: boolean;
  readings: HemingReading[];
}

// 合盘确定性契合断语(比对双盘:年命相合/四化互飞/夫妻宫呼应/相处建议)。
export interface HemingMatchReading {
  score: number; // 20-100
  level: string; // 上上缘/上等姻缘/中上可成/中平宜经营/宜慎重
  summary: string;
  sections: ReadingSection[];
}

export interface HemingResponse {
  a: HemingSide;
  b: HemingSide;
  methodology: string;
  scoreCriteria: Record<string, string>;
  reading?: HemingMatchReading; // 确定性契合断语,随合盘生成
}

// ── 古籍 ──────────────────────────────────────────────

export interface BookMeta {
  title: string;
  slug: string;
  dynasty: string;
  author: string;
  intro: string;
  wordCount: number;
  chapters: number;
  paragraphs: number;
  source?: string;
}

export interface Paragraph {
  id: string;
  idx: number;
  text: string;
  translation?: string;
  niNote?: string;
}

export interface Chapter {
  title: string;
  subtitle?: string;
  paragraphs: Paragraph[];
}

export interface Book {
  title: string;
  slug: string;
  dynasty: string;
  author: string;
  intro: string;
  wordCount: number;
  chapters: Chapter[];
}

export interface SearchHit {
  bookSlug: string;
  bookTitle: string;
  chapterTitle: string;
  chapterIdx: number;
  paragraphId: string;
  snippet: string; // 含 <mark>,后端已转义
  text: string;
}

// ── AI ────────────────────────────────────────────────

export interface InterpretResult {
  text: string;
  provider: string;
  degraded: boolean;
}

// ── 常量(与后端口径一致)──────────────────────────────

export const BRANCHES = ["子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"] as const;
export const STEMS = ["甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"] as const;
export const HOUR_NAMES = [
  "早子时", "丑时", "寅时", "卯时", "辰时", "巳时",
  "午时", "未时", "申时", "酉时", "戌时", "亥时", "晚子时",
] as const;
