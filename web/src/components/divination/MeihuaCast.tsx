"use client";

import { createPortal } from "react-dom";

/**
 * 梅花起卦全屏仪式(对标紫微罗盘盛典),约 2.6s:
 *   ①墨幕落下,太极旋入盘心缓转 → ②先天八卦环显形环转 →
 *   ③六道爻画自下而上逐一凝定(阴阳闪烁而后定形,金芒缀之) →
 *   ④「卦成」金字收束。步进字幕:心动而占 → 数起于时 → 卦成。
 * 经 createPortal 挂 body,任何祖先 transform 无法劫持 fixed 定位;
 * prefers-reduced-motion 下动画全停(内容静置,请求本就快)。
 * 仅为仪式装饰:爻序为固定纹样,真实卦象由服务端起卦后揭示。
 */

const BAGUA = ["☰", "☴", "☵", "☶", "☷", "☳", "☲", "☱"]; // 先天卦序环布
const DECOR = [true, false, true, true, false, true]; // 装饰爻序(自下而上)

export function MeihuaCast() {
  const R = 150;
  return createPortal(
    <div
      className="mh-stage fixed inset-0 z-[80] flex flex-col items-center justify-center overflow-hidden"
      aria-hidden
      style={{
        background:
          "radial-gradient(66% 66% at 50% 44%, rgba(109,95,163,0.12), transparent 70%), rgba(4,6,12,0.96)",
      }}
    >
      <style>{`
        @keyframes mh-in { from { opacity: 0; } to { opacity: 1; } }
        .mh-stage { animation: mh-in 300ms var(--ease-out) both; }
        @keyframes mh-spin { to { transform: rotate(360deg); } }
        @keyframes mh-taiji-in {
          from { opacity: 0; transform: scale(0.5) rotate(-90deg); }
          to { opacity: 1; transform: scale(1) rotate(0deg); }
        }
        @keyframes mh-ring-in {
          from { opacity: 0; transform: scale(0.78) rotate(14deg); }
          to { opacity: 1; transform: scale(1) rotate(0deg); }
        }
        @keyframes mh-yao-flicker {
          0% { opacity: 0; transform: translateY(10px) scaleX(0.6); }
          30% { opacity: 0.55; }
          45% { opacity: 0.25; }
          60% { opacity: 0.8; }
          72% { opacity: 0.45; }
          100% { opacity: 1; transform: translateY(0) scaleX(1); }
        }
        @keyframes mh-yao-glint {
          0%, 78% { box-shadow: none; }
          88% { box-shadow: 0 0 14px var(--gold-glow), 0 0 4px var(--gold-dim); }
          100% { box-shadow: none; }
        }
        @keyframes mh-cap {
          0% { opacity: 0; transform: translateY(8px); }
          16%, 74% { opacity: 1; transform: translateY(0); }
          100% { opacity: 0; transform: translateY(-6px); }
        }
        @keyframes mh-cap-final {
          from { opacity: 0; transform: scale(0.85); letter-spacing: 0.12em; }
          to { opacity: 1; transform: scale(1); letter-spacing: 0.46em; }
        }
        @keyframes mh-flash { 0% { opacity: 0; } 45% { opacity: 0.26; } 100% { opacity: 0; } }
        @keyframes mh-taiji-dim { from { opacity: 1; } to { opacity: 0.22; } }
        @media (prefers-reduced-motion: reduce) {
          .mh-stage, .mh-stage * { animation: none !important; }
        }
      `}</style>

      {/* 卦成金芒 */}
      <span
        className="absolute inset-0"
        style={{
          background: "radial-gradient(46% 46% at 50% 46%, rgba(240,214,160,0.8), transparent 66%)",
          animation: "mh-flash 0.5s 2.15s var(--ease-out) both",
        }}
      />

      <div className="relative aspect-square w-[min(72vmin,480px)]">
        {/* 太极:旋入缓转,爻成时退隐为底纹 */}
        <div
          className="absolute inset-[24%]"
          style={{ animation: "mh-taiji-in 0.7s 0.1s var(--ease-out) both, mh-taiji-dim 0.8s 1.15s var(--ease-out) both" }}
        >
          <svg viewBox="-100 -100 200 200" className="h-full w-full" style={{ animation: "mh-spin 22s linear infinite" }}>
            <circle r="96" fill="none" stroke="var(--gold-dim)" strokeWidth="1.5" opacity="0.8" />
            <path
              d="M 0 -88 A 88 88 0 0 1 0 88 A 44 44 0 0 1 0 0 A 44 44 0 0 0 0 -88 Z"
              fill="var(--gold-dim)"
              opacity="0.5"
            />
            <circle cy="-44" r="12" fill="var(--bg, #06080f)" />
            <circle cy="44" r="12" fill="var(--gold-dim)" opacity="0.85" />
          </svg>
        </div>

        {/* 先天八卦环 */}
        <div className="absolute inset-0" style={{ animation: "mh-ring-in 0.65s 0.35s var(--ease-out) both" }}>
          <svg viewBox={`${-R - 30} ${-R - 30} ${R * 2 + 60} ${R * 2 + 60}`} className="h-full w-full">
            <g style={{ animation: "mh-spin 40s linear infinite", transformOrigin: "0 0" }}>
              <circle r={R} fill="none" stroke="var(--line-strong)" strokeWidth="1" />
              <circle r={R - 34} fill="none" stroke="var(--line)" strokeWidth="1" strokeDasharray="3 7" />
              {BAGUA.map((g, i) => {
                const a = (i * 45 - 90) * (Math.PI / 180);
                return (
                  <text
                    key={i}
                    x={(Math.cos(a) * (R - 16)).toFixed(1)}
                    y={(Math.sin(a) * (R - 16)).toFixed(1)}
                    textAnchor="middle"
                    dominantBaseline="central"
                    fontSize="17"
                    fill="var(--ink-faint)"
                  >
                    {g}
                  </text>
                );
              })}
            </g>
          </svg>
        </div>

        {/* 六爻自下而上凝定 */}
        <div className="absolute inset-x-[34%] inset-y-[31%] flex flex-col-reverse justify-between">
          {DECOR.map((yang, i) => (
            <div
              key={i}
              className="flex h-[9%] items-stretch justify-center gap-[12%]"
              style={{
                animation: `mh-yao-flicker 0.62s ${0.75 + i * 0.22}s var(--ease-out) both, mh-yao-glint 0.62s ${0.75 + i * 0.22}s linear both`,
              }}
            >
              {yang ? (
                <span className="w-full rounded-[1.5px] bg-[var(--gold)]" style={{ opacity: 0.92 }} />
              ) : (
                <>
                  <span className="flex-1 rounded-[1.5px] bg-[var(--gold)]" style={{ opacity: 0.92 }} />
                  <span className="flex-1 rounded-[1.5px] bg-[var(--gold)]" style={{ opacity: 0.92 }} />
                </>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* 步进字幕 */}
      <div className="relative mt-[3.5vmin] h-9 w-full">
        {[
          { t: "心动而占", d: 0.25 },
          { t: "数起于时", d: 1.1 },
        ].map((s) => (
          <p
            key={s.t}
            className="absolute inset-x-0 text-center text-[clamp(13px,2.5vmin,16px)] tracking-[0.34em] text-ink-secondary"
            style={{ fontFamily: "var(--font-display)", animation: `mh-cap 0.85s ${s.d}s var(--ease-inout) both` }}
          >
            {s.t}
          </p>
        ))}
        <p
          className="absolute inset-x-0 text-center text-[clamp(17px,3.4vmin,21px)] font-semibold text-gold"
          style={{
            fontFamily: "var(--font-display)",
            textShadow: "0 0 20px rgba(191,155,73,0.6)",
            animation: "mh-cap-final 0.45s 2.2s var(--ease-out) both",
          }}
        >
          卦成
        </p>
      </div>
    </div>,
    document.body
  );
}
