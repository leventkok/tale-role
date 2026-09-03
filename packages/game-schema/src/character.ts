import { z } from "zod";

export const taleSkills = [
  { id: "athletics", ability: "str" },
  { id: "acrobatics", ability: "dex" },
  { id: "stealth", ability: "dex" },
  { id: "arcana", ability: "int" },
  { id: "investigation", ability: "int" },
  { id: "history", ability: "int" },
  { id: "perception", ability: "wis" },
  { id: "insight", ability: "wis" },
  { id: "survival", ability: "wis" },
  { id: "persuasion", ability: "cha" },
  { id: "deception", ability: "cha" },
  { id: "intimidation", ability: "cha" },
] as const;

export type TaleSkillId = (typeof taleSkills)[number]["id"];

export const taleSkillIds = taleSkills.map((s) => s.id) as [TaleSkillId, ...TaleSkillId[]];

export const characterSheetSchema = z.object({
  name: z.string().min(1).max(80),
  species: z.string().max(80).optional(),
  path: z.string().max(80).optional(),
  backstory: z.string().max(4000).optional(),
  skills: z.array(z.enum(taleSkillIds)).max(3).default([]),
});

export type CharacterSheet = z.infer<typeof characterSheetSchema>;
