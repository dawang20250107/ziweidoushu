"use client";

/**
 * 起卦动效:六爻自上而下逐条落定,中心星曜金晕汇聚呼吸,伴「起卦中」。
 * 纯 CSS(内联 keyframes),单次落定 ≤1.3s;页面对总时长封顶 ~1.4s。
 * reduced-motion:@media 兜底关掉动画(且页面在该偏好下根本不挂载本组件,直接出结果)。
 */

// 装饰用的成卦轮廓(非真实卦,起卦结果另行揭示):自上而下的阴阳序。
const DECOR = [true, false, true, true, false, true];

export function CastRitual() {
  return (
    <section
      className="page-enter mt-12 flex flex-col items-center gap-6 rounded-[10px] bg-bg-raised px-6 py-16 shadow-[0_0_0_1px_var(--line)]"
      role="status"
      aria-label="起卦中"
    >
      <style>{`
        @keyframes dvn-yao-drop {
          0%   { opacity: 0; transform: translateY(-16px); }
          70%  { opacity: 1; }
          100% { opacity: 1; transform: translateY(0); }
        }
        @keyframes dvn-halo {
          0%, 100% { opacity: 0.35; transform: scale(0.82); }
          50%      { opacity: 0.9;  transform: scale(1.12); }
        }
        @keyframes dvn-dot {
          0%, 100% { opacity: 0.3; }
          50%      { opacity: 1; }
        }
        .dvn-yao { animation: dvn-yao-drop 0.5s var(--ease-out) both; }
        .dvn-halo { animation: dvn-halo 1.6s var(--ease-inout) infinite; }
        .dvn-dot { animation: dvn-dot 1.2s var(--ease-inout) infinite; }
        @media (prefers-reduced-motion: reduce) {
          .dvn-yao, .dvn-halo, .dvn-dot { animation: none; opacity: 1; transform: none; }
        }
      `}</style>

      {/* 成卦轮廓 + 金晕 */}
      <div className="relative flex h-[168px] w-[168px] items-center justify-center">
        <span
          aria-hidden
          className="dvn-halo pointer-events-none absolute inset-0 rounded-full"
          style={{ background: "radial-gradient(circle, var(--gold-glow) 0%, transparent 68%)" }}
        />
        <div className="relative flex w-[132px] flex-col gap-[9px]">
          {DECOR.map((yang, i) => (
            <div
              key={i}
              className="dvn-yao flex items-center justify-center gap-[14px]"
              style={{ animationDelay: `${i * 0.16}s` }}
            >
              {yang ? (
                <span className="h-[11px] w-full rounded-[1px]" style={{ background: "var(--ink-secondary)" }} />
              ) : (
                <>
                  <span className="h-[11px] flex-1 rounded-[1px]" style={{ background: "var(--ink-secondary)" }} />
                  <span className="h-[11px] flex-1 rounded-[1px]" style={{ background: "var(--ink-secondary)" }} />
                </>
              )}
            </div>
          ))}
        </div>
      </div>

      <p className="flex items-center gap-2 text-[14px] tracking-[0.24em] text-gold">
        <span className="dvn-dot h-1.5 w-1.5 rounded-full bg-gold" aria-hidden />
        起卦中
      </p>
    </section>
  );
}
