"use client";

import { useMemo } from "react";

/**
 * 星牌顶饰:据书 slug 确定性生成的一枚星象(asterism)。
 * 每本书点位不同、但同书恒定(哈希驱动,无随机、无 SSR 抖动)。
 * 纯内联 SVG,不引任何资源;线用 --line-strong / --gold-dim,主星着金。
 * active(在读)时:连线转金、主星点亮并罩一层 --gold-glow 微晕。
 */

function hashStr(s: string): number {
  let h = 2166136261;
  for (let i = 0; i < s.length; i++) {
    h ^= s.charCodeAt(i);
    h = Math.imul(h, 16777619);
  }
  return h >>> 0;
}

/** mulberry32:小巧确定性 PRNG,同一 seed 恒定输出。 */
function mulberry32(seed: number) {
  let a = seed >>> 0;
  return () => {
    a = (a + 0x6d2b79f5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

const W = 120;
const H = 46;

export function Constellation({
  seed,
  active = false,
  className = "",
}: {
  seed: string;
  active?: boolean;
  className?: string;
}) {
  const { points, path, lead } = useMemo(() => {
    const rand = mulberry32(hashStr(seed));
    const count = 5 + Math.floor(rand() * 3); // 5..7 颗
    const pts: { x: number; y: number }[] = [];
    for (let i = 0; i < count; i++) {
      // 按列铺开(避免连线大幅交叉),各列内加抖动,得到自然的星象漫游感
      const baseX = 10 + (W - 20) * ((i + 0.5) / count);
      const x = Math.max(8, Math.min(W - 8, baseX + (rand() - 0.5) * 22));
      const y = 9 + rand() * (H - 18);
      pts.push({ x, y });
    }
    const d = pts
      .map((p, i) => `${i === 0 ? "M" : "L"}${p.x.toFixed(1)} ${p.y.toFixed(1)}`)
      .join(" ");
    const leadIdx = Math.floor(rand() * count);
    return { points: pts, path: d, lead: leadIdx };
  }, [seed]);

  return (
    <svg
      viewBox={`0 0 ${W} ${H}`}
      className={`w-full ${className}`}
      preserveAspectRatio="xMidYMid meet"
      fill="none"
      aria-hidden
    >
      <path
        d={path}
        stroke={active ? "var(--gold-dim)" : "var(--line-strong)"}
        strokeWidth={0.7}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      {points.map((p, i) => {
        if (i === lead) {
          return (
            <g key={i}>
              <circle cx={p.x} cy={p.y} r={4.6} fill="var(--gold-glow)" opacity={active ? 1 : 0.5} />
              <circle cx={p.x} cy={p.y} r={1.9} fill={active ? "var(--gold)" : "var(--gold-dim)"} />
            </g>
          );
        }
        return <circle key={i} cx={p.x} cy={p.y} r={1.1} fill="var(--line-strong)" />;
      })}
    </svg>
  );
}
