/**
 * 报告正文极简渲染(长者叙述体):\n\n 分段;「#/##/###」开头行作小标题,
 * 「> 」开头段作古籍引文块。纯文本按段落排版,不引入 Markdown 依赖。
 */
export function ReportText({ text }: { text: string }) {
  const blocks = text
    .split(/\n{2,}/)
    .map((b) => b.trim())
    .filter(Boolean);

  return (
    <div className="prose-elder flex flex-col gap-5">
      {blocks.map((block, i) => {
        if (block.startsWith("### ")) {
          return (
            <h3 key={i} className="mt-1 font-display text-[17px] font-semibold text-ink">
              {block.slice(4).trim()}
            </h3>
          );
        }
        if (block.startsWith("## ")) {
          return (
            <h2 key={i} className="mt-1 font-display text-[20px] font-semibold text-ink">
              {block.slice(3).trim()}
            </h2>
          );
        }
        if (block.startsWith("# ")) {
          return (
            <h2 key={i} className="mt-1 font-display text-[25px] font-semibold text-ink">
              {block.slice(2).trim()}
            </h2>
          );
        }
        if (block.startsWith("> ")) {
          return (
            <blockquote
              key={i}
              className="border-l-2 border-gold-dim pl-3 font-reading text-[15px] leading-[1.9] text-ink-secondary"
            >
              {block
                .split("\n")
                .map((l) => l.replace(/^>\s?/, ""))
                .join("\n")}
            </blockquote>
          );
        }
        return (
          <p key={i} className="whitespace-pre-wrap font-reading text-[16px] leading-[1.9] text-ink-secondary">
            {block}
          </p>
        );
      })}
    </div>
  );
}
