import { authedProxy } from "@/lib/proxy";

export async function POST(request: Request) {
  const body = await request.json();
  return authedProxy("/api/v1/me/totp/disable", { method: "POST", body });
}
