import type { JudgeSection } from "@/lib/divination";

/**
 * 断语分节渲染(四线通用):金字眉题 + 楷体正文,对标紫微多维断语的规格。
 * 「逐爻细览」等多行节按行拆分呈现;无分节时渲染 null(旧档案兼容)。
 */
export function JudgeSections({ sections }: { sections?: JudgeSection[] }) {
  if (!sections || sections.length === 0) return null;
  return (
    <div className="mt-5 flex flex-col gap-4 border-t border-line pt-4">
      {sections.map((s) => (
        <section key={s.key}>
          <p className="text-[12px] font-medium tracking-[0.24em] text-gold">{s.title}</p>
          {s.text.includes("\n") ? (
            <div className="mt-1.5 flex flex-col gap-1">
              {s.text.split("\n").map((line, i) => (
                <p key={i} className="tnum font-reading text-[14px] leading-[1.9] text-ink-secondary">
                  {line}
                </p>
              ))}
            </div>
          ) : (
            <p className="mt-1.5 font-reading text-[15px] leading-[1.9] text-ink-secondary">{s.text}</p>
          )}
        </section>
      ))}
    </div>
  );
}
