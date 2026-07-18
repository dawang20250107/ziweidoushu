"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { AUTH_EVENT, currentUser } from "@/lib/auth";
import {
  createProfile,
  RELATIONS,
  ProfileError,
  type BirthRequest,
  type Relation,
} from "@/lib/profiles";

/**
 * 保存档案按钮(完全自包含,零外部依赖)。
 * 点击:未登录→提示并链接 /login;已登录→弹出小型表单(label/relation/isDefault)→ POST /profiles。
 * 视觉与排盘页工具按钮一致(次级按钮样式)。主线负责把它接进排盘页。
 */
export function SaveProfileButton({
  birth,
  className = "",
}: {
  birth: BirthRequest;
  className?: string;
}) {
  const [signedIn, setSignedIn] = useState(false);
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);

  // 表单态
  const [label, setLabel] = useState("");
  const [relation, setRelation] = useState<Relation>("self");
  const [isDefault, setIsDefault] = useState(false);
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState("");
  const [limit, setLimit] = useState(false); // profile_limit 达上限

  useEffect(() => {
    const sync = () => setSignedIn(!!currentUser());
    sync();
    window.addEventListener(AUTH_EVENT, sync);
    return () => window.removeEventListener(AUTH_EVENT, sync);
  }, []);

  // 点击外部关闭浮层
  useEffect(() => {
    if (!open) return;
    function onDown(e: MouseEvent) {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) setOpen(false);
    }
    window.addEventListener("mousedown", onDown);
    return () => window.removeEventListener("mousedown", onDown);
  }, [open]);

  function toggle() {
    if (open) {
      setOpen(false);
      return;
    }
    // 每次打开重置表单,label 默认取 birth.name 或「我的命盘」
    setLabel(birth.name?.trim() || "我的命盘");
    setRelation("self");
    setIsDefault(false);
    setError("");
    setLimit(false);
    setDone(false);
    setOpen(true);
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    const name = label.trim();
    if (!name) {
      setError("请填写档案名称");
      return;
    }
    setBusy(true);
    setError("");
    setLimit(false);
    try {
      await createProfile({ ...birth, label: name, relation, isDefault });
      setDone(true);
      // 成功打勾反馈后自动收起
      window.setTimeout(() => setOpen(false), 1200);
    } catch (err) {
      if (err instanceof ProfileError && err.code === "profile_limit") {
        setLimit(true);
      } else {
        setError(err instanceof Error ? err.message : "保存失败,请稍后重试");
      }
    } finally {
      setBusy(false);
    }
  }

  const triggerCls =
    "inline-flex min-h-[44px] items-center gap-1.5 rounded-[6px] bg-bg-raised px-4 py-2 text-[14px] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] transition-colors hover:text-gold hover:shadow-[inset_0_0_0_1px_var(--gold-dim)] disabled:cursor-not-allowed disabled:opacity-50";

  return (
    <div ref={rootRef} className={`relative inline-block ${className}`}>
      <button type="button" onClick={toggle} aria-expanded={open} className={triggerCls}>
        <BookmarkIcon />
        保存档案
      </button>

      {open && (
        <div className="absolute left-0 top-full z-50 mt-2 w-72 max-w-[80vw] rounded-[10px] bg-bg-overlay p-4 shadow-[0_0_0_1px_var(--line-strong),0_8px_24px_rgba(0,0,0,0.35)]">
          {!signedIn ? (
            <div className="text-center">
              <p className="text-[14px] text-ink-secondary">登录后即可保存命盘档案。</p>
              <Link
                href="/login?next=/chart"
                className="mt-3 inline-flex min-h-[44px] items-center justify-center rounded-[6px] bg-gold px-5 py-2 text-[14px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
              >
                去登录
              </Link>
            </div>
          ) : done ? (
            <div className="flex items-center justify-center gap-2 py-4 text-[14px] text-ok">
              <CheckIcon />
              已保存到档案库
            </div>
          ) : limit ? (
            <div className="text-center">
              <p className="text-[14px] text-ink">档案数量已达上限</p>
              <p className="mt-1.5 text-[13px] leading-relaxed text-ink-secondary">
                免费版最多保存 3 份档案,升级后可保存更多。
              </p>
              <Link
                href="/pricing"
                className="mt-3 inline-flex min-h-[44px] items-center justify-center rounded-[6px] bg-gold px-5 py-2 text-[14px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
              >
                升级解锁
              </Link>
            </div>
          ) : (
            <form onSubmit={submit} className="flex flex-col gap-3">
              <label className="flex flex-col gap-1">
                <span className="text-[12px] text-ink-faint">档案名称</span>
                <input
                  value={label}
                  onChange={(e) => setLabel(e.target.value)}
                  maxLength={20}
                  autoFocus
                  className="rounded-[6px] bg-bg px-3 py-2 text-[15px] text-ink shadow-[inset_0_0_0_1px_var(--line)] outline-none transition-shadow focus:shadow-[inset_0_0_0_1px_var(--gold-dim)]"
                />
              </label>
              <label className="flex flex-col gap-1">
                <span className="text-[12px] text-ink-faint">关系</span>
                <select
                  value={relation}
                  onChange={(e) => setRelation(e.target.value as Relation)}
                  className="rounded-[6px] bg-bg px-3 py-2 text-[15px] text-ink shadow-[inset_0_0_0_1px_var(--line)] outline-none transition-shadow focus:shadow-[inset_0_0_0_1px_var(--gold-dim)]"
                >
                  {RELATIONS.map((r) => (
                    <option key={r.key} value={r.key}>
                      {r.label}
                    </option>
                  ))}
                </select>
              </label>
              <label className="flex min-h-[44px] cursor-pointer items-center gap-2 text-[14px] text-ink-secondary">
                <input
                  type="checkbox"
                  checked={isDefault}
                  onChange={(e) => setIsDefault(e.target.checked)}
                  className="h-4 w-4 accent-gold"
                />
                设为默认档案
              </label>

              {error && <p className="text-[13px] text-danger">{error}</p>}

              <div className="mt-1 flex items-center justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setOpen(false)}
                  className="min-h-[44px] rounded-[6px] px-3 py-2 text-[14px] text-ink-secondary transition-colors hover:text-ink"
                >
                  取消
                </button>
                <button
                  type="submit"
                  disabled={busy}
                  className="min-h-[44px] rounded-[6px] bg-gold px-5 py-2 text-[14px] font-medium text-[#161206] transition-colors hover:bg-gold-bright disabled:opacity-50"
                >
                  {busy ? "保存中…" : "保存"}
                </button>
              </div>
            </form>
          )}
        </div>
      )}
    </div>
  );
}

function BookmarkIcon() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" aria-hidden>
      <path d="M6 3h12a1 1 0 0 1 1 1v17l-7-4-7 4V4a1 1 0 0 1 1-1z" strokeLinejoin="round" />
    </svg>
  );
}

function CheckIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden>
      <path d="M5 13l4 4L19 7" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
