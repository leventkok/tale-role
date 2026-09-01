"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslations } from "next-intl";

const emptyStats = { str: 3, dex: 3, con: 3, int: 3, wis: 3, cha: 3 };
const SKILLS = ["str", "dex", "con", "int", "wis", "cha"] as const;

type Stats = Record<(typeof SKILLS)[number], number>;

type Character = {
  user_id: string;
  name: string;
  hp: number;
  stats?: Stats;
};

type Room = {
  id: string;
  name: string;
  host_id: string;
  dice_system: string;
  started: boolean;
  turn_order: string[];
  presence: { user_id: string; role: string }[];
  characters: Character[];
  turns: {
    actor_id: string;
    kind: string;
    notes?: string;
    rolls?: number[];
    total?: number;
    success?: boolean | null;
  }[];
};

function kindLabel(kind: string, t: (key: "kindPass" | "kindWait" | "kindAction") => string) {
  if (kind === "pass") return t("kindPass");
  if (kind === "wait") return t("kindWait");
  return t("kindAction");
}

export function TableClient({ roomId }: { roomId: string }) {
  const t = useTranslations("table");
  const [room, setRoom] = useState<Room | null>(null);
  const [me, setMe] = useState<string | null>(null);
  const [name, setName] = useState("Adventurer");
  const [stats, setStats] = useState<Stats>(emptyStats);
  const [notes, setNotes] = useState("");
  const [skill, setSkill] = useState<(typeof SKILLS)[number]>("str");
  const [copied, setCopied] = useState(false);
  const [busy, setBusy] = useState(false);
  const logRef = useRef<HTMLDivElement>(null);

  async function refresh() {
    const res = await fetch(`/api/rooms/${roomId}`, { cache: "no-store" });
    if (res.ok) {
      setRoom(await res.json());
    }
  }

  useEffect(() => {
    void fetch("/api/me", { cache: "no-store" })
      .then((r) => (r.ok ? r.json() : null))
      .then((data: { id?: string } | null) => {
        if (data?.id) setMe(data.id);
      });
  }, []);

  useEffect(() => {
    void refresh();
    const id = setInterval(() => void refresh(), 2500);
    return () => clearInterval(id);
  }, [roomId]);

  useEffect(() => {
    const el = logRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, [room?.turns.length]);

  const names = useMemo(() => {
    const map = new Map<string, string>();
    for (const ch of room?.characters ?? []) {
      map.set(ch.user_id, ch.name);
    }
    return map;
  }, [room]);

  const mine = room?.characters.find((c) => c.user_id === me);
  const total = SKILLS.reduce((sum, key) => sum + stats[key], 0);
  const leaked = room?.presence.some((p) => p.role === "system_admin") ?? false;
  const isHost = Boolean(me && room && room.host_id === me);
  const maxHp = mine?.stats ? 8 + mine.stats.con : 14;
  const hpPct = mine ? Math.max(8, Math.min(100, Math.round((mine.hp / maxHp) * 100))) : 0;

  function labelFor(id: string) {
    if (id === me) return names.get(id) ?? t("you");
    return names.get(id) ?? id.slice(0, 8);
  }

  async function saveCharacter(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    await fetch(`/api/rooms/${roomId}/characters`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, stats }),
    });
    setBusy(false);
    await refresh();
  }

  async function start() {
    setBusy(true);
    await fetch(`/api/rooms/${roomId}/start`, { method: "POST" });
    setBusy(false);
    await refresh();
  }

  async function act(kind: string) {
    setBusy(true);
    await fetch(`/api/rooms/${roomId}/turns`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ kind, skill, notes, dc: 12 }),
    });
    setNotes("");
    setBusy(false);
    await refresh();
  }

  async function copyId() {
    await navigator.clipboard.writeText(roomId);
    setCopied(true);
    window.setTimeout(() => setCopied(false), 1600);
  }

  if (!room) {
    return (
      <div className="stage-loading" role="status">
        {t("loading")}
      </div>
    );
  }

  return (
    <section className="stage">
      <aside className="story">
        <div className="room-head">
          <div>
            <h2>{t("storyteller")}</h2>
            <p className="muted" style={{ margin: 0 }}>
              {room.name}
            </p>
          </div>
          <span className="pill">
            {t("dice")} {room.dice_system}
          </span>
        </div>
        <div className="portrait">{t("sceneSoon")}</div>
        <div className="scene">
          <button className="ghost copy" type="button" onClick={() => void copyId()}>
            {copied ? t("copied") : t("copyId")}
          </button>
        </div>
      </aside>

      <div className="log">
        <h2>{t("order")}</h2>
        {room.started && room.turn_order.length > 0 ? (
          <p className="muted" style={{ margin: 0 }}>
            {room.turn_order.map((id) => labelFor(id)).join(" → ")}
          </p>
        ) : (
          <p className="muted" style={{ margin: 0 }}>
            {isHost ? t("statHint") : t("awaitStart")}
          </p>
        )}
        <div className="chronicle" ref={logRef}>
          {room.turns.length === 0 ? (
            <p className="muted">{t("waiting")}</p>
          ) : (
            room.turns.map((turn, idx) => {
              const cls =
                turn.success === true ? "ok" : turn.success === false ? "miss" : "";
              return (
                <article className={`turn ${cls}`} key={`${turn.actor_id}-${idx}`}>
                  <div className="kind">{kindLabel(turn.kind, t)}</div>
                  <strong>{labelFor(turn.actor_id)}</strong>
                  {turn.notes ? <p style={{ margin: "0.35rem 0 0" }}>{turn.notes}</p> : null}
                  {turn.rolls?.length ? (
                    <p className="dice-pips" style={{ margin: "0.4rem 0 0" }}>
                      {turn.rolls.join(" + ")}
                      {typeof turn.total === "number" ? ` → ${turn.total}` : ""}
                      {turn.success === true
                        ? ` · ${t("hit")}`
                        : turn.success === false
                          ? ` · ${t("missLabel")}`
                          : ""}
                    </p>
                  ) : null}
                </article>
              );
            })
          )}
        </div>
        {leaked ? (
          <p className="alert" role="alert">
            admin leak
          </p>
        ) : null}
      </div>

      <aside className="sheet">
        {mine ? (
          <>
            <h2>{mine.name}</h2>
            <div className="hp">
              <span className="muted">
                {t("hp")} {mine.hp}
              </span>
              <div className="hp-bar">
                <span style={{ ["--hp"]: `${hpPct}%` } as React.CSSProperties} />
              </div>
            </div>
            <div className="stats">
              {SKILLS.map((key) => (
                <div className="stat" key={key}>
                  <span className="muted">{key}</span>
                  <b>{mine.stats?.[key] ?? "—"}</b>
                </div>
              ))}
            </div>
          </>
        ) : (
          <form onSubmit={saveCharacter}>
            <h2>{t("character")}</h2>
            <p className="muted">{t("awaitSheet")}</p>
            <label>
              {t("character")}
              <input value={name} onChange={(e) => setName(e.target.value)} required />
            </label>
            <div className="stats">
              {SKILLS.map((key) => (
                <label className="stat" key={key}>
                  {key}
                  <input
                    type="number"
                    min={1}
                    max={6}
                    value={stats[key]}
                    onChange={(e) => setStats({ ...stats, [key]: Number(e.target.value) })}
                  />
                </label>
              ))}
            </div>
            <p className={total === 18 ? "stat-total" : "stat-total bad"}>{t("statTotal", { n: total })}</p>
            <button type="submit" disabled={busy || total !== 18}>
              {t("saveCharacter")}
            </button>
          </form>
        )}

        <h2 style={{ marginTop: "1.25rem" }}>{t("presence")}</h2>
        <ul className="presence">
          {room.presence.map((p) => (
            <li className={p.user_id === me ? "you" : undefined} key={p.user_id}>
              <span>{labelFor(p.user_id)}</span>
              <span className="muted">{p.user_id === room.host_id ? t("gm") : p.role}</span>
            </li>
          ))}
        </ul>
      </aside>

      <div className="act">
        {room.started ? (
          <div className="act-row">
            <label>
              {t("skill")}
              <select value={skill} onChange={(e) => setSkill(e.target.value as (typeof SKILLS)[number])}>
                {SKILLS.map((s) => (
                  <option key={s} value={s}>
                    {s}
                  </option>
                ))}
              </select>
            </label>
            <label>
              {t("notes")}
              <input value={notes} onChange={(e) => setNotes(e.target.value)} placeholder={t("notes")} />
            </label>
            <div className="btn-row">
              <button type="button" disabled={busy || !mine} onClick={() => void act("action")}>
                {t("roll")}
              </button>
              <button className="ghost" type="button" disabled={busy || !mine} onClick={() => void act("pass")}>
                {t("pass")}
              </button>
              <button className="ghost" type="button" disabled={busy || !mine} onClick={() => void act("wait")}>
                {t("wait")}
              </button>
            </div>
          </div>
        ) : (
          <div className="btn-row">
            {isHost ? (
              <button type="button" disabled={busy || room.characters.length === 0} onClick={() => void start()}>
                {t("start")}
              </button>
            ) : (
              <p className="muted" style={{ margin: 0 }}>
                {t("awaitStart")}
              </p>
            )}
          </div>
        )}
      </div>
    </section>
  );
}
