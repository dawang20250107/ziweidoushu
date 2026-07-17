"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { AUTH_EVENT, currentUser, logout } from "@/lib/auth";
import {
  BillingError,
  fetchEntitlements,
  fetchMe,
  fetchOrders,
  fetchProducts,
  type EntitlementsResponse,
  type MeUser,
  type Order,
  type Product,
} from "@/lib/billing";
import { AccountUserCard } from "@/components/billing/AccountUserCard";
import { EntitlementsPanel } from "@/components/billing/EntitlementsPanel";
import { OrderList } from "@/components/billing/OrderList";
import { AccountBlockSkeleton } from "@/components/billing/skeletons";

type Status = "loading" | "unauth" | "error" | "ready";

/**
 * 账户页:用户卡 + 权益区 + 订单区。
 * 未登录居中引导登录;客户端取数,三态齐全。
 */
export default function AccountPage() {
  const router = useRouter();
  const [status, setStatus] = useState<Status>("loading");
  const [user, setUser] = useState<MeUser | null>(null);
  const [ent, setEnt] = useState<EntitlementsResponse | null>(null);
  const [orders, setOrders] = useState<Order[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!currentUser()) {
      setStatus("unauth");
      return;
    }
    let cancelled = false;
    setStatus("loading");
    Promise.all([fetchMe(), fetchEntitlements(), fetchOrders()])
      .then(([me, entResp, ordResp]) => {
        if (cancelled) return;
        setUser(me.user);
        setEnt(entResp);
        setOrders(ordResp.orders);
        setStatus("ready");
      })
      .catch((e) => {
        if (cancelled) return;
        if (e instanceof BillingError && e.status === 401) {
          setStatus("unauth");
        } else {
          setError(e instanceof BillingError ? e.message : "账户信息加载失败,请稍后重试");
          setStatus("error");
        }
      });
    // 产品名映射(best-effort,失败不阻断)
    fetchProducts()
      .then((r) => {
        if (!cancelled) setProducts(r.products);
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, []);

  // 感知外部登出(如顶栏退出),同步为未登录
  useEffect(() => {
    const sync = () => {
      if (!currentUser()) setStatus("unauth");
    };
    window.addEventListener(AUTH_EVENT, sync);
    return () => window.removeEventListener(AUTH_EVENT, sync);
  }, []);

  const titleOf = useMemo(() => {
    const map = new Map(products.map((p) => [p.id, p.title]));
    return (productId: string) => map.get(productId);
  }, [products]);

  async function handleLogout() {
    await logout();
    router.push("/");
  }

  return (
    <div className="mx-auto max-w-3xl px-5 py-14 md:py-24">
      <header className="mb-12 md:mb-14">
        <p className="text-[12px] font-medium tracking-[0.24em] text-gold">账户 · 观星台</p>
        <h1 className="mt-3 font-display text-[31px] font-semibold text-ink sm:text-[39px]">我的账户</h1>
      </header>

      {status === "loading" && (
        <div className="flex flex-col gap-6 md:gap-8">
          <AccountBlockSkeleton lines={2} />
          <AccountBlockSkeleton lines={3} />
          <AccountBlockSkeleton lines={4} />
        </div>
      )}

      {status === "unauth" && (
        <div className="rounded-[10px] bg-bg-raised px-6 py-20 text-center shadow-[0_0_0_1px_var(--line)]">
          <p className="font-display text-xl font-semibold text-ink">登录后查看账户</p>
          <p className="mt-3 text-[14px] leading-relaxed text-ink-secondary">会员权益、次数余额与订单记录都在这里。</p>
          <Link
            href="/login?next=/account"
            className="glow-gold mt-8 inline-flex min-h-[44px] items-center justify-center rounded-[6px] bg-gold px-6 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
          >
            去登录
          </Link>
        </div>
      )}

      {status === "error" && (
        <div className="rounded-[6px] bg-bg-raised px-4 py-3 text-[14px] text-danger shadow-[0_0_0_1px_var(--danger)]">
          {error}
        </div>
      )}

      {status === "ready" && user && ent && (
        <div className="flex flex-col gap-6 md:gap-8">
          <AccountUserCard user={user} onLogout={handleLogout} />
          <EntitlementsPanel data={ent} />
          <OrderList orders={orders} titleOf={titleOf} />
        </div>
      )}
    </div>
  );
}
