import { NextResponse } from "next/server";
import { apiBase, getSessionToken } from "@/lib/session";

export async function GET(request: Request) {
  const token = await getSessionToken();
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }
  const roomId = new URL(request.url).searchParams.get("room_id") ?? "";
  const upstream = await fetch(`${apiBase()}/api/v1/admin/live?room_id=${encodeURIComponent(roomId)}`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  const data = await upstream.json().catch(() => ({}));
  return NextResponse.json(data, { status: upstream.status });
}
