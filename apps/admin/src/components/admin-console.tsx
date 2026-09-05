"use client";

import { useEffect, useState } from "react";

type Runtime = { adapter_id?: string; prompt_pack?: string; inference?: string };
type Trace = {
  at?: string;
  room_id?: string;
  prompt_pack?: string;
  adapter_id?: string;
  redacted_prompt?: string;
  mechanic_intent?: unknown;
  narrative_excerpt?: string;
};
type Pack = { id: string; en: string; tr: string };
type Lobby = {
  id: string;
  name: string;
  join_mode: string;
  started: boolean;
  completed: boolean;
  seats: number;
};

function lobbyStatus(lobby: Lobby) {
  if (lobby.completed) {
    return { label: "Ended", className: "status ended" };
  }
  if (lobby.started) {
    return { label: "Live", className: "status live" };
  }
  return { label: "Waiting", className: "status waiting" };
}

export function AdminConsole() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [needOtp, setNeedOtp] = useState(false);
  const [needMfa, setNeedMfa] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [runtime, setRuntime] = useState<Runtime | null>(null);
  const [traces, setTraces] = useState<Trace[]>([]);
  const [packs, setPacks] = useState<Pack[]>([]);
  const [lobbies, setLobbies] = useState<Lobby[]>([]);
  const [edit, setEdit] = useState<Pack | null>(null);
  const [busyRoom, setBusyRoom] = useState<string | null>(null);

  async function load() {
    const rt = await fetch("/api/admin/runtime", { cache: "no-store" });
    if (rt.status === 401) {
      setRuntime(null);
      return;
    }
    if (rt.status === 403) {
      setRuntime(null);
      setError("not a spectator");
      return;
    }
    if (rt.ok) {
      setRuntime((await rt.json()) as Runtime);
      const [tr, pk, lb] = await Promise.all([
        fetch("/api/admin/traces", { cache: "no-store" }),
        fetch("/api/admin/packs", { cache: "no-store" }),
        fetch("/api/admin/lobbies", { cache: "no-store" }),
      ]);
      if (tr.ok) {
        const data = (await tr.json()) as { traces?: Trace[] };
        setTraces(data.traces ?? []);
      }
      if (pk.ok) {
        const data = (await pk.json()) as { packs?: Pack[] };
        setPacks(data.packs ?? []);
      }
      if (lb.ok) {
        const data = (await lb.json()) as { lobbies?: Lobby[] };
        setLobbies(data.lobbies ?? []);
      }
    }
  }

  useEffect(() => {
    void load();
  }, []);

  useEffect(() => {
    if (!runtime) {
      return;
    }
    const id = window.setInterval(() => {
      void load();
    }, 4000);
    return () => window.clearInterval(id);
  }, [runtime]);

  async function login(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    const path = needMfa ? "/api/auth/totp/verify" : needOtp ? "/api/auth/otp/verify" : "/api/auth/login";
    const body = needMfa ? { email, password, code } : needOtp ? { email, code } : { email, password };
    const res = await fetch(path, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    const data = (await res.json().catch(() => ({}))) as {
      token?: string;
      otp_required?: boolean;
      mfa_required?: boolean;
      error?: string;
    };
    if (data.token) {
      setError("session leaked");
      return;
    }
    if (data.otp_required) {
      setNeedOtp(true);
      setNeedMfa(false);
      return;
    }
    if (data.mfa_required) {
      setNeedMfa(true);
      setNeedOtp(false);
      setCode("");
      return;
    }
    if (!res.ok) {
      setError(typeof data.error === "string" ? data.error : "forbidden");
      return;
    }
    await load();
  }

  async function swap(pack: string, adapter: string) {
    await fetch("/api/admin/runtime", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ prompt_pack: pack, adapter_id: adapter }),
    });
    await load();
  }

  async function savePack(e: React.FormEvent) {
    e.preventDefault();
    if (!edit) {
      return;
    }
    await fetch("/api/admin/packs", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(edit),
    });
    setEdit(null);
    await load();
  }

  async function closeLobby(lobby: Lobby) {
    if (
      !window.confirm(
        `End "${lobby.name}"? Players will see the tale as closed. Lantern XP is not granted from spectator closure.`,
      )
    ) {
      return;
    }
    setBusyRoom(lobby.id);
    const res = await fetch(`/api/admin/rooms/${encodeURIComponent(lobby.id)}/close`, { method: "POST" });
    setBusyRoom(null);
    if (!res.ok) {
      setError("could not close lobby");
      return;
    }
    setError(null);
    await load();
  }

  if (!runtime) {
    return (
      <form onSubmit={login}>
        <p>Sign in with the spectator email. Players never see this origin.</p>
        <label>
          Email
          <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        </label>
        {needOtp ? (
          <label>
            OTP
            <input inputMode="numeric" maxLength={6} value={code} onChange={(e) => setCode(e.target.value)} required />
          </label>
        ) : (
          <label>
            Password
            <input type="password" minLength={8} value={password} onChange={(e) => setPassword(e.target.value)} required />
          </label>
        )}
        {needMfa ? (
          <label>
            Authenticator
            <input inputMode="numeric" maxLength={6} value={code} onChange={(e) => setCode(e.target.value)} required />
          </label>
        ) : null}
        {error ? <p className="alert">{error}</p> : null}
        <button type="submit">Continue</button>
      </form>
    );
  }

  const pack = runtime.prompt_pack || "v1";
  const adapter = runtime.adapter_id === "hub" ? "hub" : "stub";
  const openLobbies = lobbies.filter((lobby) => !lobby.completed);
  const endedLobbies = lobbies.filter((lobby) => lobby.completed);

  return (
    <div>
      <section className="panel">
        <h2>Runtime</h2>
        <div className="runtime-bar">
          <span>
            Adapter <code>{runtime.adapter_id}</code>
          </span>
          <span>
            Pack <code>{runtime.prompt_pack}</code>
          </span>
          <span>
            Inference <code>{runtime.inference}</code>
          </span>
        </div>
        <div className="row">
          <button type="button" onClick={() => void swap("v1", adapter)}>
            Use v1
          </button>
          <button type="button" onClick={() => void swap("v1-terse", adapter)}>
            Use v1-terse
          </button>
          <button type="button" onClick={() => void swap(pack, "hub")}>
            Adapter hub
          </button>
          <button type="button" onClick={() => void swap(pack, "stub")}>
            Adapter stub
          </button>
        </div>
      </section>

      <section className="panel">
        <h2>Live lobbies</h2>
        <p className="hint">Open tables refresh every 4s. Ending a lobby closes it for players without granting lantern XP.</p>
        {openLobbies.length === 0 ? <p className="empty">No open lobbies.</p> : null}
        <ul className="lobby-list">
          {openLobbies.map((lobby) => {
            const status = lobbyStatus(lobby);
            return (
              <li key={lobby.id}>
                <div className="lobby-head">
                  <div>
                    <strong>{lobby.name}</strong>
                    <div className="lobby-meta">
                      <span className={status.className}>{status.label}</span>
                      <span>{lobby.seats} seated</span>
                      <span>{lobby.join_mode}</span>
                      <span>
                        <code>{lobby.id}</code>
                      </span>
                    </div>
                  </div>
                  <button
                    type="button"
                    className="danger"
                    disabled={busyRoom === lobby.id}
                    onClick={() => void closeLobby(lobby)}
                  >
                    End lobby
                  </button>
                </div>
              </li>
            );
          })}
        </ul>
        {endedLobbies.length > 0 ? (
          <details className="ended-fold">
            <summary>
              Recently ended
              <span className="ended-count">{endedLobbies.length}</span>
            </summary>
            <ul className="lobby-list">
              {endedLobbies.map((lobby) => (
                <li key={lobby.id}>
                  <strong>{lobby.name}</strong>
                  <div className="lobby-meta">
                    <span className="status ended">Ended</span>
                    <span>{lobby.seats} seated</span>
                    <span>
                      <code>{lobby.id}</code>
                    </span>
                  </div>
                </li>
              ))}
            </ul>
          </details>
        ) : null}
      </section>

      <section className="panel">
        <h2>Traces</h2>
        <p className="hint">Mechanic JSON stays off the player table. Next turn uses the pack text below.</p>
        <ul className="traces">
          {traces.length === 0 ? <li className="empty">No turns yet.</li> : null}
          {traces
            .slice()
            .reverse()
            .map((tr, idx) => (
              <li key={`${tr.at}-${idx}`}>
                <strong>
                  {tr.prompt_pack} / {tr.adapter_id}
                </strong>
                <div>{tr.narrative_excerpt || tr.redacted_prompt}</div>
                {tr.mechanic_intent ? <pre>{JSON.stringify(tr.mechanic_intent, null, 2)}</pre> : null}
              </li>
            ))}
        </ul>
      </section>

      <section className="panel">
        <h2>Pack context</h2>
        <ul className="pack-list">
          {packs.map((p) => (
            <li key={p.id}>
              <div className="pack-actions">
                <button type="button" className="ghost" onClick={() => setEdit({ ...p })}>
                  Edit {p.id}
                </button>
              </div>
              <pre>{p.en}</pre>
            </li>
          ))}
        </ul>
        {edit ? (
          <form onSubmit={savePack}>
            <p>
              Editing <code>{edit.id}</code>
            </p>
            <label>
              English
              <textarea rows={4} value={edit.en} onChange={(e) => setEdit({ ...edit, en: e.target.value })} />
            </label>
            <label>
              Turkish
              <textarea rows={4} value={edit.tr} onChange={(e) => setEdit({ ...edit, tr: e.target.value })} />
            </label>
            <div className="row">
              <button type="submit">Save pack</button>
              <button type="button" className="ghost" onClick={() => setEdit(null)}>
                Cancel
              </button>
            </div>
          </form>
        ) : null}
      </section>

      {error ? <p className="alert">{error}</p> : null}

      <footer className="console-footer">
        <form action="/api/auth/logout" method="post">
          <button type="submit" className="ghost">
            Sign out
          </button>
        </form>
      </footer>
    </div>
  );
}
