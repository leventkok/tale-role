"use client";

import { useState } from "react";
import { useRouter } from "@/i18n/routing";
import { useTranslations } from "next-intl";

export function JoinForm({ initialRoomId = "" }: { initialRoomId?: string }) {
  const t = useTranslations("table");
  const router = useRouter();
  const [roomId, setRoomId] = useState(initialRoomId);
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const join = await fetch(`/api/rooms/${roomId}/join`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ password }),
    });
    if (!join.ok) {
      setError("forbidden");
      return;
    }
    router.push(`/table/${roomId}`);
  }

  return (
    <form onSubmit={onSubmit}>
      <label>
        {t("roomId")}
        <input value={roomId} onChange={(e) => setRoomId(e.target.value)} required />
      </label>
      <label>
        {t("password")}
        <input value={password} onChange={(e) => setPassword(e.target.value)} />
      </label>
      {error ? (
        <p className="alert" role="alert">
          {error}
        </p>
      ) : null}
      <button type="submit">{t("join")}</button>
    </form>
  );
}
