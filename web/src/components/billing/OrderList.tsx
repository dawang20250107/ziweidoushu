import { formatDateTime, formatPrice, type Order } from "@/lib/billing";
import { OrderStatusBadge } from "@/components/billing/badges";

/**
 * 订单区:近 50 条。每行标题(产品名,取不到则回退单号)、状态徽标、金额、时间。
 */
export function OrderList({
  orders,
  titleOf,
}: {
  orders: Order[];
  titleOf: (productId: string) => string | undefined;
}) {
  return (
    <div className="rounded-[10px] bg-bg-raised p-6 shadow-[0_0_0_1px_var(--line)]">
      <h2 className="font-display text-lg font-semibold text-ink">订单记录</h2>

      {orders.length === 0 ? (
        <p className="mt-4 rounded-[6px] px-4 py-8 text-center text-[14px] text-ink-faint shadow-[inset_0_0_0_1px_var(--line)]">
          还没有订单。
        </p>
      ) : (
        <ul className="mt-2 divide-y divide-line">
          {orders.map((o) => (
            <li
              key={o.id}
              className="flex flex-wrap items-center justify-between gap-x-4 gap-y-1.5 py-4"
            >
              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <span className="truncate text-[15px] text-ink">
                    {titleOf(o.productId) ?? "订单"}
                  </span>
                  <OrderStatusBadge status={o.status} />
                </div>
                <p className="mt-0.5 text-[12px] text-ink-faint tnum">
                  单号 {o.orderNo} · {formatDateTime(o.createdAt)}
                </p>
              </div>
              <span className="font-display text-[16px] text-ink tnum">
                {formatPrice(o.amountCents)}
              </span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
