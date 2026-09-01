import { z } from "zod";

export const mechanicIntentSchema = z.object({
  kind: z.enum(["action", "pass", "wait"]),
  skill: z.string().min(1).optional(),
  targetId: z.string().min(1).optional(),
  dc: z.number().int().min(1).max(40).optional(),
  inventoryOp: z
    .object({
      itemId: z.string().min(1),
      op: z.enum(["use", "equip", "drop", "give"]),
    })
    .optional(),
  notes: z.string().max(500).optional(),
});

export type MechanicIntent = z.infer<typeof mechanicIntentSchema>;

export const gameStatePatchSchema = z.object({
  hp: z.record(z.string(), z.number().int()).optional(),
  inventory: z.record(z.string(), z.array(z.string())).optional(),
  conditions: z.record(z.string(), z.array(z.string())).optional(),
  turnOrder: z.array(z.string()).optional(),
});

export type GameStatePatch = z.infer<typeof gameStatePatchSchema>;

export const resolutionSchema = z.object({
  intent: mechanicIntentSchema,
  diceSystem: z.enum(["d20", "2d6"]),
  rolls: z.array(z.number().int()),
  total: z.number(),
  success: z.boolean().nullable(),
  patch: gameStatePatchSchema,
});

export type Resolution = z.infer<typeof resolutionSchema>;
