"use client";

import { useEffect, useRef, useState } from "react";
import { fetchHeming, ApiError } from "@/lib/api";
import type { HemingResponse, HemingSide } from "@/lib/types";
import { ChartBoard } from "@/components/chart/ChartBoard";
import { Markdown } from "@/components/chat/Markdown";
import { BirthFields, toBirthInfo, type BirthValue } from "@/components/heming/BirthFields";
import { HemingReadingPanel } from "@/components/heming/HemingReadingPanel";
import { prefersReducedMotion } from "@/components/divination/useReducedMotion";

const DEFAULT_A: BirthValue = { name: "", date: "1990-06-15", hour: 6, gender: "female" };
const DEFAULT_B: BirthValue = { name: "", date: "1988-03-02", hour: 7, gender: "male" };

/** 星级顺序(评分标准展示用)。 */
const SCORE_ORDER = ["五星", "四星", "三星", "二星", "一星"];

/** 合盘:倪师双宫联参 —— 双方命盘 + 夫妻宫断语 + 方法论。 */
export default function HemingPage() {
  const [a, setA] = useState<BirthValue>(DEFAULT_A);
  const [b, setB] = useState<BirthValue>(DEFAULT_B);
  const [result, setResult] = useState<HemingResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [showMethod, setShowMethod] = useState(false);
  // 农历换算中/失败(任一方):禁提交,防竞态提交旧公历值
  const [pendA, setPendA] = useState(false);
  const [pendB, setPendB] = useState(false);
  // 收束镜头交接(与排盘/占卜同套):光圈聚拢结果区 + 滚入视口
  const [cue, setCue] = useState(0);
  const [spot, setSpot] = useState(false);
  const resultRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (cue === 0) return;
    resultRef.current?.scrollIntoView({
      behavior: prefersReducedMotion() ? "auto" : "smooth",
      block: "start",
    });
  }, [cue]);

  async function run() {
    const ba = toBirthInfo(a);
    const bb = toBirthInfo(b);
    if (!ba || !bb) {
      setError("请填写双方 1900-2100 之间的有效公历生日");
      return;
    }
    setLoading(true);
    setError("");
    try {
      setResult(await fetchHeming(ba, bb));
      setCue((c) => c + 1);
      if (!prefersReducedMotion()) {
        setSpot(true);
        setTimeout(() => setSpot(false), 1100);
      }
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "合盘失败,请稍后重试");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="mx-auto max-w-6xl px-4 py-14 md:py-24">
      <p className="text-[12px] font-medium tracking-[0.24em] text-gold">合 盘 · 双宫联参</p>
      <h1 className="mt-3 font-display text-[31px] font-semibold sm:text-[39px]">双人合盘</h1>
      <p className="mt-3 max-w-2xl text-[15px] leading-relaxed text-ink-secondary md:text-[16px]">
        依倪师口径:看婚姻必夫妻宫与福德宫同参。输入双方生辰,得到两张命盘、
        夫妻宫主星断语与合盘方法论。
      </p>

      <div className="mt-10 flex flex-col gap-5 md:flex-row">
        <BirthFields title="甲方" value={a} onChange={setA} onPendingChange={setPendA} />
        <BirthFields title="乙方" value={b} onChange={setB} onPendingChange={setPendB} />
      </div>

      {error && (
        <p className="mt-5 rounded-[6px] bg-bg-raised px-4 py-3 text-[14px] text-danger shadow-[0_0_0_1px_var(--danger)]">
          {error}
        </p>
      )}

      <button
        type="button"
        onClick={run}
        disabled={loading || pendA || pendB}
        className="glow-gold mt-6 inline-flex min-h-[44px] items-center rounded-[6px] bg-gold px-8 py-3 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright disabled:opacity-50 disabled:shadow-none"
      >
        {loading ? "合盘中…" : pendA || pendB ? "换算中…" : "开始合盘"}
      </button>

      {spot && <div className="cast-spot" aria-hidden />}

      {result && (
        <div ref={resultRef} className="mt-12 flex flex-col gap-8 scroll-mt-24 md:mt-16">
          {/* 合盘契合度(确定性,比对双盘);结果区各块与全站同套滚动聚焦节奏 */}
          <div className="reveal">
            <HemingReadingPanel reading={result.reading} />
          </div>

          {/* 夫妻宫断语(双方并列) */}
          <div className="reveal grid gap-6 lg:grid-cols-2">
            <SidePanel title={a.name || "甲方"} side={result.a} />
            <SidePanel title={b.name || "乙方"} side={result.b} />
          </div>

          {/* 缘分评级标准 */}
          <section className="reveal rounded-[10px] bg-bg-raised p-6 shadow-[0_0_0_1px_var(--line)] md:p-8">
            <h2 className="font-display text-[20px] font-semibold">缘分评级标准</h2>
            <p className="mt-1.5 text-[13px] leading-relaxed text-ink-faint">
              倪师体系的合盘参照系,对照双方夫妻宫/福德宫星情自评。
            </p>
            <dl className="mt-4 flex flex-col gap-2">
              {SCORE_ORDER.filter((k) => result.scoreCriteria[k]).map((k) => (
                <div key={k} className="flex gap-3 rounded-[6px] bg-bg px-3 py-2 shadow-[inset_0_0_0_1px_var(--line)]">
                  <dt className="w-10 shrink-0 font-display text-gold">{k}</dt>
                  <dd className="text-[14px] leading-relaxed text-ink-secondary">{result.scoreCriteria[k]}</dd>
                </div>
              ))}
            </dl>
          </section>

          {/* 双方盘面(移动端显式 1 列:auto 轨道会被盘面 min-w 撑开导致整页溢出) */}
          <div className="reveal grid grid-cols-1 gap-8 xl:grid-cols-2">
            {([["甲方", result.a, a], ["乙方", result.b, b]] as const).map(([label, side, v]) => (
              <section key={label}>
                <h3 className="mb-3 font-display text-[15px] text-ink-secondary">
                  {v.name || label} 命盘
                </h3>
                <div className="overflow-x-auto">
                  <div className="min-w-[560px]">
                    <ChartBoard
                      chart={side.chart}
                      density="simple"
                      selectedBranch={null}
                      onSelectBranch={() => {}}
                    />
                  </div>
                </div>
              </section>
            ))}
          </div>

          {/* 方法论(折叠) */}
          <section className="reveal rounded-[10px] bg-bg-raised p-6 shadow-[0_0_0_1px_var(--line)] md:p-8">
            <button
              type="button"
              onClick={() => setShowMethod((s) => !s)}
              aria-expanded={showMethod}
              className="flex min-h-[44px] w-full items-center justify-between text-left"
            >
              <span className="font-display text-[20px] font-semibold">合盘方法论(倪海厦体系)</span>
              <span className="text-[13px] text-gold">{showMethod ? "收起" : "展开"}</span>
            </button>
            {showMethod && (
              <div className="prose-reader mt-4 text-[14px] leading-relaxed text-ink-secondary">
                <Markdown text={result.methodology} />
              </div>
            )}
          </section>

          <p className="text-center text-[12px] text-ink-faint">
            合盘结果为传统文化参考,婚恋决策请以现实相处为准。
          </p>
        </div>
      )}
    </div>
  );
}

