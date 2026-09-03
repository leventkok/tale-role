"use client";

import { useEffect, useState } from "react";
import QRCode from "qrcode";

export function TotpQr({ value }: { value: string }) {
  const [svg, setSvg] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    void QRCode.toString(value, {
      type: "svg",
      margin: 1,
      width: 192,
      color: { dark: "#1a120c", light: "#ffffff" },
    }).then((out) => {
      if (!cancelled) {
        setSvg(out);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [value]);

  if (!svg) {
    return <div className="totp-qr totp-qr-wait" aria-hidden />;
  }
  return (
    <div
      className="totp-qr"
      role="img"
      aria-label="QR"
      dangerouslySetInnerHTML={{ __html: svg }}
    />
  );
}
