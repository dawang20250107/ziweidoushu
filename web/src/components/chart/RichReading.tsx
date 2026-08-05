import { Fragment } from "react";

/**
 * 断语富文本渲染:把确定性断语里的结构标记转成视觉层级,
 * 让信息量大的段落可扫读——
 *   【…】 双主星组合 / 应验档 / 宫名  → 金色强调
 *   ★…    运限引动本命格局          → 独立高亮要点
 *   ——     综论/引导前缀              → 轻微断行留白
 * 不改后端文本,纯呈现层增强。
 */

// 高亮 【…】 标记。
function withLabels(text: string, keyBase: string) {
  const parts = text.split(/(【[^】]*】)/g);
  return parts.map((p, i) =>
    p.startsWith("【") && p.endsWith("】") ? (
      <span key={`${keyBase}-${i}`} className="font-medium text-gold">
        {p}
      </span>
    ) : (
      <Fragment key={`${keyBase}-${i}`}>{p}</Fragment>
    ),
  );
}

export function RichReading({ text, className = "" }: { text: string; className?: string }) {
  // 以 ★ 分出「引动格局」高亮要点,其余为正文流。
  const segments = text.split(/(★[^★]*)/g).filter(Boolean);
  return (
    <div className={`flex flex-col gap-2 ${className}`}>
      {segments.map((seg, i) =>
        seg.startsWith("★") ? (
          <p
            key={i}
            className="rounded-[6px] bg-[var(--gold-glow)] px-3 py-2 text-[13.5px] leading-[1.85] text-ink shadow-[inset_0_0_0_1px_var(--gold-dim)]"
          >
            <span className="mr-1 text-gold">★</span>
            {withLabels(seg.slice(1), `star-${i}`)}
          </p>
        ) : (
          <p key={i} className="text-[14px] leading-[1.95] text-ink-secondary">
            {withLabels(seg, `body-${i}`)}
          </p>
        ),
      )}
    </div>
  );
}

// LEVEL_ACCENT 维度吉凶 → 左缘强调色(用于详情卡)。
export const LEVEL_ACCENT: Record<string, string> = {
  good: "var(--ok)",
  caution: "var(--danger)",
  neutral: "var(--line-strong)",
};
