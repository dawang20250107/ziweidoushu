"use client";

import { Suspense, useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { streamInterpret, ApiError } from "@/lib/api";
import type { BirthInfo } from "@/lib/types";
import { SubjectBar } from "@/components/chat/SubjectBar";
import { TopicChips, type Topic } from "@/components/chat/TopicChips";
import { MessageList, type ChatMessage } from "@/components/chat/MessageList";
import { Composer } from "@/components/chat/Composer";

const BIRTH_KEY = "ziwei-birth";

let seq = 0;
const uid = () => `m${Date.now().toString(36)}-${seq++}`;

function readBirth(): BirthInfo | null {
  try {
    const raw = localStorage.getItem(BIRTH_KEY);
    return raw ? (JSON.parse(raw) as BirthInfo) : null;
  } catch {
    return null;
  }
}

function ChatInner() {
  const searchParams = useSearchParams();
  const [birth, setBirth] = useState<BirthInfo | null>(null);
  const [loaded, setLoaded] = useState(false);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState("");
  const [busy, setBusy] = useState(false);
  const abortRef = useRef<AbortController | null>(null);

  // 启动:读取排盘页写入的命主生辰
  useEffect(() => {
    setBirth(readBirth());
    setLoaded(true);
  }, []);

  // 跨页契约:?q=… 进入即填入输入框(不自动发送)
  useEffect(() => {
    const q = searchParams.get("q");
    if (q) setInput(q);
  }, [searchParams]);

  const send = useCallback(
    async (opts: { topic?: string; question?: string; display: string }) => {
      if (busy || !birth) return;

      const aiId = uid();
      setMessages((prev) => [
        ...prev,
        { id: uid(), role: "user", content: opts.display },
        { id: aiId, role: "assistant", content: "", streaming: true },
      ]);
      setBusy(true);

      const controller = new AbortController();
      abortRef.current = controller;

      try {
        const result = await streamInterpret(
          { ...birth, topic: opts.topic, question: opts.question },
          (delta) => {
            setMessages((prev) =>
              prev.map((m) =>
                m.id === aiId ? { ...m, content: m.content + delta } : m,
              ),
            );
          },
          controller.signal,
        );
        setMessages((prev) =>
          prev.map((m) =>
            m.id === aiId
              ? { ...m, streaming: false, provider: result.provider, degraded: result.degraded }
              : m,
          ),
        );
      } catch (e) {
        const aborted = e instanceof DOMException && e.name === "AbortError";
        setMessages((prev) =>
          prev.map((m) => {
            if (m.id !== aiId) return m;
            if (aborted) {
              return { ...m, streaming: false, content: m.content || "已中止生成。" };
            }
            return {
              ...m,
              streaming: false,
              error: true,
              content: e instanceof ApiError ? e.message : "解读失败,请稍后重试。",
            };
          }),
        );
      } finally {
        setBusy(false);
        abortRef.current = null;
      }
    },
    [birth, busy],
  );

  const sendInput = useCallback(() => {
    const q = input.trim();
    if (!q || busy) return;
    setInput("");
    void send({ question: q, display: q });
  }, [input, busy, send]);

  const sendTopic = useCallback(
    (t: Topic) => {
      if (busy) return;
      void send({ topic: t.key, display: t.label });
    },
    [busy, send],
  );

  const stop = useCallback(() => abortRef.current?.abort(), []);
  const clear = useCallback(() => {
    if (busy) return;
    setMessages([]);
  }, [busy]);

  // 首帧未读完 localStorage 前不渲染,避免引导卡闪现
  if (!loaded) {
    return <div className="h-[calc(100dvh-3.5rem)]" />;
  }

  if (!birth) {
    return (
      <div className="mx-auto max-w-md px-4 py-24 md:py-32">
        <div className="rounded-[10px] bg-bg-raised px-6 py-16 text-center shadow-[0_0_0_1px_var(--line)]">
          <p className="font-display text-xl font-semibold text-ink">先去排盘,才能问星</p>
          <p className="mt-3 text-[14px] leading-relaxed text-ink-secondary">
            问星以你的命盘为据。先完成排盘,系统会记住你的生辰,再来与命盘对话。
          </p>
          <Link
            href="/chart"
            className="glow-gold mt-8 inline-flex min-h-[44px] items-center justify-center rounded-[6px] bg-gold px-6 py-2.5 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright"
          >
            去排盘
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto flex h-[calc(100dvh-3.5rem)] max-w-3xl flex-col px-4">
      <div className="flex flex-col gap-3 pt-6">
        <SubjectBar birth={birth} canClear={messages.length > 0 && !busy} onClear={clear} />
        <TopicChips onPick={sendTopic} disabled={busy} />
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto py-4">
        <MessageList messages={messages} />
      </div>

      <div className="pb-4">
        <Composer value={input} onChange={setInput} onSend={sendInput} onStop={stop} busy={busy} />
        <p className="mt-2 text-center text-[11px] text-ink-faint">
          内容为传统文化与娱乐参考,不构成医疗投资建议。
        </p>
      </div>
    </div>
  );
}

/** 问星:排盘与 AI 一体化入口。use client + useSearchParams 需 Suspense 包裹。 */
export default function ChatPage() {
  return (
    <Suspense fallback={<div className="h-[calc(100dvh-3.5rem)]" />}>
      <ChatInner />
    </Suspense>
  );
}
