"use client";

import { Suspense, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { sendCode, verifyCode } from "@/lib/auth";

const fieldCls =
  "w-full rounded-[6px] bg-bg px-3 py-2.5 text-[15px] text-ink shadow-[inset_0_0_0_1px_var(--line)] focus:shadow-[inset_0_0_0_1px_var(--gold-dim)] outline-none transition-shadow tnum";

function LoginForm() {
  const router = useRouter();
  const params = useSearchParams();
  const [phone, setPhone] = useState("");
  const [code, setCode] = useState("");
  const [step, setStep] = useState<"phone" | "code">("phone");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [devHint, setDevHint] = useState("");
  const [countdown, setCountdown] = useState(0);

  async function handleSend() {
    setBusy(true);
    setError("");
    try {
      const resp = await sendCode(phone);
      setStep("code");
      if (resp.devCode) {
        setDevHint(resp.devCode);
        setCode(resp.devCode);
      }
      setCountdown(60);
      const timer = setInterval(() => {
        setCountdown((c) => {
          if (c <= 1) clearInterval(timer);
          return c - 1;
        });
      }, 1000);
    } catch (e) {
      setError(e instanceof Error ? e.message : "发送失败");
    } finally {
      setBusy(false);
    }
  }

  async function handleVerify(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      await verifyCode(phone, code);
      router.replace(params.get("next") || "/chart");
    } catch (err) {
      setError(err instanceof Error ? err.message : "登录失败");
    } finally {
      setBusy(false);
    }
  }

  const phoneValid = /^1\d{10}$/.test(phone);

  return (
    <div className="mx-auto max-w-sm px-4 py-20 md:py-28">
      <p className="mb-3 text-center text-[12px] font-medium tracking-[0.24em] text-gold">观星台</p>
      <h1 className="mb-10 text-center font-display text-[25px] font-semibold sm:text-[31px]">手机号登录</h1>

      <form onSubmit={handleVerify} className="flex flex-col gap-5">
        <label className="flex flex-col gap-1.5">
          <span className="text-[12px] text-ink-faint">手机号</span>
          <input
            type="tel"
            inputMode="numeric"
            autoComplete="tel"
            placeholder="11 位手机号"
            maxLength={11}
            value={phone}
            onChange={(e) => setPhone(e.target.value.replace(/\D/g, ""))}
            className={fieldCls}
          />
        </label>

        {step === "code" && (
          <label className="flex flex-col gap-1.5">
            <span className="text-[12px] text-ink-faint">验证码</span>
            <input
              type="text"
              inputMode="numeric"
              autoComplete="one-time-code"
              placeholder="6 位验证码"
              maxLength={6}
              value={code}
              onChange={(e) => setCode(e.target.value.replace(/\D/g, ""))}
              className={fieldCls}
            />
          </label>
        )}

        {devHint && (
          <p className="rounded-[4px] bg-bg-raised px-3 py-2 text-[12px] text-warn shadow-[inset_0_0_0_1px_var(--line)]">
            开发模式:验证码 <span className="tnum font-medium">{devHint}</span>(已自动填入,接入短信服务商后此提示消失)
          </p>
        )}
        {error && <p className="text-[13px] text-danger">{error}</p>}

        {step === "phone" ? (
          <button
            type="button"
            disabled={!phoneValid || busy}
            onClick={handleSend}
            className="glow-gold inline-flex min-h-[44px] items-center justify-center rounded-[6px] bg-gold px-6 py-3 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright disabled:opacity-40 disabled:shadow-none"
          >
            {busy ? "发送中…" : "获取验证码"}
          </button>
        ) : (
          <>
            <button
              type="submit"
              disabled={code.length !== 6 || busy}
              className="glow-gold inline-flex min-h-[44px] items-center justify-center rounded-[6px] bg-gold px-6 py-3 text-[15px] font-medium text-[#161206] transition-colors hover:bg-gold-bright disabled:opacity-40 disabled:shadow-none"
            >
              {busy ? "登录中…" : "登录 / 注册"}
            </button>
            <button
              type="button"
              disabled={countdown > 0 || busy}
              onClick={handleSend}
              className="text-[13px] text-ink-secondary transition-colors hover:text-gold disabled:opacity-40"
            >
              {countdown > 0 ? `${countdown}s 后可重发` : "重新发送验证码"}
            </button>
          </>
        )}
      </form>

      <p className="mt-8 text-center text-[12px] leading-relaxed text-ink-faint">
        登录即代表同意以传统文化研究为目的使用本平台。
        <br />
        未注册手机号将自动创建账号。
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
      <LoginForm />
    </Suspense>
  );
}
