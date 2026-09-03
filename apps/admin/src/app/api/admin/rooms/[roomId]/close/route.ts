import { NextResponse } from "next/server";
import { apiBase, getSessionToken } from "@/lib/session";

export async function POST(_request: Request, context: { params: Promise<{ roomId: string }> }) {
  const token = await getSessionToken();
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }
  const { roomId } = await context.params;
  const upstream = await fetch(`${apiBase()}/api/v1/admin/rooms/${encodeURIComponent(roomId)}/close`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  const data = await upstream.json().catch(() => ({}));
  return NextResponse.json(data, { status: upstream.status });
}
