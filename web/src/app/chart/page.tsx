"use client";

import { useCallback, useEffect, useState } from "react";
import { fetchChart, fetchHoroscope, ApiError } from "@/lib/api";
import type { BirthInfo, ChartResponse, Horoscope } from "@/lib/types";
import { DENSITY_LABELS, type Density } from "@/lib/chart-helpers";
import { BirthForm } from "@/components/chart/BirthForm";
import { ChartBoard } from "@/components/chart/ChartBoard";
import { DetailPanel } from "@/components/chart/DetailPanel";
import { TimelineBar, type TimelineSelection } from "@/components/chart/TimelineBar";

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

  // 时间轴选择 → 拉运限(目标日取该年 7 月 15 日午时,避开农历年界)
  useEffect(() => {
    if (!birth || timeline.year == null) {
      setHoroscope(null);
      return;
    }
    let cancelled = false;
    fetchHoroscope(birth, { year: timeline.year, month: 7, day: 15, hour: 6 })
      .then((resp) => {
        if (!cancelled) setHoroscope(resp.horoscope);
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof ApiError ? e.message : "运限计算失败");
      });
    return () => {
      cancelled = true;
    };
  }, [birth, timeline.year]);

  return (
    <div className="mx-auto max-w-6xl px-4 py-6">
      <div className="mb-4 flex flex-wrap items-end justify-between gap-3">
        <BirthForm loading={loading} onSubmit={runChart} />
        {data && (
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
                  density === d ? "bg-gold font-medium text-[#161206]" : "bg-bg-raised text-ink-secondary hover:text-ink",
                ].join(" ")}
              >
                {DENSITY_LABELS[d]}
              </button>
            ))}
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
          <TimelineBar chart={data.chart} selection={timeline} onChange={setTimeline} />
          <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_320px]">
            <div className="overflow-x-auto">
              <div className="min-w-[640px]">
                <ChartBoard
                  chart={data.chart}
                  density={density}
                  selectedBranch={selectedBranch}
                  onSelectBranch={setSelectedBranch}
                  horoscope={horoscope}
                />
              </div>
            </div>
            <DetailPanel chart={data.chart} patterns={data.patterns ?? []} selectedBranch={selectedBranch} />
          </div>
        </div>
      )}
    </div>
  );
}
