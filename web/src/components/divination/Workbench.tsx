"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { AUTH_EVENT, currentUser } from "@/lib/auth";
import {
  castMeihua,
  castLiuYao,
  divineAI,
  divineLiuYaoAI,
  fetchDivinationCredits,
  DivinationError,
  type CastInput,
  type MeihuaResult,
  type LiuYaoResult,
} from "@/lib/divination";
import { HexagramView } from "@/components/divination/HexagramView";
import { MeihuaCast } from "@/components/divination/MeihuaCast";
import { LiuYaoCast } from "@/components/divination/LiuYaoCast";
import { StepShake } from "@/components/divination/StepShake";
import { LiuYaoPan } from "@/components/divination/LiuYaoPan";
import { prefersReducedMotion } from "@/components/divination/useReducedMotion";
import { CreditsBadge, MethodTab, NumField, TossEntry } from "./WorkbenchParts";
import { TiYongCard, JudgeCard, JingWenCard, LiuYaoJudgeCard, LoreCard } from "./WorkbenchResultCards";
import { AiSection, type AiErr } from "./WorkbenchAiSection";

type Kind = "meihua" | "liuyao";
type CastMethod = "time" | "number" | "zi";
type LiuYaoMethod = "step" | "shake" | "tosses";

const LY_METHOD_LABEL: Record<LiuYaoMethod, string> = {
  step: "逐爻摇卦",
  shake: "铜钱摇卦",
  tosses: "手动报爻",
};

const MAX_Q = 200;
// 起卦动效总时长封顶(与各仪式编排对齐:梅花≈3.4s,六爻≈3.2s)
const CAST_ANIM_MS: Record<Kind, number> = { meihua: 3400, liuyao: 3200 };

const KIND_META: Record<Kind, { eyebrow: string; title: string; sub: string; cta: string; casting: string; aiHint: string }> = {
  meihua: {
    eyebrow: "占卜 · 心易",
    title: "梅花易数",
    sub: "一事一占,以卦观势——体用生克为纲、卦气旺衰定力度,互卦断过程、变卦断结局。",
    cta: "起卦",
    casting: "起卦中…",
    aiHint: "依断卦骨架与万物类象逐层解读",
  },
  liuyao: {
    eyebrow: "占卜 · 纳甲",
    title: "六爻纳甲",
    sub: "铜钱六掷、装卦断事——用神旺衰、动变生克、世应应期。",
    cta: "摇卦",
    casting: "摇卦中…",
    aiHint: "依用神 / 世应 / 六亲六神与动变断成败应期",
  },
};

/**
 * 占卜工作台(定占法):梅花易数 / 六爻纳甲 起卦 + 卦象展示 + 确定性断卦 + AI 深度解卦。
 * 由 /meihua 与 /liuyao 两个独立板块以 kind 固定实例化(原 /divination 问卦页拆分而来)。
 */
