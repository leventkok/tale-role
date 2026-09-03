"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { ROOM_QUERY, gql, gqlData, mapRoom, type TableRoom } from "@/lib/gql";
import { DiceDrop, PortraitMark, TypeProse } from "@/components/art/table-motion";
import { readStoredPortrait, type PortraitId } from "@/lib/portraits";

const emptyStats = { str: 3, dex: 3, con: 3, int: 3, wis: 3, cha: 3 };
const ABILITIES = ["str", "dex", "con", "int", "wis", "cha"] as const;
const TALE_SKILLS = [
  "athletics",
  "acrobatics",
  "stealth",
  "arcana",
  "investigation",
  "history",
  "perception",
  "insight",
  "survival",
  "persuasion",
  "deception",
  "intimidation",
] as const;

type Stats = Record<(typeof ABILITIES)[number], number>;
type TaleSkill = (typeof TALE_SKILLS)[number];

function kindLabel(
  kind: string,
  t: (key: "kindPass" | "kindWait" | "kindAction" | "kindSpeak" | "kindStory" | "kindInit") => string,
) {
  if (kind === "pass") return t("kindPass");
  if (kind === "wait") return t("kindWait");
  if (kind === "say") return t("kindSpeak");
  if (kind === "story") return t("kindStory");
  if (kind === "init") return t("kindInit");
  return t("kindAction");
}

