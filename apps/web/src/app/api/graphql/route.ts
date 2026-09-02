import { NextResponse } from "next/server";
import { apiBase, getSessionToken } from "@/lib/session";

export async function POST(request: Request) {
  const body = await request.json().catch(() => null);
  if (!body || typeof body !== "object" || typeof (body as { query?: unknown }).query !== "string") {
    return NextResponse.json({ error: "invalid request" }, { status: 400 });
  }
  const token = await getSessionToken();
  const headers: Record<string, string> = { "Content-Type": "application/json" };
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