export function DivinationWorkbench({ kind, homePath }: { kind: Kind; homePath: string }) {
  const [signedIn, setSignedIn] = useState<boolean | null>(null);
  const [credits, setCredits] = useState<number | null>(null); // null=未知/未登录

  // 起卦输入
  const [question, setQuestion] = useState("");
  const [method, setMethod] = useState<CastMethod>("time");
  const [ziText, setZiText] = useState(""); // 测字起卦:一或二个汉字
  const [lyYong, setLyYong] = useState(""); // 六爻显式取用(空=按问辞自动识别)
  const [lyMethod, setLyMethod] = useState<LiuYaoMethod>("step");
  const [numA, setNumA] = useState("");
  const [numB, setNumB] = useState("");
  const [lyTosses, setLyTosses] = useState<(number | null)[]>(Array(6).fill(null)); // 报爻:初爻→上爻

  // 起卦态(结果按占法各自留存)
  const [casting, setCasting] = useState(false);
  const [castError, setCastError] = useState<string | null>(null);
  const [result, setResult] = useState<MeihuaResult | null>(null);
  const [castAt, setCastAt] = useState<number | null>(null); // 时间卦:回传同一卦
  const [castNumbers, setCastNumbers] = useState<number[] | null>(null); // 数字卦:回传同一卦
  const [lyResult, setLyResult] = useState<LiuYaoResult | null>(null);
  const [lyCastAt, setLyCastAt] = useState<number | null>(null); // 六爻:tosses+castAt 回传同一卦
  // 卦档:登录起卦服务端自动存档,AI 解卦按 recordId 回填
  const [mhRecordId, setMhRecordId] = useState<string | null>(null);
  const [lyRecordId, setLyRecordId] = useState<string | null>(null);

  // AI 解卦态
  const [aiLoading, setAiLoading] = useState(false);
  const [reading, setReading] = useState<string | null>(null);
  const [aiError, setAiError] = useState<AiErr | null>(null);

  const resultRef = useRef<HTMLDivElement>(null);
  const [resultCue, setResultCue] = useState(0); // 仪式收束→结果区聚焦归位(重放入场)
  const [spot, setSpot] = useState(false); // 收束光圈:幕布落下时光聚结果区再散开
  // 起卦纪元:每次起卦递增;迟到的 AI 解读若纪元已变,不得错挂到新卦上
  const castEpoch = useRef(0);

  // 登录态同步 + 拉解卦次数
  useEffect(() => {
    const sync = () => {
      const ok = !!currentUser();
      setSignedIn(ok);
      if (ok) {
        fetchDivinationCredits()
          .then(setCredits)
          .catch(() => setCredits(null));
      } else {
        setCredits(null);
      }
    };
    sync();
    window.addEventListener(AUTH_EVENT, sync);
    return () => window.removeEventListener(AUTH_EVENT, sync);
  }, []);

  const shownResult = kind === "meihua" ? result : lyResult;

  // 结果就绪后滚入
  useEffect(() => {
    if (shownResult && !casting && resultRef.current) {
      resultRef.current.scrollIntoView({
        behavior: prefersReducedMotion() ? "auto" : "smooth",
        block: "start",
      });
    }
  }, [shownResult, casting]);

  const validNum = (s: string): number | null => {
    if (!/^\d{1,3}$/.test(s)) return null;
    const n = Number(s);
    return n >= 1 && n <= 999 ? n : null;
  };

  const ziValid = /^[\u4e00-\u9fff]{1,2}$/.test(ziText.trim());
  const inputsValid =
    kind === "meihua"
      ? method === "time" ||
        (method === "zi" ? ziValid : validNum(numA) !== null && validNum(numB) !== null)
      : lyMethod !== "tosses" || lyTosses.every((t) => t !== null);
  const canCast = question.trim().length > 0 && inputsValid && !casting;
  // 逐爻摇卦由 StepShake 自带掷爻按钮驱动,主 CTA 隐藏
  const showMainCta = !(kind === "liuyao" && lyMethod === "step");

  const cast = useCallback(async () => {
    const q = question.trim();
    if (!q || casting) return;

    let numbers: number[] | undefined;
    if (kind === "meihua" && method === "number") {
      const a = validNum(numA);
      const b = validNum(numB);
      if (a === null || b === null) {
        setCastError("请各输入一个 1-999 的整数");
        return;
      }
      numbers = [a, b];
    }
    if (kind === "liuyao" && lyMethod === "tosses" && lyTosses.some((t) => t === null)) {
      setCastError("请为六爻逐一录入背面数");
      return;
    }

    castEpoch.current += 1;
    setCasting(true);
    setCastError(null);
    setReading(null);
    setAiError(null);

    const reduced = prefersReducedMotion();
    const startedAt = Date.now();
    try {
      if (kind === "liuyao") {
        const yong = lyYong || undefined;
        const input =
          lyMethod === "tosses"
            ? { method: "tosses" as const, tosses: lyTosses.map((t) => t ?? 0), question: q, yongShen: yong }
            : { method: "shake" as const, question: q, yongShen: yong };
        const { result: r, castAt: at, recordId } = await castLiuYao(input);
        const wait = (reduced ? 0 : CAST_ANIM_MS.liuyao) - (Date.now() - startedAt);
        if (wait > 0) await new Promise((res) => setTimeout(res, wait));
        setLyResult(r);
        setLyCastAt(at);
        setLyRecordId(recordId ?? null);
      } else {
        const input: CastInput =
          method === "number"
            ? { method: "number", numbers, question: q }
            : method === "zi"
              ? { method: "zi", ziText: ziText.trim(), question: q }
              : { method: "time", question: q };
        const { result: r, castAt: at, recordId } = await castMeihua(input);
        const wait = (reduced ? 0 : CAST_ANIM_MS.meihua) - (Date.now() - startedAt);
        if (wait > 0) await new Promise((res) => setTimeout(res, wait));
        setResult(r);
        setCastAt(at);
        setCastNumbers(numbers ?? null);
        setMhRecordId(recordId ?? null);
      }
      // 仪式收束镜头交接(与排盘罗盘同套):光圈聚拢结果区,结果自微放归位
      setResultCue((c) => c + 1);
      if (!reduced) {
        setSpot(true);
        setTimeout(() => setSpot(false), 1100);
      }
    } catch (e) {
      if (kind === "liuyao") setLyResult(null);
      else setResult(null);
      setCastError(e instanceof DivinationError ? e.message : "起卦失败,请重试");
    } finally {
      setCasting(false);
    }
  }, [question, kind, method, lyMethod, numA, numB, lyTosses, casting, ziText, lyYong]);

  // 逐爻摇卦完成:按六掷记录装卦(每掷动画即仪式,不再叠加整体动效)
  const castStep = useCallback(
    async (tosses: number[]) => {
      const q = question.trim();
      if (!q || casting) return;
      castEpoch.current += 1;
      setCasting(true);
      setCastError(null);
      setReading(null);
      setAiError(null);
      try {
        const { result: r, castAt: at, recordId } = await castLiuYao({ method: "tosses", tosses, question: q, yongShen: lyYong || undefined });
        setLyResult(r);
        setLyCastAt(at);
        setLyRecordId(recordId ?? null);
        setResultCue((c) => c + 1); // 逐爻仪式在掷钱本身,收束只做归位不加光圈
      } catch (e) {
        setLyResult(null);
        setCastError(e instanceof DivinationError ? e.message : "起卦失败,请重试");
      } finally {
        setCasting(false);
      }
    },
    [question, casting, lyYong],
  );

  const divine = useCallback(async () => {
    if (aiLoading) return;
    const epoch = castEpoch.current; // 解读只属于此刻所见之卦
    setAiError(null);
    setReading(null);
    setAiLoading(true);
    try {
      if (kind === "liuyao") {
        // 关键契约:回传起卦返回的 tosses + castAt,服务端按记录重装同一卦。
        if (!lyResult?.tosses || lyCastAt == null) return;
        const q = lyResult.question ?? question.trim();
        if (!q) return;
        const { reading: rd, remainingCredits } = await divineLiuYaoAI({
          tosses: lyResult.tosses,
          castAt: lyCastAt,
          question: q,
          recordId: lyRecordId ?? undefined,
          yongShen: lyResult.yongShenOverride || undefined, // 快照:与所见同一取用
        });
        if (epoch !== castEpoch.current) return;
        setReading(rd.text);
        setCredits(remainingCredits);
      } else {
        if (!result) return;
        const q = result.question ?? question.trim();
        if (!q) return;
        // 关键契约:回传与所见「同一卦」——castAt 一律回传(卦气旺衰随月令走,
        // 数字卦卦象虽由数定,断层力度仍锚定起卦时刻);数字卦另传同组 numbers。
        const input: CastInput & { question: string; recordId?: string } =
          result.method === "number"
            ? { method: "number", numbers: castNumbers ?? result.numbers, castAt: castAt ?? undefined, question: q, recordId: mhRecordId ?? undefined }
            : result.method === "zi"
              ? { method: "zi", ziText: result.ziText, castAt: castAt ?? undefined, question: q, recordId: mhRecordId ?? undefined }
              : { method: "time", castAt: castAt ?? undefined, question: q, recordId: mhRecordId ?? undefined };
        const { reading: rd, remainingCredits } = await divineAI(input);
        if (epoch !== castEpoch.current) return;
        setReading(rd.text);
        setCredits(remainingCredits);
      }
    } catch (e) {
      if (epoch !== castEpoch.current) return; // 旧卦的失败不打扰新卦
      if (e instanceof DivinationError) {
        if (e.status === 401) setAiError({ kind: "unauth" });
        else if (e.status === 402 || e.code === "no_credits") {
          setAiError({ kind: "no_credits" });
          setCredits(0);
        } else if (e.status === 503 || e.code === "ai_unavailable") setAiError({ kind: "ai_unavailable" });
        else setAiError({ kind: "network", message: e.message });
      } else {
        setAiError({ kind: "network", message: "网络异常,请重试" });
      }
    } finally {
      setAiLoading(false);
    }
  }, [kind, result, lyResult, lyCastAt, lyRecordId, mhRecordId, aiLoading, question, castAt, castNumbers]);

  const meta = KIND_META[kind];

  return (
    <div className="mx-auto max-w-4xl px-5 py-14 md:py-20">
      {/* ── 页头 ── */}
      <header>
        <p className="text-[12px] font-medium tracking-[0.24em] text-gold">{meta.eyebrow}</p>
        <div className="mt-3 flex flex-wrap items-baseline justify-between gap-3">
          <h1 className="font-display text-[39px] font-semibold text-ink sm:text-[49px]">{meta.title}</h1>
          <span className="flex items-center gap-2.5">
            {signedIn && (
              <Link
                href="/divinations"
                className="inline-flex items-center rounded-[2px] px-2 py-1 text-[12px] tracking-[0.08em] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] transition-colors hover:text-ink"
              >
                我的卦档
              </Link>
            )}
            {signedIn && <CreditsBadge credits={credits} />}
          </span>
        </div>
        <p className="mt-3 text-[15px] leading-relaxed text-ink-secondary md:text-[16px]">{meta.sub}</p>
      </header>

      {/* ── 起卦区 ── */}
      <section className="mt-10 rounded-[10px] bg-bg-raised px-5 py-7 shadow-[0_0_0_1px_var(--line)] md:px-8 md:py-8">
        <label htmlFor="dvn-q" className="text-[13px] font-medium tracking-[0.06em] text-gold">
          心中默念所问之事
        </label>
        <div className="relative mt-3">
          <textarea
            id="dvn-q"
            value={question}
            onChange={(e) => setQuestion(e.target.value.slice(0, MAX_Q))}
            maxLength={MAX_Q}
            rows={2}
            placeholder="例如:此番转职,可否顺遂?"
            className="w-full resize-none rounded-[6px] bg-bg px-4 py-3 text-[15px] leading-relaxed text-ink shadow-[inset_0_0_0_1px_var(--line)] outline-none transition-shadow placeholder:text-ink-faint focus:shadow-[inset_0_0_0_1px_var(--gold-dim)]"
          />
          <span className="tnum pointer-events-none absolute bottom-2.5 right-3 text-[11px] text-ink-faint">
            {question.length}/{MAX_Q}
          </span>
        </div>

        {/* 起卦方式(按占法) */}
        <div className="mt-6">
          <p className="mb-3 text-[12px] font-medium tracking-[0.08em] text-gold">起卦方式</p>
          {kind === "meihua" ? (
            <>
              <div className="flex flex-wrap gap-2">
                <MethodTab active={method === "time"} onClick={() => setMethod("time")} title="以此时起卦" hint="时间卦 · 主推" />
                <MethodTab active={method === "number"} onClick={() => setMethod("number")} title="报数起卦" hint="两数 1-999" />
                <MethodTab active={method === "zi"} onClick={() => setMethod("zi")} title="测字起卦" hint="一或两字 · 端法" />
              </div>
              {method === "zi" && (
                <div className="mt-4 flex items-center gap-3">
                  <label className="flex flex-col gap-1">
                    <span className="text-[12px] text-ink-faint">心中所感之字(一或二字)</span>
                    <input
                      value={ziText}
                      onChange={(e) => setZiText(e.target.value.slice(0, 2))}
                      placeholder="如「梅」或「转职」"
                      maxLength={2}
                      className="w-40 rounded-[6px] bg-bg px-4 py-3 text-center font-display text-[22px] tracking-[0.3em] text-ink shadow-[inset_0_0_0_1px_var(--line)] outline-none transition-shadow placeholder:text-[14px] placeholder:tracking-normal placeholder:text-ink-faint focus:shadow-[inset_0_0_0_1px_var(--gold-dim)]"
                    />
                  </label>
                  <p className="mt-5 max-w-[220px] text-[11px] leading-relaxed text-ink-faint">
                    一字:字画起上卦,加时辰配下卦;两字:两仪平分。笔画依 Unihan 简体。
                  </p>
                </div>
              )}
              {method === "number" && (
                <div className="mt-4 flex items-center gap-3">
                  <NumField label="上卦数" value={numA} onChange={setNumA} />
                  <span className="mt-5 text-ink-faint" aria-hidden>
                    ·
                  </span>
                  <NumField label="下卦数" value={numB} onChange={setNumB} />
                </div>
              )}
            </>
          ) : (
            <>
              <div className="flex flex-wrap gap-2">
                <MethodTab active={lyMethod === "step"} onClick={() => setLyMethod("step")} title="逐爻摇卦" hint="铜钱六掷 · 主推" />
                <MethodTab active={lyMethod === "shake"} onClick={() => setLyMethod("shake")} title="一键摇卦" hint="六爻齐落" />
                <MethodTab active={lyMethod === "tosses"} onClick={() => setLyMethod("tosses")} title="手动报爻" hint="自摇铜钱按爻录入" />
              </div>
              {/* 所占之人事(定用神):问者自陈优先于问辞推断——用神为断卦之纲 */}
              <div className="mt-5">
                <p className="mb-2 text-[12px] font-medium tracking-[0.08em] text-gold">所占之人事(定用神)</p>
                <div className="flex flex-wrap gap-1.5">
                  {([
                    ["", "自动识别"],
                    ["世爻", "问自己"],
                    ["妻财", "求财 · 问妻"],
                    ["官鬼", "官职官司 · 问夫"],
                    ["父母", "文书屋宅 · 问长辈"],
                    ["子孙", "子女解忧"],
                    ["兄弟", "朋友同辈"],
                  ] as const).map(([v, label]) => (
                    <button
                      key={v || "auto"}
                      type="button"
                      aria-pressed={lyYong === v}
                      onClick={() => setLyYong(v)}
                      className={[
                        "rounded-[4px] px-2.5 py-1 text-[12px] transition-colors",
                        lyYong === v
                          ? "bg-[var(--gold-glow)] font-medium text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]"
                          : "bg-bg text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] hover:text-ink",
                      ].join(" ")}
                    >
                      {label}
                    </button>
                  ))}
                </div>
              </div>
              {lyMethod === "step" && (
                <StepShake
                  key={lyCastAt ?? "fresh"}
                  canThrow={question.trim().length > 0 && !casting}
                  disabledHint="先写下所问之事,方可掷爻。"
                  onComplete={castStep}
                />
              )}
              {lyMethod === "tosses" && (
                <TossEntry tosses={lyTosses} onChange={setLyTosses} />
              )}
            </>
          )}
        </div>

        {/* 起卦按钮(本区唯一金色辉光主 CTA;逐爻摇卦时由掷爻按钮承担) */}
        <div className="mt-7 flex flex-col items-start gap-2.5">
          {showMainCta && (
            <button
              type="button"
              onClick={cast}
              disabled={!canCast}
              className="glow-gold inline-flex min-h-[48px] w-full items-center justify-center rounded-[6px] bg-gold px-8 py-3 text-[16px] font-medium text-[#161206] transition-colors hover:bg-gold-bright disabled:opacity-45 disabled:shadow-none sm:w-auto"
            >
              {casting ? meta.casting : meta.cta}
            </button>
          )}
          {showMainCta && !question.trim() && (
            <p className="text-[12px] text-ink-faint">先写下所问之事,方可起卦。</p>
          )}
          {castError && <p className="text-[13px] text-danger">{castError}</p>}
        </div>
      </section>

      {/* ── 起卦动效(逐爻摇卦的仪式在掷钱本身,不再叠加) ── */}
      {casting && (kind === "liuyao" ? (lyMethod === "step" ? null : <LiuYaoCast />) : <MeihuaCast />)}

      {/* ── 卦象展示(key=resultCue:收束时整区重挂,聚焦归位重放) ── */}
      {kind === "meihua" && result && !casting && (
        <div
          ref={resultRef}
          key={`mh-${resultCue}`}
          className={["page-enter mt-12 scroll-mt-20", resultCue > 0 ? "board-focus" : ""].join(" ")}
        >
          {/* 起卦信息 */}
          <div className="flex flex-col gap-2">
            <div className="tnum flex flex-wrap items-center gap-x-2 gap-y-1 text-[12px] tracking-[0.06em] text-ink-faint">
              <span>{result.method === "time" ? "心易起卦" : "报数起卦"}</span>
              {result.lunarText && <span>· 农历 {result.lunarText}</span>}
              {result.numbers && result.numbers.length > 0 && <span>· 报数 {result.numbers.join("、")}</span>}
              {result.castBasis && <span>· {result.castBasis}</span>}
            </div>
            {result.question && (
              <p className="font-reading text-[16px] text-ink-secondary">所问:{result.question}</p>
            )}
          </div>

          {/* 本卦(大);结果区各块与全站同套滚动聚焦节奏(reveal 渐进增强) */}
          <div className="reveal mt-8 flex justify-center rounded-[10px] bg-bg-raised px-4 py-8 shadow-[0_0_0_1px_var(--line)]">
            <HexagramView hexagram={result.ben} moving={result.moving} label="本卦" emphasis />
          </div>

          {/* 互卦 / 变卦 */}
          <div className="reveal mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div className="flex justify-center rounded-[10px] bg-bg-raised px-4 py-7 shadow-[0_0_0_1px_var(--line)]">
              <HexagramView hexagram={result.hu} moving={0} label="互卦" />
            </div>
            <div className="flex justify-center rounded-[10px] bg-bg-raised px-4 py-7 shadow-[0_0_0_1px_var(--line)]">
              <HexagramView hexagram={result.bian} moving={0} label="变卦" />
            </div>
          </div>

          {/* 周易经文(公版卦辞) */}
          {result.ben.guaCi && (
            <div className="reveal">
              <JingWenCard
                rows={[
                  { label: `本卦 ${result.ben.name}`, text: result.ben.guaCi },
                  ...(result.bian.guaCi ? [{ label: `变卦 ${result.bian.name}`, text: result.bian.guaCi }] : []),
                ]}
              />
            </div>
          )}

          {/* 体用生克 */}
          <div className="reveal">
            <TiYongCard result={result} />
          </div>

          {/* 断卦骨架(确定性:卦气旺衰/体党用党/互变分层/事类/应期) */}
          {result.judgment && (
            <div className="reveal">
              <JudgeCard j={result.judgment} />
            </div>
          )}

          {/* 万物类象(体/用/变取象) */}
          {result.lore && result.lore.length > 0 && (
            <div className="reveal">
              <LoreCard lore={result.lore} />
            </div>
          )}

          {/* ── AI 深度解卦 ── */}
          <div className="reveal mt-8">
            <AiSection
              signedIn={signedIn}
              credits={credits}
              loading={aiLoading}
              reading={reading}
              error={aiError}
              hint={meta.aiHint}
              archived={!!mhRecordId}
              homePath={homePath}
              onDivine={divine}
            />
          </div>
        </div>
      )}

      {kind === "liuyao" && lyResult && !casting && (
        <div
          ref={resultRef}
          key={`ly-${resultCue}`}
          className={["page-enter mt-12 scroll-mt-20", resultCue > 0 ? "board-focus" : ""].join(" ")}
        >
          {/* 起卦信息 */}
          <div className="flex flex-col gap-2">
            <div className="tnum flex flex-wrap items-center gap-x-2 gap-y-1 text-[12px] tracking-[0.06em] text-ink-faint">
              <span>{LY_METHOD_LABEL[lyMethod]}</span>
              <span>· 六爻纳甲</span>
            </div>
            {lyResult.question && (
              <p className="font-reading text-[16px] text-ink-secondary">所问:{lyResult.question}</p>
            )}
          </div>

          {/* 装卦盘面;结果区各块与全站同套滚动聚焦节奏(reveal 渐进增强) */}
          <div className="reveal mt-8">
            <LiuYaoPan result={lyResult} />
          </div>

          {/* 周易经文:本卦辞/动爻爻辞/变卦辞(公版) */}
          {lyResult.jingWen && (
            <div className="reveal">
              <JingWenCard
                rows={[
                  { label: `本卦 ${lyResult.benName}`, text: lyResult.jingWen.benGuaCi },
                  ...(lyResult.jingWen.yaoCi ?? []).map((yc) => ({ label: "动爻", text: yc, strong: true })),
                  ...(lyResult.jingWen.yong ? [{ label: "六爻皆动", text: lyResult.jingWen.yong, strong: true }] : []),
                  ...(lyResult.jingWen.bianGuaCi && lyResult.bianName
                    ? [{ label: `变卦 ${lyResult.bianName}`, text: lyResult.jingWen.bianGuaCi }]
                    : []),
                ]}
              />
            </div>
          )}

          {/* 确定性断语骨架(免费层) */}
          {lyResult.judgment && (
            <div className="reveal">
              <LiuYaoJudgeCard j={lyResult.judgment} xingZhi={lyResult.benXingZhi} />
            </div>
          )}

          {/* ── AI 深度解卦 ── */}
          <div className="reveal mt-8">
            <AiSection
              signedIn={signedIn}
              credits={credits}
              loading={aiLoading}
              reading={reading}
              error={aiError}
              hint={meta.aiHint}
              archived={!!lyRecordId}
              homePath={homePath}
              onDivine={divine}
            />
          </div>
        </div>
      )}

      {/* 收束光圈:仪式幕布落下时光聚结果区再徐徐散开(镜头交接) */}
      {spot && <div className="cast-spot" aria-hidden />}
    </div>
  );
}
