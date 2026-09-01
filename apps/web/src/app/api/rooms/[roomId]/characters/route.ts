import { authedProxy } from "@/lib/proxy";

export async function POST(
  request: Request,
  ctx: { params: Promise<{ roomId: string }> },
) {
  const { roomId } = await ctx.params;
  const body = await request.json();
  return authedProxy(`/api/v1/rooms/${roomId}/characters`, { method: "POST", body });
}
