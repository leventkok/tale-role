import { authedProxy } from "@/lib/proxy";

export async function GET(
  _request: Request,
  ctx: { params: Promise<{ roomId: string }> },
) {
  const { roomId } = await ctx.params;
  return authedProxy(`/api/v1/rooms/${roomId}`);
}
