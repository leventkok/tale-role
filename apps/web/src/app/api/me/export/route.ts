import { authedProxy } from "@/lib/proxy";

export async function GET() {
  return authedProxy("/api/v1/me/export");
}
