import { Fragment, type ReactNode } from "react";

/**
 * 轻量 Markdown 渲染 —— 不引入任何依赖,正则逐行处理。
 *
 * 安全模型(XSS):全程不使用 dangerouslySetInnerHTML。所有用户/模型文本
 * 只作为 React 子节点渲染(React 默认对文本节点转义),解析器仅识别一个
 * 固定的语法白名单(标题 # ## ###、无序列表 -、引用 >、**粗体**、分段),
 * 并只发出固定的元素类型。任何输入文本都不可能转义成标记 —— 即使模型
 * 吐出 `<script>` 或 `<img onerror=…>`,也只会作为纯文本原样显示。
 */

/** 行内解析:仅识别 **粗体**,其余文本原样输出为受转义的文本节点。 */
function renderInline(text: string, keyBase: string): ReactNode[] {
  const nodes: ReactNode[] = [];
  const re = /\*\*(.+?)\*\*/g;
  let last = 0;
  let idx = 0;
  let m: RegExpExecArray | null;
  while ((m = re.exec(text)) !== null) {
    if (m.index > last) nodes.push(text.slice(last, m.index));
    nodes.push(
      <strong key={`${keyBase}-b${idx++}`} className="font-semibold text-ink">
        {m[1]}
      </strong>,
    );
    last = m.index + m[0].length;
  }
  if (last < text.length) nodes.push(text.slice(last));
  return nodes;
}

/** 多行文本(段落/引用)按软换行拼接。 */
function renderLines(lines: string[], keyBase: string): ReactNode[] {
  return lines.map((line, i) => (
    <Fragment key={`${keyBase}-l${i}`}>
      {i > 0 && <br />}
      {renderInline(line, `${keyBase}-l${i}`)}
    </Fragment>
  ));
}

const HEADING_CLS: Record<number, string> = {
  1: "font-display text-lg font-semibold text-ink",
  2: "font-display text-base font-semibold text-ink",
  3: "font-display text-[15px] font-semibold text-ink-secondary",
};

export function Markdown({ text, elder = false }: { text: string; elder?: boolean }) {
  const lines = text.replace(/\r\n/g, "\n").split("\n");
  const blocks: ReactNode[] = [];
  let para: string[] = [];
  let key = 0;
  let i = 0;

  const flushPara = () => {
    if (para.length === 0) return;
    const k = `p${key++}`;
    blocks.push(
      <p key={k} className="leading-relaxed text-ink">
        {renderLines(para, k)}
      </p>,
    );
    para = [];
  };

  while (i < lines.length) {
    const line = lines[i];

    // 空行 → 分段
    if (/^\s*$/.test(line)) {
      flushPara();
      i++;
      continue;
    }

    // 标题 # / ## / ###
    const heading = /^(#{1,3})\s+(.*)$/.exec(line);
    if (heading) {
      flushPara();
      const level = heading[1].length;
      const k = `h${key++}`;
      const cls = HEADING_CLS[level];
      const content = renderInline(heading[2], k);
      blocks.push(
        level === 1 ? (
          <h3 key={k} className={cls}>{content}</h3>
        ) : level === 2 ? (
          <h4 key={k} className={cls}>{content}</h4>
        ) : (
          <h5 key={k} className={cls}>{content}</h5>
        ),
      );
      i++;
      continue;
    }

    // 引用 >(可连续多行)
    if (/^\s*>\s?/.test(line)) {
      flushPara();
      const quote: string[] = [];
      while (i < lines.length && /^\s*>\s?/.test(lines[i])) {
        quote.push(lines[i].replace(/^\s*>\s?/, ""));
        i++;
      }
      const k = `q${key++}`;
      blocks.push(
        <blockquote
          key={k}
          className="border-l-2 border-gold-dim pl-3 font-reading text-ink-secondary"
        >
          {renderLines(quote, k)}
        </blockquote>,
      );
      continue;
    }

    // 无序列表 - / *(可连续多行)
    if (/^\s*[-*]\s+/.test(line)) {
      flushPara();
      const items: string[] = [];
      while (i < lines.length && /^\s*[-*]\s+/.test(lines[i])) {
        items.push(lines[i].replace(/^\s*[-*]\s+/, ""));
        i++;
      }
      const k = `ul${key++}`;
      blocks.push(
        <ul key={k} className="list-disc space-y-1 pl-5 text-ink">
          {items.map((it, li) => (
            <li key={`${k}-i${li}`} className="leading-relaxed">
              {renderInline(it, `${k}-i${li}`)}
            </li>
          ))}
        </ul>,
      );
      continue;
    }

    // 普通文本 → 累积到当前段落
    para.push(line);
    i++;
  }
  flushPara();

  return <div className={["space-y-2 text-[15px]", elder ? "prose-elder" : ""].join(" ")}>{blocks}</div>;
}
