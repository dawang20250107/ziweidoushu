"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { currentUser } from "@/lib/auth";
import { BillingError, fetchProducts, type Product } from "@/lib/billing";
import { PricingCard } from "@/components/billing/PricingCard";
import { PurchaseDialog } from "@/components/billing/PurchaseDialog";
import { PricingCardSkeleton } from "@/components/billing/skeletons";

/**
 * 定价页:订阅方案 + 次卡两组。
 * Pro 年度卡视觉主推;未登录购买跳登录,已登录进购买弹层(开发环境模拟支付)。
 */
export default function PricingPage() {
  const router = useRouter();
  const [products, setProducts] = useState<Product[]>([]);
  const [devPayEnabled, setDevPayEnabled] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [selected, setSelected] = useState<Product | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetchProducts()
      .then((r) => {
        if (cancelled) return;
        setProducts(r.products);
        setDevPayEnabled(r.devPayEnabled);
      })
      .catch((e) => {
        if (cancelled) return;
        setError(e instanceof BillingError ? e.message : "方案加载失败,请稍后重试");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const subscriptions = useMemo(
    () => products.filter((p) => p.kind === "subscription"),
    [products],
  );
  const credits = useMemo(() => products.filter((p) => p.kind === "credits"), [products]);

  // 主推卡:Pro 层里时长最长的一款(年度)
  const featuredId = useMemo(() => {
    const proSubs = subscriptions.filter((p) => p.tier === "pro" && p.durationDays);
    if (!proSubs.length) return null;
    return proSubs.reduce((a, b) => ((b.durationDays ?? 0) > (a.durationDays ?? 0) ? b : a)).id;
  }, [subscriptions]);

  function handleBuy(product: Product) {
    if (!currentUser()) {
      router.push("/login?next=/pricing");
      return;
    }
    setSelected(product);
  }

  return (
    <div className="mx-auto max-w-5xl px-5 py-8 md:py-12">
      <header className="mb-8 text-center">
        <p className="text-[12px] tracking-[0.24em] text-gold">定价 · 观星台</p>
        <h1 className="mt-2 font-display text-3xl font-semibold text-ink">选择你的方案</h1>
        <p className="mx-auto mt-2 max-w-xl text-ink-secondary">
          从排盘到深度报告,按需订阅或按次购买。数据始终归你所有。
        </p>
      </header>

      {error && (
        <p className="mb-6 rounded-[6px] bg-bg-raised px-4 py-3 text-[14px] text-danger shadow-[0_0_0_1px_var(--danger)]">
          {error}
        </p>
      )}

      {loading && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <PricingCardSkeleton key={i} />
          ))}
        </div>
      )}

      {!loading && !error && products.length === 0 && (
        <p className="py-16 text-center text-ink-faint">暂无可购买的方案。</p>
      )}

      {!loading && subscriptions.length > 0 && (
        <section className="mb-12">
          <div className="mb-5">
            <h2 className="font-display text-xl font-semibold text-ink">订阅方案</h2>
            <p className="mt-1 text-[13px] text-ink-faint">解锁全部排盘与解读能力,随时续费。</p>
          </div>
          <div className="grid gap-4 pt-2 sm:grid-cols-2 lg:grid-cols-3">
            {subscriptions.map((p) => (
              <PricingCard
                key={p.id}
                product={p}
                featured={p.id === featuredId}
                busy={selected?.id === p.id}
                onBuy={handleBuy}
              />
            ))}
          </div>
        </section>
      )}

      {!loading && credits.length > 0 && (
        <section>
          <div className="mb-5">
            <h2 className="font-display text-xl font-semibold text-ink">次卡</h2>
            <p className="mt-1 text-[13px] text-ink-faint">无需订阅,按需购买深度报告生成次数。</p>
          </div>
          <div className="grid gap-4 pt-2 sm:grid-cols-2 lg:grid-cols-3">
            {credits.map((p) => (
              <PricingCard
                key={p.id}
                product={p}
                busy={selected?.id === p.id}
                onBuy={handleBuy}
              />
            ))}
          </div>
        </section>
      )}

      {selected && (
        <PurchaseDialog
          key={selected.id}
          product={selected}
          devPayEnabled={devPayEnabled}
          onClose={() => setSelected(null)}
        />
      )}
    </div>
  );
}
