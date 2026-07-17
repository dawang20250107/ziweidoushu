"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { fetchChart, fetchHoroscope, ApiError } from "@/lib/api";
import type { BirthInfo, ChartResponse, Horoscope } from "@/lib/types";
import { DENSITY_LABELS, type Density } from "@/lib/chart-helpers";
import { BirthForm } from "@/components/chart/BirthForm";
import { ChartBoard } from "@/components/chart/ChartBoard";
import { DetailPanel } from "@/components/chart/DetailPanel";
import { TimelineBar, type TimelineSelection } from "@/components/chart/TimelineBar";
import { SiZhuPanel } from "@/components/chart/SiZhuPanel";
import { LuopanCast } from "@/components/chart/LuopanCast";
import { SaveProfileButton } from "@/components/profiles/SaveProfileButton";

/** 排盘工作台:盘面 + 运限时间轴 + 宫位详情。 */
export default function ChartPage() {
  const [birth, setBirth] = useState<BirthInfo | null>(null);
  const [data, setData] = useState<ChartResponse | null>(null);
  const [horoscope, setHoroscope] = useState<Horoscope | null>(null);
  const [timeline, setTimeline] = useState<TimelineSelection>({ year: null });
  const [density, setDensity] = useState<Density>("pro");
  const [selectedBranch, setSelectedBranch] = useState<number | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [casting, setCasting] = useState(false); // 罗盘起盘仪式中

  const [initialBirth, setInitialBirth] = useState<BirthInfo | null>(null);

  const runChart = useCallback(async (b: BirthInfo) => {
    setLoading(true);
    setError("");
    setHoroscope(null);
    setTimeline({ year: null });
    setSelectedBranch(null);
    try {
      const resp = await fetchChart(b);
      setBirth(b);
      setData(resp);
      // 跨页契约:问星页(/chat)读取最近一次排盘的生辰
      localStorage.setItem("ziwei-birth", JSON.stringify(b));
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "排盘失败,请稍后重试");
    } finally {
      setLoading(false);
    }
  }, []);

  // 手动排盘:星光击罗盘仪式(≥1.8s);恢复路径与 reduced-motion 直出
  const castChart = useCallback(async (b: BirthInfo) => {
    const reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (reduced) {
      await runChart(b);
      return;
    }
    setCasting(true);
    const minShow = new Promise((r) => setTimeout(r, 1800));
    await Promise.all([runChart(b), minShow]);
    setCasting(false);
  }, [runChart]);

  // 挂载时恢复最近一次排盘生辰并自动出盘(档案「载入排盘」/刷新续排共用 ziwei-birth 契约)
  useEffect(() => {
    try {
      const raw = localStorage.getItem("ziwei-birth");
      if (!raw) return;
      const b = JSON.parse(raw) as BirthInfo;
      if (typeof b?.year === "number" && b.year >= 1900 && b.year <= 2100 && b.gender) {
        setInitialBirth(b);
        void runChart(b);
      }
    } catch {
      // 本地数据损坏则忽略,走空白表单
    }
  }, [runChart]);

  // 时间轴选择 → 拉运限。未下钻的层取稳定默认(7 月 15 日午时,避开农历年界)。
  useEffect(() => {
    if (!birth || timeline.year == null) {
      setHoroscope(null);
      return;
    }
    let cancelled = false;
    fetchHoroscope(birth, {
      year: timeline.year,
      month: timeline.month ?? 7,
      day: timeline.day ?? 15,
      hour: timeline.hour ?? 6,
    })
      .then((resp) => {
        if (!cancelled) setHoroscope(resp.horoscope);
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof ApiError ? e.message : "运限计算失败");
      });
    return () => {
      cancelled = true;
    };
  }, [birth, timeline.year, timeline.month, timeline.day, timeline.hour]);

  // 下钻深度决定盘面叠加层
  const overlayScopes = useMemo(() => {
    const scopes: Array<"decadal" | "yearly" | "monthly" | "daily" | "hourly"> = ["decadal", "yearly"];
    if (timeline.month != null) scopes.push("monthly");
    if (timeline.day != null) scopes.push("daily");
    if (timeline.hour != null) scopes.push("hourly");
    return scopes;
  }, [timeline.month, timeline.day, timeline.hour]);

  return (
    <div className="mx-auto max-w-6xl px-4 py-6">
      <div className="mb-4 flex flex-wrap items-end justify-between gap-3">
        <BirthForm key={initialBirth ? "restored" : "blank"} initial={initialBirth ?? undefined} loading={loading} onSubmit={castChart} />
        {data && birth && (
          <div className="flex flex-wrap items-center gap-2">
            <Link
              href="/famous"
              className="inline-flex min-h-[44px] items-center rounded-[6px] bg-bg-raised px-4 py-2 text-[14px] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] transition-colors hover:text-gold hover:shadow-[inset_0_0_0_1px_var(--gold-dim)]"
            >
              名人盘库
            </Link>
            <SaveProfileButton birth={birth} />
            <div className="flex overflow-hidden rounded-[6px] shadow-[inset_0_0_0_1px_var(--line)]" role="radiogroup" aria-label="显示密度">
              {(Object.keys(DENSITY_LABELS) as Density[]).map((d) => (
                <button
                  key={d}
                  type="button"
                  role="radio"
                  aria-checked={density === d}
                  onClick={() => setDensity(d)}
                  className={[
                    "px-3 py-1.5 text-[13px] transition-colors",
                    density === d
                      ? "bg-[var(--gold-glow)] font-medium text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]"
                      : "bg-bg-raised text-ink-secondary hover:text-ink",
                  ].join(" ")}
                >
                  {DENSITY_LABELS[d]}
                </button>
              ))}
            </div>
          </div>
        )}
      </div>

      {error && (
        <p className="mb-4 rounded-[6px] bg-bg-raised px-4 py-3 text-[14px] text-danger shadow-[0_0_0_1px_var(--danger)]">
          {error}
        </p>
      )}

      {!data && !loading && (
        <div className="rounded-[6px] bg-bg-raised px-6 py-16 text-center text-ink-faint shadow-[0_0_0_1px_var(--line)]">
          <p className="font-display text-lg text-ink-secondary">输入生辰,开始排盘</p>
          <p className="mt-2 text-[13px]">安星算法与 iztro 逐字段对齐,1500+ 黄金基准回归保障。</p>
        </div>
      )}

      {data && (
        <div className="flex flex-col gap-4">
          <TimelineBar chart={data.chart} selection={timeline} horoscope={horoscope} onChange={setTimeline} />
          <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_320px]">
            <div>
              <div className="relative">
                <div className="overflow-x-auto">
                  <div className="min-w-[640px]">
                    <ChartBoard
                      chart={data.chart}
                      density={density}
                      selectedBranch={selectedBranch}
                      onSelectBranch={setSelectedBranch}
                      horoscope={horoscope}
                      overlayScopes={overlayScopes}
                    />
                  </div>
                </div>
                {/* 移动端:右缘渐隐提示盘面可横向滑动 */}
                <div
                  aria-hidden
                  className="pointer-events-none absolute inset-y-0 right-0 w-10 bg-gradient-to-l from-bg to-transparent md:hidden"
                />
              </div>
              <p className="mt-1.5 text-center text-[11px] text-ink-faint md:hidden">左右滑动查看全盘</p>
            </div>
            <DetailPanel chart={data.chart} patterns={data.patterns ?? []} selectedBranch={selectedBranch} />
          </div>

          {/* 四柱视角:八字附加层(可折叠) */}
          {data.chart.siZhu && <SiZhuPanel siZhu={data.chart.siZhu} />}
        </div>
      )}

      {casting && <LuopanCast />}
    </div>
  );
}
