"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { useRouter } from "@/i18n/routing";

type Mode = "login" | "register" | "verify";

export function AuthForm({ mode, email: initialEmail }: { mode: Mode; email?: string }) {
  const t = useTranslations("auth");
  const router = useRouter();
  const [email, setEmail] = useState(initialEmail ?? "");
  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setBusy(true);
    const path =
      mode === "login"
        ? "/api/auth/login"
        : mode === "register"
          ? "/api/auth/register"
          : "/api/auth/otp/verify";
    const body = mode === "verify" ? { email, code } : { email, password };
    const res = await fetch(path, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    const data = await res.json().catch(() => ({}));
    setBusy(false);
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
            className="otp"
            inputMode="numeric"
            pattern="[0-9]{6}"
            maxLength={6}
            autoComplete="one-time-code"
            required
            value={code}
            onChange={(e) => setCode(e.target.value)}
          />
        </label>
      )}
      {error ? (
        <p className="alert" role="alert">
          {error}
        </p>
      ) : null}
      <button type="submit" disabled={busy}>
        {mode === "verify" ? t("verify") : t("submit")}
      </button>
    </form>
  );
}
