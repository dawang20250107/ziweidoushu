import type { Tone } from "@/lib/divination";

/** 语义色调 → 徽章类(彩色文字 + 同色细描边)。单强调色约束下,仅体用/吉凶断语用语义色。 */
export function toneBadgeClass(tone: Tone): string {
  switch (tone) {
    case "gold":
      return "text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]";
    case "ok":
      return "text-ok shadow-[inset_0_0_0_1px_var(--ok)]";
    case "warn":
      return "text-warn shadow-[inset_0_0_0_1px_var(--warn)]";
    case "danger":
      return "text-danger shadow-[inset_0_0_0_1px_var(--danger)]";
  }
}

/** 语义色调 → 纯文字类。 */
export function toneTextClass(tone: Tone): string {
  switch (tone) {
    case "gold":
      return "text-gold";
    case "ok":
      return "text-ok";
    case "warn":
      return "text-warn";
    case "danger":
      return "text-danger";
  }
}
