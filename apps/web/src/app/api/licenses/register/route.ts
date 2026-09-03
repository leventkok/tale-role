import { NextResponse } from "next/server";
import { getSessionToken, gqlUpstream } from "@/lib/session";

export async function POST(request: Request) {
  const body = (await request.json()) as { device_id?: string; platform?: string };
  const token = await getSessionToken();
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }
  const result = await gqlUpstream<{ registerLicense: { id: string; deviceId: string; platform: string } }>(
    token,
    `mutation ($deviceId: String!, $platform: String) {
      registerLicense(deviceId: $deviceId, platform: $platform) { id deviceId platform }
    }`,
    { deviceId: body.device_id ?? "", platform: body.platform ?? "" },
  );
  const lic = result.data?.registerLicense;
  if (!lic?.id) {
    return NextResponse.json({ error: result.errors?.[0]?.message ?? "error" }, { status: 400 });
  }
  return NextResponse.json(
    { id: lic.id, device_id: lic.deviceId, platform: lic.platform },
    { status: 201 },
  );
}
