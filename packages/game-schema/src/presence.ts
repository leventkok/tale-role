import { z } from "zod";

/** Visible to players and GMs. system_admin never appears here. */
export const publicPresenceRoles = ["player", "gm"] as const;

export const presenceRoles = ["player", "gm", "system_admin"] as const;

export const publicPresenceRoleSchema = z.enum(publicPresenceRoles);
export const presenceRoleSchema = z.enum(presenceRoles);

export type PublicPresenceRole = z.infer<typeof publicPresenceRoleSchema>;
export type PresenceRole = z.infer<typeof presenceRoleSchema>;

export function isVisibleAtTable(role: PresenceRole): boolean {
  return role === "player" || role === "gm";
}
