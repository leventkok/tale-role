"use client";

import { useEffect, useState, type ReactNode } from "react";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { gql, gqlData } from "@/lib/gql";
import { desktopDeviceHeaders, isDesktopShell } from "@/lib/desktop-shell";

const OPEN_PATH = /\/(play|host|universe|join|table)(\/|$)/;

export function DesktopLicenseGate({ signedIn, children }: { signedIn: boolean; children: ReactNode }) {
  const t = useTranslations("account");
  const pathname = usePathname();
  const [blocked, setBlocked] = useState(false);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (!isDesktopShell() || !signedIn || !OPEN_PATH.test(pathname || "") || pathname?.includes("/account")) {
      setBlocked(false);
      return;
    }
    const device = desktopDeviceHeaders()["X-TaleRole-Device"];
    if (!device) {
      setBlocked(true);
      return;
    }
    let alive = true;
    void gql<{ licenses: { deviceId: string }[] }>(`{ licenses { deviceId } }`).then((result) => {
      if (!alive) {
        return;
      }
      const rows = gqlData(result)?.licenses ?? [];
      setBlocked(!rows.some((row) => row.deviceId === device));
    });
    return () => {
      alive = false;
    };
  }, [pathname, signedIn]);

  if (!blocked) {
    return children;
  }

  return (
    <div className="panel narrow" role="dialog" aria-labelledby="license-gate-title">
      <h2 id="license-gate-title">{t("licenseGateTitle")}</h2>
      <p className="muted">{t("licenseGateLead")}</p>
      <p>
        <button
          type="button"
          disabled={busy}
          onClick={() => {
            const headers = desktopDeviceHeaders();
            const deviceId = headers["X-TaleRole-Device"] ?? "";
            if (!deviceId) {
              return;
            }
            setBusy(true);
            void gql<{ registerLicense: { id: string } }>(
              `mutation ($deviceId: String!, $platform: String) {
                registerLicense(deviceId: $deviceId, platform: $platform) { id }
              }`,
              { deviceId, platform: headers["X-TaleRole-Platform"] ?? "" },
            ).then((result) => {
              setBusy(false);
              if (gqlData(result)?.registerLicense?.id) {
                setBlocked(false);
              }
            });
          }}
        >
          {t("registerDevice")}
        </button>
      </p>
      <p>
        <Link href="/account">{t("title")}</Link>
      </p>
    </div>
  );
}
