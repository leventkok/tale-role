import json
import unittest

from serve import (
    apply_storyteller_adapter,
    chat_prompt,
    extract_json_object,
    fallback_storyteller,
    mechanics_input,
    miss_rewrite_user,
    parse_mechanics_response,
    parse_storyteller_response,
    prose_looks_valid,
    storyteller_input,
    storyteller_user,
    table_deed,
)


class StubTokenizer:
    def apply_chat_template(self, messages, tokenize=False, add_generation_prompt=False):
        parts = []
        for msg in messages:
            parts.append(f"{msg['role']}:{msg['content']}")
        if add_generation_prompt:
            parts.append("assistant:")
        return "|".join(parts)


class ServeFormatTests(unittest.TestCase):
    def test_storyteller_input_pass_zeros_dice(self):
        locale, payload = storyteller_input(
            {
                "locale": "en",
                "kind": "pass",
                "actor_name": "Iri",
                "room_name": "Ashwood",
                "notes": "hold back",
                "rolls": [18],
                "total": 20,
                "success": True,
            }
        )
        self.assertEqual(locale, "en")
        self.assertEqual(payload["kind"], "pass")
        self.assertEqual(payload["rolls"], [])
        self.assertEqual(payload["total"], 0)
        self.assertIsNone(payload["success"])

    def test_action_prompt_carries_world_and_cast(self):
        locale, payload = storyteller_input(
            {
                "locale": "tr",
                "kind": "action",
                "actor_name": "Luther",
                "room_name": "Friday night",
                "opening": "You wake on the cold stone floor of an abandoned Shaper temple.",
                "notes": "Examine the humming carvings",
                "world_brief": "Age: first winter\nMood: wary\nLook: high fantasy\nOpening scene:\nPale blue light.",
                "cast": [{"name": "Luther", "species": "human", "path": "ranger", "backstory": "A scout of the Shapers"}],
                "rolls": [8],
                "total": 10,
                "success": False,
            }
        )
        self.assertEqual(locale, "en")
        user = storyteller_user(payload, opening=False, locale=locale)
        self.assertIn("Age: first winter", user)
        self.assertIn("Luther", user)
        self.assertIn("ranger", user)
        self.assertIn("Examine the humming carvings", user)
        self.assertNotIn("Friday night", user)
        self.assertNotIn("high-fantasy", user)
        self.assertIn("RESULT: MISS", user)
        self.assertIn("success=False", user)
        self.assertIn("Fail forward", user)
        self.assertNotIn("learns nothing useful", user)

    def test_chat_prompt_uses_template(self):
        prompt = chat_prompt(StubTokenizer(), "sys", '{"actor":"Iri"}')
        self.assertIn("system:sys", prompt)
        self.assertIn('user:{"actor":"Iri"}', prompt)
        self.assertTrue(prompt.endswith("assistant:"))

    def test_parse_storyteller_response(self):
        raw = '{"prose":"Iri waits in Ashwood under wet canvas, listening for the second creak on the stair, and does not rush the dark.","npc_lines":[]}'
        parsed = parse_storyteller_response(raw)
        self.assertIsNotNone(parsed)
        assert parsed is not None
        self.assertIn("Ashwood", parsed["prose"])

    def test_reject_short_training_wait(self):
        raw = '{"prose":"Bir sonraki çandan önce. Çığlık gelene dek bekle.","npc_lines":[]}'
        self.assertIsNone(parse_storyteller_response(raw, "tr", opening=True))

    def test_base_instruct_by_default(self):
        self.assertFalse(apply_storyteller_adapter())

    def test_reject_json_leak_as_prose(self):
        raw = '{"prose":"{\\"actor\\":\\"Lute\\"}","npc_lines":[]}'
        self.assertIsNone(parse_storyteller_response(raw))
        self.assertFalse(prose_looks_valid('{"actor":"Lute","room":"Hall"}'))

    def test_extract_json_stops_at_im_end(self):
        raw = '{"kind":"action","skill":"str","dc":12}\nnoise'
        parsed = extract_json_object(raw)
        self.assertEqual(parsed, {"kind": "action", "skill": "str", "dc": 12})

    def test_fallback_action_en(self):
        locale, payload = storyteller_input(
            {
                "locale": "en",
                "kind": "action",
                "actor_name": "Mira",
                "room_name": "Saltgate",
                "notes": "pick the lock",
                "rolls": [14],
                "total": 17,
                "success": True,
                "dice_system": "d20",
            }
        )
        out = fallback_storyteller(locale, payload, say=False)
        self.assertIn("Mira", out["prose"])
        self.assertIn("17", out["prose"])
        self.assertNotIn("[hub]", out["prose"])
        self.assertNotIn("tries to", out["prose"])

    def test_fallback_action_does_not_name_warcraft(self):
        locale, payload = storyteller_input(
            {
                "locale": "tr",
                "kind": "action",
                "actor_name": "Luther",
                "room_name": "World Of Warcraft",
                "notes": "Examine the humming carvings",
                "rolls": [8],
                "total": 10,
                "success": False,
            }
        )
        out = fallback_storyteller(locale, payload, say=False)
        self.assertIn("Luther", out["prose"])
        self.assertIn("10", out["prose"])
        self.assertNotIn("direnir", out["prose"])
        self.assertNotIn("Alet kayar", out["prose"])
        self.assertNotIn("World Of Warcraft", out["prose"])
        self.assertIn("Examine the humming carvings", out["prose"])
        self.assertNotIn("follows through", out["prose"])
        self.assertIn("misses", out["prose"])
        self.assertIn("next beat", out["prose"])
        self.assertNotIn("nothing shifts", out["prose"])

    def test_fallback_miss_does_not_echo_first_person(self):
        locale, payload = storyteller_input(
            {
                "locale": "en",
                "kind": "action",
                "actor_name": "Luther",
                "room_name": "Friday night",
                "notes": "I pick up the medallion and listen to the humming carvings, without going down the corridor.",
                "rolls": [2],
                "total": 7,
                "success": False,
            }
        )
        out = fallback_storyteller(locale, payload, say=False)
        self.assertIn("Luther", out["prose"])
        self.assertIn("misses", out["prose"])
        self.assertIn("next beat", out["prose"])
        self.assertNotIn("I pick up", out["prose"])
        self.assertEqual(table_deed(payload["notes"]), "")
        self.assertIn("Rewrite", miss_rewrite_user("THIS BEAT"))

    def test_fallback_action_does_not_name_any_lobby(self):
        locale, payload = storyteller_input(
            {
                "locale": "tr",
                "kind": "action",
                "actor_name": "Luther",
                "room_name": "Star Wars",
                "notes": "Examine the humming carvings",
                "rolls": [8],
                "total": 10,
                "success": False,
            }
        )
        out = fallback_storyteller(locale, payload, say=False)
        self.assertNotIn("Star Wars", out["prose"])
        self.assertEqual(payload["room"], "the hall")

    def test_staccato_title_salad_rejected(self):
        raw = (
            '{"prose":"Luther looks. Star Wars resists. Number 10. '
            'The tool slips. Time ends.","npc_lines":[]}'
        )
        self.assertIsNone(
            parse_storyteller_response(
                raw, "tr", table_title="Star Wars", host="Shaper temple opening"
            )
        )

    def test_story_opening_stays_in_locale(self):
        locale, payload = storyteller_input(
            {
                "locale": "tr",
                "kind": "story",
                "room_name": "Kalekarga",
                "opening": "Sis kapı eşiğinde bekler.",
                "presence_names": ["Bram", "Lute"],
                "rolls": [1],
                "total": 1,
                "success": False,
            }
        )
        self.assertEqual(payload["kind"], "story")
        self.assertEqual(payload["total"], 0)
        out = fallback_storyteller(locale, payload, say=False)
        self.assertEqual(out["prose"], "Sis kapı eşiğinde bekler.")
        self.assertNotIn("die reads", out["prose"].casefold())
        self.assertNotIn("Anlatıcı eşiğe", out["prose"])

    def test_host_english_opening_not_mashed(self):
        opening = (
            "You wake on the cold stone floor of an abandoned Shaper temple. "
            "Pale blue light leaks through the cracks."
        )
        locale, payload = storyteller_input(
            {
                "locale": "tr",
                "kind": "story",
                "room_name": "World Of Warcraft",
                "opening": opening,
                "theme_id": "high-fantasy",
                "presence_names": ["Bram"],
            }
        )
        out = fallback_storyteller(locale, payload, say=False)
        self.assertEqual(out["prose"], opening)
        self.assertNotIn("sessiz", out["prose"])
        self.assertNotIn("high-fantasy", out["prose"])

    def test_reject_english_when_locale_is_tr(self):
        raw = '{"prose":"The watch is unblinded. Hold the line. The bar splinters and yields.","npc_lines":[]}'
        self.assertIsNone(parse_storyteller_response(raw, "tr"))

    def test_reject_prior_repeat(self):
        raw = '{"prose":"Night holds Kalekarga. Bram stands at the threshold. The tale begins before anyone moves.","npc_lines":[]}'
        prior = ["Night holds Kalekarga. Bram stands at the threshold. The tale begins before anyone moves."]
        self.assertIsNone(parse_storyteller_response(raw, "en", prior))

    def test_continue_scene_is_not_prior_repeat(self):
        opening = (
            "You wake on the cold stone floor of an abandoned Shaper temple. "
            "Pale blue light leaks through the cracks in the ceiling, catching the ancient carvings."
        )
        raw = (
            '{"prose":"Luther keeps his distance as pale blue light slides over the stranger\'s runes. '
            "The staff ticks once and goes dark. Wake stays cold in his palm. "
            'He learns nothing more. The figure does not answer.","npc_lines":[]}'
        )
        parsed = parse_storyteller_response(raw, "en", [opening])
        self.assertIsNotNone(parsed)

    def test_reject_hit_voice_on_failed_roll(self):
        prose = (
            "Luther picks up the medallion and studies it for a moment, feeling the subtle "
            "vibrations emanating from the carvings grow stronger. He lets out a low whistle, "
            "recognizing the power at work here. 'Interesting,' he murmurs. Without hesitation, "
            "he begins to trace a complex pattern along the wall with his finger, following the "
            "hum like a map. The others watch warily as the temperature drops slightly, making "
            "the hairs on their arms stand on end. Metal clatters against stone somewhere deeper "
            "within the temple, growing louder. 'We should move quickly,' Luther says, voice "
            "steady despite the creeping unease in the air. 'Whatever made these carvings is "
            "still awake, and I have a bad feeling about that.'"
        )
        raw = json.dumps({"prose": prose, "npc_lines": []})
        self.assertIsNone(parse_storyteller_response(raw, "en", success=False))
        self.assertIsNotNone(parse_storyteller_response(raw, "en", success=True))

    def test_fallback_say_tr(self):
        locale, payload = storyteller_input(
            {
                "locale": "tr",
                "kind": "say",
                "actor_name": "Bram",
                "room_name": "Kalekarga",
                "notes": "kapıyı açın",
            }
        )
        out = fallback_storyteller(locale, payload, say=True)
        self.assertIn("Bram", out["prose"])
        self.assertIn("kapıyı açın", out["prose"])

    def test_mechanics_parse_rejects_invented_state(self):
        raw = json.dumps({"kind": "action", "skill": "str", "dc": 12, "rolls": [10]})
        self.assertIsNone(parse_mechanics_response(raw, "pick lock"))

    def test_mechanics_input_maps_say_to_wait(self):
        payload = mechanics_input({"kind": "say", "notes": "hello"})
        self.assertEqual(payload["kind"], "wait")


if __name__ == "__main__":
    unittest.main()
