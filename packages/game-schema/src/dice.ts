import { z } from "zod";

export const diceSystems = ["d20", "2d6"] as const;

export const diceSystemSchema = z.enum(diceSystems);

export type DiceSystem = z.infer<typeof diceSystemSchema>;

/** Tale Core default. Universes may select `2d6`. */
export const defaultDiceSystem: DiceSystem = "d20";