export function TableClient({ roomId }: { roomId: string }) {
  const t = useTranslations("table");
  const skillsT = useTranslations("skills");
  const locale = useLocale();
  const [room, setRoom] = useState<TableRoom | null>(null);
  const [me, setMe] = useState<string | null>(null);
  const [myPortrait, setMyPortrait] = useState<PortraitId>("warden");
  const [name, setName] = useState("");
  const [species, setSpecies] = useState("");
  const [path, setPath] = useState("");
  const [backstory, setBackstory] = useState("");
  const [picks, setPicks] = useState<string[]>([]);
  const [stats, setStats] = useState<Stats>(emptyStats);
  const [notes, setNotes] = useState("");
  const [skill, setSkill] = useState<string>("athletics");
  const [copied, setCopied] = useState(false);
  const [busy, setBusy] = useState(false);
  const logRef = useRef<HTMLDivElement>(null);

  async function refresh() {
    const result = await gql<{ room: Parameters<typeof mapRoom>[0] | null }>(ROOM_QUERY, { id: roomId });
    const data = gqlData(result);
    if (data?.room) {
      setRoom(mapRoom(data.room));
    }
  }

  useEffect(() => {
    void gql<{ me: { id: string; portraitId?: string } | null }>("{ me { id portraitId } }").then((result) => {
      const data = gqlData(result);
      if (data?.me?.id) setMe(data.me.id);
      if (data?.me?.portraitId) setMyPortrait(data.me.portraitId as PortraitId);
      else setMyPortrait(readStoredPortrait());
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
  const total = ABILITIES.reduce((sum, key) => sum + stats[key], 0);
  const leaked = room?.presence.some((p) => p.role === "system_admin") ?? false;
  const isHost = Boolean(me && room && room.host_id === me);
  const maxHp = 8 + (mine?.stats?.con ?? 3) + (mine?.level ?? 1);
  const hpPct = mine ? Math.max(8, Math.min(100, Math.round((mine.hp / maxHp) * 100))) : 0;
  const xpNeed = 100 * (mine?.level ?? 1);
  const xpPct = mine ? Math.max(4, Math.min(100, Math.round(((mine.xp ?? 0) / xpNeed) * 100))) : 0;
  const myTurn = Boolean(room?.started && me && room.current_actor_id === me);
  const allSeated = Boolean(room && room.characters.length > 0 && room.characters.every((c) => c.has_initiative));

  function labelFor(id: string) {
    if (id === "storyteller") return t("storyteller");
    if (id === me) return names.get(id) ?? t("you");
    return names.get(id) ?? id.slice(0, 8);
  }

  function toggleSkill(id: string) {
    setPicks((prev) => {
      if (prev.includes(id)) {
        return prev.filter((row) => row !== id);
      }
      if (prev.length >= 3) {
        return prev;
      }
      return [...prev, id];
    });
  }

  async function saveCharacter(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    await gql(
      `mutation ($roomId: ID!, $name: String!, $stats: StatsInput!, $species: String, $path: String, $backstory: String, $skills: [String!]) {
        setCharacter(roomId: $roomId, name: $name, stats: $stats, species: $species, path: $path, backstory: $backstory, skills: $skills)
      }`,
      { roomId, name, stats, species, path, backstory, skills: picks },
    );
    setBusy(false);
    await refresh();
  }

  async function start() {
    setBusy(true);
    await gql(`mutation ($roomId: ID!, $locale: String) { startRoom(roomId: $roomId, locale: $locale) }`, {
      roomId,
      locale,
    });
    setBusy(false);
    await refresh();
  }

  async function rollInit() {
    setBusy(true);
    await gql(`mutation ($roomId: ID!) { rollInitiative(roomId: $roomId) }`, { roomId });
    setBusy(false);
    await refresh();
  }

  async function act(kind: string) {
    setBusy(true);
    await gql(
      `mutation ($roomId: ID!, $kind: String, $skill: String, $notes: String, $locale: String) {
        actTurn(roomId: $roomId, kind: $kind, skill: $skill, notes: $notes, dc: 12, locale: $locale) { kind total }
      }`,
      { roomId, kind, skill, notes, locale },
    );
    setNotes("");
    setBusy(false);
    await refresh();
  }

  async function copyId() {
    const web = `${window.location.origin}/${locale}/join/${roomId}`;
    const app = `talerole://join/${roomId}`;
    await navigator.clipboard.writeText(`${web}\n${app}`);
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
      {myTurn ? (
        <p className="your-turn" role="status">
          {t("yourTurn")}
        </p>
      ) : null}
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
        {room.scene?.image_svg ? (
          <div className="portrait has-art">
            <img
              alt=""
              src={`data:image/svg+xml;charset=utf-8,${encodeURIComponent(room.scene.image_svg)}`}
            />
            <p>{t("sceneStub")}</p>
          </div>
        ) : (
          <div className="portrait">{t("sceneSoon")}</div>
        )}
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
            {t("initHint")}
          </p>
        )}
        <div className="chronicle" ref={logRef}>
          {room.turns.length === 0 ? (
            <p className="muted">{t("waiting")}</p>
          ) : (
            room.turns.map((turn, idx) => {
              const cls =
                turn.success === true ? "ok" : turn.success === false ? "miss" : "";
              const last = idx === room.turns.length - 1;
              return (
                <article className={`turn ${cls}`} key={`${turn.actor_id}-${idx}`}>
                  <div className="kind">{kindLabel(turn.kind, t)}</div>
                  <strong>{labelFor(turn.actor_id)}</strong>
                  {turn.notes ? <p style={{ margin: "0.35rem 0 0" }}>{turn.notes}</p> : null}
                  {turn.narrative?.prose ? (
                    last && turn.kind === "story" ? (
                      <TypeProse text={turn.narrative.prose} />
                    ) : (
                      <p className="prose">{turn.narrative.prose}</p>
                    )
                  ) : null}
                  {turn.rolls?.length ? (
                    last ? (
                      <DiceDrop rolls={turn.rolls} />
                    ) : (
                      <p className="dice-pips" style={{ margin: "0.4rem 0 0" }}>
                        {turn.rolls.join(" + ")}
                        {typeof turn.total === "number" ? ` → ${turn.total}` : ""}
                      </p>
                    )
                  ) : null}
                  {turn.success === true ? ` · ${t("hit")}` : turn.success === false ? ` · ${t("missLabel")}` : ""}
                </article>
              );
            })
          )}
        </div>
        {leaked ? (
          <p className="alert" role="alert">
            {t("adminLeak")}
          </p>
        ) : null}
      </div>

      <aside className="sheet">
        {mine ? (
          <article className="profile-card you">
            <PortraitMark name={mine.name} you portraitId={myPortrait} />
            <div>
              <h2>{mine.name}</h2>
              {[mine.path, mine.species].filter(Boolean).length ? (
                <p className="muted">{[mine.path, mine.species].filter(Boolean).join(" · ")}</p>
              ) : null}
              <p className="muted">{t("level", { n: mine.level ?? 1 })}</p>
              {mine.backstory ? <p className="sheet-story">{mine.backstory}</p> : null}
              <div className="hp">
                <span className="muted">
                  {t("hp")} {mine.hp}
                </span>
                <div className="hp-bar">
                  <span style={{ ["--hp"]: `${hpPct}%` } as React.CSSProperties} />
                </div>
              </div>
              <div className="hp">
                <span className="muted">
                  {t("xp")} {mine.xp ?? 0} / {xpNeed}
                </span>
                <div className="hp-bar xp">
                  <span style={{ ["--hp"]: `${xpPct}%` } as React.CSSProperties} />
                </div>
              </div>
              <div className="stats">
                {ABILITIES.map((key) => (
                  <div className="stat" key={key}>
                    <span className="muted">{key}</span>
                    <b>{mine.stats?.[key] ?? "—"}</b>
                  </div>
                ))}
              </div>
              {mine.skills?.length ? (
                <p className="muted skill-pills">
                  {mine.skills.map((id) => skillsT(id as TaleSkill)).join(" · ")}
                </p>
              ) : null}
            </div>
          </article>
        ) : (
          <form className="char-sheet" onSubmit={saveCharacter}>
            <h2>{t("character")}</h2>
            <p className="muted">{t("awaitSheet")}</p>
            <label>
              {t("character")}
              <input value={name} onChange={(e) => setName(e.target.value)} required />
            </label>
            <label>
              {t("species")}
              <span className="hint">{t("speciesHint")}</span>
              <input value={species} onChange={(e) => setSpecies(e.target.value)} />
            </label>
            <label>
              {t("path")}
              <span className="hint">{t("pathHint")}</span>
              <input value={path} onChange={(e) => setPath(e.target.value)} />
            </label>
            <label>
              {t("backstory")}
              <span className="hint">{t("backstoryHint")}</span>
              <textarea value={backstory} onChange={(e) => setBackstory(e.target.value)} rows={4} />
            </label>
            <div className="stats">
              {ABILITIES.map((key) => (
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
            <fieldset className="skill-picks">
              <legend>
                {t("skills")} · {t("skillsCount", { n: picks.length })}
              </legend>
              <p className="hint">{t("skillsHint")}</p>
              <div className="skill-grid">
                {TALE_SKILLS.map((id) => (
                  <label key={id} className={picks.includes(id) ? "on" : undefined}>
                    <input
                      type="checkbox"
                      checked={picks.includes(id)}
                      onChange={() => toggleSkill(id)}
                    />
                    {skillsT(id)}
                  </label>
                ))}
              </div>
            </fieldset>
            <button type="submit" disabled={busy || total !== 18 || picks.length !== 3}>
              {t("saveCharacter")}
            </button>
          </form>
        )}

        <h2 style={{ marginTop: "1.25rem" }}>{t("presence")}</h2>
        <ul className="roster">
          {room.presence.map((p) => {
            const ch = room.characters.find((c) => c.user_id === p.user_id);
            return (
              <li className={`profile-card mini ${p.user_id === me ? "you" : ""}`} key={p.user_id}>
                  <PortraitMark
                    name={ch?.name ?? p.user_id}
                    you={p.user_id === me}
                    portraitId={p.user_id === me ? myPortrait : undefined}
                  />
                <div>
                  <strong>{labelFor(p.user_id)}</strong>
                  <span className="muted">{p.user_id === room.host_id ? t("gm") : t("level", { n: ch?.level ?? 1 })}</span>
                </div>
              </li>
            );
          })}
        </ul>
      </aside>

      <div className="act">
        {room.started ? (
          myTurn ? (
            <div className="act-row">
              <label>
                {t("skill")}
                <select value={skill} onChange={(e) => setSkill(e.target.value)}>
                  {TALE_SKILLS.map((s) => (
                    <option key={s} value={s}>
                      {skillsT(s)}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                {t("notes")}
                <input value={notes} onChange={(e) => setNotes(e.target.value)} placeholder={t("notes")} />
              </label>
              <div className="btn-row">
                <button type="button" disabled={busy || !mine || !notes.trim()} onClick={() => void act("say")}>
                  {t("speak")}
                </button>
                <button type="button" disabled={busy || !mine} onClick={() => void act("action")}>
                  {t("roll")}
                </button>
                <button className="ghost" type="button" disabled={busy || !mine} onClick={() => void act("pass")}>
                  {t("pass")}
                </button>
              </div>
            </div>
          ) : (
            <p className="muted" style={{ margin: 0 }}>
              {t("waitTurn")}
            </p>
          )
        ) : (
          <div className="btn-row">
            {mine && !mine.has_initiative ? (
              <button type="button" disabled={busy} onClick={() => void rollInit()}>
                {t("rollInit")}
              </button>
            ) : null}
            {mine?.has_initiative && !allSeated ? <p className="muted">{t("awaitInit")}</p> : null}
            {isHost && allSeated ? (
              <button type="button" disabled={busy} onClick={() => void start()}>
                {t("start")}
              </button>
            ) : null}
            {!mine ? <p className="muted">{t("awaitSheet")}</p> : null}
          </div>
        )}
      </div>
    </section>
  );
}
