/**
 * 占卜(梅花易数 + 小六壬)API client + 类型契约。
 * 后端统一信封:{ ok, data } / { ok: false, error: { code, message } }。
 * 起卦与卦象展示免费(匿名 fetch);AI 深度解卦按次付费(authFetch,消耗 divination)。
 *
 * 起卦由服务端推导,客户端不可伪造。关键契约:AI 解卦须回传与用户所见「同一卦」——
 *   时间卦:记住起卦返回的 castAt,原样回传;数字卦:回传同一组 numbers。
 */
import { authFetch, currentUser, ensureAccess } from "@/lib/auth";

const BASE = process.env.NEXT_PUBLIC_API_BASE ?? "";

// ── 类型(与 Go 后端 meihua 包逐字段对应)──────────────

/** 八卦(先天数 1-8)。lines 为三爻,自下而上,true=阳爻。 */
export interface Trigram {
  num: number;
  name: string; // 乾/兑/离/震/巽/坎/艮/坤
  symbol: string; // ☰☱☲☳☴☵☶☷
  nature: string; // 天/泽/火/雷/风/水/山/地
  element: string; // 金/木/水/火/土
  lines: boolean[]; // ×3
}

/** 六十四卦。lines 为六爻,自下而上(1-6 爻),true=阳爻。 */
export interface Hexagram {
  name: string;
  upper: Trigram;
  lower: Trigram;
  lines: boolean[]; // ×6
  guaCi?: string; // 《周易》卦辞(公版经文)
}

/** 体用生克关系。 */
export type Relation = "用生体" | "比和" | "体克用" | "体生用" | "用克体";

/** 一次梅花起卦的完整卦象。 */
/** 断语分节(四线通用:免费确定性层的呈现单元)。 */
export interface JudgeSection {
  key: string;
  title: string;
  text: string;
}

export interface MeihuaJudgment {
  level: "good" | "neutral" | "caution";
  score: number;
  tiQi: string; // 体卦月令旺衰:旺/相/休/囚/死
  conclusion: string;
  points: string[];
  topic: string;
  topicNote?: string;
  yingQi: string;
  sections?: JudgeSection[]; // 分节深断(卦象总论/体用之辨/过程与结局/类象取应/应期)
}

export interface MeihuaRoleLore {
  role: string; // 体卦/用卦/变卦
  name: string;
  renlun: string;
  shenti: string;
  dongwu: string;
  jingwu: string;
  fangwei: string;
  tianshi: string;
  xing: string;
}

export interface MeihuaResult {
  method: "time" | "number" | "zi";
  question?: string;
  lunarText?: string; // 时间卦:「午年六月初四日申时」
  numbers?: number[]; // 数字卦原始数
  castBasis?: string; // 起数依据(如「问辞12字起上卦,加申时数9配下卦」)
  ziText?: string; // 测字起卦的原字(一或二字)
  ziStrokes?: number[]; // 各字笔画数(Unihan 简体口径)
  ben: Hexagram; // 本卦
  hu: Hexagram; // 互卦
  bian: Hexagram; // 变卦
  judgment?: MeihuaJudgment; // 断卦骨架(体用总诀确定性推演)
  lore?: MeihuaRoleLore[]; // 万物类象(体/用/变)
  moving: number; // 动爻 1-6
  tiTrigram: Trigram; // 体卦
  yongTrigram: Trigram; // 用卦
  tiIsUpper: boolean; // 体在上卦
  relation: Relation;
  verdict: string; // 吉凶倾向一句话
}

/** 小六壬六位之一。 */
export interface LiuRenPos {
  name: string; // 大安/留连/速喜/赤口/小吉/空亡
  luck: "吉" | "凶" | "平";
  meaning: string;
}

