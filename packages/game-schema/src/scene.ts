import { z } from "zod";
import { themeIdSchema } from "./universe";

export const sceneCardSchema = z.object({
  themeId: themeIdSchema,
  visualPrompt: z.string().min(1).max(2000),
  inference: z.literal("stub"),
});

export type SceneCard = z.infer<typeof sceneCardSchema>;
