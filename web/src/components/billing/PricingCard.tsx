import type { Product } from "@/lib/billing";
import {
  benefitsFor,
  creditLabel,
  discountPercent,
  durationLabel,
  formatPrice,
  TIER_LABEL,
} from "@/lib/billing";

/** 权益点前的对勾图标(线性单色,取 token 金,不用 emoji)。 */
function Check() {
  return (
    <svg
      viewBox="0 0 16 16"
      width="16"
      height="16"
      fill="none"
      className="mt-0.5 shrink-0 text-gold"
      aria-hidden
    >
      <path
        d="M3.5 8.5l3 3 6-7"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

/**
 * 定价卡:订阅 / 次卡通用。
 * featured=true 时做视觉主推(金色强调边框 + 「最划算」徽标 + 实心金按钮)。
 */
export function PricingCard({
  product,
  featured = false,
  busy = false,
  onBuy,
}: {
  product: Product;
  featured?: boolean;
  busy?: boolean;
  onBuy: (product: Product) => void;
}) {
  const isSub = product.kind === "subscription";
  const eyebrow = isSub
    ? TIER_LABEL[product.tier ?? "pro"]
    : creditLabel(product.creditType);
  const off = discountPercent(product.priceCents, product.originalPriceCents);
  const benefits = benefitsFor(product);

  return (
    <div
      className={[
        "lift relative flex flex-col rounded-[10px] bg-bg-raised p-6 md:p-8",
        featured
          ? "shadow-[0_0_0_2px_var(--gold),0_0_24px_var(--gold-glow)]"
          : "shadow-[0_0_0_1px_var(--line)]",
      ].join(" ")}
    >
      {featured && (
        <span className="absolute -top-3 left-6 inline-flex items-center rounded-[2px] bg-gold px-2 py-0.5 text-[11px] font-semibold tracking-[0.08em] text-[#161206] md:left-8">
          最划算
        </span>
      )}

      <p className="text-[12px] font-medium tracking-[0.08em] text-gold">{eyebrow}</p>
      <h3 className="mt-2 font-display text-[20px] font-semibold text-ink">{product.title}</h3>

      {/* 价格 */}
      <div className="mt-5 flex items-baseline gap-2">
        <span className="font-display text-3xl font-semibold text-ink tnum">
          {formatPrice(product.priceCents)}
        </span>
        {product.originalPriceCents && product.originalPriceCents > product.priceCents && (
          <span className="text-[14px] text-ink-faint line-through tnum">
            {formatPrice(product.originalPriceCents)}
          </span>
        )}
        {off > 0 && (
          <span className="rounded-[2px] px-1.5 py-0.5 text-[11px] font-medium text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]">
            省 {off}%
          </span>
        )}
      </div>

      {/* 计量说明 */}
      <p className="mt-2 text-[13px] text-ink-secondary">
        {isSub
          ? `有效期 ${durationLabel(product.durationDays)}`
          : `含 ${product.creditAmount ?? 0} 次${creditLabel(product.creditType)}`}
      </p>

      {/* 权益点 */}
      <ul className="mt-6 flex flex-1 flex-col gap-3">
        {benefits.map((b) => (
          <li key={b} className="flex items-start gap-2.5 text-[14px] leading-relaxed text-ink-secondary">
            <Check />
            <span>{b}</span>
          </li>
        ))}
      </ul>

      {/* 购买 */}
      <button
        type="button"
        disabled={busy}
        onClick={() => onBuy(product)}
        className={[
          "mt-8 min-h-[44px] w-full rounded-[6px] px-6 py-3 text-[15px] font-medium transition-colors disabled:opacity-40",
          featured
            ? "glow-gold bg-gold text-[#161206] hover:bg-gold-bright"
            : "bg-transparent text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)] hover:bg-[var(--gold-glow)]",
        ].join(" ")}
      >
        {busy ? "处理中…" : isSub ? "立即开通" : "立即购买"}
      </button>
    </div>
  );
}
