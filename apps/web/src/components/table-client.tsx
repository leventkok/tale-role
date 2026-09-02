"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";

const emptyStats = { str: 3, dex: 3, con: 3, int: 3, wis: 3, cha: 3 };

type Room = {
  id: string;
  name: string;
  host_id: string;
  dice_system: string;
  started: boolean;
  turn_order: string[];
  presence: { user_id: string; role: string }[];
  characters: { user_id: string; name: string; hp: number }[];
  turns: { actor_id: string; kind: string; rolls?: number[]; total?: number; success?: boolean | null }[];
};

export function TableClient({ roomId }: { roomId: string }) {
  const t = useTranslations("table");
  const [room, setRoom] = useState<Room | null>(null);
  const [name, setName] = useState("Adventurer");
  const [stats, setStats] = useState(emptyStats);
  const [notes, setNotes] = useState("");
  const [skill, setSkill] = useState("str");

  async function refresh() {
    const res = await fetch(`/api/rooms/${roomId}`, { cache: "no-store" });
    if (res.ok) {
      setRoom(await res.json());
    }
  }

  useEffect(() => {
    void refresh();
    const id = setInterval(() => void refresh(), 2500);
    return () => clearInterval(id);
  }, [roomId]);

  async function saveCharacter(e: React.FormEvent) {
    e.preventDefault();
    await fetch(`/api/rooms/${roomId}/characters`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, stats }),
    });
    await refresh();
  }

  async function start() {
    await fetch(`/api/rooms/${roomId}/start`, { method: "POST" });
    await refresh();
  }

  async function act(kind: string) {
    await fetch(`/api/rooms/${roomId}/turns`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ kind, skill, notes, dc: 12 }),
    });
    await refresh();
  }

  if (!room) {
    return <p>…</p>;
  }

  const leaked = room.presence.some((p) => p.role === "system_admin");

  return (
    <section>
      <h1>{room.name}</h1>
      <p>
        {t("dice")}: {room.dice_system} · id {room.id}
      </p>
      {leaked ? <p role="alert">admin leak</p> : null}
      <h2>{t("presence")}</h2>
      <ul>
        {room.presence.map((p) => (
          <li key={p.user_id}>
            {p.role} · {p.user_id.slice(0, 8)}
          </li>
        ))}
      </ul>
      <form onSubmit={saveCharacter}>
        <label>
          {t("character")}
          <input value={name} onChange={(e) => setName(e.target.value)} required />
        </label>
        <p>{t("statHint")}</p>
        {(["str", "dex", "con", "int", "wis", "cha"] as const).map((key) => (
          <label key={key}>
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
        <button type="submit">{t("saveCharacter")}</button>
      </form>
      {!room.started ? (
        <button type="button" onClick={() => void start()}>
          {t("start")}
        </button>
      ) : (
        <>
          <p>
            {t("order")}: {room.turn_order.map((id) => id.slice(0, 8)).join(" → ")}
          </p>
          <label>
            {t("skill")}
            <select value={skill} onChange={(e) => setSkill(e.target.value)}>
              {["str", "dex", "con", "int", "wis", "cha"].map((s) => (
                <option key={s}>{s}</option>
              ))}
            </select>
          </label>
          <label>
            {t("notes")}
            <input value={notes} onChange={(e) => setNotes(e.target.value)} />
          </label>
          <button type="button" onClick={() => void act("action")}>
            {t("roll")}
          </button>
          <button type="button" onClick={() => void act("pass")}>
            {t("pass")}
          </button>
          <button type="button" onClick={() => void act("wait")}>
            {t("wait")}
          </button>
          <ul>
            {room.turns.map((turn, idx) => (
              <li key={idx}>
                {turn.kind} {turn.rolls?.join(",")} {turn.total ?? ""}{" "}
                {turn.success === true ? "ok" : turn.success === false ? "miss" : ""}
              </li>
            ))}
          </ul>
        </>
      )}
    </section>
  );
}
