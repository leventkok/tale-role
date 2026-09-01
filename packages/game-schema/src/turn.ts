import { z } from "zod";
import { resolutionSchema } from "./intent";

export const narrativeTurnSchema = z.object({
  locale: z.enum(["en", "tr"]),
  prose: z.string().min(1),
  npcLines: z.array(
    z.object({
      npcId: z.string().min(1),
      text: z.string().min(1),
    }),
  ),
  visualPrompt: z.string().max(2000).optional(),
});

export type NarrativeTurn = z.infer<typeof narrativeTurnSchema>;

/** Player/GM websocket payload. Must not include system_admin identities. */
export const publicTurnEventSchema = z.object({
  type: z.literal("turn"),
  roomId: z.string().min(1),
  actorId: z.string().min(1),
  narrative: narrativeTurnSchema,
  resolution: resolutionSchema.pick({
    diceSystem: true,
    rolls: true,
    total: true,
    success: true,
  }),
});

export type PublicTurnEvent = z.infer<typeof publicTurnEventSchema>;

/** Admin spectator channel only. */
export const adminTurnEventSchema = publicTurnEventSchema
  .omit({ type: true })
  .extend({
    type: z.literal("admin_turn"),
    mechanicIntent: z.unknown(),
    promptPackVersion: z.string().min(1),
    adapterId: z.string().min(1).optional(),
  });

export type AdminTurnEvent = z.infer<typeof adminTurnEventSchema>;
