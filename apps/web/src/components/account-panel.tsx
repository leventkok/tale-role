"use client";

import { useEffect, useState, type MouseEvent } from "react";
import { useTranslations } from "next-intl";
import { useRouter } from "@/i18n/routing";
import { gql, gqlData } from "@/lib/gql";
import { isDesktopShell, useDesktopBridge } from "@/lib/desktop-shell";
import { TotpQr } from "@/components/totp-qr";
import { AppearanceControls } from "@/components/appearance-controls";
import { ProfilePortrait } from "@/components/art/profile-portrait";
import { defaultPortraitId, portraitIds, normalizePortraitId, readStoredPortrait, writeStoredPortrait, type PortraitId } from "@/lib/portraits";

type License = { id: string; device_id: string; platform: string; created_at?: string };

export function AccountPanel() {
  const t = useTranslations("account");
  const router = useRouter();
  const [confirm, setConfirm] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [totpEnabled, setTotpEnabled] = useState(false);
  const [secret, setSecret] = useState<string | null>(null);
  const [otpauth, setOtpauth] = useState<string | null>(null);
  const [totpCode, setTotpCode] = useState("");
  const [copied, setCopied] = useState(false);
  const [licenses, setLicenses] = useState<License[]>([]);
  const [lanternXp, setLanternXp] = useState(0);
  const [lanternLevel, setLanternLevel] = useState(1);
  const [portrait, setPortrait] = useState<PortraitId>(defaultPortraitId);
  const desktop = useDesktopBridge();
  const lanternNeed = 100 * lanternLevel;
  const lanternPct = Math.max(4, Math.min(100, Math.round((lanternXp / lanternNeed) * 100)));

  async function refresh() {
    setPortrait(readStoredPortrait());
    const result = await gql<{
      me: { totpEnabled: boolean; lanternXp?: number; lanternLevel?: number; portraitId?: string } | null;
      licenses: { id: string; deviceId: string; platform: string; createdAt?: string }[];
    }>(`{ me { totpEnabled lanternXp lanternLevel portraitId } licenses { id deviceId platform createdAt } }`);
    const data = gqlData(result);
    if (data?.me) {
      setTotpEnabled(Boolean(data.me.totpEnabled));
      setLanternXp(data.me.lanternXp ?? 0);
      setLanternLevel(data.me.lanternLevel && data.me.lanternLevel > 0 ? data.me.lanternLevel : 1);
      if (data.me.portraitId) {
        const id = normalizePortraitId(data.me.portraitId);
        setPortrait(id);
        writeStoredPortrait(id);
      }
    }
    setLicenses(
      (data?.licenses ?? []).map((row) => ({
        id: row.id,
        device_id: row.deviceId,
        platform: row.platform,
        created_at: row.createdAt,
      })),
    );
  }

  useEffect(() => {
    void refresh();
  }, []);

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

  async function onBeginTotp() {
    setBusy(true);
    setError(null);
    const res = await fetch("/api/me/totp/begin", { method: "POST" });
    const data = (await res.json().catch(() => ({}))) as { secret?: string; otpauth_url?: string };
    setBusy(false);
    if (!res.ok || !data.secret) {
      setError("error");
      return;
    }
    setSecret(data.secret);
    setOtpauth(data.otpauth_url ?? null);
  }

  async function onConfirmTotp(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    const res = await fetch("/api/me/totp/confirm", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ code: totpCode }),
    });
    setBusy(false);
    if (!res.ok) {
      setError("error");
      return;
    }
    setSecret(null);
    setOtpauth(null);
    setTotpCode("");
    setTotpEnabled(true);
  }

  async function onDisableTotp(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    const res = await fetch("/api/me/totp/disable", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ code: totpCode }),
    });
    setBusy(false);
    if (!res.ok) {
      setError("error");
      return;
    }
    setTotpCode("");
    setTotpEnabled(false);
  }

  async function onRegisterDevice() {
    if (!desktop) {
      return;
    }
    setBusy(true);
    setError(null);
    const result = await gql<{ registerLicense: { id: string } }>(
      `mutation ($deviceId: String!, $platform: String) {
        registerLicense(deviceId: $deviceId, platform: $platform) { id }
      }`,
      { deviceId: desktop.deviceId, platform: desktop.platform },
    );
    setBusy(false);
    if (!gqlData(result)?.registerLicense?.id) {
      setError("error");
      return;
    }
    await refresh();
  }

  async function onRevokeDevice(id: string) {
    if (!window.confirm(t("disconnectConfirm"))) {
      return;
    }
    setBusy(true);
    setError(null);
    const result = await gql<{ revokeLicense: boolean }>(
      `mutation ($id: ID!) { revokeLicense(id: $id) }`,
      { id },
    );
    setBusy(false);
    if (!gqlData(result)?.revokeLicense) {
      setError("error");
      return;
    }
    await fetch("/api/auth/logout", { method: "POST" });
    router.push("/");
    router.refresh();
  }

  function onOpenDesktop(e: MouseEvent<HTMLAnchorElement>) {
    if (isDesktopShell()) {
      e.preventDefault();
    }
  }

  async function onErase(e: React.FormEvent) {
    e.preventDefault();
    if (confirm !== "DELETE") {
      return;
    }
    setBusy(true);
    setError(null);
    const result = await gql<{ eraseMe: boolean }>(`mutation { eraseMe }`);
    if (!gqlData(result)?.eraseMe) {
      setBusy(false);
      setError("error");
      return;
    }
    await fetch("/api/auth/logout", { method: "POST" });
    router.push("/");
    router.refresh();
  }

  async function pickPortrait(id: PortraitId) {
    setPortrait(id);
    writeStoredPortrait(id);
    await gql(`mutation ($id: String!) { setPortrait(id: $id) }`, { id });
  }

  async function copySecret() {
    if (!secret) {
      return;
    }
    await navigator.clipboard.writeText(secret);
    setCopied(true);
    window.setTimeout(() => setCopied(false), 1600);
  }

  return (
    <div className="account-stack">
      <section>
        <h2>{t("prefs")}</h2>
        <p className="muted">{t("prefsHint")}</p>
        <AppearanceControls signOut />
      </section>
      <section>
        <h2>{t("portrait")}</h2>
        <p className="muted">{t("portraitHint")}</p>
        <div className="portrait-picker">
          {portraitIds.map((id) => (
            <button
              key={id}
              type="button"
              className={portrait === id ? "portrait-pick on" : "portrait-pick"}
              aria-pressed={portrait === id}
              onClick={() => void pickPortrait(id)}
            >
              <ProfilePortrait id={id} />
              <span>{t(`portraits.${id}`)}</span>
            </button>
          ))}
        </div>
      </section>
      <section>
        <h2>{t("lantern")}</h2>
        <p className="muted">{t("lanternHint")}</p>
        <p className="muted">{t("lanternLevel", { n: lanternLevel })}</p>
        <div className="hp">
          <span className="muted">{t("lanternXp", { n: lanternXp, need: lanternNeed })}</span>
          <div className="hp-bar xp">
            <span style={{ ["--hp"]: `${lanternPct}%` } as React.CSSProperties} />
          </div>
        </div>
      </section>
      <section>
        <h2>{t("totpTitle")}</h2>
        <p className="muted">{totpEnabled ? t("totpOn") : t("totpOff")}</p>
        {!totpEnabled && !secret ? (
          <p>
            <button type="button" disabled={busy} onClick={() => void onBeginTotp()}>
              {t("totpBegin")}
            </button>
          </p>
        ) : null}
        {secret ? (
          <form className="totp-setup" onSubmit={(e) => void onConfirmTotp(e)}>
            <p className="muted">{t("totpScan")}</p>
            {otpauth ? <TotpQr value={otpauth} /> : null}
            <p className="muted">{t("totpManual")}</p>
            <code className="totp-secret">{secret.match(/.{1,4}/g)?.join(" ") ?? secret}</code>
            <div className="btn-row">
              <button className="ghost copy" type="button" onClick={() => void copySecret()}>
                {copied ? t("totpCopied") : t("totpCopy")}
              </button>
            </div>
            <label>
              {t("totpCode")}
              <input
                className="otp"
                inputMode="numeric"
                pattern="[0-9]{6}"
                maxLength={6}
                autoComplete="one-time-code"
                required
                value={totpCode}
                onChange={(e) => setTotpCode(e.target.value)}
              />
            </label>
            <button type="submit" disabled={busy}>
              {t("totpConfirm")}
            </button>
          </form>
        ) : null}
        {totpEnabled ? (
          <form onSubmit={(e) => void onDisableTotp(e)}>
            <label>
              {t("totpCode")}
              <input
                className="otp"
                inputMode="numeric"
                pattern="[0-9]{6}"
                maxLength={6}
                autoComplete="one-time-code"
                required
                value={totpCode}
                onChange={(e) => setTotpCode(e.target.value)}
              />
            </label>
            <button className="ghost" type="submit" disabled={busy}>
              {t("totpDisable")}
            </button>
          </form>
        ) : null}
      </section>
      <section>
        <h2>{t("devices")}</h2>
        {licenses.length === 0 ? <p className="muted">{t("noDevices")}</p> : null}
        <ul className="device-list">
          {licenses.map((row) => (
            <li key={row.id} className="device-row">
              <a className="device-open" href="talerole://open" onClick={onOpenDesktop} aria-label={t("openDesktop")}>
                {row.platform} · {row.device_id}
                {desktop && row.device_id === desktop.deviceId ? ` · ${t("thisDevice")}` : ""}
              </a>
              <button type="button" className="ghost" disabled={busy} onClick={() => void onRevokeDevice(row.id)}>
                {t("disconnectDevice")}
              </button>
            </li>
          ))}
        </ul>
        {desktop ? (
          <p>
            <button type="button" disabled={busy} onClick={() => void onRegisterDevice()}>
              {t("registerDevice")}
            </button>
          </p>
        ) : (
          <p className="muted">{t("desktopOnly")}</p>
        )}
      </section>
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
