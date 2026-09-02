"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { useRouter } from "@/i18n/routing";

type Mode = "login" | "register" | "verify";

export function AuthForm({ mode, email: initialEmail }: { mode: Mode; email?: string }) {
  const t = useTranslations("auth");
  const nav = useTranslations("nav");
  const router = useRouter();
  const [email, setEmail] = useState(initialEmail ?? "");
  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    const path =
      mode === "login"
        ? "/api/auth/login"
        : mode === "register"
          ? "/api/auth/register"
          : "/api/auth/otp/verify";
    const body =
      mode === "verify" ? { email, code } : { email, password };
    const res = await fetch(path, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    const data = await res.json().catch(() => ({}));
    if (data.token) {
      setError("session leaked");
      return;
    }
    if (!res.ok) {
      if (data.otp_required || mode === "register") {
        router.push(`/verify?email=${encodeURIComponent(email)}`);
        return;
      }
      setError(typeof data.error === "string" ? data.error : "unauthorized");
      return;
    }
    if (mode === "register" || data.otp_required) {
      router.push(`/verify?email=${encodeURIComponent(email)}`);
      return;
    }
    router.push("/");
    router.refresh();
  }

  return (
    <form onSubmit={onSubmit}>
      <label>
        {t("email")}
        <input
          type="email"
          autoComplete="email"
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />
      </label>
      {mode !== "verify" ? (
        <label>
          {t("password")}
          <input
            type="password"
            autoComplete={mode === "login" ? "current-password" : "new-password"}
            minLength={8}
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </label>
      ) : (
        <label>
          {t("otp")}
          <input
            inputMode="numeric"
            pattern="[0-9]{6}"
            maxLength={6}
            required
            value={code}
            onChange={(e) => setCode(e.target.value)}
          />
        </label>
      )}
      {mode === "verify" ? <p>{t("otpHint")}</p> : null}
      {error ? <p role="alert">{error}</p> : null}
      <button type="submit">{mode === "verify" ? t("verify") : t("submit")}</button>
      {mode === "login" ? <p>{t("needAccount")}</p> : null}
      {mode === "register" ? <p>{t("haveAccount")}</p> : null}
      <span hidden>{nav("signIn")}</span>
    </form>
  );
}
