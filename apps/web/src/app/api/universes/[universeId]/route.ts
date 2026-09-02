import { authedProxy } from "@/lib/proxy";

export async function GET(
  _request: Request,
  ctx: { params: Promise<{ universeId: string }> },
) {
  const { universeId } = await ctx.params;
  return authedProxy(`/api/v1/universes/${universeId}`);
}
