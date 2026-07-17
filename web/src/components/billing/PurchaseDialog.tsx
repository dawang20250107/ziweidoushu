"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import {
  BillingError,
  createOrder,
  creditLabel,
  devPayOrder,
  durationLabel,
  formatPrice,
  TIER_LABEL,
  type Order,
  type Product,
} from "@/lib/billing";

type Step = "coming_soon" | "creating" | "confirm" | "paying" | "success" | "error";

/** 关闭图标(线性单色)。 */
function CloseIcon() {
  return (
    <svg viewBox="0 0 16 16" width="16" height="16" fill="none" aria-hidden>
      <path d="M4 4l8 8M12 4l-8 8" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
    </svg>
  );
}

/** 成功对勾(线性单色,金)。 */
function SuccessMark() {
  return (
    <svg viewBox="0 0 44 44" width="44" height="44" fill="none" className="text-gold" aria-hidden>
      <circle cx="22" cy="22" r="20" stroke="currentColor" strokeWidth="1.5" opacity="0.4" />
      <path
        d="M13 22.5l6 6 12-13"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

/** 购买该产品到账后的权益文案。 */
function fulfilledSummary(product: Product): string {
  if (product.kind === "subscription") {
    return `${TIER_LABEL[product.tier ?? "pro"]} 会员已开通,有效期 ${durationLabel(product.durationDays)}。`;
  }
  return `已到账 ${product.creditAmount ?? 0} 次${creditLabel(product.creditType)}。`;
}

/**
 * 购买弹层:承接下单 → 模拟支付 → 成功态的完整流程。
 * - devPayEnabled=false:直接展示「支付渠道接入中」,不建单。
 * - devPayEnabled=true:建单 → 确认层「开发环境模拟支付」→ dev-pay → 成功态。
 * 每次选中新产品时由父级以 key 重挂,状态自动重置。
 */
export function PurchaseDialog({
  product,
  devPayEnabled,
  onClose,
}: {
  product: Product;
  devPayEnabled: boolean;
  onClose: () => void;
}) {
  const [step, setStep] = useState<Step>(devPayEnabled ? "creating" : "coming_soon");
  const [order, setOrder] = useState<Order | null>(null);
  const [error, setError] = useState("");

  // 建单(仅在渠道可用时)
  useEffect(() => {
    if (!devPayEnabled) return;
    let cancelled = false;
    createOrder(product.id)
      .then((r) => {
        if (cancelled) return;
        setOrder(r.order);
        // 渠道以下单响应为准:若服务端此刻关闭渠道,退回接入中提示
        setStep(r.devPayEnabled ? "confirm" : "coming_soon");
      })
      .catch((e) => {
        if (cancelled) return;
        setError(e instanceof BillingError ? e.message : "下单失败,请稍后重试");
        setStep("error");
      });
    return () => {
      cancelled = true;
    };
  }, [product.id, devPayEnabled]);

  // Esc 关闭
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  const pay = useCallback(async () => {
    if (!order) return;
    setStep("paying");
    setError("");
    try {
      const r = await devPayOrder(order.id);
      setOrder(r.order);
      setStep("success");
    } catch (e) {
      if (e instanceof BillingError && e.code === "dev_pay_disabled") {
        setError("模拟支付渠道未启用。微信 / 支付宝支付正在接入中。");
        setStep("coming_soon");
      } else {
        setError(e instanceof BillingError ? e.message : "支付失败,请稍后重试");
        setStep("error");
      }
    }
  }, [order]);

  // 支付进行中禁止点遮罩关闭,避免误中断
  const dismissable = step !== "paying" && step !== "creating";

  return (
    <div
      className="fixed inset-0 z-50 flex items-end justify-center bg-bg/80 p-4 backdrop-blur-sm sm:items-center"
      onClick={dismissable ? onClose : undefined}
      role="presentation"
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-label="购买"
        onClick={(e) => e.stopPropagation()}
        className="w-full max-w-md rounded-[10px] bg-bg-overlay p-6 shadow-[0_0_0_1px_var(--line-strong),0_8px_24px_rgba(0,0,0,0.3)]"
      >
        {/* 头部 */}
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-[12px] tracking-[0.16em] text-gold">
              {product.kind === "subscription" ? "订阅开通" : "次卡购买"}
            </p>
            <h2 className="mt-1 font-display text-xl font-semibold text-ink">{product.title}</h2>
          </div>
          {dismissable && (
            <button
              type="button"
              onClick={onClose}
              aria-label="关闭"
              className="-mr-1 -mt-1 grid h-9 w-9 place-items-center rounded-[6px] text-ink-faint transition-colors hover:bg-bg-raised hover:text-ink"
            >
              <CloseIcon />
            </button>
          )}
        </div>

        <div className="mt-5">
          {step === "creating" && (
            <p className="py-6 text-center text-[14px] text-ink-secondary" aria-live="polite">
              正在创建订单…
            </p>
          )}

          {step === "confirm" && (
            <div className="flex flex-col gap-4">
              <div className="rounded-[6px] bg-bg-raised px-4 py-3 shadow-[inset_0_0_0_1px_var(--line)]">
                <div className="flex items-baseline justify-between">
                  <span className="text-[13px] text-ink-secondary">应付金额</span>
                  <span className="font-display text-2xl font-semibold text-ink tnum">
                    {formatPrice(order?.amountCents ?? product.priceCents)}
                  </span>
                </div>
              </div>
              <p className="rounded-[4px] bg-bg-raised px-3 py-2 text-[12px] text-warn shadow-[inset_0_0_0_1px_var(--line)]">
                开发环境模拟支付:点击「确认支付」即刻到账,不产生真实扣款(接入微信 /
                支付宝后此层替换为正式收银台)。
              </p>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={onClose}
                  className="flex-1 rounded-[6px] px-6 py-3 text-[15px] font-medium text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] transition-colors hover:text-ink"
                >
                  取消
                </button>
                <button
                  type="button"
                  onClick={pay}
                  className="flex-1 rounded-[6px] bg-gold px-6 py-3 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
                >
                  确认支付
                </button>
              </div>
            </div>
          )}

          {step === "paying" && (
            <p className="py-6 text-center text-[14px] text-ink-secondary" aria-live="polite">
              支付处理中…
            </p>
          )}

          {step === "success" && (
            <div className="flex flex-col items-center gap-3 text-center">
              <SuccessMark />
              <p className="font-display text-lg font-semibold text-ink">支付成功</p>
              <p className="text-[14px] text-ink-secondary">{fulfilledSummary(product)}</p>
              <div className="mt-3 flex w-full gap-3">
                <button
                  type="button"
                  onClick={onClose}
                  className="flex-1 rounded-[6px] px-6 py-3 text-[15px] font-medium text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] transition-colors hover:text-ink"
                >
                  继续浏览
                </button>
                <Link
                  href="/account"
                  className="flex-1 rounded-[6px] bg-gold px-6 py-3 text-center text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
                >
                  查看权益
                </Link>
              </div>
            </div>
          )}

          {step === "coming_soon" && (
            <div className="flex flex-col gap-4">
              <p className="text-[14px] leading-relaxed text-ink-secondary">
                {error || "微信 / 支付宝支付正在接入中,敬请期待。"}
              </p>
              {order && (
                <p className="text-[13px] text-ink-faint">
                  已为你保留待支付订单,可在
                  <Link href="/account" className="mx-1 text-gold-dim hover:text-gold">
                    账户页
                  </Link>
                  查看。
                </p>
              )}
              <button
                type="button"
                onClick={onClose}
                className="rounded-[6px] px-6 py-3 text-[15px] font-medium text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] transition-colors hover:text-ink"
              >
                知道了
              </button>
            </div>
          )}

          {step === "error" && (
            <div className="flex flex-col gap-4">
              <p className="text-[14px] leading-relaxed text-danger">{error || "出错了,请稍后重试。"}</p>
              <button
                type="button"
                onClick={onClose}
                className="rounded-[6px] px-6 py-3 text-[15px] font-medium text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] transition-colors hover:text-ink"
              >
                关闭
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
