import { NextResponse } from "next/server";
import { apiBase, SESSION_COOKIE, sessionCookieOptions } from "@/lib/session";

type Upstream = Record<string, unknown>;

export async function POST(request: Request) {
  const body = await request.json();
  const upstream = await fetch(`${apiBase()}/api/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const data = (await upstream.json().catch(() => ({}))) as Upstream;
  const token = data.token;
  delete data.token;
  const res = NextResponse.json(data, { status: upstream.status });
  if (typeof token === "string" && token.length > 0) {
    res.cookies.set(SESSION_COOKIE, token, sessionCookieOptions());
  }
  return res;
}
