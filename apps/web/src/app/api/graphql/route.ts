import { NextResponse } from "next/server";
import { apiBase, getSessionToken } from "@/lib/session";

function deviceHeaders(incoming: Headers): Record<string, string> {
  const out: Record<string, string> = {};
  const device = incoming.get("x-talerole-device");
  if (device) {
    out["X-TaleRole-Device"] = device;
  }
  const platform = incoming.get("x-talerole-platform");
  if (platform) {
    out["X-TaleRole-Platform"] = platform;
  }
  return out;
}

export async function POST(request: Request) {
  const body = await request.json().catch(() => null);
  if (!body || typeof body !== "object" || typeof (body as { query?: unknown }).query !== "string") {
    return NextResponse.json({ error: "invalid request" }, { status: 400 });
  }
  const token = await getSessionToken();
  const headers: Record<string, string> = { "Content-Type": "application/json", ...deviceHeaders(request.headers) };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  const upstream = await fetch(`${apiBase()}/graphql`, {
    method: "POST",
    headers,
    body: JSON.stringify(body),
    cache: "no-store",
  });
  const data = await upstream.json().catch(() => ({}));
  return NextResponse.json(data, { status: upstream.status });
}
