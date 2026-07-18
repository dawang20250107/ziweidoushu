"use client";

/**
 * 六爻摇卦动效:三枚古钱翻掷(rotateY 连翻 + 起落),下方六爻位逐一落定点亮。
 * 纯 CSS(内联 keyframes);页面对总时长封顶 ~1.6s,reduced-motion 下页面不挂载本组件。
 */

const COINS = [0, 1, 2];
const SLOTS = ["初", "二", "三", "四", "五", "上"];

export function ShakeRitual() {
  return (
    <section
      className="page-enter mt-12 flex flex-col items-center gap-7 rounded-[10px] bg-bg-raised px-6 py-16 shadow-[0_0_0_1px_var(--line)]"
      role="status"
      aria-label="摇卦中"
    >
      <style>{`
        @keyframes ly-coin-flip {
          0%   { transform: translateY(0) rotateY(0deg); }
          35%  { transform: translateY(-22px) rotateY(540deg); }
          70%  { transform: translateY(0) rotateY(900deg); }
          100% { transform: translateY(0) rotateY(1080deg); }
        }
        @keyframes ly-slot-on {
          0%   { opacity: 0.25; box-shadow: inset 0 0 0 1px var(--line); }
          100% { opacity: 1; box-shadow: inset 0 0 0 1px var(--gold-dim); }
        }
        @keyframes ly-halo {
          0%, 100% { opacity: 0.3; transform: scale(0.85); }
          50%      { opacity: 0.85; transform: scale(1.1); }
        }
        .ly-coin { animation: ly-coin-flip 1.1s var(--ease-inout) infinite; transform-style: preserve-3d; }
        .ly-slot { animation: ly-slot-on 0.3s var(--ease-out) both; }
        .ly-halo { animation: ly-halo 1.5s var(--ease-inout) infinite; }
        @media (prefers-reduced-motion: reduce) {
          .ly-coin, .ly-slot, .ly-halo { animation: none; opacity: 1; transform: none; }
        }
      `}</style>

      {/* 三枚古钱 + 金晕 */}
      <div className="relative flex h-[120px] items-center justify-center" style={{ perspective: 480 }}>
        <span
          aria-hidden
          className="ly-halo pointer-events-none absolute -inset-6 rounded-full"
          style={{ background: "radial-gradient(circle, var(--gold-glow) 0%, transparent 68%)" }}
        />
        <div className="relative flex items-center gap-5">
          {COINS.map((i) => (
            <span
              key={i}
              aria-hidden
              className="ly-coin flex h-12 w-12 items-center justify-center rounded-full"
              style={{
                animationDelay: `${i * 0.14}s`,
                background: "radial-gradient(circle at 35% 30%, var(--gold-glow), transparent 70%)",
                boxShadow: "inset 0 0 0 1.5px var(--gold-dim)",
              }}
            >
              {/* 方孔 */}
              <span className="h-3.5 w-3.5 rounded-[1px]" style={{ boxShadow: "inset 0 0 0 1.5px var(--gold-dim)" }} />
            </span>
          ))}
        </div>
      </div>

      {/* 六爻位逐一落定(自下而上依次点亮) */}
      <div className="flex items-center gap-2.5" aria-hidden>
        {SLOTS.map((name, i) => (
          <span
            key={name}
            className="ly-slot flex h-8 w-8 items-center justify-center rounded-[4px] bg-bg text-[11px] text-ink-secondary"
            style={{ animationDelay: `${0.15 + i * 0.22}s` }}
          >
            {name}
          </span>
        ))}
      </div>

      <p className="text-[14px] tracking-[0.24em] text-gold">摇卦中 · 六爻齐落</p>
    </section>
  );
}
