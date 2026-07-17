"use client";

import { useEffect, useRef, useState } from "react";
import { fetchFamousList, fetchFamousChart, ApiError } from "@/lib/api";
import type { Chart, FamousPerson, Pattern } from "@/lib/types";
import { HOUR_NAMES } from "@/lib/types";
import { ChartBoard } from "@/components/chart/ChartBoard";
import { DetailPanel } from "@/components/chart/DetailPanel";

interface FamousDetail {
  person: FamousPerson;
  chart: Chart;
  patterns: Pattern[];
}

/** 名人盘库:真实生辰实证盘例,点选即看盘。 */
export default function FamousPage() {
  const [persons, setPersons] = useState<FamousPerson[] | null>(null);
  const [detail, setDetail] = useState<FamousDetail | null>(null);
  const [selectedBranch, setSelectedBranch] = useState<number | null>(null);
  const [loadingID, setLoadingID] = useState("");
  const [error, setError] = useState("");
  const detailRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    fetchFamousList()
      .then((r) => setPersons(r.persons))
      .catch((e) => setError(e instanceof ApiError ? e.message : "名人库加载失败"));
  }, []);

  async function open(p: FamousPerson) {
    setLoadingID(p.id);
    setError("");
    try {
      const d = await fetchFamousChart(p.id);
      setDetail(d);
      setSelectedBranch(null);
      // 移动端列表较长,选中后滚到盘面
      requestAnimationFrame(() => detailRef.current?.scrollIntoView({ behavior: "smooth", block: "start" }));
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "盘例加载失败");
    } finally {
      setLoadingID("");
    }
  }

  const categories = persons ? [...new Set(persons.map((p) => p.category))] : [];

  return (
    <div className="mx-auto max-w-6xl px-4 py-8">
      <p className="text-[12px] tracking-[0.24em] text-gold">盘 例 · 实证研究</p>
      <h1 className="mt-1 font-display text-3xl font-semibold">名人盘库</h1>
      <p className="mt-2 max-w-2xl text-[14px] leading-relaxed text-ink-secondary">
        以公开生辰起盘的名人命例,对照其人生轨迹研习星曜格局的实证表达。
      </p>

      {error && (
        <p className="mt-4 rounded-[6px] bg-bg-raised px-4 py-3 text-[14px] text-danger shadow-[0_0_0_1px_var(--danger)]">
          {error}
        </p>
      )}

      {!persons && !error && (
        <div className="mt-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }, (_, i) => (
            <div key={i} className="h-32 animate-pulse rounded-[10px] bg-bg-raised" />
          ))}
        </div>
      )}

      {persons && (
        <div className="mt-6 flex flex-col gap-6">
          {categories.map((cat) => (
            <section key={cat}>
              <h2 className="mb-2 text-[13px] tracking-[0.12em] text-ink-faint">{cat}</h2>
              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {persons.filter((p) => p.category === cat).map((p) => {
                  const active = detail?.person.id === p.id;
                  return (
                    <button
                      key={p.id}
                      type="button"
                      onClick={() => open(p)}
                      aria-pressed={active}
                      className={[
                        "flex flex-col rounded-[10px] bg-bg-raised p-4 text-left transition-shadow",
                        active
                          ? "shadow-[0_0_0_2px_var(--gold)]"
                          : "shadow-[0_0_0_1px_var(--line)] hover:shadow-[0_0_0_1px_var(--gold-dim)]",
                      ].join(" ")}
                    >
                      <div className="flex items-baseline justify-between gap-2">
                        <span className="font-display text-[17px] font-semibold">{p.name}</span>
                        <span className="tnum text-[12px] text-ink-faint">
                          {p.year} · {HOUR_NAMES[p.hour]}
                        </span>
                      </div>
                      <p className="mt-1 text-[13px] text-ink-secondary">{p.description}</p>
                      <p className="mt-2 line-clamp-2 text-[12px] leading-relaxed text-ink-faint">{p.notable}</p>
                      {loadingID === p.id && <span className="mt-2 text-[12px] text-gold">起盘中…</span>}
                    </button>
                  );
                })}
              </div>
            </section>
          ))}
        </div>
      )}

      {detail && (
        <div ref={detailRef} className="mt-10 scroll-mt-20">
          <div className="mb-3 flex flex-wrap items-baseline gap-x-3 gap-y-1">
            <h2 className="font-display text-2xl font-semibold">{detail.person.name} 命盘</h2>
            <span className="text-[13px] text-ink-secondary">
              {detail.person.year}-{detail.person.month}-{detail.person.day} {HOUR_NAMES[detail.person.hour]} ·{" "}
              {detail.person.gender === "male" ? "男" : "女"}命
            </span>
          </div>
          <p className="mb-4 max-w-3xl text-[13px] leading-relaxed text-ink-faint">{detail.person.notable}</p>
          <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_320px]">
            <div className="overflow-x-auto">
              <div className="min-w-[640px]">
                <ChartBoard
                  chart={detail.chart}
                  density="pro"
                  selectedBranch={selectedBranch}
                  onSelectBranch={setSelectedBranch}
                />
              </div>
            </div>
            <DetailPanel chart={detail.chart} patterns={detail.patterns} selectedBranch={selectedBranch} />
          </div>
        </div>
      )}
    </div>
  );
}
