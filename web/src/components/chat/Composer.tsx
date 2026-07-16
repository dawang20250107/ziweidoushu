"use client";

import { useEffect, useRef } from "react";

function SendIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
      <path d="M12 19V5" />
      <path d="M5 12l7-7 7 7" />
    </svg>
  );
}

function StopIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" aria-hidden>
      <rect x="6" y="6" width="12" height="12" rx="1.5" />
    </svg>
  );
}

/** 输入框:Enter 发送 / Shift+Enter 换行;发送中可中止。 */
export function Composer({
  value,
  onChange,
  onSend,
  onStop,
  busy,
}: {
  value: string;
  onChange: (v: string) => void;
  onSend: () => void;
  onStop: () => void;
  busy: boolean;
}) {
  const ref = useRef<HTMLTextAreaElement>(null);

  // 自适应高度(1~6 行)
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    el.style.height = "0px";
    el.style.height = `${Math.min(el.scrollHeight, 160)}px`;
  }, [value]);

  function onKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    // 输入法组字过程中的 Enter 不发送(中文输入关键)
    if (e.key === "Enter" && !e.shiftKey && !e.nativeEvent.isComposing) {
      e.preventDefault();
      if (!busy) onSend();
    }
  }

  const canSend = value.trim().length > 0;

  return (
    <div className="flex items-end gap-2 rounded-[10px] bg-bg-raised p-2 shadow-[0_0_0_1px_var(--line)] focus-within:shadow-[0_0_0_1px_var(--gold-dim)]">
      <textarea
        ref={ref}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={onKeyDown}
        rows={1}
        placeholder="就此命盘发问…（Enter 发送，Shift+Enter 换行）"
        className="max-h-40 flex-1 resize-none bg-transparent px-2 py-1.5 text-[15px] leading-relaxed text-ink outline-none placeholder:text-ink-faint"
        aria-label="向命盘提问"
      />
      {busy ? (
        <button
          type="button"
          onClick={onStop}
          aria-label="中止生成"
          className="flex h-9 w-9 shrink-0 items-center justify-center rounded-[6px] bg-bg text-ink-secondary shadow-[inset_0_0_0_1px_var(--line-strong)] transition-colors hover:text-ink"
        >
          <StopIcon />
        </button>
      ) : (
        <button
          type="button"
          onClick={onSend}
          disabled={!canSend}
          aria-label="发送"
          className="flex h-9 w-9 shrink-0 items-center justify-center rounded-[6px] bg-gold text-[#161206] transition-colors hover:bg-gold-bright disabled:cursor-not-allowed disabled:opacity-40"
        >
          <SendIcon />
        </button>
      )}
    </div>
  );
}
