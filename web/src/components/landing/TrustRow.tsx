/** 信任区:三个可核对的事实。font-display 强调值 + 小字说明。 */
const FACTS = [
  { value: "1566", caption: "例黄金基准,逐字段校验" },
  { value: "70+", caption: "经典格局,自动判定" },
  { value: "《天纪》", caption: "倪海厦体系口径,全程统一" },
] as const;

export function TrustRow() {
  return (
    <section className="mx-auto max-w-6xl px-4 py-16 md:py-24">
      <div
        className="grid gap-px overflow-hidden rounded-[10px] sm:grid-cols-3"
        style={{ boxShadow: "inset 0 0 0 1px var(--line)", background: "var(--line)" }}
      >
        {FACTS.map((f) => (
          <div key={f.caption} className="bg-bg px-6 py-10 text-center md:py-12">
            <div className="tnum font-display text-[39px] font-bold leading-none text-gold md:text-[49px]">
              {f.value}
            </div>
            <p className="mt-4 text-[13px] leading-relaxed text-ink-secondary">{f.caption}</p>
          </div>
        ))}
      </div>
    </section>
  );
}
