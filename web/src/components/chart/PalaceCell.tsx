"use client";

import type { HoroscopeScope, Palace, Star } from "@/lib/types";
import {
  branchName, brightnessVar, groupStars, sihuaBadgeStyle, stemName,
  type Density,
} from "@/lib/chart-helpers";

/**
 * 竖排星柱(古典盘式):星名自上而下一柱一星,柱脚缀亮度小字与四化印。
 * 传统命盘星曜皆竖排,竖柱既省横向空间(多星宫不再折行),又与古籍版式同气。
 */
/** 三档密度 → 竖柱字号/柱距:简洁疏朗大字、专业均衡、大师紧凑纳杂曜 */
const MAJOR_SIZE: Record<Density, string> = {
  simple: "text-[19px]",
  pro: "text-[16px]",
  master: "text-[15px]",
};
const ASSIST_SIZE: Record<Density, string> = {
  simple: "text-[12.5px]",
  pro: "text-[12.5px]",
  master: "text-[11.5px]",
};
const COL_GAP: Record<Density, string> = {
  simple: "gap-x-3",
  pro: "gap-x-2",
  master: "gap-x-1.5",
};

function StarColumn({ star, size, fontCls }: { star: Star; size: "lg" | "md"; fontCls: string }) {
  // 庙旺主星带星光辉晕:亮度语义从「颜色」升级为「颜色 + 光」
  const glow =
    size === "lg" && (star.brightness === "庙" || star.brightness === "旺")
      ? { textShadow: "0 0 10px var(--gold-glow), 0 0 18px var(--gold-glow)" }
      : undefined;
  const color = brightnessVar(star.brightness);
  return (
    <span className="inline-flex flex-col items-center gap-0.5">
      {/* 逐字 block 竖叠(不依赖 writing-mode 的垂直字体度量,跨端稳定) */}
      <span
        className={
          size === "lg" ? `star-major font-display font-semibold ${fontCls}` : fontCls
        }
        style={{ color, ...glow }}
      >
        {star.name.split("").map((ch, i) => (
          <span key={i} className="block text-center leading-[1.12]">
            {ch}
          </span>
        ))}
      </span>
      {star.brightness && (
        <span className="text-[9.5px] leading-none opacity-75" style={{ color }}>
          {star.brightness}
        </span>
      )}
      {star.siHua && (
        <span
          className="rounded-[2px] px-[3px] py-px text-[9.5px] font-semibold leading-none"
          style={sihuaBadgeStyle(star.siHua)}
        >
          {star.siHua}
        </span>
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
  /** 格局联动:悬停格局卡/徽章时本宫为其关联宫,点亮金晕 */
  patternGlow?: boolean;
  /** 选宫聚焦:他宫被选且本宫不在其三方四正时降暗,让焦点结构浮出 */
  dimmed?: boolean;
  /** 运限叠加:各激活层在此宫的流曜与运限宫名 */
  overlayStars?: Star[];
  overlayNames?: { scope: string; name: string }[];
  enterDelay?: number;
  onSelect: (branch: number) => void;
}

export function PalaceCell({
  palace, density, selected, inSanFang, patternGlow = false, dimmed = false,
  overlayStars, overlayNames, enterDelay = 0, onSelect,
}: PalaceCellProps) {
  const { major, assist, adjective } = groupStars(palace.stars);

  return (
    <button
      type="button"
      onClick={() => onSelect(palace.branch)}
      aria-pressed={selected}
      aria-label={`${palace.name},${branchName(palace.branch)}宫`}
      className={[
        "palace-cell palace-enter relative flex min-h-[124px] w-full flex-col rounded-[6px] p-2 pb-1.5 text-left",
        "bg-bg-raised transition-[box-shadow,opacity,filter] duration-300",
        palace.isMingGong && !dimmed ? "ming-breathe" : "",
        selected
          ? "shadow-[0_0_0_2px_var(--gold),var(--glow-gold)]"
          : patternGlow
            ? "shadow-[0_0_0_1.5px_var(--gold),0_0_20px_rgba(217,179,108,0.18)]"
            : inSanFang
              ? "shadow-[0_0_0_1px_var(--gold-dim),0_0_14px_rgba(217,179,108,0.07)]"
              : "shadow-[0_0_0_1px_var(--line)] hover:shadow-[0_0_0_1px_var(--gold-dim),0_0_16px_rgba(217,179,108,0.08)]",
        dimmed ? "opacity-50 saturate-[0.8]" : "opacity-100",
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

      {/* 星区:主星大柱 + 辅星小柱并排竖排(古典盘式,身宫徽标右上避让;柱距随密度档) */}
      <div className={["flex flex-wrap items-start gap-y-1", COL_GAP[density], palace.isShenGong ? "pr-6" : ""].join(" ")}>
        {major.map((s) => (
          <StarColumn key={s.name} star={s} size="lg" fontCls={MAJOR_SIZE[density]} />
        ))}
        {density !== "simple" &&
          assist.map((s) => <StarColumn key={s.name} star={s} size="md" fontCls={ASSIST_SIZE[density]} />)}
        {major.length === 0 && palace.borrowedStars && palace.borrowedStars.length > 0 && (
          <>
            <span className="mt-0.5 text-[10px] leading-none text-ink-faint">借</span>
            {palace.borrowedStars.map((n) => (
              <span key={n} className="text-[12.5px] text-ink-secondary opacity-80">
                {n.split("").map((ch, i) => (
                  <span key={i} className="block text-center leading-[1.12]">
                    {ch}
                  </span>
                ))}
              </span>
            ))}
          </>
        )}
      </div>

      {/* 杂曜行(大师档;含年支系补充杂曜大耗/龙德/劫煞) */}
      {density === "master" && (adjective.length > 0 || (palace.extraStars?.length ?? 0) > 0) && (
        <div className="mt-0.5 flex flex-wrap gap-x-1.5 text-[11px] leading-tight text-ink-faint">
          {adjective.map((s) => (
            <span key={s.name}>{s.name}</span>
          ))}
          {(palace.extraStars ?? []).map((s) => (
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
