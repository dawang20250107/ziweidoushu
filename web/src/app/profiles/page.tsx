"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { AUTH_EVENT, currentUser } from "@/lib/auth";
import {
  listProfiles,
  deleteProfile,
  setDefaultProfile,
  writeRecentBirth,
  ProfileError,
  type Profile,
} from "@/lib/profiles";
import { ProfileCard } from "@/components/profiles/ProfileCard";
import { ProfileSkeleton } from "@/components/profiles/ProfileSkeleton";

/** 命盘档案库:档案卡片网格 + 载入排盘 / 设为默认 / 删除。 */
export default function ProfilesPage() {
  const router = useRouter();
  const [signedIn, setSignedIn] = useState<boolean | null>(null); // null=未判定
  const [profiles, setProfiles] = useState<Profile[]>([]);
  const [limit, setLimit] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [busyId, setBusyId] = useState<string | null>(null); // 正在操作的档案

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const r = await listProfiles();
      setProfiles(r.profiles);
      setLimit(r.limit);
    } catch (e) {
      setError(e instanceof ProfileError ? e.message : "档案加载失败,请稍后重试");
    } finally {
      setLoading(false);
    }
  }, []);

  // 登录态感知:登录后加载档案
  useEffect(() => {
    const sync = () => {
      const ok = !!currentUser();
      setSignedIn(ok);
      if (ok) void load();
      else setLoading(false);
    };
    sync();
    window.addEventListener(AUTH_EVENT, sync);
    return () => window.removeEventListener(AUTH_EVENT, sync);
  }, [load]);

  // 载入排盘:写回 ziwei-birth 后跳排盘页重排
  const handleLoad = useCallback(
    (p: Profile) => {
      writeRecentBirth(p.birthInput);
      router.push("/chart");
    },
    [router],
  );

  const handleSetDefault = useCallback(async (p: Profile) => {
    setBusyId(p.id);
    setError("");
    try {
      await setDefaultProfile(p.id);
      setProfiles((prev) => prev.map((x) => ({ ...x, isDefault: x.id === p.id })));
    } catch (e) {
      setError(e instanceof ProfileError ? e.message : "设置默认失败,请稍后重试");
    } finally {
      setBusyId(null);
    }
  }, []);

  const handleDelete = useCallback(async (p: Profile) => {
    setBusyId(p.id);
    setError("");
    try {
      await deleteProfile(p.id);
      setProfiles((prev) => prev.filter((x) => x.id !== p.id));
    } catch (e) {
      setError(e instanceof ProfileError ? e.message : "删除失败,请稍后重试");
    } finally {
      setBusyId(null);
    }
  }, []);

  // 未登录:居中提示 + 去登录
  if (signedIn === false) {
    return (
      <div className="mx-auto max-w-md px-4 py-24 md:py-32">
        <div className="rounded-[10px] bg-bg-raised px-6 py-16 text-center shadow-[0_0_0_1px_var(--line)]">
          <p className="font-display text-xl font-semibold text-ink">登录后查看命盘档案库</p>
          <p className="mt-3 text-[14px] leading-relaxed text-ink-secondary">
            档案库为你保存命主生辰,随时一键载入排盘。
          </p>
          <Link
            href="/login?next=/profiles"
            className="glow-gold mt-8 inline-flex min-h-[44px] items-center rounded-[6px] bg-gold px-6 py-2.5 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
          >
            去登录
          </Link>
        </div>
      </div>
    );
  }

  const count = profiles.length;
  const atLimit = limit > 0 && count >= limit;

  return (
    <div className="mx-auto max-w-5xl px-5 py-14 md:py-24">
      <header className="mb-12 md:mb-14">
        <p className="text-[12px] font-medium tracking-[0.24em] text-gold">命盘 · 档案库</p>
        <div className="mt-3 flex flex-wrap items-baseline justify-between gap-3">
          <h1 className="font-display text-[31px] font-semibold text-ink sm:text-[39px]">档案库</h1>
          {!loading && (
            <span className="tnum text-[13px] text-ink-faint">
              {limit > 0 ? `${count}/${limit}` : `${count} 份档案`}
            </span>
          )}
        </div>
        <p className="mt-3 text-[15px] leading-relaxed text-ink-secondary md:text-[16px]">保存命主生辰,随时一键载入排盘。</p>
      </header>

      {/* 免费层达上限:升级横幅 */}
      {!loading && atLimit && (
        <div className="mb-6 flex flex-wrap items-center justify-between gap-3 rounded-[6px] bg-bg-raised px-4 py-3 shadow-[inset_0_0_0_1px_var(--gold-dim)]">
          <p className="text-[14px] text-ink-secondary">
            已达免费版档案上限({limit} 份)。升级后可保存更多命盘。
          </p>
          <Link
            href="/pricing"
            className="inline-flex min-h-[44px] shrink-0 items-center rounded-[6px] bg-gold px-4 py-2 text-[14px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
          >
            升级解锁
          </Link>
        </div>
      )}

      {error && (
        <p className="mb-6 rounded-[6px] bg-bg-raised px-4 py-3 text-[14px] text-danger shadow-[0_0_0_1px_var(--danger)]">
          {error}
        </p>
      )}

      {loading && (
        <div className="grid gap-5 sm:grid-cols-2 md:gap-6">
          {Array.from({ length: 3 }).map((_, i) => (
            <ProfileSkeleton key={i} />
          ))}
        </div>
      )}

      {!loading && count > 0 && (
        <div className="grid gap-5 sm:grid-cols-2 md:gap-6">
          {profiles.map((p) => (
            <ProfileCard
              key={p.id}
              profile={p}
              busy={busyId === p.id}
              onLoad={handleLoad}
              onSetDefault={handleSetDefault}
              onDelete={handleDelete}
            />
          ))}
        </div>
      )}

      {/* 空态 */}
      {!loading && !error && count === 0 && (
        <div className="rounded-[10px] bg-bg-raised px-6 py-20 text-center shadow-[0_0_0_1px_var(--line)]">
          <p className="font-display text-xl font-semibold text-ink">还没有命盘档案</p>
          <p className="mt-3 text-[14px] leading-relaxed text-ink-secondary">
            去排盘,把命主生辰保存为第一份档案,以后一键载入。
          </p>
          <Link
            href="/chart"
            className="glow-gold mt-8 inline-flex min-h-[44px] items-center rounded-[6px] bg-gold px-6 py-2.5 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
          >
            去排盘并保存
          </Link>
        </div>
      )}
    </div>
  );
}
