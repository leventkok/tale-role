import { NextResponse } from "next/server";
import { apiFetch, SESSION_COOKIE, sessionCookieOptions } from "@/lib/session";

type Upstream = Record<string, unknown>;

function withOptionalSession(status: number, data: Upstream) {
  const payload = { ...data };
  const token = payload.token;
  delete payload.token;
  const res = NextResponse.json(payload, { status });
  if (typeof token === "string" && token.length > 0) {
    res.cookies.set(SESSION_COOKIE, token, sessionCookieOptions());
  }
  return res;
}

export async function POST(request: Request) {
  const body = await request.json();
  const upstream = await apiFetch("/api/v1/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const data = (await upstream.json().catch(() => ({}))) as Upstream;
  return withOptionalSession(upstream.status, data);
}
