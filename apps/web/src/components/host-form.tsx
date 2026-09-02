"use client";

import { useState } from "react";
import { useRouter } from "@/i18n/routing";
import { useTranslations } from "next-intl";

export function HostForm() {
  const t = useTranslations("table");
  const router = useRouter();
  const [name, setName] = useState("Ashwood");
  const [joinMode, setJoinMode] = useState("link");
  const [password, setPassword] = useState("");
  const [dice, setDice] = useState("d20");
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const res = await fetch("/api/rooms", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, join_mode: joinMode, password, dice_system: dice }),
    });
    const data = await res.json();
    if (!res.ok) {
      setError(data.error ?? "error");
      return;
    }
    router.push(`/table/${data.id}`);
  }

  return (
    <form onSubmit={onSubmit}>
      <label>
        {t("roomName")}
        <input value={name} onChange={(e) => setName(e.target.value)} required />
      </label>
      <label>
        {t("dice")}
        <select value={dice} onChange={(e) => setDice(e.target.value)}>
          <option value="d20">d20</option>
          <option value="2d6">2d6</option>
        </select>
      </label>
      <label>
        {t("joinMode")}
        <select value={joinMode} onChange={(e) => setJoinMode(e.target.value)}>
          <option value="link">{t("joinLink")}</option>
          <option value="password">{t("joinPassword")}</option>
        </select>
      </label>
      {joinMode === "password" ? (
        <label>
          {t("password")}
          <input value={password} onChange={(e) => setPassword(e.target.value)} required />
        </label>
      ) : null}
      {error ? <p role="alert">{error}</p> : null}
      <button type="submit">{t("create")}</button>
    </form>
  );
}
