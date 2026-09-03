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

export function AdminConsole() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [needOtp, setNeedOtp] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [runtime, setRuntime] = useState<Runtime | null>(null);
  const [traces, setTraces] = useState<Trace[]>([]);
  const [packs, setPacks] = useState<Pack[]>([]);
  const [edit, setEdit] = useState<Pack | null>(null);

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
      const [tr, pk] = await Promise.all([
        fetch("/api/admin/traces", { cache: "no-store" }),
        fetch("/api/admin/packs", { cache: "no-store" }),
      ]);
      if (tr.ok) {
        const data = (await tr.json()) as { traces?: Trace[] };
        setTraces(data.traces ?? []);
      }
      if (pk.ok) {
        const data = (await pk.json()) as { packs?: Pack[] };
        setPacks(data.packs ?? []);
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
    const path = needOtp ? "/api/auth/otp/verify" : "/api/auth/login";
    const body = needOtp ? { email, code } : { email, password };
    const res = await fetch(path, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    const data = (await res.json().catch(() => ({}))) as { token?: string; otp_required?: boolean; error?: string };
    if (data.token) {
      setError("session leaked");
      return;
    }
    if (data.otp_required) {
      setNeedOtp(true);
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
        {error ? <p className="alert">{error}</p> : null}
        <button type="submit">Continue</button>
      </form>
    );
  }

  const pack = runtime.prompt_pack || "v1";
  const adapter = runtime.adapter_id === "hub" ? "hub" : "stub";

  return (
    <div>
      <p>
        Adapter <code>{runtime.adapter_id}</code> · pack <code>{runtime.prompt_pack}</code> · inference{" "}
        <code>{runtime.inference}</code>
      </p>
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
      <p>Live traces refresh every 4s. Mechanic JSON stays off the player table. Next turn uses the pack text below.</p>
      <ul className="traces">
        {traces.length === 0 ? <li>No turns yet.</li> : null}
        {traces
          .slice()
          .reverse()
          .map((tr, idx) => (
            <li key={`${tr.at}-${idx}`}>
              <strong>
                {tr.prompt_pack} / {tr.adapter_id}
              </strong>
              <div>{tr.narrative_excerpt || tr.redacted_prompt}</div>
              {tr.mechanic_intent ? <pre>{JSON.stringify(tr.mechanic_intent)}</pre> : null}
            </li>
          ))}
      </ul>
      <h2>Pack context</h2>
      <ul>
        {packs.map((p) => (
          <li key={p.id}>
            <button type="button" className="ghost" onClick={() => setEdit({ ...p })}>
              Edit {p.id}
            </button>
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
          <button type="submit">Save pack</button>
        </form>
      ) : null}
      <form action="/api/auth/logout" method="post">
        <button type="submit">Sign out</button>
      </form>
    </div>
  );
}