/** 小六壬占算结果。 */
export interface XiaoLiuRenResult {
  question?: string;
  lunarText: string; // 「正月初一子时」
  steps: string[]; // 逐步落位名:月/日/时,有问辞再加问数一步
  qNum?: number; // 问数(问辞字数,第四跳步数;缺省=无问辞正时课)
  result: LiuRenPos;
  path: LiuRenPos[]; // 三步完整落位 ×3
  sections?: JudgeSection[]; // 分节深断(掐指路径/落宫详断/途中之象)
}

// ── 六爻纳甲(与 Go 后端 liuyao 包逐字段对应)──────────

/** 六爻一爻(装卦后)。bianYao 仅动爻有,为变卦对应爻。 */
export interface LiuYaoYao {
  pos: number; // 1-6 自下而上
  yang: boolean;
  moving: boolean;
  stem: string; // 纳甲天干
  branch: string; // 纳甲地支
  element: string; // 金/木/水/火/土
  liuQin: string; // 六亲:父母/兄弟/子孙/妻财/官鬼
  liuShen: string; // 六神:青龙/朱雀/勾陈/腾蛇/白虎/玄武
  isShi: boolean; // 世
  isYing: boolean; // 应
  bianYao?: LiuYaoYao;
  monthState?: string; // 对月建旺衰:旺/相/休/囚/死
  yuePo?: boolean; // 月破
  xunKong?: boolean; // 旬空
  dayRelation?: string; // 日辰对爻:临/冲/合/扶/生/克/泄/耗
  anDong?: boolean; // 暗动
  riPo?: boolean; // 日破
  dayStage?: string; // 对日辰四态:长生/帝旺/墓/绝
  bianRelation?: string; // 动爻之变:化进神/化退神/伏吟/反吟/化长生/化墓/化绝/化合/回头生/回头克
}

/** 六爻确定性断语(用神旺衰/伏神/卦性/动变/世应/应期,免费层)。 */
export interface LiuYaoJudgment {
  conclusion: string;
  level: "good" | "neutral" | "caution";
  yongShen: string; // 用神状态摘要
  yingQi: string; // 应期提示(含具体地支)
  points: string[];
  sections?: JudgeSection[]; // 分节深断(取用/旺衰/元忌/动变/世应/逐爻/应期)
}

/** 用神不上卦时之伏神(本宫首卦纳甲取)。 */
export interface LiuYaoFuShen {
  liuQin: string;
  branch: string;
  element: string;
  pos: number;
  fei: string; // 飞神支
  canOut: boolean;
  note: string;
  chuFuRi: string;
}

/** 六爻装卦结果。yaos 自下而上(index 0 = 初爻)。 */
export interface LiuYaoResult {
  question?: string;
  lunarText: string; // 「六月初三日(甲子日)」
  dayStem: string;
  dayBranch: string;
  monthJian: string; // 月建地支
  riJian: string; // 日辰地支
  benName: string;
  bianName?: string; // 有动爻才有
  palace: string; // 「乾宫」
  palaceSeq: string; // 八纯卦/一世卦…游魂卦/归魂卦
  yaos: LiuYaoYao[]; // ×6
  movingNums: number[]; // 动爻位置(空=静卦)
  tosses?: number[]; // 摇卦原始记录(每爻背面数 0-3),回传同一卦的凭据
  yongShen?: string; // 用神建议(六亲名或「世爻」)
  yongShenBasis?: string;
  yongShenOverride?: string; // 问者显式指明的取用(快照回传解卦用) // 经义依据
  yongShenPos?: number[]; // 用神所在爻位;空=不上卦
  yuanShen?: string; // 元神(生用神者)
  yuanShenPos?: number[];
  jiShen?: string; // 忌神(克用神者)
  jiShenPos?: number[];
  chouShen?: string; // 仇神(生忌克元者)
  benXingZhi?: string; // 本卦卦性:六冲/六合
  bianXingZhi?: string; // 变卦卦性
  fuShen?: LiuYaoFuShen; // 用神不上卦时之伏神
  judgment?: LiuYaoJudgment; // 确定性断语骨架
  jingWen?: LiuYaoJingWen; // 《周易》经文层
}

