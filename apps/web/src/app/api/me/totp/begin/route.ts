import { authedProxy } from "@/lib/proxy";

export async function POST() {
  return authedProxy("/api/v1/me/totp/begin", { method: "POST", body: {} });
}
