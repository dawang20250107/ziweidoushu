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

function UserBubble({ text }: { text: string }) {
  return (
    <div className="msg-in flex justify-end">
      <div className="max-w-[85%] whitespace-pre-wrap break-words rounded-[6px] bg-gold px-4 py-2.5 text-[15px] leading-relaxed text-[#161206]">
        {text}
      </div>
    </div>
  );
}

function AssistantBubble({ m }: { m: ChatMessage }) {
  const showFootnote = !m.streaming && !m.error && (m.degraded || !!m.provider);
  return (
    <div className="msg-in flex justify-start">
      <div
        className={[
          "max-w-[85%] rounded-[6px] bg-bg-raised px-4 py-3 shadow-[0_0_0_1px_var(--line)]",
          m.streaming ? "bubble-streaming" : "",
        ].join(" ")}
      >
        {m.error ? (
          <p className="text-[14px] leading-relaxed text-danger">{m.content}</p>
        ) : (
          <>
            {m.content && (
              <div className={m.streaming ? "stream-md" : undefined}>
                <Markdown text={m.content} elder />
              </div>
            )}
            {m.streaming && !m.content && (
              <span className="flex items-center gap-2 text-[13px] text-ink-faint">
                正在推演命盘
                <span className="flex items-center gap-1" aria-hidden>
                  <span className="think-dot" />
                  <span className="think-dot" style={{ animationDelay: "0.15s" }} />
                  <span className="think-dot" style={{ animationDelay: "0.3s" }} />
                </span>
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
  const listRef = useRef<HTMLDivElement>(null);
  const stickRef = useRef(true); // 是否贴底:贴底才跟随流式滚动
  const hasMessages = messages.length > 0;

  // 贴底跟踪:用户上翻回看(离底 > 80px)时,流式增量不再强拉到底——
  // 修「AI 长答期间无法回看上文」;发出新提问则恢复贴底。
  useEffect(() => {
    if (!hasMessages) return;
    const container = listRef.current?.parentElement;
    if (!container) return;
    const onScroll = () => {
      stickRef.current =
        container.scrollHeight - container.scrollTop - container.clientHeight < 80;
    };
    container.addEventListener("scroll", onScroll, { passive: true });
    return () => container.removeEventListener("scroll", onScroll);
  }, [hasMessages]);

  // 新提问恢复贴底:发问时用户气泡与助手占位同帧追加(末条是 assistant),
  // 故以「用户消息数量增长」判定,而非看末条角色。
  const userCountRef = useRef(0);
  useEffect(() => {
    const users = messages.filter((m) => m.role === "user").length;
    if (users > userCountRef.current) stickRef.current = true;
    userCountRef.current = users;
    if (stickRef.current) endRef.current?.scrollIntoView({ block: "end" });
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
    <div ref={listRef} className="flex flex-col gap-4">
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
