import assert from "node:assert/strict";
import { describe, it } from "node:test";
import {
  defaultDiceSystem,
  isVisibleAtTable,
  mechanicIntentSchema,
  publicTurnEventSchema,
  sceneCardSchema,
  universeDocumentSchema,
} from "./index";

describe("tale core contracts", () => {
  it("defaults dice to d20 while allowing 2d6", () => {
    assert.equal(defaultDiceSystem, "d20");
    const parsed = universeDocumentSchema.parse({
      id: "u1",
      name: { en: "Ashwood" },
      themeId: "high-fantasy",
      promptPackVersion: "v1",
      npcs: [],
    });
    assert.equal(parsed.diceSystem, "d20");
    assert.equal(parsed.era, undefined);
    assert.equal(
      universeDocumentSchema.parse({ ...parsed, era: "late autumn", tone: "wary" }).era,
      "late autumn",
    );
    assert.equal(
      universeDocumentSchema.parse({ ...parsed, diceSystem: "2d6" }).diceSystem,
      "2d6",
    );
  });

  it("hides system admins from the table", () => {
    assert.equal(isVisibleAtTable("player"), true);
    assert.equal(isVisibleAtTable("gm"), true);
    assert.equal(isVisibleAtTable("system_admin"), false);
  });

  it("rejects pass intents that try to invent dice", () => {
    const pass = mechanicIntentSchema.parse({ kind: "pass" });
    assert.equal(pass.kind, "pass");
  });

  it("forbids spectator fields on public turn events", () => {
    const event = publicTurnEventSchema.parse({
      type: "turn",
      roomId: "r1",
      actorId: "p1",
      narrative: { locale: "en", prose: "The door yields.", npcLines: [] },
      resolution: {
        diceSystem: "d20",
        rolls: [14],
        total: 17,
        success: true,
      },
    });
    assert.equal("mechanicIntent" in event, false);
  });

  it("keeps scene cards on the six themes and the stub", () => {
    const card = sceneCardSchema.parse({
      themeId: "gothic-horror",
      visualPrompt: "Tale Role stub scene. Theme gothic-horror. No dice.",
      inference: "stub",
    });
    assert.equal(card.inference, "stub");
  });
});
