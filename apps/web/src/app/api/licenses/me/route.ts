import { NextResponse } from "next/server";
import { getSessionToken, gqlUpstream } from "@/lib/session";

export async function GET() {
  const token = await getSessionToken();
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }
  const result = await gqlUpstream<{
    licenses: { id: string; deviceId: string; platform: string; createdAt?: string }[];
  }>(token, `{ licenses { id deviceId platform createdAt } }`);
  if (result.errors?.length) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }
  return NextResponse.json({
    licenses: (result.data?.licenses ?? []).map((row) => ({
      id: row.id,
      device_id: row.deviceId,
      platform: row.platform,
      created_at: row.createdAt,
    })),
  });
}
