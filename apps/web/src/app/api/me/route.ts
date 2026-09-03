import { NextResponse } from "next/server";
import { SESSION_COOKIE, getSessionToken, gqlUpstream, sessionCookieOptions } from "@/lib/session";

export async function GET() {
  const token = await getSessionToken();
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }
  const result = await gqlUpstream<{ me: { id: string; email: string; verified: boolean; totpEnabled: boolean } | null }>(
    token,
    "{ me { id email verified totpEnabled } }",
  );
  const me = result.data?.me;
  if (!me) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }
  return NextResponse.json({
    id: me.id,
    email: me.email,
    verified: me.verified,
    totp_enabled: me.totpEnabled,
  });
}

export async function DELETE() {
  const token = await getSessionToken();
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }
  const result = await gqlUpstream<{ eraseMe: boolean }>(token, "mutation { eraseMe }");
  const res = NextResponse.json(
    result.errors?.length ? { error: "unauthorized" } : { ok: true, erased: true },
    { status: result.data?.eraseMe ? 200 : 401 },
  );
  res.cookies.set(SESSION_COOKIE, "", { ...sessionCookieOptions(), maxAge: 0 });
  return res;
}
