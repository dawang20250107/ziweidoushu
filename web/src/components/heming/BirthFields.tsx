"use client";

import { useEffect, useState } from "react";
import type { BirthInfo, Gender } from "@/lib/types";
import { HOUR_NAMES } from "@/lib/types";
import { currentUser, AUTH_EVENT } from "@/lib/auth";
import { listProfiles, type Profile, type BirthRequest } from "@/lib/profiles";

const fieldCls =
  "rounded-[6px] bg-bg px-3 py-2 text-[15px] text-ink shadow-[inset_0_0_0_1px_var(--line)] focus:shadow-[inset_0_0_0_1px_var(--gold-dim)] outline-none transition-shadow";

/** 合盘单方生辰值(受控)。 */
export interface BirthValue {
  name: string;
  date: string; // yyyy-mm-dd
  hour: number;
  gender: Gender;
}

export function toBirthInfo(v: BirthValue): BirthInfo | null {
  const [y, m, d] = v.date.split("-").map(Number);
  if (!y || !m || !d || y < 1900 || y > 2100) return null;
  return { year: y, month: m, day: d, hour: v.hour, gender: v.gender, name: v.name || undefined };
}

export function fromBirthRequest(b: BirthRequest): BirthValue {
  return {
    name: b.name ?? "",
    date: `${b.year}-${String(b.month).padStart(2, "0")}-${String(b.day).padStart(2, "0")}`,
    hour: b.hour,
    gender: b.gender,
  };
}

/** 合盘一方的生辰输入卡(无提交按钮,受控;登录后可从档案一键填入)。 */
export function BirthFields({
  title, value, onChange,
}: {
  title: string;
  value: BirthValue;
  onChange: (v: BirthValue) => void;
}) {
  const [profiles, setProfiles] = useState<Profile[]>([]);

  // 登录后拉档案供快速填入(未登录静默跳过)
  useEffect(() => {
    let cancelled = false;
    const load = () => {
      if (!currentUser()) {
        setProfiles([]);
        return;
      }
      listProfiles()
        .then((r) => {
          if (!cancelled) setProfiles(r.profiles);
        })
        .catch(() => {});
    };
    load();
    window.addEventListener(AUTH_EVENT, load);
    return () => {
      cancelled = true;
      window.removeEventListener(AUTH_EVENT, load);
    };
  }, []);

  return (
    <div className="flex-1 rounded-[10px] bg-bg-raised p-5 shadow-[0_0_0_1px_var(--line)] md:p-6">
      <div className="mb-4 flex items-center justify-between gap-2">
        <span className="font-display text-[15px] font-medium text-gold">{title}</span>
        {profiles.length > 0 && (
          <select
            value=""
            onChange={(e) => {
              const p = profiles.find((x) => x.id === e.target.value);
              if (p) onChange(fromBirthRequest(p.birthInput));
            }}
            className="max-w-36 rounded-[4px] bg-bg px-2 py-1 text-[12px] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] outline-none"
            aria-label="从档案填入"
          >
            <option value="">从档案填入…</option>
            {profiles.map((p) => (
              <option key={p.id} value={p.id}>
                {p.label}
              </option>
            ))}
          </select>
        )}
      </div>

      <div className="grid grid-cols-2 gap-3">
        <label className="flex flex-col gap-1">
          <span className="text-[12px] text-ink-faint">称呼(可选)</span>
          <input
            value={value.name}
            onChange={(e) => onChange({ ...value, name: e.target.value })}
            placeholder={title}
            maxLength={12}
            className={fieldCls}
          />
        </label>
        <label className="flex flex-col gap-1">
          <span className="text-[12px] text-ink-faint">公历生日</span>
          <input
            type="date"
            value={value.date}
            min="1900-01-01"
            max="2100-12-31"
            required
            onChange={(e) => onChange({ ...value, date: e.target.value })}
            className={`${fieldCls} tnum`}
          />
        </label>
        <label className="flex flex-col gap-1">
          <span className="text-[12px] text-ink-faint">时辰</span>
          <select
            value={value.hour}
            onChange={(e) => onChange({ ...value, hour: Number(e.target.value) })}
            className={fieldCls}
          >
            {HOUR_NAMES.map((label, idx) => (
              <option key={idx} value={idx}>
                {label}
              </option>
            ))}
          </select>
        </label>
        <div className="flex flex-col gap-1">
          <span className="text-[12px] text-ink-faint">性别</span>
          <div className="flex overflow-hidden rounded-[6px] shadow-[inset_0_0_0_1px_var(--line)]">
            {(["male", "female"] as const).map((g) => (
              <button
                key={g}
                type="button"
                aria-pressed={value.gender === g}
                onClick={() => onChange({ ...value, gender: g })}
                className={[
                  "flex-1 px-4 py-2 text-[14px] transition-colors",
                  value.gender === g
                    ? "bg-gold font-medium text-[#161206]"
                    : "bg-bg text-ink-secondary hover:text-ink",
                ].join(" ")}
              >
                {g === "male" ? "男" : "女"}
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
