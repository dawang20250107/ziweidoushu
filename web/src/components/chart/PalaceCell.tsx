"use client";

import type { HoroscopeScope, Palace, Star } from "@/lib/types";
import {
  branchName, brightnessVar, groupStars, sihuaVar, stemName,
  type Density,
} from "@/lib/chart-helpers";

/** 星名 + 四化徽章(本命实心;流曜空心由 StarBadge outline 表达) */
function StarGlyph({ star, size }: { star: Star; size: "lg" | "md" }) {
  return (
    <span
      className={size === "lg" ? "font-display text-[17px] font-semibold leading-tight" : "text-[13px] leading-tight"}
      style={{ color: brightnessVar(star.brightness) }}
    >
      {star.name}
      {star.siHua && (
        <sup
          className="ml-px rounded-[2px] px-[3px] text-[10px] font-semibold not-italic"
          style={{ background: sihuaVar(star.siHua), color: "var(--bg)" }}
        >
          {star.siHua}
        </sup>
      )}
    </span>
  );
}

/** 流曜徽章(空心描边,与本命星区分) */
function FlowStarBadge({ name }: { name: string }) {
  return (
    <span className="rounded-[2px] px-1 text-[11px] leading-[18px] text-gold-dim shadow-[inset_0_0_0_1px_var(--gold-dim)]">
      {name}
    </span>
  );
}

export interface PalaceCellProps {
  palace: Palace;
  density: Density;
  selected: boolean;
  inSanFang: boolean;
  /** 运限叠加:各激活层在此宫的流曜与运限宫名 */
  overlayStars?: Star[];
  overlayNames?: { scope: string; name: string }[];
  enterDelay?: number;
  onSelect: (branch: number) => void;
}

export function PalaceCell({
  palace, density, selected, inSanFang, overlayStars, overlayNames, enterDelay = 0, onSelect,
}: PalaceCellProps) {
  const { major, assist, adjective } = groupStars(palace.stars);

  return (
    <button
      type="button"
      onClick={() => onSelect(palace.branch)}
      aria-pressed={selected}
      aria-label={`${palace.name},${branchName(palace.branch)}宫`}
      className={[
        "palace-enter relative flex min-h-[124px] flex-col rounded-[6px] p-2 pb-1.5 text-left transition-shadow",
        "bg-bg-raised",
        selected
          ? "shadow-[0_0_0_2px_var(--gold)]"
          : inSanFang
            ? "shadow-[0_0_0_1px_var(--gold-dim)]"
            : "shadow-[0_0_0_1px_var(--line)] hover:shadow-[0_0_0_1px_var(--line-strong)]",
      ].join(" ")}
      style={{
        animationDelay: `${enterDelay}ms`,
        backgroundImage: palace.isMingGong
          ? "linear-gradient(var(--gold-glow), transparent 60%)"
          : undefined,
      }}
    >
      {palace.isShenGong && (
        <span className="absolute right-1.5 top-1.5 rounded-[2px] border border-gold-dim px-1 text-[10px] leading-4 text-gold">
          身
        </span>
      )}

      {/* 主星行 */}
      <div className="flex flex-wrap gap-x-2.5 gap-y-0.5">
        {major.map((s) => (
          <StarGlyph key={s.name} star={s} size="lg" />
        ))}
        {major.length === 0 && palace.borrowedStars && palace.borrowedStars.length > 0 && (
          <span className="text-[13px] text-ink-faint">
            借<span className="ml-1 text-ink-secondary">{palace.borrowedStars.join(" ")}</span>
          </span>
        )}
      </div>

      {/* 辅星行(专业/大师档) */}
      {density !== "simple" && assist.length > 0 && (
        <div className="mt-0.5 flex flex-wrap gap-x-2 gap-y-0.5">
          {assist.map((s) => (
            <StarGlyph key={s.name} star={s} size="md" />
          ))}
        </div>
      )}

      {/* 杂曜行(大师档) */}
      {density === "master" && adjective.length > 0 && (
        <div className="mt-0.5 flex flex-wrap gap-x-1.5 text-[11px] leading-tight text-ink-faint">
          {adjective.map((s) => (
            <span key={s.name}>{s.name}</span>
          ))}
        </div>
      )}

      {/* 运限流曜(空心徽章) */}
      {overlayStars && overlayStars.length > 0 && (
        <div className="mt-1 flex flex-wrap gap-1">
          {overlayStars.map((s, i) => (
            <FlowStarBadge key={`${i}-${s.name}`} name={s.name} />
          ))}
        </div>
      )}

      <div className="flex-1" />

      {/* 大师档:长生/博士十二神 */}
      {density === "master" && (palace.changsheng12 || palace.boshi12) && (
        <div className="mb-1 flex gap-2 text-[10px] text-ink-faint">
          {palace.changsheng12 && <span>{palace.changsheng12}</span>}
          {palace.boshi12 && <span>{palace.boshi12}</span>}
        </div>
      )}

      {/* 底栏:宫名 · 干支 · 大限 + 运限宫名 */}
      <div className="border-t border-line pt-1">
        <div className="flex items-baseline justify-between gap-1">
          <span className="font-display text-[13px] font-semibold text-ink">{palace.name}</span>
          <span className="text-[11px] text-ink-secondary">
            {stemName(palace.stem)}
            {branchName(palace.branch)}
            <span className="tnum ml-1 text-ink-faint">
              {palace.daXianStart}-{palace.daXianEnd}
            </span>
          </span>
        </div>
        {overlayNames && overlayNames.length > 0 && (
          <div className="mt-0.5 flex flex-wrap gap-x-2 text-[10px] text-gold-dim">
            {overlayNames.map((o) => (
              <span key={o.scope}>
                {o.scope}·{o.name}
              </span>
            ))}
          </div>
        )}
      </div>
    </button>
  );
}
