import assert from "node:assert/strict";
import { readFileSync, readdirSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { describe, it } from "node:test";
import { z } from "zod";
import { mechanicIntentSchema, narrativeTurnSchema } from "./index";

const root = join(dirname(fileURLToPath(import.meta.url)), "../../../llm/datasets/synthetic");
const emailRe = /[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}/i;

const storytellerRow = z.object({
  locale: z.enum(["en", "tr"]),
  input: z.object({
    actor: z.string().min(1),
    room: z.string().min(1),
    kind: z.enum(["action", "pass", "wait"]),
    dice: z.enum(["d20", "2d6"]),
    rolls: z.array(z.number().int()),
    total: z.number(),
    success: z.boolean().nullable(),
    notes: z.string(),
    presence: z.array(z.string().min(1)),
  }),
  output: z.object({
    prose: z.string().min(1),
    npc_lines: z.array(z.object({ npc_id: z.string().min(1), text: z.string().min(1) })),
  }),
});

function loadJsonl(name: string): unknown[] {
  const raw = readFileSync(join(root, name), "utf8");
  return raw
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => JSON.parse(line) as unknown);
}

function leak(s: string) {
  return emailRe.test(s) || s.includes("system_admin");
}

describe("own-model synthetic datasets", () => {
  it("ships jsonl files", () => {
    const names = readdirSync(root);
    assert.ok(names.includes("storyteller.jsonl"));
    assert.ok(names.includes("mechanics.jsonl"));
  });

  it("storyteller rows match the narrative contract and do not leak pii", () => {
    for (const row of loadJsonl("storyteller.jsonl")) {
      const parsed = storytellerRow.parse(row);
      narrativeTurnSchema.parse({
        locale: parsed.locale,
        prose: parsed.output.prose,
        npcLines: parsed.output.npc_lines.map((l) => ({ npcId: l.npc_id, text: l.text })),
      });
      const blob = JSON.stringify(parsed);
      assert.equal(leak(blob), false);
      if (parsed.input.kind === "action" && parsed.input.success === true) {
        assert.match(parsed.output.prose, new RegExp(String(parsed.input.total)));
      }
    }
  });

  it("mechanics rows are schema JSON and never include dice", () => {
    for (const row of loadJsonl("mechanics.jsonl")) {
      const parsed = z
        .object({
          locale: z.enum(["en", "tr"]),
          input: z.object({
            notes: z.string(),
            kind: z.enum(["action", "pass", "wait"]),
            skill: z.string(),
          }),
          output: mechanicIntentSchema,
        })
        .parse(row);
      const blob = JSON.stringify(parsed.output);
      assert.equal("rolls" in parsed.output, false);
      assert.equal("hp" in parsed.output, false);
      assert.equal(leak(blob), false);
      assert.equal(parsed.output.kind, parsed.input.kind);
    }
  });
});
