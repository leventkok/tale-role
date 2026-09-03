"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { useRouter } from "@/i18n/routing";
import { gql, gqlData } from "@/lib/gql";

type Lobby = {
  id: string;
  name: string;
  joinMode: string;
  started: boolean;
  seats: number;
};

export function LobbyBoard() {
  const t = useTranslations("lobby");
  const table = useTranslations("table");
  const router = useRouter();
  const [rows, setRows] = useState<Lobby[] | null>(null);
  const [invite, setInvite] = useState("");
  const [passwords, setPasswords] = useState<Record<string, string>>({});
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState<string | null>(null);

  async function refresh() {
    const result = await gql<{ lobbies: Lobby[] }>(`{ lobbies { id name joinMode started seats } }`);
    setRows(gqlData(result)?.lobbies ?? []);
  }

  useEffect(() => {
    void refresh();
    const id = setInterval(() => void refresh(), 4000);
    return () => clearInterval(id);
  }, []);

  async function sit(id: string, password?: string) {
    setBusy(id);
    setError(null);
    const result = await gql<{ joinRoom: boolean }>(
      `mutation ($roomId: ID!, $password: String) { joinRoom(roomId: $roomId, password: $password) }`,
      { roomId: id, password: password || null },
    );
    setBusy(null);
    if (!gqlData(result)?.joinRoom) {
      setError(table("joinForbidden"));
      return;
    }
    router.push(`/table/${id}`);
  }

  const privateRoom = (mode: string) => mode === "password";

  return (
    <div className="lobby">
      {rows === null ? (
        <p className="muted">{table("loading")}</p>
      ) : rows.length === 0 ? (
        <p className="muted">{t("empty")}</p>
      ) : (
        <ul className="lobby-list">
          {rows.map((row) => (
            <li className="card lobby-card" key={row.id}>
              <div>
                <h2>{row.name}</h2>
                <p className="muted">
                  <span className="pill">{privateRoom(row.joinMode) ? table("private") : table("public")}</span>{" "}
                  {row.started ? table("inPlay") : table("gathering")} · {table("seats", { n: row.seats })}
                </p>
              </div>
              {privateRoom(row.joinMode) ? (
                <div className="lobby-join">
                  <label>
                    {table("needPassword")}
                    <input
                      type="password"
                      value={passwords[row.id] ?? ""}
                      onChange={(e) => setPasswords((prev) => ({ ...prev, [row.id]: e.target.value }))}
                    />
                  </label>
                  <button
                    type="button"
                    disabled={busy === row.id || !(passwords[row.id] ?? "").trim()}
                    onClick={() => void sit(row.id, passwords[row.id])}
                  >
                    {table("join")}
                  </button>
                </div>
              ) : (
                <button type="button" disabled={busy === row.id} onClick={() => void sit(row.id)}>
                  {table("join")}
                </button>
              )}
            </li>
          ))}
        </ul>
      )}

      <form
        className="panel lobby-invite"
        onSubmit={(e) => {
          e.preventDefault();
          const id = invite.trim();
          if (id) void sit(id);
        }}
      >
        <h2>{t("invite")}</h2>
        <p className="muted">{t("inviteHint")}</p>
        <label>
          {table("roomId")}
          <input value={invite} onChange={(e) => setInvite(e.target.value)} />
        </label>
        <button type="submit" disabled={!invite.trim()}>
          {table("join")}
        </button>
      </form>
      {error ? (
        <p className="alert" role="alert">
          {error}
        </p>
      ) : null}
    </div>
  );
}
