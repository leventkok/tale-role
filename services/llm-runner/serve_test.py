import json
import unittest

from serve import (
    chat_prompt,
    extract_json_object,
    fallback_storyteller,
    mechanics_input,
    parse_mechanics_response,
    parse_storyteller_response,
    prose_looks_valid,
    storyteller_input,
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

    def test_chat_prompt_uses_template(self):
        prompt = chat_prompt(StubTokenizer(), "sys", '{"actor":"Iri"}')
        self.assertIn("system:sys", prompt)
        self.assertIn('user:{"actor":"Iri"}', prompt)
        self.assertTrue(prompt.endswith("assistant:"))

    def test_parse_storyteller_response(self):
        raw = '{"prose":"Iri waits in Ashwood, listening, and does not roll. The table waits.","npc_lines":[]}'
        parsed = parse_storyteller_response(raw)
        self.assertIsNotNone(parsed)
        assert parsed is not None
        self.assertIn("Ashwood", parsed["prose"])
        self.assertEqual(parsed["npc_lines"], [])

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
