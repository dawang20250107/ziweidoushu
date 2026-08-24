"use client";

import { createPortal } from "react-dom";

/**
 * 梅花起卦全屏仪式(盛典版),约 3.4s:
 *   ①墨幕落下,梅瓣自四野飘旋汇入盘心(观梅之意)→ ②太极旋入、
 *   先天八卦环显形环转 → ③击盘脉冲,六道爻画自下而上逐一凝定 →
 *   ④动爻金脉冲点睛 → ⑤远环震荡、金芒一闪、「卦成」收束。
 * 字幕:心动而占 → 观梅取数 → 定动爻 → 卦成。
 * 经 createPortal 挂 body,任何祖先 transform 无法劫持 fixed 定位;
 * prefers-reduced-motion 下动画全停(内容静置,请求本就快)。
 * 仅为仪式装饰:爻序为固定纹样,真实卦象由服务端起卦后揭示。
 */

const BAGUA = ["☰", "☴", "☵", "☶", "☷", "☳", "☲", "☱"]; // 先天卦序环布
const DECOR = [true, false, true, true, false, true]; // 装饰爻序(自下而上)
const MOVING_YAO = 2; // 装饰动爻位(第三爻,金脉冲点睛)

/** 梅瓣:黄金角散布全屏,自四野旋入盘心(确定性方位/延迟)。 */
const PETALS = Array.from({ length: 16 }, (_, i) => {
  const a = (i * 137.5 * Math.PI) / 180;
  const d = 36 + (i % 5) * 8;
  return {
    dx: (Math.cos(a) * d).toFixed(1),
    dy: (Math.sin(a) * d).toFixed(1),
    delay: ((i % 8) * 0.09).toFixed(2),
    rot: (i * 137.5) % 360 | 0,
    big: i % 3 === 0,
  };
});

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
        .mh-stage { animation: mh-in 300ms var(--ease-out) both, mh-quake 0.45s 1.15s var(--ease-out) both; }
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
        @keyframes mh-flash { 0% { opacity: 0; } 45% { opacity: 0.3; } 100% { opacity: 0; } }
        @keyframes mh-petal {
          0% { transform: translate(calc(var(--px) * 1vw), calc(var(--py) * 1vh)) rotate(var(--pr)) scale(0.5); opacity: 0; }
          14% { opacity: 0.95; }
          66% { transform: translate(0, 0) rotate(calc(var(--pr) + 200deg)) scale(1); opacity: 0.85; }
          78%, 100% { transform: translate(0, 0) rotate(calc(var(--pr) + 230deg)) scale(0.3); opacity: 0; }
        }
        @keyframes mh-quake {
          0%, 100% { translate: 0 0; }
          22% { translate: 6px -4px; }
          46% { translate: -5px 3px; }
          70% { translate: 3px 2px; }
        }
        @keyframes mh-pulse { from { transform: scale(0.4); opacity: 0.85; } to { transform: scale(2); opacity: 0; } }
        @keyframes mh-pulse-far { from { transform: scale(0.12); opacity: 0.5; } to { transform: scale(1); opacity: 0; } }
        @keyframes mh-glow { from { opacity: 0; } to { opacity: 0.5; } }
        @keyframes mh-moving-ring {
          0% { transform: scale(0.7); opacity: 0; }
          30% { opacity: 0.9; }
          100% { transform: scale(1.5); opacity: 0; }
        }
        @keyframes mh-taiji-dim { from { opacity: 1; } to { opacity: 0.22; } }
        @media (prefers-reduced-motion: reduce) {
          .mh-stage, .mh-stage * { animation: none !important; }
        }
      `}</style>

      {/* 梅瓣自四野飘旋汇入(观梅之意) */}
      {PETALS.map((p, i) => (
        <span
          key={i}
          className="absolute left-1/2 top-1/2"
          style={{
            // @ts-expect-error 自定义变量
            "--px": p.dx,
            "--py": p.dy,
            "--pr": `${p.rot}deg`,
            width: p.big ? 13 : 9,
            height: p.big ? 11 : 8,
            marginLeft: p.big ? -6 : -4,
            marginTop: p.big ? -5 : -4,
            borderRadius: "62% 4% 62% 62%",
            background: "linear-gradient(135deg, rgba(240,196,186,0.92), rgba(217,179,108,0.7))",
            boxShadow: "0 0 8px rgba(240,196,186,0.45)",
            animation: `mh-petal 1.45s ${p.delay}s var(--ease-inout) both`,
          }}
        />
      ))}

      {/* 击盘后的全屏远环震荡 */}
      <span
        className="absolute left-1/2 top-1/2 h-[120vmax] w-[120vmax] -translate-x-1/2 -translate-y-1/2 rounded-full"
        style={{
          boxShadow: "inset 0 0 0 1.5px var(--gold-dim), inset 0 0 56px rgba(191,155,73,0.22)",
          animation: "mh-pulse-far 1s 2.55s var(--ease-out) both",
        }}
      />

      {/* 卦成金芒 */}
      <span
        className="absolute inset-0"
        style={{
          background: "radial-gradient(46% 46% at 50% 46%, rgba(240,214,160,0.8), transparent 66%)",
          animation: "mh-flash 0.5s 2.9s var(--ease-out) both",
        }}
      />

      <div className="relative aspect-square w-[min(72vmin,480px)]">
        {/* 击盘:盘心辉光洇开 + 双重脉冲扩散 */}
        <span
          className="absolute inset-0 rounded-full"
          style={{
            background: "radial-gradient(50% 50% at 50% 50%, rgba(191,155,73,0.2), transparent 70%)",
            animation: "mh-glow 0.8s 1.2s var(--ease-out) both",
          }}
        />
        <span
          className="absolute inset-[8%] rounded-full"
          style={{
            boxShadow: "0 0 0 1.5px var(--gold-dim), var(--glow-gold-strong)",
            animation: "mh-pulse 0.9s 1.15s var(--ease-out) both",
          }}
        />
        <span
          className="absolute inset-[8%] rounded-full"
          style={{
            boxShadow: "0 0 0 1px var(--gold-dim)",
            animation: "mh-pulse 0.8s 1.38s var(--ease-out) both",
          }}
        />

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

        {/* 六爻自下而上凝定;动爻位另加金脉冲点睛 */}
        <div className="absolute inset-x-[34%] inset-y-[31%] flex flex-col-reverse justify-between">
          {DECOR.map((yang, i) => (
            <div
              key={i}
              className="relative flex h-[9%] items-stretch justify-center gap-[12%]"
              style={{
                animation: `mh-yao-flicker 0.62s ${0.95 + i * 0.22}s var(--ease-out) both, mh-yao-glint 0.62s ${0.95 + i * 0.22}s linear both`,
              }}
            >
              {i === MOVING_YAO && (
                <span
                  className="pointer-events-none absolute -inset-x-[10%] -inset-y-[55%] rounded-full"
                  style={{
                    boxShadow: "0 0 0 1.5px var(--gold), 0 0 22px rgba(240,214,160,0.5)",
                    animation: "mh-moving-ring 0.85s 2.35s var(--ease-out) both",
                  }}
                />
              )}
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
          { t: "心动而占", d: 0.3 },
          { t: "观梅取数", d: 1.15 },
          { t: "定动爻", d: 2.05 },
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
            animation: "mh-cap-final 0.45s 2.95s var(--ease-out) both",
          }}
        >
          卦成
        </p>
      </div>
    </div>,
    document.body
  );
}
