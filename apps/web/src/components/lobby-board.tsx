"use client";

import { useEffect, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { useRouter } from "@/i18n/routing";
import { gql, gqlData } from "@/lib/gql";

const TABLE_MAX_SEATS = 8;

type Lobby = {
  id: string;
  name: string;
  universeName?: string | null;
  joinMode: string;
  started: boolean;
  seats: number;
  startedAt?: string | null;
};

function formatDuration(startedAt: string | null | undefined, nowMs: number) {
  if (!startedAt) {
    return null;
  }
  const start = Date.parse(startedAt);
  if (Number.isNaN(start)) {
    return null;
  }
  const secs = Math.max(0, Math.floor((nowMs - start) / 1000));
  const h = Math.floor(secs / 3600);
  const m = Math.floor((secs % 3600) / 60);
  const s = secs % 60;
  if (h > 0) {
    return `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
  }
  return `${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}

export function LobbyBoard() {
  const t = useTranslations("lobby");
  const table = useTranslations("table");
  const locale = useLocale();
  const router = useRouter();
  const [rows, setRows] = useState<Lobby[] | null>(null);
  const [invite, setInvite] = useState("");
  const [passwords, setPasswords] = useState<Record<string, string>>({});
  const [expanded, setExpanded] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState<string | null>(null);
  const [now, setNow] = useState(() => Date.now());

  async function refresh() {
    const result = await gql<{ lobbies: Lobby[] }>(
      `query ($locale: String) { lobbies(locale: $locale) { id name universeName joinMode started seats startedAt } }`,
      { locale },
    );
    setRows(gqlData(result)?.lobbies ?? []);
  }

  useEffect(() => {
    void refresh();
    const id = setInterval(() => void refresh(), 4000);
    return () => clearInterval(id);
  }, [locale]);

  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000);
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

  function openRow(row: Lobby) {
    if (row.joinMode === "password") {
      setExpanded((cur) => (cur === row.id ? null : row.id));
      return;
    }
    void sit(row.id);
  }

  return (
    <div className="lobby">
      {rows === null ? (
        <p className="muted">{table("loading")}</p>
      ) : rows.length === 0 ? (
        <p className="muted">{t("empty")}</p>
      ) : (
        <div className="panel lobby-browser">
          <div className="lobby-table-head" aria-hidden="true">
            <span>{t("colUniverse")}</span>
            <span>{t("colStatus")}</span>
            <span>{t("colPlayers")}</span>
            <span>{t("colDuration")}</span>
          </div>
          <ul className="lobby-table">
            {rows.map((row) => {
              const duration = row.started ? formatDuration(row.startedAt, now) : null;
              const isPrivate = row.joinMode === "password";
              return (
                <li key={row.id} className={expanded === row.id ? "expanded" : undefined}>
                  <button
                    type="button"
                    className="lobby-row"
                    disabled={busy === row.id}
                    onClick={() => openRow(row)}
                  >
                    <span className="lobby-universe">
                      <strong>{row.universeName || row.name}</strong>
                      {row.universeName && row.universeName !== row.name ? (
                        <span className="muted lobby-room-name">{row.name}</span>
                      ) : null}
                    </span>
                    <span className={`lobby-status ${row.started ? "live" : "waiting"}`}>
                      {row.started ? table("inPlay") : table("gathering")}
                    </span>
                    <span className="lobby-players">
                      {row.seats} / {TABLE_MAX_SEATS}
                    </span>
                    <span className="lobby-duration">{duration ?? t("noDuration")}</span>
                  </button>
                  {isPrivate && expanded === row.id ? (
                    <div className="lobby-private">
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
                  ) : null}
                </li>
              );
            })}
          </ul>
        </div>
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