/** 一方的夫妻宫断语面板。 */
function SidePanel({ title, side }: { title: string; side: HemingSide }) {
  return (
    <section className="rounded-[10px] bg-bg-raised p-6 shadow-[0_0_0_1px_var(--line)] md:p-8">
      <div className="flex items-center gap-2">
        <h2 className="font-display text-[20px] font-semibold">{title} · 夫妻宫</h2>
        {side.fuqiBorrowed && (
          <span className="rounded-[2px] px-1.5 py-0.5 text-[11px] text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]">
            空宫借对宫
          </span>
        )}
      </div>
      <div className="mt-2 flex flex-wrap gap-1.5">
        {side.fuqiStars.map((s) => (
          <span key={s} className="rounded-[4px] bg-bg px-2 py-0.5 font-display text-[14px] text-ink shadow-[inset_0_0_0_1px_var(--line-strong)]">
            {s}
          </span>
        ))}
      </div>

      <div className="mt-4 flex flex-col gap-3">
        {side.readings.map((r) => (
          <details key={r.star} className="group rounded-[6px] bg-bg p-3 shadow-[inset_0_0_0_1px_var(--line)]">
            <summary className="flex cursor-pointer list-none items-baseline justify-between gap-2">
              <span>
                <span className="font-display text-[15px] font-medium text-ink">{r.star}</span>
                <span className="ml-2 text-[13px] text-ink-secondary">{r.summary}</span>
              </span>
              <span className="shrink-0 text-[12px] text-ink-faint transition-transform group-open:rotate-90">›</span>
            </summary>
            <dl className="mt-3 flex flex-col gap-2 text-[13px] leading-relaxed">
              {([
                ["吉", r.good],
                ["凶", r.bad],
                ["配偶特质", r.spouseTraits],
                ["时机", r.timing],
              ] as const).map(([label, text]) =>
                text ? (
                  <div key={label} className="flex gap-2">
                    <dt className="w-14 shrink-0 text-ink-faint">{label}</dt>
                    <dd className="text-ink-secondary">{text}</dd>
                  </div>
                ) : null,
              )}
              {r.niQuote && (
                <blockquote className="mt-1 border-l-2 border-gold-dim pl-3 text-ink-secondary">
                  倪师:「{r.niQuote}」
                </blockquote>
              )}
            </dl>
          </details>
        ))}
      </div>
    </section>
  );
}
