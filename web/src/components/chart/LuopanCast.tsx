"use client";

import { createPortal } from "react-dom";

/**
 * 排盘仪式动效(全屏盛典版),约 3.6s + 0.5s 收场:
 *   ①天幕降下,星辰自全屏四野汇入盘心,流星两道掠空 → ②罗盘四环错落显形
 *   (二十八宿/地支/八卦/齿轮) → ③星光击盘:脉冲扩散、金波沿地支环扫过一周
 *   → ④四化印「禄权科忌」自盘心飞临四方 → ⑤远环震荡至屏幕边缘、金芒一闪、「盘成」。
 * 步进字幕与真实安星次序同构:定五行局 → 安十四主星 → 起生年四化。
 * 经 createPortal 挂到 body:任何祖先的 transform/filter 都无法劫持 fixed 定位,
 * 保证始终以视口为中心全屏呈现。prefers-reduced-motion 下由调用方直接跳过本组件。
 */

const BRANCHES = ["子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"];
const BAGUA = ["☰", "☱", "☲", "☳", "☴", "☵", "☶", "☷"];
const XIU = "角亢氐房心尾箕斗牛女虚危室壁奎娄胃昴毕觜参井鬼柳星张翼轸".split("");

/** 汇聚星辰(26 粒,黄金角散布全屏,确定性方位/延迟)。距离单位 vw/vh:自屏缘飞入。 */
const MOTES = Array.from({ length: 26 }, (_, i) => {
  const a = (i * 137.5 * Math.PI) / 180;
  const d = 34 + (i % 5) * 7;
  return {
    dx: (Math.cos(a) * d).toFixed(1),
    dy: (Math.sin(a) * d).toFixed(1),
    delay: ((i % 9) * 0.07).toFixed(2),
    r: 1.5 + (i % 3),
    gold: i % 4 === 0,
  };
});

/** 流星两道:恒定倾角平移掠空。 */
const SHOOTS = [
  { fx: "-58vw", fy: "-34vh", tx: "16vw", ty: "6vh", rot: "26deg", delay: 0.45 },
  { fx: "52vw", fy: "-42vh", tx: "-12vw", ty: "-4vh", rot: "152deg", delay: 1.5 },
];

/** 四化印:自盘心飞临四方(色随四化 token)。 */
const HUAS = [
  { ch: "禄", color: "var(--sihua-lu)", hx: "-25vmin", hy: "-17vmin" },
  { ch: "权", color: "var(--sihua-quan)", hx: "25vmin", hy: "-17vmin" },
  { ch: "科", color: "var(--sihua-ke)", hx: "-25vmin", hy: "17vmin" },
  { ch: "忌", color: "var(--sihua-ji)", hx: "25vmin", hy: "17vmin" },
];

/** 步进字幕:与安星算法真实次序同构。 */
const STEPS = [
  { t: "定五行局", delay: 0.55 },
  { t: "安十四主星", delay: 1.5 },
  { t: "起生年四化", delay: 2.45 },
];

