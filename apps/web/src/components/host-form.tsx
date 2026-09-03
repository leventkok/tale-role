"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Link, useRouter } from "@/i18n/routing";

import { gql, gqlData } from "@/lib/gql";

type Uni = { id: string; name_en: string; theme_id: string; dice_system: string };

export function HostForm({ universeId }: { universeId?: string }) {
  const t = useTranslations("table");
  const u = useTranslations("universe");
  const router = useRouter();
  const [name, setName] = useState("");
  const [joinMode, setJoinMode] = useState("link");
  const [password, setPassword] = useState("");
  const [dice, setDice] = useState("d20");
  const [universe, setUniverse] = useState(universeId ?? "");
  const [catalog, setCatalog] = useState<Uni[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    void gql<{ universes: { id: string; nameEn: string; themeId: string; diceSystem: string }[] }>(
      `{ universes { id nameEn themeId diceSystem } }`,
    ).then((result) => {
      const rows = (gqlData(result)?.universes ?? []).map((row) => ({
        id: row.id,
        name_en: row.nameEn,
        theme_id: row.themeId,
        dice_system: row.diceSystem,
      }));
      setCatalog(rows);
      if (universeId && rows.some((row) => row.id === universeId)) {
        setUniverse(universeId);
      }
    });
  }, [universeId]);

  useEffect(() => {
    const selected = catalog.find((row) => row.id === universe);
    if (selected) {
      setDice(selected.dice_system);
      if (!name) setName(selected.name_en);
    }
  }, [universe, catalog, name]);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const result = await gql<{ createRoom: { id: string } }>(
      `mutation ($name: String, $joinMode: String, $password: String, $diceSystem: String, $universeId: ID) {
        createRoom(name: $name, joinMode: $joinMode, password: $password, diceSystem: $diceSystem, universeId: $universeId) { id }
      }`,
      {
        name,
        joinMode,
        password: password || null,
        diceSystem: dice,
        universeId: universe || null,
      },
    );
    const created = gqlData(result)?.createRoom;
    if (!created?.id) {
      setError(result.errors?.[0]?.message ?? "error");
      return;
    }
    router.push(`/table/${created.id}`);
  }

  return (
    <form onSubmit={onSubmit}>
      <label>
        {u("pick")}
        <select value={universe} onChange={(e) => setUniverse(e.target.value)}>
          <option value="">{u("none")}</option>
          {catalog.map((row) => (
            <option key={row.id} value={row.id}>
              {row.name_en} · {row.theme_id}
            </option>
          ))}
        </select>
      </label>
      <p className="muted">
        <Link href="/universe/new">{u("create")}</Link>
      </p>
      <label>
        {t("roomName")}
        <input value={name} onChange={(e) => setName(e.target.value)} required={universe === ""} />
      </label>
      <label>
        {t("dice")}
        <select value={dice} onChange={(e) => setDice(e.target.value)} disabled={Boolean(universe)}>
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
      {error ? (
        <p className="alert" role="alert">
          {error}
        </p>
      ) : null}
      <button type="submit">{t("create")}</button>
    </form>
  );
}
