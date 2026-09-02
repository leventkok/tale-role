import { NextResponse } from "next/server";
import { SESSION_COOKIE, apiBase, getSessionToken, sessionCookieOptions } from "@/lib/session";
import { authedProxy } from "@/lib/proxy";

export async function GET() {
  const token = await getSessionToken();
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }
  const upstream = await fetch(`${apiBase()}/api/v1/me`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  const data = await upstream.json().catch(() => ({}));
  return NextResponse.json(data, { status: upstream.status });
}

export async function DELETE() {
  const res = await authedProxy("/api/v1/me", { method: "DELETE" });
  res.cookies.set(SESSION_COOKIE, "", { ...sessionCookieOptions(), maxAge: 0 });
  return res;
}
