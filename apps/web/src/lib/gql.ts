export type GqlError = { message: string };
export type GqlResult<T> = { data?: T; errors?: GqlError[] };

export const ROOM_QUERY = `
query Room($id: ID!) {
  room(id: $id) {
    id name hostId diceSystem joinMode started universeId themeId
    turnOrder
    presence { userId role }
    characters { userId name hp stats { str dex con int wis cha } }
    turns { actorId kind notes rolls total success prose }
    scene { themeId visualPrompt imageSvg inference }
  }
}
`;

export type GqlRoom = {
  id: string;
  name: string;
  hostId: string;
  diceSystem: string;
  joinMode: string;
  started: boolean;
  universeId?: string | null;
  themeId?: string | null;
  turnOrder: string[];
  presence: { userId: string; role: string }[];
  characters: {
    userId: string;
    name: string;
    hp: number;
    stats?: { str: number; dex: number; con: number; int: number; wis: number; cha: number } | null;
  }[];
  turns: {
    actorId?: string | null;
    kind: string;
    notes?: string | null;
    rolls: number[];
    total: number;
    success?: boolean | null;
    prose?: string | null;
  }[];
  scene?: {
    themeId?: string | null;
    visualPrompt?: string | null;
    imageSvg?: string | null;
    inference?: string | null;
  } | null;
};

export type TableRoom = {
  id: string;
  name: string;
  host_id: string;
  dice_system: string;
  started: boolean;
  turn_order: string[];
  presence: { user_id: string; role: string }[];
  characters: {
    user_id: string;
    name: string;
    hp: number;
    stats?: { str: number; dex: number; con: number; int: number; wis: number; cha: number };
  }[];
  theme_id?: string;
  scene?: {
    theme_id: string;
    visual_prompt?: string;
    image_svg?: string;
    inference?: string;
  };
  turns: {
    actor_id: string;
    kind: string;
    notes?: string;
    rolls?: number[];
    total?: number;
    success?: boolean | null;
    narrative?: { prose?: string };
  }[];
};

export function mapRoom(row: GqlRoom): TableRoom {
  return {
    id: row.id,
    name: row.name,
    host_id: row.hostId,
    dice_system: row.diceSystem,
    started: row.started,
    turn_order: row.turnOrder ?? [],
    presence: (row.presence ?? []).map((p) => ({ user_id: p.userId, role: p.role })),
    characters: (row.characters ?? []).map((ch) => ({
      user_id: ch.userId,
      name: ch.name,
      hp: ch.hp,
      stats: ch.stats ?? undefined,
    })),
    theme_id: row.themeId ?? undefined,
    scene: row.scene
      ? {
          theme_id: row.scene.themeId ?? "",
          visual_prompt: row.scene.visualPrompt ?? undefined,
          image_svg: row.scene.imageSvg ?? undefined,
          inference: row.scene.inference ?? undefined,
        }
      : undefined,
    turns: (row.turns ?? []).map((t) => ({
      actor_id: t.actorId ?? "",
      kind: t.kind,
      notes: t.notes ?? undefined,
      rolls: t.rolls,
      total: t.total,
      success: t.success,
      narrative: t.prose ? { prose: t.prose } : undefined,
    })),
  };
}

export async function gql<T>(query: string, variables?: Record<string, unknown>): Promise<GqlResult<T>> {
  const res = await fetch("/api/graphql", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ query, variables }),
    cache: "no-store",
  });
  return (await res.json().catch(() => ({}))) as GqlResult<T>;
}

export function gqlData<T>(result: GqlResult<T>): T | null {
  if (!result.data || (result.errors && result.errors.length > 0)) {
    return null;
  }
  return result.data;
}
