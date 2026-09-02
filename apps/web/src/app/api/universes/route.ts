import { authedProxy } from "@/lib/proxy";

export async function GET() {
  return authedProxy("/api/v1/universes");
}

export async function POST(request: Request) {
  const body = await request.json();
  return authedProxy("/api/v1/universes", { method: "POST", body });
}
