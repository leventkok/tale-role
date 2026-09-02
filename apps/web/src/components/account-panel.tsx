"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { useRouter } from "@/i18n/routing";

export function AccountPanel() {
  const t = useTranslations("account");
  const router = useRouter();
  const [confirm, setConfirm] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onExport() {
    setBusy(true);
    setError(null);
    const res = await fetch("/api/me/export", { cache: "no-store" });
    setBusy(false);
    if (!res.ok) {
      setError("error");
      return;
    }
    const blob = await res.blob();
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "talerole-export.json";
    a.click();
    URL.revokeObjectURL(url);
  }

  async function onErase(e: React.FormEvent) {
    e.preventDefault();
    if (confirm !== "DELETE") {
      return;
    }
    setBusy(true);
    setError(null);
    const res = await fetch("/api/me", { method: "DELETE" });
    if (!res.ok) {
      setBusy(false);
      setError("error");
      return;
    }
    await fetch("/api/auth/logout", { method: "POST" });
    router.push("/");
    router.refresh();
  }

  return (
    <div>
      <p>
        <button type="button" disabled={busy} onClick={() => void onExport()}>
          {t("export")}
        </button>
      </p>
      <form onSubmit={(e) => void onErase(e)}>
        <label>
          {t("confirm")}
          <input value={confirm} onChange={(e) => setConfirm(e.target.value)} autoComplete="off" />
        </label>
        {error ? (
          <p className="alert" role="alert">
            {error}
          </p>
        ) : null}
        <button className="ghost" type="submit" disabled={busy || confirm !== "DELETE"}>
          {t("erase")}
        </button>
      </form>
    </div>
  );
}