export function LuopanCast({ leaving = false }: { leaving?: boolean }) {
  const R = 150; // SVG 坐标半径(容器按 vmin 缩放)
  return createPortal(
    <div
      className="lp-stage fixed inset-0 z-[80] flex flex-col items-center justify-center overflow-hidden"
      data-leaving={leaving || undefined}
      style={{
        background:
          "radial-gradient(70% 70% at 50% 44%, rgba(109,95,163,0.14), transparent 70%), radial-gradient(120% 90% at 50% 110%, rgba(191,155,73,0.06), transparent 60%), rgba(4,6,12,0.965)",
      }}
      aria-hidden
    >
      <style>{`
        @keyframes lp-in { from { opacity: 0; } to { opacity: 1; } }
        @keyframes lp-out { from { opacity: 1; transform: scale(1); } to { opacity: 0; transform: scale(1.045); } }
        .lp-stage { animation: lp-in 320ms var(--ease-out) both, lp-quake 0.5s 1.26s var(--ease-out) both; }
        .lp-stage[data-leaving] { animation: lp-out 480ms var(--ease-inout) both; }
        @keyframes lp-quake {
          0%, 100% { translate: 0 0; }
          18% { translate: 7px -5px; }
          38% { translate: -6px 4px; }
          58% { translate: 4px 2px; }
          78% { translate: -2px -2px; }
        }
        @keyframes lp-spin { to { transform: rotate(360deg); } }
        @keyframes lp-spin-rev { to { transform: rotate(-360deg); } }
        @keyframes lp-star-fall {
          0% { transform: translate(var(--sx), -52vh) scale(0.6); opacity: 0; }
          12% { opacity: 1; }
          62% { transform: translate(0, 0) scale(1); opacity: 1; }
          68% { opacity: 0; }
          100% { transform: translate(0, 0); opacity: 0; }
        }
        @keyframes lp-mote {
          0% { transform: translate(calc(var(--mx) * 1vw), calc(var(--my) * 1vh)) scale(0.2); opacity: 0; }
          16% { opacity: 1; }
          62% { transform: translate(0, 0) scale(1); opacity: 0.9; }
          70%, 100% { transform: translate(0, 0) scale(0.2); opacity: 0; }
        }
        @keyframes lp-shoot {
          0% { transform: translate(var(--fx), var(--fy)) rotate(var(--rot)); opacity: 0; }
          14% { opacity: 0.9; }
          100% { transform: translate(var(--tx), var(--ty)) rotate(var(--rot)); opacity: 0; }
        }
        @keyframes lp-pulse { from { transform: scale(0.35); opacity: 0.9; } to { transform: scale(2.1); opacity: 0; } }
        @keyframes lp-pulse-far { from { transform: scale(0.1); opacity: 0.5; } to { transform: scale(1); opacity: 0; } }
        @keyframes lp-ring-in {
          from { opacity: 0; transform: scale(0.72) rotate(-16deg); }
          to { opacity: 1; transform: scale(1) rotate(0deg); }
        }
        @keyframes lp-branch-glow { 0%, 40% { fill: var(--ink-faint); } 56% { fill: var(--gold-bright); } 74%, 100% { fill: var(--gold); } }
        @keyframes lp-sweep { from { transform: rotate(0deg); opacity: 1; } to { transform: rotate(360deg); opacity: 0; } }
        @keyframes lp-hit {
          0% { transform: scale(0.55); opacity: 0; }
          28% { opacity: 0.6; }
          100% { transform: scale(1.18); opacity: 0; }
        }
        @keyframes lp-hub { 0% { opacity: 0.3; } 42% { opacity: 1; } 100% { opacity: 0.85; } }
        @keyframes lp-glow { from { opacity: 0; } to { opacity: 0.55; } }
        @keyframes lp-hua {
          0% { transform: translate(0, 0) scale(0.3); opacity: 0; }
          28% { transform: translate(calc(var(--hx) * 0.86), calc(var(--hy) * 0.86)) scale(1.12); opacity: 1; }
          46% { transform: translate(var(--hx), var(--hy)) scale(1); opacity: 1; }
          82% { opacity: 0.95; }
          100% { transform: translate(var(--hx), calc(var(--hy) - 2.2vmin)) scale(0.96); opacity: 0; }
        }
        @keyframes lp-cap {
          0% { opacity: 0; transform: translateY(10px); }
          14%, 74% { opacity: 1; transform: translateY(0); }
          100% { opacity: 0; transform: translateY(-8px); }
        }
        @keyframes lp-cap-final {
          from { opacity: 0; transform: scale(0.82); letter-spacing: 0.1em; }
          to { opacity: 1; transform: scale(1); letter-spacing: 0.5em; }
        }
        @keyframes lp-flash { 0% { opacity: 0; } 45% { opacity: 0.32; } 100% { opacity: 0; } }
        @keyframes lp-disk-in { from { opacity: 0; transform: scale(0.9); } to { opacity: 1; transform: none; } }
      `}</style>

      {/* 星辰汇聚:全屏四野的星粒被罗盘引力吸入盘心 */}
      {MOTES.map((m, i) => (
        <span
          key={i}
          className="absolute left-1/2 top-1/2 rounded-full"
          style={{
            // @ts-expect-error 自定义变量
            "--mx": m.dx,
            "--my": m.dy,
            width: m.r * 2,
            height: m.r * 2,
            marginLeft: -m.r,
            marginTop: -m.r,
            background: m.gold ? "var(--gold-bright)" : "#cdd6f0",
            boxShadow: m.gold ? "0 0 10px var(--gold)" : "0 0 7px rgba(205,214,240,0.8)",
            animation: `lp-mote 1.5s ${m.delay}s var(--ease-inout) both`,
          }}
        />
      ))}

      {/* 流星掠空 */}
      {SHOOTS.map((s, i) => (
        <span
          key={i}
          className="absolute left-1/2 top-1/2 h-[1.5px] w-36 rounded-full"
          style={{
            // @ts-expect-error 自定义变量
            "--fx": s.fx,
            "--fy": s.fy,
            "--tx": s.tx,
            "--ty": s.ty,
            "--rot": s.rot,
            background: "linear-gradient(to right, transparent, rgba(240,214,160,0.95))",
            animation: `lp-shoot 0.95s ${s.delay}s var(--ease-inout) both`,
          }}
        />
      ))}

      {/* 击中后的全屏远环震荡 */}
      <span
        className="absolute left-1/2 top-1/2 h-[120vmax] w-[120vmax] -translate-x-1/2 -translate-y-1/2 rounded-full"
        style={{
          boxShadow: "inset 0 0 0 1.5px var(--gold-dim), inset 0 0 60px rgba(191,155,73,0.25)",
          animation: "lp-pulse-far 1.05s 2.55s var(--ease-out) both",
        }}
      />

      {/* 盘成金芒一闪 */}
      <span
        className="absolute inset-0"
        style={{
          background: "radial-gradient(52% 52% at 50% 46%, rgba(240,214,160,0.85), transparent 68%)",
          animation: "lp-flash 0.55s 3.12s var(--ease-out) both",
        }}
      />

      {/* 罗盘主体:随视口缩放,永居屏心 */}
      <div
        className="relative aspect-square w-[min(78vmin,560px)]"
        style={{ animation: "lp-disk-in 340ms var(--ease-out) both" }}
      >
        {/* 击盘后盘心辉光洇开 */}
        <span
          className="absolute inset-0 rounded-full"
          style={{
            background: "radial-gradient(50% 50% at 50% 50%, rgba(191,155,73,0.22), transparent 70%)",
            animation: "lp-glow 0.9s 1.35s var(--ease-out) both",
          }}
        />

        {/* 星光三道,错落坠向盘心 */}
        {[0, 1, 2].map((i) => (
          <span
            key={i}
            className="absolute left-1/2 top-1/2 h-24 w-[2px] rounded-full"
            style={{
              // @ts-expect-error 自定义变量
              "--sx": `${[-140, 60, 170][i]}px`,
              background: "linear-gradient(to bottom, transparent, var(--gold-bright))",
              transformOrigin: "center bottom",
              animation: `lp-star-fall 1.35s ${0.1 + i * 0.14}s var(--ease-inout) both`,
              marginLeft: -1,
              marginTop: -96,
            }}
          />
        ))}

        {/* 击中瞬间:盘心金芒迸发 + 双重脉冲扩散 */}
        <span
          className="absolute inset-[10%] rounded-full"
          style={{
            background: "radial-gradient(50% 50% at 50% 50%, rgba(240,214,160,0.7), transparent 62%)",
            animation: "lp-hit 0.5s 1.26s var(--ease-out) both",
          }}
        />
        <span
          className="absolute inset-[6%] rounded-full"
          style={{
            boxShadow: "0 0 0 1.5px var(--gold-dim), var(--glow-gold-strong)",
            animation: "lp-pulse 0.95s 1.25s var(--ease-out) both",
          }}
        />
        <span
          className="absolute inset-[6%] rounded-full"
          style={{
            boxShadow: "0 0 0 1px var(--gold-dim)",
            animation: "lp-pulse 0.85s 1.5s var(--ease-out) both",
          }}
        />

        <svg viewBox="-200 -200 400 400" className="absolute inset-0 h-full w-full">
          {/* 最外环:二十八宿缓转(观星台意象) */}
          <g style={{ animation: "lp-ring-in 0.65s 0.3s var(--ease-out) both" }}>
            <g style={{ animation: "lp-spin 60s linear infinite" }}>
              <circle r={R + 36} fill="none" stroke="var(--line)" strokeWidth="1" opacity="0.7" />
              {XIU.map((x, i) => {
                const a = ((i * 360) / 28 - 90) * (Math.PI / 180);
                return (
                  <text
                    key={i}
                    x={(Math.cos(a) * (R + 25)).toFixed(1)}
                    y={(Math.sin(a) * (R + 25)).toFixed(1)}
                    textAnchor="middle"
                    dominantBaseline="central"
                    fontSize="9.5"
                    fill="var(--ink-faint)"
                    opacity="0.75"
                    style={{ fontFamily: "var(--font-display)" }}
                  >
                    {x}
                  </text>
                );
              })}
            </g>
          </g>

          {/* 外环显形 → 十二地支顺转、依次点亮 */}
          <g style={{ animation: "lp-ring-in 0.6s 0.4s var(--ease-out) both" }}>
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
                    style={{ animation: `lp-branch-glow 1.6s ${1.05 + i * 0.055}s both`, fontFamily: "var(--font-display)" }}
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

          {/* 点亮波:击中后金弧沿地支环扫一周。
              基线 opacity:0 + fill forwards:延迟期内完全隐形(fill both 会把 from 帧
              提前定格,金弧会在开场就挂在盘上)。 */}
          <g
            style={{
              opacity: 0,
              animation: "lp-sweep 1.1s 1.3s var(--ease-inout) forwards",
              transformOrigin: "0 0",
            }}
          >
            <circle
              r={R - 13}
              fill="none"
              stroke="var(--gold-bright)"
              strokeWidth="22"
              strokeLinecap="round"
              strokeDasharray={`${(R - 13) * 0.9} ${(R - 13) * 7}`}
              opacity="0.38"
            />
          </g>

          {/* 中环显形 → 八卦逆转 */}
          <g style={{ animation: "lp-ring-in 0.6s 0.52s var(--ease-out) both" }}>
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
          <g style={{ animation: "lp-ring-in 0.6s 0.64s var(--ease-out) both" }}>
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
          <circle r="10" fill="var(--gold)" opacity="0.25" style={{ animation: "lp-hub 1.4s 1.2s var(--ease-out) both" }} />
          <circle r="6" fill="var(--gold)" style={{ animation: "lp-hub 1.4s 1.2s var(--ease-out) both" }} />
        </svg>

        {/* 四化印飞临四方 */}
        {HUAS.map((h, i) => (
          <span
            key={h.ch}
            className="absolute left-1/2 top-1/2 flex items-center justify-center rounded-full font-semibold"
            style={{
              // @ts-expect-error 自定义变量
              "--hx": h.hx,
              "--hy": h.hy,
              width: "clamp(30px, 7vmin, 46px)",
              height: "clamp(30px, 7vmin, 46px)",
              marginLeft: "clamp(-23px, -3.5vmin, -15px)",
              marginTop: "clamp(-23px, -3.5vmin, -15px)",
              fontSize: "clamp(15px, 3.4vmin, 22px)",
              color: h.color,
              boxShadow: `inset 0 0 0 1.5px ${h.color}, 0 0 18px color-mix(in srgb, ${h.color} 55%, transparent)`,
              background: `color-mix(in srgb, ${h.color} 14%, rgba(4,6,12,0.8))`,
              fontFamily: "var(--font-display)",
              animation: `lp-hua 1.35s ${1.95 + i * 0.12}s var(--ease-inout) both`,
            }}
          >
            {h.ch}
          </span>
        ))}
      </div>

      {/* 步进字幕:与安星次序同构,末了「盘成」 */}
      <div className="relative mt-[4vmin] h-9 w-full">
        {STEPS.map((s) => (
          <p
            key={s.t}
            className="absolute inset-x-0 text-center text-[clamp(13px,2.6vmin,16px)] tracking-[0.34em] text-ink-secondary"
            style={{ fontFamily: "var(--font-display)", animation: `lp-cap 0.95s ${s.delay}s var(--ease-inout) both` }}
          >
            {s.t}
          </p>
        ))}
        <p
          className="absolute inset-x-0 text-center text-[clamp(17px,3.6vmin,22px)] font-semibold text-gold"
          style={{
            fontFamily: "var(--font-display)",
            textShadow: "0 0 22px rgba(191,155,73,0.65)",
            animation: "lp-cap-final 0.5s 3.3s var(--ease-out) both",
          }}
        >
          盘成
        </p>
      </div>
    </div>,
    document.body
  );
}
