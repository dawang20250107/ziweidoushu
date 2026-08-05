"use client";

import { createPortal } from "react-dom";

/**
 * 大六壬起课全屏仪式(对标紫微罗盘/梅花/六爻盛典),约 3.2s:
 *   ①墨幕落下,地盘十二支环隐现 → ②天盘环携月将旋转就位(减速收束)→
 *   ③贵人金点落位一闪 → ④初中末三传自中宫逐一凝定 → ⑤「课成」金芒。
 * 字幕:月将加时 → 天将布位 → 课成。环上支序为装饰定式,真实课象由
 * 服务端起课后揭示。createPortal 挂 body;reduced-motion 全停。
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
        .lr-stage { animation: lr-in 300ms var(--ease-out) both; }
        @keyframes lr-ring-in { from { opacity: 0; transform: scale(0.86); } to { opacity: 1; transform: none; } }
        @keyframes lr-spin {
          0% { transform: rotate(300deg); opacity: 0; }
          18% { opacity: 1; }
          100% { transform: rotate(0deg); opacity: 1; }
        }
        @keyframes lr-gui {
          0%, 55% { opacity: 0; transform: scale(0.4); }
          70% { opacity: 1; transform: scale(1.5); box-shadow: 0 0 22px var(--gold-glow), 0 0 6px var(--gold); }
          100% { opacity: 0.9; transform: scale(1); box-shadow: 0 0 10px var(--gold-glow); }
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
        @keyframes lr-flash { 0% { opacity: 0; } 45% { opacity: 0.24; } 100% { opacity: 0; } }
        @media (prefers-reduced-motion: reduce) {
          .lr-stage, .lr-stage * { animation: none !important; }
        }
      `}</style>

      {/* 课成金芒 */}
      <span
        className="absolute inset-0"
        style={{
          background: "radial-gradient(46% 46% at 50% 46%, rgba(240,214,160,0.8), transparent 66%)",
          animation: "lr-flash 0.5s 2.75s var(--ease-out) both",
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
        {/* 天盘环(旋转就位) */}
        <div className="absolute inset-0" style={{ animation: "lr-spin 1.7s 0.35s var(--ease-out) both" }}>
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
        {/* 贵人金点(落于天门之侧,装饰位) */}
        <span
          className="absolute h-[7px] w-[7px] -translate-x-1/2 -translate-y-1/2 rounded-full bg-gold"
          style={{ ...ringPos(11, 32), animation: "lr-gui 2.6s 0s var(--ease-inout) both" }}
        />
        {/* 三传凝定 */}
        <div className="absolute inset-0 flex items-center justify-center gap-[4vmin]">
          {["初", "中", "末"].map((t, i) => (
            <span
              key={t}
              className="text-[clamp(17px,3.6vmin,23px)] font-semibold text-gold"
              style={{
                fontFamily: "var(--font-display)",
                textShadow: "0 0 16px rgba(191,155,73,0.55)",
                animation: `lr-chuan 0.4s ${2.0 + i * 0.28}s var(--ease-out) both`,
              }}
            >
              {t}
            </span>
          ))}
        </div>
      </div>

      {/* 字幕 */}
      <div className="relative mt-[3.5vmin] h-9 w-full">
        {[
          { t: "月将加时", d: 0.25 },
          { t: "天将布位", d: 1.45 },
        ].map((s) => (
          <p
            key={s.t}
            className="absolute inset-x-0 text-center text-[clamp(13px,2.5vmin,16px)] tracking-[0.34em] text-ink-secondary"
            style={{ fontFamily: "var(--font-display)", animation: `lr-cap 0.95s ${s.d}s var(--ease-inout) both` }}
          >
            {s.t}
          </p>
        ))}
        <p
          className="absolute inset-x-0 text-center text-[clamp(17px,3.4vmin,21px)] font-semibold text-gold"
          style={{
            fontFamily: "var(--font-display)",
            textShadow: "0 0 20px rgba(191,155,73,0.6)",
            animation: "lr-cap-final 0.45s 2.8s var(--ease-out) both",
          }}
        >
          课成
        </p>
      </div>
    </div>,
    document.body
  );
}
