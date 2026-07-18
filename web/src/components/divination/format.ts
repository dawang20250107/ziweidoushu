/** 卦档展示辅助。 */

export const DIVINATION_KIND_LABEL: Record<string, string> = {
  meihua: "梅花易数",
  liuyao: "六爻纳甲",
  xiaoliuren: "小六壬",
};

/** ISO 时间 → 「2026-07-18 09:30」。 */
export function formatDivinationTime(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  const p = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`;
}
