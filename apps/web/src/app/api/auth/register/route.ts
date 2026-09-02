import { NextResponse } from "next/server";
import { apiBase } from "@/lib/session";

export async function POST(request: Request) {
  const body = await request.json();
  const upstream = await fetch(`${apiBase()}/api/v1/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const data = await upstream.json().catch(() => ({}));
  return NextResponse.json(data, { status: upstream.status });
}
