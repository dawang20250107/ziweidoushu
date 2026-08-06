"use client";

import { Suspense, useCallback, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { loginEmail, registerEmail, resetPassword, sendEmailCode, type EmailCodePurpose } from "@/lib/auth";

const fieldCls =
  "w-full rounded-[6px] bg-bg px-3 py-2.5 text-[15px] text-ink shadow-[inset_0_0_0_1px_var(--line)] focus:shadow-[inset_0_0_0_1px_var(--gold-dim)] outline-none transition-shadow";

type Mode = "login" | "register" | "reset";

const MODE_META: Record<Mode, { title: string; cta: string; busy: string }> = {
  login: { title: "邮箱登录", cta: "登录", busy: "登录中…" },
  register: { title: "邀请注册", cta: "注册并登录", busy: "注册中…" },
  reset: { title: "找回密码", cta: "重置密码", busy: "重置中…" },
};

/**
 * 内测鉴权(邮箱通道):
 * 登录 = 邮箱 + 密码(strict 模式陌生设备再要一次邮箱验证码);
 * 注册 = 邮箱验证码 + 自设密码 + 邀请码;找回 = 邮箱验证码 + 新密码。
 */
function AuthForm() {
  const router = useRouter();
  const params = useSearchParams();
  const [mode, setMode] = useState<Mode>("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [invite, setInvite] = useState("");
  const [needVerify, setNeedVerify] = useState(false); // 登录态:陌生设备升级
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [devHint, setDevHint] = useState("");
  const [countdown, setCountdown] = useState(0);

  const emailValid = /^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(email.trim());
  const pwValid = password.length >= 8 && /[a-zA-Z]/.test(password) && /\d/.test(password);
  const meta = MODE_META[mode];

  const switchMode = (m: Mode) => {
    setMode(m);
    setError("");
    setNotice("");
    setCode("");
    setNeedVerify(false);
    setDevHint("");
  };

  const startCountdown = () => {
    setCountdown(60);
    const timer = setInterval(() => {
      setCountdown((c) => {
        if (c <= 1) clearInterval(timer);
        return c - 1;
      });
    }, 1000);
  };

  const handleSendCode = useCallback(async () => {
    const purpose: EmailCodePurpose = mode === "register" ? "register" : mode === "reset" ? "reset" : "login";
    setBusy(true);
    setError("");
    try {
      const resp = await sendEmailCode(email.trim(), purpose);
      setNotice("验证码已发送至邮箱,10 分钟内有效。");
      if (resp.devCode) {
        setDevHint(resp.devCode);
        setCode(resp.devCode);
      }
      startCountdown();
    } catch (e) {
      setError(e instanceof Error ? e.message : "发送失败");
    } finally {
      setBusy(false);
    }
  }, [email, mode]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      if (mode === "login") {
        const r = await loginEmail({
          email: email.trim(),
          password,
          code: needVerify ? code : undefined,
        });
        if (r.needVerify) {
          // 环境检测:陌生设备,引导邮箱验证码升级
          setNeedVerify(true);
          setNotice("检测到新设备登录,请用邮箱验证码确认一次。");
          setBusy(false);
          return;
        }
        router.replace(params.get("next") || "/chart");
      } else if (mode === "register") {
        await registerEmail({ email: email.trim(), code, password, invite: invite.trim() || undefined });
        router.replace(params.get("next") || "/chart");
      } else {
        await resetPassword({ email: email.trim(), code, newPassword: password });
        setNotice("密码已重置,请用新密码登录。");
        switchMode("login");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "操作失败");
    } finally {
      setBusy(false);
    }
  }

  const needCode = mode !== "login" || needVerify;
  const canSubmit =
    emailValid && pwValid && !busy && (!needCode || code.length === 6) && (mode !== "register" || invite.trim() !== "");

  return (
    <div className="mx-auto max-w-sm px-4 py-20 md:py-28">
      <p className="mb-3 text-center text-[12px] font-medium tracking-[0.24em] text-gold">观星台 · 内测</p>
      <h1 className="mb-8 text-center font-display text-[25px] font-semibold sm:text-[31px]">{meta.title}</h1>

      {/* 模式切换 */}
      <div className="mb-6 flex overflow-hidden rounded-[6px] shadow-[inset_0_0_0_1px_var(--line)]" role="tablist">
        {(
          [
            ["login", "登录"],
            ["register", "邀请注册"],
            ["reset", "找回密码"],
          ] as const
        ).map(([m, label]) => (
          <button
            key={m}
            type="button"
            role="tab"
            aria-selected={mode === m}
            onClick={() => switchMode(m)}
            className={[
              "min-h-[40px] flex-1 text-[13px] transition-colors",
              mode === m ? "bg-[var(--gold-glow)] font-medium text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)]" : "text-ink-secondary hover:text-ink",
            ].join(" ")}
          >
            {label}
          </button>
        ))}
      </div>

      <form onSubmit={handleSubmit} className="flex flex-col gap-5">
        <label className="flex flex-col gap-1.5">
          <span className="text-[12px] text-ink-faint">邮箱</span>
          <input
            type="email"
            autoComplete="email"
            placeholder="you@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className={fieldCls}
          />
        </label>

        {mode === "register" && (
          <label className="flex flex-col gap-1.5">
            <span className="text-[12px] text-ink-faint">邀请码(内测)</span>
            <input
              type="text"
              placeholder="ZW-XXXXXXXX"
              value={invite}
              onChange={(e) => setInvite(e.target.value.toUpperCase())}
              className={`${fieldCls} tnum`}
            />
          </label>
        )}

        <label className="flex flex-col gap-1.5">
          <span className="text-[12px] text-ink-faint">{mode === "reset" ? "新密码" : "密码"}</span>
          <input
            type="password"
            autoComplete={mode === "login" ? "current-password" : "new-password"}
            placeholder="8 位以上,含字母与数字"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className={fieldCls}
          />
        </label>

        {needCode && (
          <label className="flex flex-col gap-1.5">
            <span className="text-[12px] text-ink-faint">邮箱验证码</span>
            <div className="flex gap-2">
              <input
                type="text"
                inputMode="numeric"
                autoComplete="one-time-code"
                placeholder="6 位验证码"
                maxLength={6}
                value={code}
                onChange={(e) => setCode(e.target.value.replace(/\D/g, ""))}
                className={`${fieldCls} tnum flex-1`}
              />
              <button
                type="button"
                disabled={!emailValid || countdown > 0 || busy}
                onClick={handleSendCode}
                className="shrink-0 rounded-[6px] px-3 text-[13px] text-gold shadow-[inset_0_0_0_1px_var(--gold-dim)] transition-shadow hover:shadow-[inset_0_0_0_1px_var(--gold)] disabled:opacity-40"
              >
                {countdown > 0 ? `${countdown}s` : "发验证码"}
              </button>
            </div>
          </label>
        )}

        {devHint && (
          <p className="rounded-[4px] bg-bg-raised px-3 py-2 text-[12px] text-warn shadow-[inset_0_0_0_1px_var(--line)]">
            开发模式:验证码 <span className="tnum font-medium">{devHint}</span>(已自动填入,配置 SMTP 后此提示消失)
          </p>
        )}
        {notice && !error && <p className="text-[13px] text-ok">{notice}</p>}
        {error && <p className="text-[13px] text-danger">{error}</p>}

        <button
          type="submit"
          disabled={!canSubmit}
          className="glow-gold inline-flex min-h-[44px] items-center justify-center rounded-[6px] bg-gold px-6 py-3 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright disabled:opacity-40 disabled:shadow-none"
        >
          {busy ? meta.busy : meta.cta}
        </button>
      </form>

      <p className="mt-8 text-center text-[12px] leading-relaxed text-ink-faint">
        内测阶段凭邀请码注册;登录即代表同意以传统文化研究为目的使用本平台。
        <Link href="/" className="ml-1 text-gold-dim hover:text-gold">
          返回首页
        </Link>
      </p>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense>
      <AuthForm />
    </Suspense>
  );
}
