"use client";

/**
 * 排盘仪式动效:星光自天而降击打罗盘,罗盘齿环转动、十二地支依次点亮,
 * 击中瞬间脉冲扩散,随后盘面级联入场(palace-enter 承接)。
 * 全程约 1.8s;prefers-reduced-motion 下由调用方直接跳过本组件。
 */

const BRANCHES = ["子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"];
const BAGUA = ["☰", "☱", "☲", "☳", "☴", "☵", "☶", "☷"];

export function LuopanCast() {
  const R = 150; // 视图半径
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-[rgba(5,7,14,0.86)]" aria-hidden>
      <style>{`
        @keyframes lp-spin { to { transform: rotate(360deg); } }
        @keyframes lp-spin-rev { to { transform: rotate(-360deg); } }
        @keyframes lp-star-fall {
          0% { transform: translate(var(--sx), -46vh) scale(0.6); opacity: 0; }
          12% { opacity: 1; }
          62% { transform: translate(0, 0) scale(1); opacity: 1; }
          68% { opacity: 0; }
          100% { transform: translate(0, 0); opacity: 0; }
        }
        @keyframes lp-pulse {
          0%, 55% { transform: scale(0.35); opacity: 0; }
          62% { transform: scale(0.5); opacity: 0.9; }
          100% { transform: scale(1.9); opacity: 0; }
        }
        @keyframes lp-branch-glow {
          0%, 40% { fill: var(--ink-faint); }
          58%, 100% { fill: var(--gold); }
        }
        @keyframes lp-hub {
          0%, 55% { opacity: 0.35; }
          64% { opacity: 1; }
          100% { opacity: 0.8; }
        }
        @keyframes lp-fade-in { from { opacity: 0; transform: scale(0.92); } to { opacity: 1; transform: none; } }
      `}</style>

      <div className="relative" style={{ width: R * 2 + 60, height: R * 2 + 60, animation: "lp-fade-in 280ms var(--ease-out) both" }}>
        {/* 星光三道,错落坠向盘心 */}
        {[0, 1, 2].map((i) => (
          <span
            key={i}
            className="absolute left-1/2 top-1/2 h-16 w-[2px] rounded-full"
            style={{
              // @ts-expect-error 自定义变量
              "--sx": `${[-90, 40, 110][i]}px`,
              background: "linear-gradient(to bottom, transparent, var(--gold-bright))",
              transformOrigin: "center bottom",
              animation: `lp-star-fall 1.35s ${i * 0.14}s var(--ease-inout) both`,
              marginLeft: -1,
              marginTop: -64,
            }}
          />
        ))}

        {/* 击中脉冲 */}
        <span
          className="absolute left-1/2 top-1/2 rounded-full"
          style={{
            width: R * 2, height: R * 2, marginLeft: -R, marginTop: -R,
            boxShadow: "0 0 0 1.5px var(--gold-dim), var(--glow-gold-strong)",
            animation: "lp-pulse 1.8s var(--ease-out) both",
          }}
        />

        {/* 罗盘本体 */}
        <svg
          viewBox={`${-R - 30} ${-R - 30} ${R * 2 + 60} ${R * 2 + 60}`}
          className="absolute inset-0 h-full w-full"
        >
          {/* 外环:十二地支(顺转,依次点亮) */}
          <g style={{ animation: "lp-spin 14s linear infinite" }}>
            <circle r={R} fill="none" stroke="var(--line-strong)" strokeWidth="1" />
            <circle r={R - 26} fill="none" stroke="var(--line)" strokeWidth="1" />
            {BRANCHES.map((b, i) => {
              const a = (i * 30 - 90) * (Math.PI / 180);
              return (
                <text
                  key={b}
                  x={Math.cos(a) * (R - 13)}
                  y={Math.sin(a) * (R - 13)}
                  textAnchor="middle"
                  dominantBaseline="central"
                  fontSize="13"
                  style={{ animation: `lp-branch-glow 1.8s ${0.35 + i * 0.055}s both`, fontFamily: "var(--font-display)" }}
                >
                  {b}
                </text>
              );
            })}
            {/* 刻度齿 */}
            {Array.from({ length: 60 }, (_, i) => {
              const a = (i * 6 - 90) * (Math.PI / 180);
              const r1 = R - 26, r2 = i % 5 === 0 ? R - 34 : R - 30;
              return (
                <line
                  key={i}
                  x1={Math.cos(a) * r1} y1={Math.sin(a) * r1}
                  x2={Math.cos(a) * r2} y2={Math.sin(a) * r2}
                  stroke="var(--line-strong)" strokeWidth="1"
                />
              );
            })}
          </g>

          {/* 中环:八卦(逆转) */}
          <g style={{ animation: "lp-spin-rev 9s linear infinite" }}>
            <circle r={R - 46} fill="none" stroke="var(--line)" strokeWidth="1" strokeDasharray="3 6" />
            {BAGUA.map((g, i) => {
              const a = (i * 45 - 90) * (Math.PI / 180);
              return (
                <text
                  key={g}
                  x={Math.cos(a) * (R - 60)}
                  y={Math.sin(a) * (R - 60)}
                  textAnchor="middle"
                  dominantBaseline="central"
                  fontSize="14"
                  fill="var(--ink-faint)"
                >
                  {g}
                </text>
              );
            })}
          </g>

          {/* 内环齿轮(顺转快) */}
          <g style={{ animation: "lp-spin 5s linear infinite" }}>
            <circle r={R - 82} fill="none" stroke="var(--gold-dim)" strokeWidth="1" />
            {Array.from({ length: 24 }, (_, i) => {
              const a = (i * 15) * (Math.PI / 180);
              return (
                <line
                  key={i}
                  x1={Math.cos(a) * (R - 82)} y1={Math.sin(a) * (R - 82)}
                  x2={Math.cos(a) * (R - 76)} y2={Math.sin(a) * (R - 76)}
                  stroke="var(--gold-dim)" strokeWidth="2"
                />
              );
            })}
          </g>

          {/* 盘心 */}
          <circle r="7" fill="var(--gold)" style={{ animation: "lp-hub 1.8s var(--ease-out) both" }} />
        </svg>

        <p className="absolute inset-x-0 -bottom-10 text-center text-[13px] tracking-[0.24em] text-ink-faint">
          观星定盘
        </p>
      </div>
    </div>
  );
}
