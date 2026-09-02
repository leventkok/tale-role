import { authedProxy } from "@/lib/proxy";

export async function POST(
  _request: Request,
  ctx: { params: Promise<{ roomId: string }> },
) {
  const { roomId } = await ctx.params;
  return authedProxy(`/api/v1/rooms/${roomId}/start`, { method: "POST", body: {} });
}