/** 《周易》经文层(公版):yaoCi 与动爻同序,文本带爻题。 */
export interface LiuYaoJingWen {
  benGuaCi: string;
  bianGuaCi?: string;
  yaoCi?: string[];
  yong?: string; // 六爻皆动:乾用九/坤用六
}

/** AI 解卦读物。 */
export interface DivineReading {
  text: string;
  provider: string;
}

// ── 卦档(占卜记录存档)────────────────────────────────

/** 卦档记录。列表态无 payload/reading;详情态 payload 为对应卦象 JSON。 */
export interface DivinationRecord {
  id: string;
  kind: "meihua" | "liuyao" | "xiaoliuren" | "daliuren";
  question: string;
  summary: string; // 「地天泰 → 山风蛊」/「泽火革 · 用克体」/「速喜 · 吉」/「元首课 · 三传辰申子」
  payload?: MeihuaResult | LiuYaoResult | XiaoLiuRenResult | DaLiuRenResult;
  reading?: string;
  readingProvider?: string;
  hasReading: boolean;
  castAt: string;
  createdAt: string;
}

// ── 错误类型 ──────────────────────────────────────────

/** 带业务错误码的异常(no_credits / ai_unavailable / cast_failed …)。 */
export class DivinationError extends Error {
  constructor(
    public code: string,
    message: string,
    public status: number,
  ) {
    super(message);
  }
}

interface Envelope<T> {
  ok: boolean;
  data?: T;
  error?: { code: string; message: string };
}

/** 解析统一信封;非 2xx 或 ok:false 抛 DivinationError。 */
async function parse<T>(res: Response): Promise<T> {
  let body: Envelope<T>;
  try {
    body = (await res.json()) as Envelope<T>;
  } catch {
    throw new DivinationError("network", `请求失败(${res.status})`, res.status);
  }
  if (!res.ok || !body.ok) {
    throw new DivinationError(
      body.error?.code ?? "unknown",
      body.error?.message ?? `请求失败(${res.status})`,
      res.status,
    );
  }
  return body.data as T;
}

const JSON_HEADERS = { "Content-Type": "application/json" };

// ── 起卦入参 ──────────────────────────────────────────

export interface CastInput {
  method: "time" | "number" | "zi";
  numbers?: number[]; // 数字卦:[n,n] 或 [n,n,n]
  ziText?: string; // 测字卦:一或二个汉字(端法笔画起数)
  castAt?: number; // unix 秒:解卦回传同一卦必带(数字卦断层旺衰亦锚定起卦时刻)
  question?: string;
}

export interface LiuYaoCastInput {
  method: "shake" | "tosses"; // 服务端摇卦 | 报爻(自摇铜钱录入)
  tosses?: number[]; // ×6 每爻背面数 0-3,自下而上
  castAt?: number; // unix 秒(回传同一卦时用)
  question?: string;
  yongShen?: string; // 显式取用:世爻/妻财/官鬼/父母/子孙/兄弟(空=按问辞推断)
}

// ── 接口 ──────────────────────────────────────────────

/** 带上登录态(若有)的起卦请求头:登录用户起卦自动存入卦档,匿名照常起卦。 */
async function castHeaders(): Promise<Record<string, string>> {
  if (!currentUser()) return JSON_HEADERS;
  const token = await ensureAccess().catch(() => null);
  return token ? { ...JSON_HEADERS, Authorization: `Bearer ${token}` } : JSON_HEADERS;
}

/** 梅花易数起卦(免费,匿名可用;登录则自动存档并返回 recordId)。 */
export async function castMeihua(
  input: CastInput,
): Promise<{ result: MeihuaResult; castAt: number; recordId?: string }> {
  const res = await fetch(`${BASE}/api/v1/divination/meihua`, {
    method: "POST",
    headers: await castHeaders(),
    body: JSON.stringify(input),
  });
  return parse(res);
}

