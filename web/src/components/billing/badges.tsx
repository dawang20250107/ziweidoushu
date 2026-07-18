import type { OrderStatus, Tier } from "@/lib/billing";
import { ORDER_STATUS_LABEL, TIER_LABEL } from "@/lib/billing";

/**
 * 会员层级徽标。free=描边弱化,pro/master=金色描边强调。
 * 器物感近方角(2px),单强调色。
 */
export function TierBadge({ tier }: { tier: Tier }) {
  const gold = tier !== "free";
  return (
    <span
      className={[
        "inline-flex items-center rounded-[2px] px-1.5 py-0.5 text-[11px] font-medium tracking-[0.08em]",
        gold
          ? "text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]"
          : "text-ink-faint shadow-[inset_0_0_0_1px_var(--line)]",
      ].join(" ")}
    >
      {TIER_LABEL[tier]}
    </span>
  );
}

// 订单状态 → 语义色。fulfilled/paid 用成功绿,created/paying 用警告赭,
// closed 用弱文本,refunded 用信息青。四化色永不复用,此处仅系统语义色。
const STATUS_TONE: Record<OrderStatus, string> = {
  created: "text-warn shadow-[inset_0_0_0_1px_var(--warn)]",
  paying: "text-warn shadow-[inset_0_0_0_1px_var(--warn)]",
  paid: "text-ok shadow-[inset_0_0_0_1px_var(--ok)]",
  fulfilled: "text-ok shadow-[inset_0_0_0_1px_var(--ok)]",
  closed: "text-ink-faint shadow-[inset_0_0_0_1px_var(--line)]",
  refunded: "text-info shadow-[inset_0_0_0_1px_var(--info)]",
};

/** 订单状态徽标(中文 + 语义色描边)。 */
export function OrderStatusBadge({ status }: { status: OrderStatus }) {
  return (
    <span
      className={[
        "inline-flex items-center rounded-[2px] px-1.5 py-0.5 text-[12px] font-medium",
        STATUS_TONE[status] ?? "text-ink-faint shadow-[inset_0_0_0_1px_var(--line)]",
      ].join(" ")}
    >
      {ORDER_STATUS_LABEL[status] ?? status}
    </span>
  );
}
