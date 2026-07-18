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
  RELATION_TONE,
  TOSS_OPTIONS,
  DivinationError,
  type CastInput,
  type MeihuaResult,
  type LiuYaoResult,
} from "@/lib/divination";
import { ReportText } from "@/components/profiles/ReportText";
import { HexagramView } from "@/components/divination/HexagramView";
import { CastRitual } from "@/components/divination/CastRitual";
import { ShakeRitual } from "@/components/divination/ShakeRitual";
import { StepShake } from "@/components/divination/StepShake";
import { LiuYaoPan } from "@/components/divination/LiuYaoPan";
import { XiaoLiuRen } from "@/components/divination/XiaoLiuRen";
import { toneBadgeClass } from "@/components/divination/tone";
import { prefersReducedMotion } from "@/components/divination/useReducedMotion";

type Kind = "meihua" | "liuyao";
type CastMethod = "time" | "number";
type LiuYaoMethod = "step" | "shake" | "tosses";

const LY_METHOD_LABEL: Record<LiuYaoMethod, string> = {
  step: "逐爻摇卦",
  shake: "铜钱摇卦",
  tosses: "手动报爻",
};

type AiErr =
  | { kind: "unauth" }
  | { kind: "no_credits" }
  | { kind: "ai_unavailable" }
  | { kind: "network"; message: string };

const MAX_Q = 200;
// 起卦动效总时长封顶(六爻六位落定稍长)
const CAST_ANIM_MS: Record<Kind, number> = { meihua: 1400, liuyao: 1600 };

const KIND_META: Record<Kind, { eyebrow: string; cta: string; casting: string; aiHint: string }> = {
  meihua: {
    eyebrow: "占卜 · 梅花易数",
    cta: "起卦",
    casting: "起卦中…",
    aiHint: "依本卦 / 互卦 / 变卦与体用生克逐层解读",
  },
  liuyao: {
    eyebrow: "占卜 · 六爻纳甲",
    cta: "摇卦",
    casting: "摇卦中…",
    aiHint: "依用神 / 世应 / 六亲六神与动变断成败应期",
  },
};