/** 六爻起卦(免费,匿名可用;登录则自动存档并返回 recordId)。 */
export async function castLiuYao(
  input: LiuYaoCastInput,
): Promise<{ result: LiuYaoResult; castAt: number; recordId?: string }> {
  const res = await fetch(`${BASE}/api/v1/divination/liuyao`, {
    method: "POST",
    headers: await castHeaders(),
    body: JSON.stringify(input),
  });
  return parse(res);
}

/** 大六壬起课结果(天地盘/四课/三传/课体 + 确定性断语)。 */
export interface DaLiuRenKe {
  lower: string;
  upper: string;
}

export interface DaLiuRenJudgment {
  conclusion: string;
  level: "good" | "neutral" | "caution";
  keTypeText: string;
  sanChuan: string[];
  points: string[];
  sections?: JudgeSection[]; // 分节深断(课体详解/三传始末/天将所临/应期推算)
}

export interface DaLiuRenResult {
  dayStem: string;
  dayBranch: string;
  hourBranch: string;
  monthGen: string; // 月将
  tianPan: string[]; // 地盘子起十二位上所乘天盘之神
  ke: DaLiuRenKe[]; // 四课
  chuan: string[]; // 三传(初/中/末)
  keType: string; // 课体
  tianJiang?: string[]; // 地盘十二位所乘天将
  chuanJiang?: string[]; // 三传所乘天将
  guiIsDay?: boolean;
  xunKong?: string[]; // 旬空两支
  chuanDunGan?: string[]; // 三传旬遁干(传落空亡为空串)
  baoShu?: number; // 活时报数(正时无)
  hourNote?: string; // 「活时·报数7」
  judgment?: DaLiuRenJudgment;
  nianMing?: { birthYear: number; branch: string; shangShen: string; jiang: string }; // 年命上神
}

/** 大六壬起课(免费,匿名可用;登录则自动存入卦档)。 */
export async function castDaLiuRen(
  input?: { question?: string; castAt?: number; baoShu?: number; birthYear?: number },
): Promise<{ result: DaLiuRenResult; castAt: number; recordId?: string }> {
  const res = await fetch(`${BASE}/api/v1/divination/daliuren`, {
    method: "POST",
    headers: await castHeaders(),
    body: JSON.stringify(input ?? {}),
  });
  return parse(res);
}

/** AI 深度解课(大六壬,需登录,消耗 1 次)。 */
export async function divineDaLiuRenAI(
  input: { castAt: number; question: string; recordId?: string; baoShu?: number; birthYear?: number },
): Promise<{ reading: DivineReading; remainingCredits: number }> {
  const res = await authFetch(`${BASE}/api/v1/ai/divine`, {
    method: "POST",
    headers: JSON_HEADERS,
    body: JSON.stringify({ kind: "daliuren", ...input }),
  });
  return parse(res);
}

/** 小六壬快占(免费,匿名可用;登录则自动存入卦档)。 */
export async function castXiaoLiuRen(
  input?: { question?: string; castAt?: number },
): Promise<{ result: XiaoLiuRenResult; recordId?: string }> {
  const res = await fetch(`${BASE}/api/v1/divination/xiaoliuren`, {
    method: "POST",
    headers: await castHeaders(),
    body: JSON.stringify(input ?? {}),
  });
  return parse(res);
}

