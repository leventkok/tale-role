import { z } from "zod";
import { diceSystemSchema, defaultDiceSystem } from "./dice";

export const alignments = ["good", "evil", "neutral"] as const;

export const themeIds = [
  "high-fantasy",
  "gothic-horror",
  "space-opera",
  "cyber-noir",
  "post-apocalyptic",
  "fairytale",
] as const;

export const themeIdSchema = z.enum(themeIds);

export type ThemeId = z.infer<typeof themeIdSchema>;

export const localizedTextSchema = z.object({
  en: z.string().min(1),
  tr: z.string().min(1).optional(),
});

export const universeDocumentSchema = z.object({
  id: z.string().min(1),
  name: localizedTextSchema,
  themeId: themeIdSchema,
  diceSystem: diceSystemSchema.default(defaultDiceSystem),
  rulesetId: z.string().min(1).default("tale-core"),
  npcs: z.array(
    z.object({
      id: z.string().min(1),
      name: localizedTextSchema,
      alignment: z.enum(alignments),
      voice: z.string().max(500).optional(),
    }),
  ),
  promptPackVersion: z.string().min(1),
  contentRating: z.enum(["everyone", "teen", "mature"]).default("teen"),
  era: z.string().max(120).optional(),
  tone: z.string().max(120).optional(),
  taboos: z.string().max(1000).optional(),
});

export type UniverseDocument = z.infer<typeof universeDocumentSchema>;
