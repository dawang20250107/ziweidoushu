"use client";

import { useCallback, useState } from "react";
import { castXiaoLiuRen, luckTone, DivinationError, type XiaoLiuRenResult } from "@/lib/divination";
import { toneBadgeClass, toneTextClass } from "./tone";
import { JudgeSections } from "./JudgeSections";

/**
 * 小六壬快占(独立轻区块):一键掐指起算,三步路径依次点亮 + 结果断语。
 * 免费、可反复;不占用 AI 解卦次数。本区块 CTA 用金描边(非辉光),
 * 辉光主 CTA 留给起卦与 AI 解卦。
 */
export function XiaoLiuRen() {
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<XiaoLiuRenResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [nonce, setNonce] = useState(0); // 每次起算换 key,重放点亮动效

  const run = useCallback(async () => {
    if (loading) return;
    setLoading(true);
    setError(null);
    try {
      const { result: r } = await castXiaoLiuRen();
      setResult(r);
      setNonce((n) => n + 1);
    } catch (e) {
      setError(e instanceof DivinationError ? e.message : "起算失败,请重试");
    } finally {
      setLoading(false);
    }
  }, [loading]);

  return (
    <section className="mt-14 rounded-[10px] bg-bg-raised px-5 py-8 shadow-[0_0_0_1px_var(--line)] md:px-8">
      <style>{`
        @keyframes dvn-light {
          0%   { opacity: 0; transform: translateY(6px) scale(0.96); }
          100% { opacity: 1; transform: none; }
        }
        .dvn-step { animation: dvn-light 0.34s var(--ease-out) both; }
        @media (prefers-reduced-motion: reduce) {
          .dvn-step { animation: none; opacity: 1; transform: none; }
        }
      `}</style>

      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <p className="text-[12px] font-medium tracking-[0.24em] text-gold">小六壬 · 急事速占</p>
          <h2 className="mt-2 font-display text-[25px] font-semibold text-ink">掐指一算</h2>
          <p className="mt-2 text-[14px] leading-relaxed text-ink-secondary">
            急事当下起算,掐指三步定吉凶缓急。免费,可反复。
          </p>
        </div>
        <button
          type="button"
          onClick={run}
          disabled={loading}
          className="inline-flex min-h-[44px] shrink-0 items-center rounded-[6px] bg-bg-raised px-5 py-2.5 text-[15px] font-medium text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)] transition-shadow hover:shadow-[inset_0_0_0_1px_var(--gold)] disabled:opacity-50"
        >
          {loading ? "掐指中…" : result ? "再算一次" : "急事速占"}
        </button>
      </div>

      {error && (
        <p className="mt-6 rounded-[6px] bg-bg px-4 py-3 text-[14px] text-danger shadow-[inset_0_0_0_1px_var(--danger)]">
          {error}
        </p>
      )}

      {result && (
        <div className="mt-7 flex flex-col gap-6" key={nonce}>
          <p className="tnum text-[12px] tracking-[0.06em] text-ink-faint">起算 · {result.lunarText}</p>

          {/* 掌诀六宫环:金线自大安起沿环游走,月/日/时三落宫依次点亮,终宫金芒 */}
          <PalmRing path={result.steps} />

          {/* 掐指三步:依次点亮 */}
          <div className="flex items-stretch gap-2">
            {result.path.map((pos, i) => {
              const tone = luckTone(pos.luck);
              const last = i === result.path.length - 1;
              return (
                <div key={i} className="flex flex-1 items-stretch gap-2">
                  <div
                    className={`dvn-step flex flex-1 flex-col items-center justify-center gap-1.5 rounded-[6px] bg-bg px-2 py-4 text-center shadow-[inset_0_0_0_1px_var(--line)] ${
                      last ? "shadow-[inset_0_0_0_1px_var(--gold-dim)]" : ""
                    }`}
                    style={{ animationDelay: `${i * 0.28}s` }}
                  >
                    <span className={`font-display text-[18px] font-semibold ${last ? "text-gold" : "text-ink"}`}>
                      {pos.name}
                    </span>
                    <span className={`text-[11px] ${toneTextClass(tone)}`}>{pos.luck}</span>
                  </div>
                  {!last && (
                    <span className="flex items-center text-ink-faint" aria-hidden>
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
                        <path d="M9 6l6 6-6 6" strokeLinecap="round" strokeLinejoin="round" />
                      </svg>
                    </span>
                  )}
                </div>
              );
            })}
          </div>

          {/* 结果断语 */}
          <div className="flex flex-col gap-3 rounded-[6px] bg-bg px-4 py-4 shadow-[inset_0_0_0_1px_var(--line)]">
            <div className="flex items-center gap-3">
              <span className="font-display text-[22px] font-semibold text-ink">{result.result.name}</span>
              <span
                className={`inline-flex items-center rounded-[2px] px-2 py-0.5 text-[12px] tracking-[0.08em] ${toneBadgeClass(
                  luckTone(result.result.luck),
                )}`}
              >
                {result.result.luck}
              </span>
            </div>
            <p className="text-[14px] leading-relaxed text-ink-secondary">{result.result.meaning}</p>
            {/* 分节深断:掐指路径/落宫详断/途中之象(免费确定性层) */}
            <JudgeSections sections={result.sections} />
          </div>
        </div>
      )}
    </section>
  );
}