/** AI 深度解卦(需登录,消耗 1 次 divination)。question 必填;卦象参数须与所见同一卦;recordId 用于卦档回填。 */
export async function divineAI(
  input: CastInput & { question: string; recordId?: string },
): Promise<{ result: MeihuaResult; reading: DivineReading; remainingCredits: number }> {
  const res = await authFetch(`${BASE}/api/v1/ai/divine`, {
    method: "POST",
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
  return parse(res);
}

/**
 * AI 六爻解卦(需登录,消耗 1 次 divination)。
 * 同一卦契约:必须回传起卦返回的 tosses + castAt(method 固定 "tosses"),服务端按记录重装此卦。
 */
export async function divineLiuYaoAI(
  input: { tosses: number[]; castAt: number; question: string; recordId?: string; yongShen?: string },
): Promise<{ result: LiuYaoResult; reading: DivineReading; remainingCredits: number }> {
  const res = await authFetch(`${BASE}/api/v1/ai/divine`, {
    method: "POST",
    headers: JSON_HEADERS,
    body: JSON.stringify({ kind: "liuyao", method: "tosses", ...input }),
  });
  return parse(res);
}

/** 卦档列表(需登录)。kind 缺省为全部占法。 */
export async function listDivinations(
  limit = 50,
  offset = 0,
  kind?: DivinationRecord["kind"],
): Promise<{ records: DivinationRecord[]; total: number }> {
  const kindQ = kind ? `&kind=${kind}` : "";
  const res = await authFetch(`${BASE}/api/v1/me/divinations?limit=${limit}&offset=${offset}${kindQ}`);
  const data = await parse<{ records?: DivinationRecord[]; total?: number }>(res);
  return { records: data.records ?? [], total: data.total ?? 0 };
}

/** 卦档详情(需登录,含卦象与解卦全文)。 */
export async function getDivination(id: string): Promise<DivinationRecord> {
  const res = await authFetch(`${BASE}/api/v1/me/divinations/${id}`);
  const data = await parse<{ record: DivinationRecord }>(res);
  return data.record;
}

/** 删除卦档(需登录,仅本人)。 */
export async function deleteDivination(id: string): Promise<void> {
  const res = await authFetch(`${BASE}/api/v1/me/divinations/${id}`, { method: "DELETE" });
  await parse(res);
}

/** 剩余解卦次数(需登录)。读 data.credits.divination,缺省 0。 */
export async function fetchDivinationCredits(): Promise<number> {
  const res = await authFetch(`${BASE}/api/v1/me/entitlements`);
  const data = await parse<{ credits?: Record<string, number | undefined> }>(res);
  return data.credits?.divination ?? 0;
}

// ── 展示辅助(纯函数,页面与组件共用)────────────────

/** 语义色调:金(gold)/成(ok)/戒(warn)/危(danger)。 */
export type Tone = "gold" | "ok" | "warn" | "danger";

/** 体用生克 → 吉凶断语 + 语义色。 */
export const RELATION_TONE: Record<Relation, { label: string; tone: Tone }> = {
  用生体: { label: "用生体 · 大吉", tone: "gold" },
  比和: { label: "比和 · 吉", tone: "ok" },
  体克用: { label: "体克用 · 小吉", tone: "ok" },
  体生用: { label: "体生用 · 小凶", tone: "warn" },
  用克体: { label: "用克体 · 凶", tone: "danger" },
};

/** 小六壬吉凶 → 语义色。 */
export function luckTone(luck: LiuRenPos["luck"]): Tone {
  if (luck === "吉") return "ok";
  if (luck === "凶") return "danger";
  return "warn"; // 平
}

/** 卦名之上下卦类象:如「上乾天 · 下兑泽」。 */
export function trigramLine(h: Hexagram): string {
  return `上${h.upper.name}${h.upper.nature} · 下${h.lower.name}${h.lower.nature}`;
}

/** 五行 → 语义色变量(与四柱视角同映射:木青 火朱 土赭 金曜 水墨蓝)。 */
export const ELEMENT_VAR: Record<string, string> = {
  木: "var(--ok)",
  火: "var(--danger)",
  土: "var(--warn)",
  金: "var(--gold)",
  水: "var(--info)",
};

/** 摇卦背面数 → 爻象文案(1背少阳 2背少阴 3背老阳动 0背老阴动)。 */
export const TOSS_OPTIONS: { backs: number; label: string; hint: string }[] = [
  { backs: 1, label: "一背", hint: "少阳" },
  { backs: 2, label: "两背", hint: "少阴" },
  { backs: 3, label: "三背", hint: "老阳 · 动" },
  { backs: 0, label: "无背", hint: "老阴 · 动" },
];
