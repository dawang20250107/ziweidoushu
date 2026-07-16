"use client";

/** 受控检索框:书架上回车跳检索页,检索页里即时改写查询。 */
export function SearchBox({
  value,
  onChange,
  onSubmit,
  autoFocus,
  placeholder = "检索全文…",
}: {
  value: string;
  onChange: (v: string) => void;
  onSubmit?: (v: string) => void;
  autoFocus?: boolean;
  placeholder?: string;
}) {
  return (
    <form
      role="search"
      onSubmit={(e) => {
        e.preventDefault();
        onSubmit?.(value);
      }}
      className="flex items-center gap-2 rounded-[6px] bg-bg-raised px-3 py-2 shadow-[inset_0_0_0_1px_var(--line)] transition-shadow focus-within:shadow-[inset_0_0_0_1px_var(--gold-dim)]"
    >
      <svg
        aria-hidden
        viewBox="0 0 20 20"
        className="h-4 w-4 shrink-0 text-ink-faint"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.6"
      >
        <circle cx="9" cy="9" r="6" />
        <path d="M13.5 13.5 18 18" strokeLinecap="round" />
      </svg>
      <input
        type="search"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        // eslint-disable-next-line jsx-a11y/no-autofocus
        autoFocus={autoFocus}
        placeholder={placeholder}
        aria-label="检索古籍全文"
        className="min-w-0 flex-1 bg-transparent text-[15px] text-ink outline-none placeholder:text-ink-faint [&::-webkit-search-cancel-button]:appearance-none"
      />
      {value && (
        <button
          type="button"
          onClick={() => onChange("")}
          aria-label="清空检索"
          className="shrink-0 text-[13px] text-ink-faint transition-colors hover:text-ink"
        >
          清除
        </button>
      )}
    </form>
  );
}
