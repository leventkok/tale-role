import { authedProxy } from "@/lib/proxy";

export async function POST(
  request: Request,
  ctx: { params: Promise<{ roomId: string }> },
) {
  const { roomId } = await ctx.params;
  const body = await request.json().catch(() => ({}));
  return authedProxy(`/api/v1/rooms/${roomId}/join`, { method: "POST", body });
}