const RING_ORDER = ["大安", "留连", "速喜", "赤口", "小吉", "空亡"];

/**
 * 掌诀六宫环(掐指仪式可视化):
 * 金线自「大安」起,按掐指真实跳序沿环游走(月→日→时逐宫顺数),
 * 三处落宫依次点亮,终宫金芒定格。reduced-motion 下静态呈现。
 */
function PalmRing({ path }: { path: [string, string, string] | string[] }) {
  const C = 110; // 画布半宽
  const R = 78; // 环半径
  const idxOf = (name: string) => Math.max(0, RING_ORDER.indexOf(name));
  const nodeXY = (i: number) => {
    const a = ((i * 60 - 90) * Math.PI) / 180;
    return [C + R * Math.cos(a), C + R * Math.sin(a)] as const;
  };
  // 掐指跳序:自大安(0)顺行至月落宫,再至日落宫、时落宫(逐宫 60° 小弧)
  const hops: number[] = [0];
  let cur = 0;
  for (const name of path) {
    const target = idxOf(name);
    while (cur !== target) {
      cur = (cur + 1) % 6;
      hops.push(cur);
    }
  }
  const landing = new Map<number, number>(); // 节点 → 到达时刻(hop 序)
  {
    let c2 = 0;
    let step = 0;
    landing.set(0, 0);
    for (const name of path) {
      const target = idxOf(name);
      while (c2 !== target) {
        c2 = (c2 + 1) % 6;
        step++;
      }
      landing.set(c2, step);
    }
  }
  const finalIdx = idxOf(path[path.length - 1]);
  // 游走弧线:逐段 60° 圆弧(顺时针 sweep=1)
  let d = "";
  for (let i = 0; i < hops.length; i++) {
    const [x, y] = nodeXY(hops[i]);
    d += i === 0 ? `M ${x.toFixed(1)} ${y.toFixed(1)}` : ` A ${R} ${R} 0 0 1 ${x.toFixed(1)} ${y.toFixed(1)}`;
  }
  const hopMs = 170;
  const total = (hops.length - 1) * hopMs;

  return (
    <div className="flex justify-center">
      <svg viewBox={`0 0 ${C * 2} ${C * 2}`} className="w-[min(64vw,260px)]" aria-hidden>
        <style>{`
          @keyframes xlr-trail { from { stroke-dashoffset: 1; } to { stroke-dashoffset: 0; } }
          @keyframes xlr-node-in { from { opacity: 0.25; } to { opacity: 1; } }
          @keyframes xlr-flare {
            0% { r: 12; opacity: 0.7; }
            100% { r: 30; opacity: 0; }
          }
          .xlr-trail { stroke-dasharray: 1; stroke-dashoffset: 1; animation: xlr-trail ${total}ms linear forwards; }
          @media (prefers-reduced-motion: reduce) {
            .xlr-trail { animation: none; stroke-dashoffset: 0; }
            .xlr-node { animation: none !important; opacity: 1 !important; }
            .xlr-flare { display: none; }
          }
        `}</style>
        <circle cx={C} cy={C} r={R} fill="none" stroke="var(--line)" strokeWidth="1" />
        {/* 游走金线(按真实跳数描迹) */}
        {hops.length > 1 && (
          <path
            d={d}
            pathLength={1}
            className="xlr-trail"
            fill="none"
            stroke="var(--gold)"
            strokeWidth="2.5"
            strokeLinecap="round"
            opacity="0.75"
          />
        )}
        {/* 终宫金芒 */}
        <circle
          cx={nodeXY(finalIdx)[0]}
          cy={nodeXY(finalIdx)[1]}
          r="12"
          fill="none"
          stroke="var(--gold)"
          strokeWidth="1.5"
          className="xlr-flare"
          style={{ animation: `xlr-flare 900ms ${total + 100}ms var(--ease-out) both` }}
        />
        {/* 六宫节点 */}
        {RING_ORDER.map((name, i) => {
          const [x, y] = nodeXY(i);
          const visited = landing.has(i);
          const isFinal = i === finalIdx;
          const delay = visited ? (landing.get(i) ?? 0) * hopMs : 0;
          return (
            <g
              key={name}
              className="xlr-node"
              style={visited ? { animation: `xlr-node-in 300ms ${delay}ms var(--ease-out) both` } : { opacity: 0.4 }}
            >
              <circle
                cx={x}
                cy={y}
                r={isFinal ? 8 : 5.5}
                fill={visited ? "var(--gold)" : "var(--line-strong)"}
                opacity={isFinal ? 1 : 0.8}
              />
              <text
                x={x}
                y={y + (y > C ? 24 : -16)}
                textAnchor="middle"
                fontSize="13"
                fill={isFinal ? "var(--gold)" : visited ? "var(--ink)" : "var(--ink-faint)"}
                style={{ fontFamily: "var(--font-display)", fontWeight: isFinal ? 600 : 400 }}
              >
                {name}
              </text>
            </g>
          );
        })}
      </svg>
    </div>
  );
}
