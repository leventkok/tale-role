import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { defaultLocale, isLocale, locales, resolveLocale } from "./locales";
import { messagesFor } from "./catalog";

describe("locale registry", () => {
  it("defaults to english", () => {
    assert.equal(defaultLocale, "en");
    assert.deepEqual(locales, ["en", "tr"]);
  });

  it("prefers an explicit user locale", () => {
    assert.equal(resolveLocale("tr", "en-US"), "tr");
  });

  it("falls back to a supported browser language", () => {
    assert.equal(resolveLocale(undefined, "tr-TR"), "tr");
  });

  it("falls back to en when the browser language is unknown", () => {
    assert.equal(resolveLocale(null, "de-DE"), "en");
  });

  it("rejects unknown tags", () => {
    assert.equal(isLocale("de"), false);
    assert.equal(isLocale("en"), true);
  });

  it("loads catalogs for both shipped locales", () => {
    assert.equal(messagesFor("en").app.name, "Tale Role");
    assert.equal(messagesFor("tr").nav.play, "Oyna");
    assert.equal(messagesFor("en").nav.privacy, "Privacy");
  });
});
