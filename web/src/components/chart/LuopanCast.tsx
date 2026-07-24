"use client";

/**
 * 排盘仪式动效(盛宴版):
 *   ①星辰自四方汇聚盘心 + 三道星光坠落 → ②罗盘三环错落显形(缩放旋入)
 *   → ③击中脉冲扩散、金色点亮波沿地支环扫过一周 → ④盘心辉光呼吸、宫格承接入场。
 * 全程约 1.9s;prefers-reduced-motion 下由调用方直接跳过本组件。
 */

const BRANCHES = ["子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"];
const BAGUA = ["☰", "☱", "☲", "☳", "☴", "☵", "☶", "☷"];

/** 汇聚星辰(14 粒,确定性方位/距离/延迟)。 */
const MOTES = Array.from({ length: 14 }, (_, i) => {
  const a = (i * 137.5 * Math.PI) / 180; // 黄金角散布
  const d = 190 + (i % 5) * 46;
  return {
    dx: (Math.cos(a) * d).toFixed(1),
    dy: (Math.sin(a) * d).toFixed(1),
    delay: (i % 7) * 0.06,
    r: 1.5 + (i % 3),
  };
});

export function LuopanCast() {
  const R = 150; // 视图半径
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      style={{
        background:
          "radial-gradient(60% 60% at 50% 46%, rgba(109,95,163,0.10), transparent 70%), rgba(5,7,14,0.88)",
      }}
      aria-hidden
    >
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
        @keyframes lp-mote {
          0% { transform: translate(var(--dx), var(--dy)) scale(0.2); opacity: 0; }
          18% { opacity: 1; }
          58% { transform: translate(0, 0) scale(1); opacity: 0.9; }
          66%, 100% { transform: translate(0, 0) scale(0.2); opacity: 0; }
        }
        @keyframes lp-pulse {
          0%, 55% { transform: scale(0.35); opacity: 0; }
          62% { transform: scale(0.5); opacity: 0.9; }
          100% { transform: scale(1.9); opacity: 0; }
        }
        @keyframes lp-ring-in {
          from { opacity: 0; transform: scale(0.72) rotate(-16deg); }
          to { opacity: 1; transform: scale(1) rotate(0deg); }
        }
        @keyframes lp-branch-glow {
          0%, 40% { fill: var(--ink-faint); text-shadow: none; }
          58%, 100% { fill: var(--gold); }
        }
        @keyframes lp-sweep { from { transform: rotate(0deg); opacity: 0.9; } to { transform: rotate(360deg); opacity: 0; } }
        @keyframes lp-hub {
          0%, 55% { opacity: 0.35; }
          64% { opacity: 1; }
          100% { opacity: 0.85; }
        }
        @keyframes lp-title {
          from { opacity: 0; letter-spacing: 0.1em; }
          40% { opacity: 1; }
          to { opacity: 0.9; letter-spacing: 0.34em; }
        }
        @keyframes lp-fade-in { from { opacity: 0; transform: scale(0.92); } to { opacity: 1; transform: none; } }
      `}</style>

      <div
        className="relative"
        style={{ width: R * 2 + 60, height: R * 2 + 60, animation: "lp-fade-in 280ms var(--ease-out) both" }}
      >
        {/* 星辰汇聚:四方星粒被罗盘引力吸入盘心 */}
        {MOTES.map((m, i) => (
          <span
            key={i}
            className="absolute left-1/2 top-1/2 rounded-full"
            style={{
              // @ts-expect-error 自定义变量
              "--dx": `${m.dx}px`,
              "--dy": `${m.dy}px`,
              width: m.r * 2,
              height: m.r * 2,
              marginLeft: -m.r,
              marginTop: -m.r,
              background: i % 4 === 0 ? "var(--gold-bright)" : "#cdd6f0",
              boxShadow: i % 4 === 0 ? "0 0 8px var(--gold)" : "0 0 6px rgba(205,214,240,0.8)",
              animation: `lp-mote 1.3s ${m.delay}s var(--ease-inout) both`,
            }}
          />
        ))}

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
          {/* 外环显形 → 十二地支顺转、依次点亮 */}
          <g style={{ animation: "lp-ring-in 0.6s 0.05s var(--ease-out) both" }}>
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
          </g>

          {/* 点亮波:击中后金弧沿地支环扫一周 */}
          <g style={{ animation: "lp-sweep 1.0s 0.62s var(--ease-inout) both", transformOrigin: "0 0" }}>
            <circle
              r={R - 13}
              fill="none"
              stroke="var(--gold-bright)"
              strokeWidth="22"
              strokeLinecap="round"
              strokeDasharray={`${(R - 13) * 0.7} ${(R - 13) * 7}`}
              opacity="0.16"
            />
          </g>

          {/* 中环显形 → 八卦逆转 */}
          <g style={{ animation: "lp-ring-in 0.6s 0.16s var(--ease-out) both" }}>
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
          </g>

          {/* 内环显形 → 齿轮顺转快 */}
          <g style={{ animation: "lp-ring-in 0.6s 0.27s var(--ease-out) both" }}>
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
          </g>

          {/* 盘心:击中后辉光常驻 */}
          <circle r="10" fill="var(--gold)" opacity="0.25" style={{ animation: "lp-hub 1.8s var(--ease-out) both" }} />
          <circle r="6" fill="var(--gold)" style={{ animation: "lp-hub 1.8s var(--ease-out) both" }} />
        </svg>

        <p
          className="absolute inset-x-0 -bottom-10 text-center text-[13px] text-ink-secondary"
          style={{ animation: "lp-title 1.6s 0.3s var(--ease-out) both" }}
        >
          观星定盘
        </p>
      </div>
    </div>
  );
}
