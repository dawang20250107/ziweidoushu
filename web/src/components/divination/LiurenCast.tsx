"use client";

import { createPortal } from "react-dom";

/**
 * 大六壬起课全屏仪式(盛典版),约 3.6s:
 *   ①墨幕落下,十二辰之气(星粒)自四野汇入盘心 → ②地盘十二支环隐现,
 *   天盘环携月将旋转就位 → ③骤停一瞬:金弧沿地支环扫过一周、舞台震颤,
 *   贵人金点落位一闪 → ④四课立柱于盘下逐一升起 → ⑤初中末三传自上凝定 →
 *   ⑥远环震荡、金芒一闪、「课成」收束。
 * 字幕:月将加时 → 天将布位 → 四课三传 → 课成。环上支序为装饰定式,
 * 真实课象由服务端起课后揭示。createPortal 挂 body;reduced-motion 全停。
 */

const BRANCHES = ["子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"];

// 环上定位:午居上顺时针(南上式盘),半径以百分比计
function ringPos(i: number, radius: number) {
  const deg = ((i - 6 + 12) % 12) * 30; // 午(6)在正上
  const rad = (deg * Math.PI) / 180;
  return {
    left: `${50 + radius * Math.sin(rad)}%`,
    top: `${50 - radius * Math.cos(rad)}%`,
  };
}

/** 十二辰之气:黄金角散布全屏,自四野汇入盘心(确定性方位/延迟)。 */
const STARS = Array.from({ length: 12 }, (_, i) => {
  const a = (i * 137.5 * Math.PI) / 180;
  const d = 34 + (i % 4) * 9;
  return {
    dx: (Math.cos(a) * d).toFixed(1),
    dy: (Math.sin(a) * d).toFixed(1),
    delay: ((i % 6) * 0.08).toFixed(2),
    big: i % 4 === 0,
  };
});