/** 问卦:梅花易数 / 六爻纳甲起卦 + 卦象展示 + AI 深度解卦,附小六壬快占。 */
export default function DivinationPage() {
  const [signedIn, setSignedIn] = useState<boolean | null>(null);
  const [credits, setCredits] = useState<number | null>(null); // null=未知/未登录

  // 起卦输入
  const [question, setQuestion] = useState("");
  const [kind, setKind] = useState<Kind>("meihua");
  const [method, setMethod] = useState<CastMethod>("time");
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

  // 切换占法:清 AI 态与错误(各占法卦象留存)
  const switchKind = (k: Kind) => {
    if (k === kind) return;
    setKind(k);
    setCastError(null);
    setReading(null);
    setAiError(null);
  };

  const validNum = (s: string): number | null => {
    if (!/^\d{1,3}$/.test(s)) return null;
    const n = Number(s);
    return n >= 1 && n <= 999 ? n : null;
  };

  const inputsValid =
    kind === "meihua"
      ? method === "time" || (validNum(numA) !== null && validNum(numB) !== null)
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

    setCasting(true);
    setCastError(null);
    setReading(null);
    setAiError(null);

    const reduced = prefersReducedMotion();
    const startedAt = Date.now();
    try {
      if (kind === "liuyao") {
        const input =
          lyMethod === "tosses"
            ? { method: "tosses" as const, tosses: lyTosses.map((t) => t ?? 0), question: q }
            : { method: "shake" as const, question: q };
        const { result: r, castAt: at, recordId } = await castLiuYao(input);
        const wait = (reduced ? 0 : CAST_ANIM_MS.liuyao) - (Date.now() - startedAt);
        if (wait > 0) await new Promise((res) => setTimeout(res, wait));
        setLyResult(r);
        setLyCastAt(at);
        setLyRecordId(recordId ?? null);
      } else {
        const input: CastInput =
          method === "number" ? { method: "number", numbers, question: q } : { method: "time", question: q };
        const { result: r, castAt: at, recordId } = await castMeihua(input);
        const wait = (reduced ? 0 : CAST_ANIM_MS.meihua) - (Date.now() - startedAt);
        if (wait > 0) await new Promise((res) => setTimeout(res, wait));
        setResult(r);
        setCastAt(at);
        setCastNumbers(numbers ?? null);
        setMhRecordId(recordId ?? null);
      }
    } catch (e) {
      if (kind === "liuyao") setLyResult(null);
      else setResult(null);
      setCastError(e instanceof DivinationError ? e.message : "起卦失败,请重试");
    } finally {
      setCasting(false);
    }
  }, [question, kind, method, lyMethod, numA, numB, lyTosses, casting]);

  // 逐爻摇卦完成:按六掷记录装卦(每掷动画即仪式,不再叠加整体动效)
  const castStep = useCallback(
    async (tosses: number[]) => {
      const q = question.trim();
      if (!q || casting) return;
      setCasting(true);
      setCastError(null);
      setReading(null);
      setAiError(null);
      try {
        const { result: r, castAt: at, recordId } = await castLiuYao({ method: "tosses", tosses, question: q });
        setLyResult(r);
        setLyCastAt(at);
        setLyRecordId(recordId ?? null);
      } catch (e) {
        setLyResult(null);
        setCastError(e instanceof DivinationError ? e.message : "起卦失败,请重试");
      } finally {
        setCasting(false);
      }
    },
    [question, casting],
  );

  const divine = useCallback(async () => {
    if (aiLoading) return;
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
        });
        setReading(rd.text);
        setCredits(remainingCredits);
      } else {
        if (!result) return;
        const q = result.question ?? question.trim();
        if (!q) return;
        // 关键契约:回传与所见「同一卦」——时间卦传 castAt,数字卦传 numbers。
        const input: CastInput & { question: string; recordId?: string } =
          result.method === "number"
            ? { method: "number", numbers: castNumbers ?? result.numbers, question: q, recordId: mhRecordId ?? undefined }
            : { method: "time", castAt: castAt ?? undefined, question: q, recordId: mhRecordId ?? undefined };
        const { reading: rd, remainingCredits } = await divineAI(input);
        setReading(rd.text);
        setCredits(remainingCredits);
      }
    } catch (e) {
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
          <h1 className="font-display text-[39px] font-semibold text-ink sm:text-[49px]">问卦</h1>
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
        <p className="mt-3 text-[15px] leading-relaxed text-ink-secondary md:text-[16px]">
          一事一占,以卦观势。心念既定,起卦以问。
        </p>
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

        {/* 占法 */}
        <div className="mt-6">
          <p className="mb-3 text-[12px] font-medium tracking-[0.08em] text-gold">占法</p>
          <div className="flex flex-wrap gap-2">
            <MethodTab
              active={kind === "meihua"}
              onClick={() => switchKind("meihua")}
              title="梅花易数"
              hint="心易时数 · 体用生克"
            />
            <MethodTab
              active={kind === "liuyao"}
              onClick={() => switchKind("liuyao")}
              title="六爻纳甲"
              hint="铜钱摇卦 · 装卦断事"
            />
          </div>
        </div>

        {/* 起卦方式(按占法) */}
        <div className="mt-6">
          <p className="mb-3 text-[12px] font-medium tracking-[0.08em] text-gold">起卦方式</p>
          {kind === "meihua" ? (
            <>
              <div className="flex flex-wrap gap-2">
                <MethodTab active={method === "time"} onClick={() => setMethod("time")} title="以此时起卦" hint="时间卦 · 主推" />
                <MethodTab active={method === "number"} onClick={() => setMethod("number")} title="报数起卦" hint="两数 1-999" />
              </div>
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
      {casting && (kind === "liuyao" ? (lyMethod === "step" ? null : <ShakeRitual />) : <CastRitual />)}

      {/* ── 卦象展示 ── */}
      {kind === "meihua" && result && !casting && (
        <div ref={resultRef} className="page-enter mt-12 scroll-mt-20">
          {/* 起卦信息 */}
          <div className="flex flex-col gap-2">
            <div className="tnum flex flex-wrap items-center gap-x-2 gap-y-1 text-[12px] tracking-[0.06em] text-ink-faint">
              <span>{result.method === "time" ? "时间起卦" : "报数起卦"}</span>
              {result.lunarText && <span>· 农历 {result.lunarText}</span>}
              {result.numbers && result.numbers.length > 0 && <span>· 报数 {result.numbers.join("、")}</span>}
            </div>
            {result.question && (
              <p className="font-reading text-[16px] text-ink-secondary">所问:{result.question}</p>
            )}
          </div>

          {/* 本卦(大) */}
          <div className="mt-8 flex justify-center rounded-[10px] bg-bg-raised px-4 py-8 shadow-[0_0_0_1px_var(--line)]">
            <HexagramView hexagram={result.ben} moving={result.moving} label="本卦" emphasis />
          </div>

          {/* 互卦 / 变卦 */}
          <div className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div className="flex justify-center rounded-[10px] bg-bg-raised px-4 py-7 shadow-[0_0_0_1px_var(--line)]">
              <HexagramView hexagram={result.hu} moving={0} label="互卦" />
            </div>
            <div className="flex justify-center rounded-[10px] bg-bg-raised px-4 py-7 shadow-[0_0_0_1px_var(--line)]">
              <HexagramView hexagram={result.bian} moving={0} label="变卦" />
            </div>
          </div>

          {/* 体用生克 */}
          <TiYongCard result={result} />

          {/* ── AI 深度解卦 ── */}
          <div className="mt-8">
            <AiSection
              signedIn={signedIn}
              credits={credits}
              loading={aiLoading}
              reading={reading}
              error={aiError}
              hint={meta.aiHint}
              archived={!!mhRecordId}
              onDivine={divine}
            />
          </div>
        </div>
      )}

      {kind === "liuyao" && lyResult && !casting && (
        <div ref={resultRef} className="page-enter mt-12 scroll-mt-20">
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

          {/* 装卦盘面 */}
          <div className="mt-8">
            <LiuYaoPan result={lyResult} />
          </div>

          {/* ── AI 深度解卦 ── */}
          <div className="mt-8">
            <AiSection
              signedIn={signedIn}
              credits={credits}
              loading={aiLoading}
              reading={reading}
              error={aiError}
              hint={meta.aiHint}
              archived={!!lyRecordId}
              onDivine={divine}
            />
          </div>
        </div>
      )}

      {/* ── 小六壬快占 ── */}
      <XiaoLiuRen />
    </div>
  );
}

/** 剩余解卦次数徽标。 */
function CreditsBadge({ credits }: { credits: number | null }) {
  if (credits == null) return null;
  if (credits <= 0) {
    return (
      <Link
        href="/pricing"
        className="tnum inline-flex items-center rounded-[2px] px-2 py-1 text-[12px] tracking-[0.08em] text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)] transition-colors hover:bg-bg-raised"
      >
        解卦次数不足 · 去购买
      </Link>
    );
  }
  return (
    <span className="tnum inline-flex items-center rounded-[2px] px-2 py-1 text-[12px] tracking-[0.08em] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)]">
      解卦剩余 {credits} 次
    </span>
  );
}

/** 起卦方式分段项。 */
function MethodTab({
  active,
  onClick,
  title,
  hint,
}: {
  active: boolean;
  onClick: () => void;
  title: string;
  hint: string;
}) {
  return (
    <button
      type="button"
      aria-pressed={active}
      onClick={onClick}
      className={[
        "flex min-h-[44px] flex-col items-start gap-0.5 rounded-[6px] px-4 py-2.5 text-left transition-shadow",
        active
          ? "bg-[var(--gold-glow)] shadow-[inset_0_0_0_1px_var(--gold-dim)]"
          : "bg-bg shadow-[inset_0_0_0_1px_var(--line)] hover:shadow-[inset_0_0_0_1px_var(--line-strong)]",
      ].join(" ")}
    >
      <span className={`text-[14px] font-medium ${active ? "text-gold" : "text-ink"}`}>{title}</span>
      <span className="text-[11px] text-ink-faint">{hint}</span>
    </button>
  );
}

/** 报数输入(1-999,仅数字)。 */
function NumField({ label, value, onChange }: { label: string; value: string; onChange: (v: string) => void }) {
  return (
    <label className="flex flex-col gap-1.5">
      <span className="text-[11px] text-ink-faint">{label}</span>
      <input
        type="text"
        inputMode="numeric"
        value={value}
        onChange={(e) => onChange(e.target.value.replace(/\D/g, "").slice(0, 3))}
        placeholder="1-999"
        className="tnum w-24 rounded-[6px] bg-bg px-3 py-2.5 text-center text-[16px] text-ink shadow-[inset_0_0_0_1px_var(--line)] outline-none transition-shadow placeholder:text-ink-faint focus:shadow-[inset_0_0_0_1px_var(--gold-dim)]"
      />
    </label>
  );
}

const YAO_NAMES = ["初爻", "二爻", "三爻", "四爻", "五爻", "上爻"];

/** 六爻报爻录入:初爻在上(先摇先录),每爻四选一(背面数)。 */
function TossEntry({ tosses, onChange }: { tosses: (number | null)[]; onChange: (t: (number | null)[]) => void }) {
  const set = (i: number, backs: number) => {
    const next = [...tosses];
    next[i] = backs;
    onChange(next);
  };
  return (
    <div className="mt-4 flex flex-col gap-2">
      <p className="text-[12px] leading-relaxed text-ink-faint">
        以三枚铜钱自摇六次,自初爻起逐次录入每掷的背面枚数(字面朝上不计)。
      </p>
      {YAO_NAMES.map((name, i) => (
        <div key={name} className="flex items-center gap-2.5">
          <span className="w-9 shrink-0 text-[12px] text-ink-secondary">{name}</span>
          <div className="flex flex-1 flex-wrap gap-1.5">
            {TOSS_OPTIONS.map((opt) => {
              const active = tosses[i] === opt.backs;
              return (
                <button
                  key={opt.backs}
                  type="button"
                  aria-pressed={active}
                  onClick={() => set(i, opt.backs)}
                  className={[
                    "flex min-h-[38px] flex-col items-center justify-center rounded-[4px] px-2.5 py-1 transition-shadow",
                    active
                      ? "bg-[var(--gold-glow)] shadow-[inset_0_0_0_1px_var(--gold-dim)]"
                      : "bg-bg shadow-[inset_0_0_0_1px_var(--line)] hover:shadow-[inset_0_0_0_1px_var(--line-strong)]",
                  ].join(" ")}
                >
                  <span className={`text-[12px] leading-tight ${active ? "text-gold" : "text-ink"}`}>{opt.label}</span>
                  <span className="text-[10px] leading-tight text-ink-faint">{opt.hint}</span>
                </button>
              );
            })}
          </div>
        </div>
      ))}
    </div>
  );
}

/** 体用生克卡。 */
function TiYongCard({ result }: { result: MeihuaResult }) {
  const rel = RELATION_TONE[result.relation];
  return (
    <div className="mt-4 rounded-[10px] bg-bg-raised px-5 py-6 shadow-[0_0_0_1px_var(--line)] md:px-8">
      <p className="text-[12px] font-medium tracking-[0.24em] text-gold">体用生克</p>
      <div className="mt-4 grid grid-cols-2 gap-3">
        <TiYongCell
          role="体"
          position={result.tiIsUpper ? "上卦" : "下卦"}
          trigram={result.tiTrigram.name}
          element={result.tiTrigram.element}
        />
        <TiYongCell
          role="用"
          position={result.tiIsUpper ? "下卦" : "上卦"}
          trigram={result.yongTrigram.name}
          element={result.yongTrigram.element}
        />
      </div>
      <div className="mt-5 flex flex-wrap items-center gap-3">
        <span
          className={`inline-flex items-center rounded-[2px] px-2.5 py-1 text-[13px] font-medium tracking-[0.06em] ${toneBadgeClass(
            rel.tone,
          )}`}
        >
          {rel.label}
        </span>
        <p className="flex-1 text-[14px] leading-relaxed text-ink-secondary">{result.verdict}</p>
      </div>
    </div>
  );
}

function TiYongCell({
  role,
  position,
  trigram,
  element,
}: {
  role: string;
  position: string;
  trigram: string;
  element: string;
}) {
  return (
    <div className="flex flex-col gap-1 rounded-[6px] bg-bg px-4 py-4 shadow-[inset_0_0_0_1px_var(--line)]">
      <span className="text-[11px] tracking-[0.1em] text-ink-faint">
        {role} · {position}
      </span>
      <div className="flex items-baseline gap-2">
        <span className="font-display text-[25px] font-semibold text-ink">{trigram}</span>
        <span className="text-[13px] text-ink-secondary">五行 · {element}</span>
      </div>
    </div>
  );
}

/** AI 深度解卦区:登录/次数/加载/错误/结果分态。 */
function AiSection({
  signedIn,
  credits,
  loading,
  reading,
  error,
  hint,
  archived,
  onDivine,
}: {
  signedIn: boolean | null;
  credits: number | null;
  loading: boolean;
  reading: string | null;
  error: AiErr | null;
  hint: string;
  archived?: boolean;
  onDivine: () => void;
}) {
  // 已出结果
  if (reading != null && !loading) {
    return (
      <article className="rounded-[10px] bg-bg-raised px-6 py-8 shadow-[0_0_0_1px_var(--line)] md:px-10 md:py-10">
        <p className="mb-5 text-[12px] font-medium tracking-[0.24em] text-gold">AI 深度解卦</p>
        <ReportText text={reading} />
        <div className="mt-6 flex flex-wrap items-center justify-between gap-2 border-t border-line pt-4">
          <span className="flex flex-wrap items-center gap-x-3 gap-y-1">
            {credits != null && <span className="tnum text-[12px] text-ink-faint">解卦剩余 {credits} 次</span>}
            {archived && (
              <Link href="/divinations" className="text-[12px] text-gold transition-opacity hover:opacity-80">
                已存入卦档 · 查看
              </Link>
            )}
          </span>
          <span className="text-[11px] leading-relaxed text-ink-faint">占卜为传统文化参考,不构成决策建议。</span>
        </div>
      </article>
    );
  }

  // 加载骨架(解卦较慢,>20s 常见)
  if (loading) {
    return (
      <div
        className="rounded-[10px] bg-bg-raised px-6 py-8 shadow-[0_0_0_1px_var(--line)]"
        role="status"
        aria-label="AI 正在解卦"
      >
        <div className="h-4 w-24 animate-pulse rounded-[2px] bg-line" aria-hidden />
        <div className="mt-5 flex flex-col gap-3">
          {[94, 100, 86, 96, 72].map((w, i) => (
            <div key={i} className="h-4 animate-pulse rounded-[2px] bg-line" style={{ width: `${w}%` }} aria-hidden />
          ))}
        </div>
        <p className="mt-6 text-[12px] text-ink-faint">AI 正在依卦象逐层解读,通常需 20 秒以上,请勿离开…</p>
      </div>
    );
  }

  // 错误分态
  if (error) {
    if (error.kind === "unauth") {
      return (
        <PromptBar tone="gold" text="登录后即可 AI 深度解卦。" action={{ href: "/login?next=/divination", label: "去登录" }} />
      );
    }
    if (error.kind === "no_credits") {
      return (
        <PromptBar tone="gold" text="解卦次数不足,购买次卡后再试。" action={{ href: "/pricing", label: "去购买" }} />
      );
    }
    if (error.kind === "ai_unavailable") {
      return <PromptBar tone="warn" text="AI 服务暂不可用,本次未扣次数,请稍后再试。" retry={onDivine} />;
    }
    return <PromptBar tone="danger" text={error.message || "解卦失败,请重试。"} retry={onDivine} />;
  }

  // 初始:登录 / 次数 / 解卦按钮(本区唯一金色辉光主 CTA)
  if (signedIn === false) {
    return (
      <Link
        href="/login?next=/divination"
        className="glow-gold inline-flex min-h-[48px] w-full items-center justify-center rounded-[6px] bg-gold px-7 py-3 text-[16px] font-medium text-[#161206] transition-colors hover:bg-gold-bright sm:w-auto"
      >
        登录后 AI 深度解卦
      </Link>
    );
  }
  if (signedIn && credits === 0) {
    return <PromptBar tone="gold" text="解卦次数不足,购买次卡后即可深度解卦。" action={{ href: "/pricing", label: "去购买" }} />;
  }
  return (
    <div className="flex flex-col items-start gap-2.5">
      <button
        type="button"
        onClick={onDivine}
        className="glow-gold inline-flex min-h-[48px] w-full items-center justify-center rounded-[6px] bg-gold px-7 py-3 text-[16px] font-medium text-[#161206] transition-colors hover:bg-gold-bright sm:w-auto"
      >
        AI 深度解卦(消耗 1 次)
      </button>
      <p className="text-[12px] text-ink-faint">
        {hint}
        {credits != null ? `,当前剩余 ${credits} 次` : ""}。
      </p>
    </div>
  );
}

/** 提示条:登录/购买/重试。 */
function PromptBar({
  tone,
  text,
  action,
  retry,
}: {
  tone: "gold" | "warn" | "danger";
  text: string;
  action?: { href: string; label: string };
  retry?: () => void;
}) {
  const ring =
    tone === "gold"
      ? "shadow-[inset_0_0_0_1px_var(--gold-dim)]"
      : tone === "warn"
        ? "shadow-[inset_0_0_0_1px_var(--warn)]"
        : "shadow-[inset_0_0_0_1px_var(--danger)]";
  const textColor = tone === "gold" ? "text-ink-secondary" : tone === "warn" ? "text-warn" : "text-danger";
  return (
    <div className={`flex flex-wrap items-center justify-between gap-3 rounded-[6px] bg-bg-raised px-4 py-3 ${ring}`}>
      <p className={`text-[14px] ${textColor}`}>{text}</p>
      {action && (
        <Link
          href={action.href}
          className="inline-flex min-h-[44px] shrink-0 items-center rounded-[6px] bg-gold px-4 py-2 text-[14px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
        >
          {action.label}
        </Link>
      )}
      {retry && (
        <button
          type="button"
          onClick={retry}
          className="inline-flex min-h-[44px] shrink-0 items-center rounded-[6px] bg-bg px-4 py-2 text-[14px] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] transition-colors hover:text-ink"
        >
          重试
        </button>
      )}
    </div>
  );
}
