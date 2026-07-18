/**
 * 报告正文极简渲染:\n\n 分段;「## 」/「# 」开头行作小标题。
 * 纯文本按段落排版,不引入 Markdown 依赖。
 */
export function ReportText({ text }: { text: string }) {
  const blocks = text
    .split(/\n{2,}/)
    .map((b) => b.trim())
    .filter(Boolean);

  return (
    <div className="flex flex-col gap-5">
      {blocks.map((block, i) => {
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
        return (
          <p key={i} className="whitespace-pre-wrap font-reading text-[16px] leading-[1.9] text-ink-secondary">
            {block}
          </p>
        );
      })}
    </div>
  );
}