export function LiurenCast() {
  return createPortal(
    <div
      className="lr-stage fixed inset-0 z-[80] flex flex-col items-center justify-center overflow-hidden"
      aria-hidden
      style={{
        background:
          "radial-gradient(66% 66% at 50% 44%, rgba(109,95,163,0.12), transparent 70%), rgba(4,6,12,0.96)",
      }}
    >
      <style>{`
        @keyframes lr-in { from { opacity: 0; } to { opacity: 1; } }
        .lr-stage { animation: lr-in 300ms var(--ease-out) both, lr-quake 0.4s 1.8s var(--ease-out) both; }
        @keyframes lr-quake {
          0%, 100% { translate: 0 0; }
          24% { translate: 5px -4px; }
          48% { translate: -4px 3px; }
          72% { translate: 3px 2px; }
        }
        @keyframes lr-star {
          0% { transform: translate(calc(var(--sx) * 1vw), calc(var(--sy) * 1vh)) scale(0.4); opacity: 0; }
          12% { opacity: 0.9; }
          70% { transform: translate(0, 0) scale(1); opacity: 0.8; }
          100% { transform: translate(0, 0) scale(0.2); opacity: 0; }
        }
        @keyframes lr-ring-in { from { opacity: 0; transform: scale(0.86); } to { opacity: 1; transform: none; } }
        @keyframes lr-spin {
          0% { transform: rotate(300deg); opacity: 0; }
          18% { opacity: 1; }
          100% { transform: rotate(0deg); opacity: 1; }
        }
        @keyframes lr-sweep {
          0% { transform: rotate(0deg); opacity: 0; }
          14% { opacity: 0.9; }
          86% { opacity: 0.9; }
          100% { transform: rotate(360deg); opacity: 0; }
        }
        @keyframes lr-gui {
          0% { opacity: 0; transform: scale(0.4); }
          40% { opacity: 1; transform: scale(1.5); box-shadow: 0 0 22px var(--gold-glow), 0 0 6px var(--gold); }
          100% { opacity: 0.9; transform: scale(1); box-shadow: 0 0 10px var(--gold-glow); }
        }
        @keyframes lr-ke {
          0% { opacity: 0; transform: translateY(12px) scaleY(0.4); }
          60% { opacity: 1; transform: translateY(0) scaleY(1.08); }
          100% { opacity: 1; transform: none; }
        }
        @keyframes lr-chuan {
          0% { opacity: 0; transform: translateY(10px) scale(0.7); }
          60% { opacity: 1; transform: translateY(0) scale(1.12); }
          100% { opacity: 1; transform: none; }
        }
        @keyframes lr-cap {
          0% { opacity: 0; transform: translateY(8px); }
          16%, 74% { opacity: 1; transform: none; }
          100% { opacity: 0; transform: translateY(-6px); }
        }
        @keyframes lr-cap-final {
          from { opacity: 0; transform: scale(0.85); letter-spacing: 0.12em; }
          to { opacity: 1; transform: scale(1); letter-spacing: 0.46em; }
        }
        @keyframes lr-pulse-far { from { transform: scale(0.12); opacity: 0.5; } to { transform: scale(1); opacity: 0; } }
        @keyframes lr-flash { 0% { opacity: 0; } 45% { opacity: 0.26; } 100% { opacity: 0; } }
        @media (prefers-reduced-motion: reduce) {
          .lr-stage, .lr-stage * { animation: none !important; }
        }
      `}</style>

      {/* 十二辰之气自四野汇入盘心 */}
      {STARS.map((s, i) => (
        <span
          key={i}
          className="absolute left-1/2 top-1/2 rounded-full"
          style={{
            // @ts-expect-error 自定义变量
            "--sx": s.dx,
            "--sy": s.dy,
            width: s.big ? 6 : 4,
            height: s.big ? 6 : 4,
            marginLeft: s.big ? -3 : -2,
            marginTop: s.big ? -3 : -2,
            background: s.big ? "var(--gold)" : "rgba(214,196,240,0.9)",
            boxShadow: s.big ? "0 0 10px var(--gold-glow)" : "0 0 8px rgba(163,141,219,0.55)",
            animation: `lr-star 1.3s ${s.delay}s var(--ease-inout) both`,
          }}
        />
      ))}

      {/* 课成后的全屏远环震荡 */}
      <span
        className="absolute left-1/2 top-1/2 h-[120vmax] w-[120vmax] -translate-x-1/2 -translate-y-1/2 rounded-full"
        style={{
          boxShadow: "inset 0 0 0 1.5px var(--gold-dim), inset 0 0 56px rgba(191,155,73,0.22)",
          animation: "lr-pulse-far 1s 2.95s var(--ease-out) both",
        }}
      />

      {/* 课成金芒 */}
      <span
        className="absolute inset-0"
        style={{
          background: "radial-gradient(46% 46% at 50% 46%, rgba(240,214,160,0.8), transparent 66%)",
          animation: "lr-flash 0.5s 3.1s var(--ease-out) both",
        }}
      />

      {/* 式盘双环 */}
      <div className="relative aspect-square w-[min(64vmin,380px)]">
        {/* 地盘环(静) */}
        <div className="absolute inset-0" style={{ animation: "lr-ring-in 0.6s 0.1s var(--ease-out) both" }}>
          <span
            className="absolute inset-0 rounded-full"
            style={{ boxShadow: "inset 0 0 0 1px var(--line-strong), inset 0 0 40px rgba(109,95,163,0.10)" }}
          />
          {BRANCHES.map((b, i) => (
            <span
              key={b}
              className="absolute -translate-x-1/2 -translate-y-1/2 text-[clamp(12px,2.3vmin,15px)] text-ink-faint"
              style={{ ...ringPos(i, 46), fontFamily: "var(--font-display)" }}
            >
              {b}
            </span>
          ))}
        </div>
        {/* 天盘环(旋转就位,骤停收束) */}
        <div className="absolute inset-0" style={{ animation: "lr-spin 1.45s 0.35s var(--ease-out) both" }}>
          {BRANCHES.map((b, i) => (
            <span
              key={b}
              className="absolute -translate-x-1/2 -translate-y-1/2 text-[clamp(15px,3vmin,20px)] font-medium text-gold"
              style={{ ...ringPos(i, 32), fontFamily: "var(--font-display)", opacity: 0.92 }}
            >
              {b}
            </span>
          ))}
        </div>
        {/* 骤停一瞬:金弧沿地支环扫过一周 */}
        <svg viewBox="-100 -100 200 200" className="absolute inset-0 h-full w-full">
          <circle
            r="78"
            fill="none"
            stroke="var(--gold)"
            strokeWidth="2.5"
            strokeLinecap="round"
            pathLength={1}
            strokeDasharray="0.22 0.78"
            style={{
              transformOrigin: "0 0",
              filter: "drop-shadow(0 0 6px rgba(240,214,160,0.7))",
              animation: "lr-sweep 0.7s 1.78s var(--ease-inout) both",
            }}
          />
        </svg>
        {/* 贵人金点(骤停后落位一闪,天门之侧装饰位) */}
        <span
          className="absolute h-[7px] w-[7px] -translate-x-1/2 -translate-y-1/2 rounded-full bg-gold"
          style={{ ...ringPos(11, 32), animation: "lr-gui 0.8s 1.95s var(--ease-inout) both" }}
        />
        {/* 四课立柱:上下神两位一柱,盘下逐一升起 */}
        <div className="absolute inset-x-0 top-[60%] flex items-start justify-center gap-[3vmin]">
          {[0, 1, 2, 3].map((k) => (
            <div
              key={k}
              className="flex origin-bottom flex-col gap-[0.8vmin]"
              style={{ animation: `lr-ke 0.45s ${2.1 + k * 0.13}s var(--ease-out) both` }}
            >
              <span
                className="h-[2.6vmin] max-h-[15px] w-[2.6vmin] max-w-[15px] rounded-[2px] bg-gold"
                style={{ opacity: 0.88, boxShadow: "0 0 10px rgba(191,155,73,0.4)" }}
              />
              <span
                className="h-[2.6vmin] max-h-[15px] w-[2.6vmin] max-w-[15px] rounded-[2px]"
                style={{ boxShadow: "inset 0 0 0 1px var(--line-strong)" }}
              />
            </div>
          ))}
        </div>
        {/* 三传凝定(居上,与四课分列) */}
        <div className="absolute inset-x-0 top-[42%] flex -translate-y-1/2 items-center justify-center gap-[4vmin]">
          {["初", "中", "末"].map((t, i) => (
            <span
              key={t}
              className="text-[clamp(17px,3.6vmin,23px)] font-semibold text-gold"
              style={{
                fontFamily: "var(--font-display)",
                textShadow: "0 0 16px rgba(191,155,73,0.55)",
                animation: `lr-chuan 0.4s ${2.55 + i * 0.24}s var(--ease-out) both`,
              }}
            >
              {t}
            </span>
          ))}
        </div>
      </div>

      {/* 步进字幕 */}
      <div className="relative mt-[3.5vmin] h-9 w-full">
        {[
          { t: "月将加时", d: 0.3 },
          { t: "天将布位", d: 1.5 },
          { t: "四课三传", d: 2.35 },
        ].map((s) => (
          <p
            key={s.t}
            className="absolute inset-x-0 text-center text-[clamp(13px,2.5vmin,16px)] tracking-[0.34em] text-ink-secondary"
            style={{ fontFamily: "var(--font-display)", animation: `lr-cap 0.85s ${s.d}s var(--ease-inout) both` }}
          >
            {s.t}
          </p>
        ))}
        <p
          className="absolute inset-x-0 text-center text-[clamp(17px,3.4vmin,21px)] font-semibold text-gold"
          style={{
            fontFamily: "var(--font-display)",
            textShadow: "0 0 20px rgba(191,155,73,0.6)",
            animation: "lr-cap-final 0.45s 3.15s var(--ease-out) both",
          }}
        >
          课成
        </p>
      </div>
    </div>,
    document.body
  );
}
