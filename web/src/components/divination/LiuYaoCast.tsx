"use client";

import { createPortal } from "react-dom";

/**
 * 六爻一键摇卦全屏仪式(对标紫微/梅花盛典),约 2.8s:
 *   ①墨幕落下,三枚铜钱悬于盘心翻转 → ②六轮掷落:每轮钱影一闪、
 *   一道爻画自下而上凝定(阴阳未定闪烁而后定形)→ ③「卦成」金芒收束。
 * 字幕:诚心默祷 → 铜钱六掷 → 卦成。装饰爻序为固定纹样,真实卦象由
 * 服务端摇卦后揭示。createPortal 挂 body;reduced-motion 全停。
 */

const DECOR = [true, false, false, true, true, false]; // 装饰爻序(自下而上)

export function LiuYaoCast() {
  return createPortal(
    <div
      className="ly-stage fixed inset-0 z-[80] flex flex-col items-center justify-center overflow-hidden"
      aria-hidden
      style={{
        background:
          "radial-gradient(66% 66% at 50% 44%, rgba(109,95,163,0.12), transparent 70%), rgba(4,6,12,0.96)",
      }}
    >
      <style>{`
        @keyframes ly-in { from { opacity: 0; } to { opacity: 1; } }
        .ly-stage { animation: ly-in 300ms var(--ease-out) both; }
        @keyframes ly-coin {
          0%, 100% { transform: rotateY(0deg) translateY(0); }
          25% { transform: rotateY(540deg) translateY(-14px); }
          50% { transform: rotateY(1080deg) translateY(0); }
          75% { transform: rotateY(1620deg) translateY(-8px); }
        }
        @keyframes ly-coin-in { from { opacity: 0; transform: translateY(-24px) scale(0.7); } to { opacity: 1; transform: none; } }
        @keyframes ly-yao {
          0% { opacity: 0; transform: translateY(12px) scaleX(0.55); }
          35% { opacity: 0.5; }
          50% { opacity: 0.25; }
          68% { opacity: 0.85; }
          100% { opacity: 1; transform: none; }
        }
        @keyframes ly-yao-glint {
          0%, 80% { box-shadow: none; }
          90% { box-shadow: 0 0 14px var(--gold-glow), 0 0 4px var(--gold-dim); }
          100% { box-shadow: none; }
        }
        @keyframes ly-cap {
          0% { opacity: 0; transform: translateY(8px); }
          16%, 74% { opacity: 1; transform: none; }
          100% { opacity: 0; transform: translateY(-6px); }
        }
        @keyframes ly-cap-final {
          from { opacity: 0; transform: scale(0.85); letter-spacing: 0.12em; }
          to { opacity: 1; transform: scale(1); letter-spacing: 0.46em; }
        }
        @keyframes ly-flash { 0% { opacity: 0; } 45% { opacity: 0.24; } 100% { opacity: 0; } }
        @media (prefers-reduced-motion: reduce) {
          .ly-stage, .ly-stage * { animation: none !important; }
        }
      `}</style>

      {/* 卦成金芒 */}
      <span
        className="absolute inset-0"
        style={{
          background: "radial-gradient(46% 46% at 50% 46%, rgba(240,214,160,0.8), transparent 66%)",
          animation: "ly-flash 0.5s 2.35s var(--ease-out) both",
        }}
      />

      {/* 三枚铜钱:悬转不歇,掷势所出 */}
      <div className="flex gap-[3.2vmin]" style={{ animation: "ly-coin-in 0.5s 0.15s var(--ease-out) both" }}>
        {[0, 1, 2].map((i) => (
          <span
            key={i}
            className="relative flex items-center justify-center rounded-full"
            style={{
              width: "clamp(34px, 7vmin, 52px)",
              height: "clamp(34px, 7vmin, 52px)",
              background: "radial-gradient(circle at 34% 30%, #e8cf9a, #b3924f 58%, #8a6c33)",
              boxShadow: "0 0 18px rgba(191,155,73,0.35), inset 0 0 0 2px rgba(90,68,26,0.55)",
              animation: `ly-coin 1.1s ${0.35 + i * 0.12}s var(--ease-inout) infinite`,
            }}
          >
            <span
              className="block"
              style={{
                width: "26%",
                height: "26%",
                background: "rgba(6,8,15,0.92)",
                boxShadow: "inset 0 0 0 1px rgba(90,68,26,0.7)",
              }}
            />
          </span>
        ))}
      </div>

      {/* 六爻自下而上逐轮凝定 */}
      <div className="mt-[4.5vmin] flex w-[min(46vmin,300px)] flex-col-reverse gap-[1.6vmin]">
        {DECOR.map((yang, i) => (
          <div
            key={i}
            className="flex h-[clamp(9px,1.9vmin,13px)] items-stretch justify-center gap-[12%]"
            style={{
              animation: `ly-yao 0.5s ${0.55 + i * 0.3}s var(--ease-out) both, ly-yao-glint 0.5s ${0.55 + i * 0.3}s linear both`,
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

      {/* 字幕 */}
      <div className="relative mt-[3.5vmin] h-9 w-full">
        {[
          { t: "诚心默祷", d: 0.2 },
          { t: "铜钱六掷", d: 1.15 },
        ].map((s) => (
          <p
            key={s.t}
            className="absolute inset-x-0 text-center text-[clamp(13px,2.5vmin,16px)] tracking-[0.34em] text-ink-secondary"
            style={{ fontFamily: "var(--font-display)", animation: `ly-cap 0.9s ${s.d}s var(--ease-inout) both` }}
          >
            {s.t}
          </p>
        ))}
        <p
          className="absolute inset-x-0 text-center text-[clamp(17px,3.4vmin,21px)] font-semibold text-gold"
          style={{
            fontFamily: "var(--font-display)",
            textShadow: "0 0 20px rgba(191,155,73,0.6)",
            animation: "ly-cap-final 0.45s 2.4s var(--ease-out) both",
          }}
        >
          卦成
        </p>
      </div>
    </div>,
    document.body
  );
}
