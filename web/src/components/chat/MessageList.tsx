"use client";

import { useEffect, useRef } from "react";
import { Markdown } from "./Markdown";

export type ChatRole = "user" | "assistant";

export interface ChatMessage {
  id: string;
  role: ChatRole;
  /** 用户:提问文本 / 助手:流式累积的回答 */
  content: string;
  /** 助手正在流式生成 */
  streaming?: boolean;
  /** 助手回答完成后回填的 provider */
  provider?: string;
  /** 知识库降级(未配置 AI Key) */
  degraded?: boolean;
  /** 助手侧错误卡片 */
  error?: boolean;
}

/** 闪烁光标(用 Tailwind 内置 animate-pulse,不引入自定义 CSS)。 */
function Caret() {
  return (
    <span
      className="inline-block h-[15px] w-[2px] animate-pulse bg-gold align-text-bottom"
      aria-hidden
    />
  );
}

function UserBubble({ text }: { text: string }) {
  return (
    <div className="flex justify-end">
      <div className="max-w-[85%] whitespace-pre-wrap break-words rounded-[6px] bg-gold px-4 py-2.5 text-[15px] leading-relaxed text-[#161206]">
        {text}
      </div>
    </div>
  );
}

function AssistantBubble({ m }: { m: ChatMessage }) {
  const showFootnote = !m.streaming && !m.error && (m.degraded || !!m.provider);
  return (
    <div className="flex justify-start">
      <div className="max-w-[85%] rounded-[6px] bg-bg-raised px-4 py-3 shadow-[0_0_0_1px_var(--line)]">
        {m.error ? (
          <p className="text-[14px] leading-relaxed text-danger">{m.content}</p>
        ) : (
          <>
            {m.content && <Markdown text={m.content} />}
            {m.streaming && (
              <span className="mt-1 flex items-center gap-2 text-[13px] text-ink-faint">
                {!m.content && <span>正在推演命盘…</span>}
                <Caret />
              </span>
            )}
            {showFootnote && (
              <p className="mt-2.5 text-[11px] text-ink-faint">
                {m.degraded
                  ? "知识库规则版,配置 AI Key 可获深度解读"
                  : `模型:${m.provider}`}
              </p>
            )}
          </>
        )}
      </div>
    </div>
  );
}

export function MessageList({ messages }: { messages: ChatMessage[] }) {
  const endRef = useRef<HTMLDivElement>(null);

  // 每次消息或流式增量更新后,滚动到底部跟随对话
  useEffect(() => {
    endRef.current?.scrollIntoView({ block: "end" });
  }, [messages]);

  if (messages.length === 0) {
    return (
      <div className="flex h-full flex-col items-center justify-center px-6 text-center">
        <p className="font-display text-xl font-semibold text-ink-secondary">与你的命盘对话</p>
        <p className="mt-3 max-w-sm text-[14px] leading-relaxed text-ink-faint">
          选择上方主题速问,或直接输入你的疑问 —— 关于性情、姻缘、事业、财帛与流年,皆可就盘发问。
        </p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      {messages.map((m) =>
        m.role === "user" ? (
          <UserBubble key={m.id} text={m.content} />
        ) : (
          <AssistantBubble key={m.id} m={m} />
        ),
      )}
      <div ref={endRef} />
    </div>
  );
}
