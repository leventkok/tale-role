import { z } from "zod";

export const portraitIds = ["warden", "blade", "sage", "ranger", "oath"] as const;

export type PortraitId = (typeof portraitIds)[number];

export const defaultPortraitId: PortraitId = "warden";

export const portraitIdSchema = z.enum(portraitIds);
