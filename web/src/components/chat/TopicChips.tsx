"use client";

/** 主题速问:点击即以该 topic 发起解读(question 留空)。 */
export interface Topic {
  key: string;
  label: string;
}

export const TOPICS: Topic[] = [
  { key: "overview", label: "命格总览" },
  { key: "love", label: "感情婚姻" },
  { key: "career", label: "事业" },
  { key: "wealth", label: "财运" },
  { key: "health", label: "健康" },
];

export function TopicChips({
  onPick,
  disabled,
}: {
  onPick: (topic: Topic) => void;
  disabled?: boolean;
}) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      <span className="mr-1 text-[11px] tracking-[0.08em] text-ink-faint">主题速问</span>
      {TOPICS.map((t) => (
        <button
          key={t.key}
          type="button"
          disabled={disabled}
          onClick={() => onPick(t)}
          className="rounded-[6px] bg-bg-raised px-3 py-1.5 text-[13px] text-ink-secondary shadow-[inset_0_0_0_1px_var(--line)] transition-colors hover:text-gold hover:shadow-[inset_0_0_0_1px_var(--gold-dim)] disabled:cursor-not-allowed disabled:opacity-50"
        >
          {t.label}
        </button>
      ))}
    </div>
  );
}
