"use client";

import { useState } from "react";
import { useRouter } from "@/i18n/routing";
import { useTranslations } from "next-intl";
import { gql, gqlData } from "@/lib/gql";

export function JoinForm({ initialRoomId = "" }: { initialRoomId?: string }) {
  const t = useTranslations("table");
  const router = useRouter();
  const [roomId, setRoomId] = useState(initialRoomId);
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const result = await gql<{ joinRoom: boolean }>(
      `mutation ($roomId: ID!, $password: String) { joinRoom(roomId: $roomId, password: $password) }`,
      { roomId, password: password || null },
    );
    if (!gqlData(result)?.joinRoom) {
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
