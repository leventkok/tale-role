import json
import os
import unittest

from serve import (
    apply_storyteller_adapter,
    chat_prompt,
    hub_has_adapter_weights,
    hub_has_full_weights,
    patch_torchao_for_peft,
    extract_json_object,
    fallback_storyteller,
    is_salad,
    mechanics_input,
    miss_rewrite_user,
    parse_mechanics_response,
    parse_storyteller_response,
    prose_looks_valid,
    stay_put_deed,
    storyteller_input,
    storyteller_system,
    storyteller_user,
    strip_engine_leak,
    table_deed,
    use_storyteller_adapter,
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

    def test_say_prompt_answers_without_changing_scene(self):
        sys = storyteller_system("en", opening=False, prior=[], kind="say", notes="Who is there?")
        self.assertIn("Do not change the scene", sys)
        self.assertNotIn("This attempt succeeded", sys)
        locale, payload = storyteller_input(
            {"locale": "en", "kind": "say", "actor_name": "Fred", "notes": "Who is there?"}
        )
        self.assertEqual(payload["kind"], "say")
        user = storyteller_user(payload, opening=False, locale=locale)
        self.assertIn("Do not change the scene", user)
        self.assertNotIn("RESULT: HIT", user)

    def test_hit_prompt_grants_deed_and_does_not_forbid_success(self):
        notes = "Sesin geldiği yere bıçak fırlatıyorum"
        sys = storyteller_system("tr", opening=False, prior=[], success=True, notes=notes)
        self.assertNotIn("Never write the attempt as a success", sys)
        self.assertNotIn("stays put", sys)
        self.assertIn("succeeded", sys)
        self.assertIn("Do not refuse", sys)
        locale, payload = storyteller_input(
            {
                "locale": "tr",
                "kind": "action",
                "actor_name": "Floc",
                "notes": notes,
                "total": 16,
                "success": True,
            }
        )
        user = storyteller_user(payload, opening=False, locale=locale)
        self.assertIn("RESULT: HIT", user)
        self.assertIn("Narrate the deed happening", user)
        self.assertNotIn("Grant the deed only", user)
        self.assertNotIn("count=", user)
        self.assertNotIn("16", user)

    def test_reject_packed_miss_salad(self):
        notes = (
            "Floc attempts a deed. Player count 8. MISS. "
            "Deed: Ayağa kalkarım ve tobamda ne olduğuna bakarım. "
            "Narrate this outcome. Never mention an opposing roll or hidden difficulty."
        )
        raw = (
            '{"prose":"Floc hamleyi kaçırır: Ayağa kalkarım ve tobamda ne olduğuna bakarım. '
            'Taş susar. Sayı 8; uzaktan bir ses sahneyi sürdürür.","npc_lines":[]}'
        )
        self.assertIsNone(parse_storyteller_response(raw, "tr", success=False, notes=notes))
        notes = "Torbanın içini kontrol ediyorum faydalı bir şey var mı ?"
        raw = (
            '{"prose":"Floc hamleyi tamamlar: Torbanın içini kontrol ediyorum faydalı bir şey var mı ? '
            'Taş cevap verir. Sayı 23; yol açılır.","npc_lines":[]}'
        )
        self.assertIsNone(parse_storyteller_response(raw, "tr", success=True, notes=notes))

    def test_strip_live_pack_and_reject_bag_salad(self):
        notes = (
            "Floc attempts a deed. Player count 9. MISS. "
            "Deed: Ayağa kalkarım ve torbanın içindekilere göz atarım. "
            "Narrate this outcome in third person. Do not paste the deed."
        )
        self.assertEqual(
            strip_engine_leak(notes),
            "Ayağa kalkarım ve torbanın içindekilere göz atarım",
        )
        locale, payload = storyteller_input(
            {
                "locale": "tr",
                "kind": "action",
                "actor_name": "Floc",
                "notes": notes,
                "total": 9,
                "success": False,
            }
        )
        self.assertEqual(payload["notes"], "Ayağa kalkarım ve torbanın içindekilere göz atarım")
        user = storyteller_user(payload, opening=False, locale=locale)
        self.assertNotIn("Player count", user)
        self.assertNotIn("9", user)
        salad = (
            "Floc hamleyi kaçırır: Ayağa kalkarım ve torbanın içindekilere göz atarım. "
            "Taş susar. Sayı 9; uzaktan bir ses sahneyi sürdürür."
        )
        self.assertTrue(is_salad(salad))
        self.assertFalse(prose_looks_valid(salad, "tr", success=False, notes=notes))
        raw = json.dumps({"prose": salad, "npc_lines": []})
        self.assertIsNone(parse_storyteller_response(raw, "tr", success=False, notes=notes))
        out = fallback_storyteller(locale, payload, say=False)
        self.assertNotIn("Sayı", out["prose"])
        self.assertNotIn("9", out["prose"])
        self.assertNotIn("Taş susar", out["prose"])
        self.assertNotIn("kalkarım", out["prose"])
        self.assertIn("kaçırır", out["prose"])

    def test_literary_train_voice_passes_the_gate(self):
        prose = (
            "Esin yerinden kıpırdamaz. Nağme avucunda netleşir, bir kelime kadar: uyan. "
            "Koridor aynı karanlıkta bekler. O karanlığa girmez."
        )
        self.assertTrue(
            prose_looks_valid(
                prose,
                "tr",
                success=True,
                notes="Yerimde kalıp taştaki nağmeyi dinlerim, koridora adım atmadan.",
            )
        )

    def test_bag_hit_is_not_etki_tutar(self):
        locale, payload = storyteller_input(
            {
                "locale": "tr",
                "kind": "action",
                "actor_name": "Floc",
                "notes": "Ayağa kalkarım ve torbanın içindekilere göz atarım.",
                "total": 14,
                "success": True,
            }
        )
        out = fallback_storyteller(locale, payload, say=False)
        self.assertIn("Floc", out["prose"])
        self.assertIn("torba", out["prose"].casefold())
        self.assertNotIn("Etki tutar", out["prose"])
        self.assertNotIn("kalkarım", out["prose"])
        self.assertNotIn("Sayı", out["prose"])
        bag = (
            "Floc ayağa kalkar. Torbanın içinde, kumaş ve soğuk bir kenar fener ışığına çıkar. "
            "Tapınağın nefesi değişmez. Sıra yine masada, torba artık açık."
        )
        self.assertTrue(
            prose_looks_valid(
                bag,
                "tr",
                success=True,
                notes="Ayağa kalkarım ve torbanın içindekilere göz atarım.",
            )
        )

    def test_miss_prompt_still_forbids_success(self):
        sys = storyteller_system(
            "en",
            opening=False,
            prior=[],
            success=False,
            notes="I throw a knife toward the sound",
        )
        self.assertIn("Never write the attempt as a success", sys)
        self.assertNotIn("stays put", sys)

    def test_stay_put_prompt_only_when_deed_stays(self):
        held = storyteller_system(
            "en",
            opening=False,
            prior=[],
            success=True,
            notes="I pick up the medallion, without going down the corridor.",
        )
        self.assertIn("stays put", held)
        move = storyteller_system(
            "en",
            opening=False,
            prior=[],
            success=True,
            notes="I throw a knife toward the sound",
        )
        self.assertNotIn("stays put", move)

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
        self.assertFalse(use_storyteller_adapter("action"))
        self.assertFalse(use_storyteller_adapter("story"))

    def test_action_infers_from_live_parameters(self):
        sys = storyteller_system(
            "tr",
            opening=False,
            prior=[],
            success=True,
            notes="Ayağa kalkarım ve torbanın içindekilere göz atarım.",
            kind="action",
        )
        self.assertIn("You do not need to have seen this object in training", sys)
        self.assertIn("Esin yerinden kıpırdamaz", sys)
        self.assertIn("Voice to copy, not facts to copy", sys)
        opening = storyteller_system("tr", opening=True, prior=[], kind="story")
        self.assertIn("You do not need to have seen this object in training", opening)
        self.assertNotIn("Esin yerinden kıpırdamaz", opening)

    def test_adapter_actions_opt_in(self):
        old = os.environ.get("TALEROLE_STORYTELLER_ADAPTER_ACTIONS")
        old_story = os.environ.get("TALEROLE_STORYTELLER_ADAPTER")
        try:
            os.environ["TALEROLE_STORYTELLER_ADAPTER"] = "1"
            os.environ.pop("TALEROLE_STORYTELLER_ADAPTER_ACTIONS", None)
            self.assertTrue(use_storyteller_adapter("story"))
            self.assertFalse(use_storyteller_adapter("action"))
            os.environ["TALEROLE_STORYTELLER_ADAPTER_ACTIONS"] = "1"
            self.assertTrue(use_storyteller_adapter("action"))
        finally:
            if old is None:
                os.environ.pop("TALEROLE_STORYTELLER_ADAPTER_ACTIONS", None)
            else:
                os.environ["TALEROLE_STORYTELLER_ADAPTER_ACTIONS"] = old
            if old_story is None:
                os.environ.pop("TALEROLE_STORYTELLER_ADAPTER", None)
            else:
                os.environ["TALEROLE_STORYTELLER_ADAPTER"] = old_story

    def test_hub_adapter_needs_weight_file(self):
        self.assertFalse(hub_has_adapter_weights({"adapter_config.json", "README.md"}))
        self.assertTrue(
            hub_has_adapter_weights({"adapter_config.json", "adapter_model.safetensors"})
        )
        self.assertTrue(
            hub_has_full_weights({"config.json", "model-00001-of-00004.safetensors"})
        )

    def test_torchao_peft_patch_is_safe(self):
        patch_torchao_for_peft()

    def test_reject_json_leak_as_prose(self):
        raw = '{"prose":"{\\"actor\\":\\"Lute\\"}","npc_lines":[]}'
        self.assertIsNone(parse_storyteller_response(raw))
        self.assertFalse(prose_looks_valid('{"actor":"Lute","room":"Hall"}'))

    def test_reject_pii_harvest(self):
        raw = '{"prose":"Luther waits in the hall and asks for your e-mail address so the tale can continue after the session.","npc_lines":[]}'
        self.assertIsNone(parse_storyteller_response(raw, "en", success=True, notes="listen"))

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
        self.assertNotIn("17", out["prose"])
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
        self.assertNotIn("10", out["prose"])
        self.assertNotIn("direnir", out["prose"])
        self.assertNotIn("Alet kayar", out["prose"])
        self.assertNotIn("World Of Warcraft", out["prose"])
        self.assertNotIn("Examine the humming carvings", out["prose"])
        self.assertNotIn("follows through", out["prose"])
        self.assertIn("misses", out["prose"])
        self.assertNotIn("nothing shifts", out["prose"])
        self.assertNotIn("the stone stays mute", out["prose"].casefold())

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
        self.assertNotIn("I pick up", out["prose"])
        self.assertNotIn("nothing shifts", out["prose"])
        self.assertEqual(table_deed(payload["notes"]), "")
        self.assertIn("Rewrite", miss_rewrite_user("THIS BEAT"))

    def test_reject_walking_when_deed_stays_put(self):
        deed = "I stay where I am and try to remember the star-path from the song, without taking a step toward the corridor."
        overshoot = (
            "With focused determination, Luther closes his eyes and allows the resonant tones to guide him. "
            "Opening his eyes, Luther notices the corridor ahead shimmering. He steps into the luminous void, "
            "feeling a gentle push against his feet. As he walks, the air thickens."
        )
        raw = json.dumps({"prose": overshoot, "npc_lines": []})
        self.assertIsNone(parse_storyteller_response(raw, "en", success=True, notes=deed))
        held = (
            "Luther stays on the cold stone and lets the song finish in his skull. "
            "The star-path is only a shape, not a map. Pale light still pools on the medallion. "
            "Down the corridor the dark waits, unentered."
        )
        ok = json.dumps({"prose": held, "npc_lines": []})
        self.assertIsNotNone(parse_storyteller_response(ok, "en", success=True, notes=deed))

    def test_stay_put_hit_stub_does_not_open_the_way(self):
        locale, payload = storyteller_input(
            {
                "locale": "en",
                "kind": "action",
                "actor_name": "Luther",
                "notes": "I stay where I am, without going down the corridor.",
                "total": 12,
                "success": True,
            }
        )
        out = fallback_storyteller(locale, payload, say=False)
        self.assertIn("holds the beat", out["prose"])
        self.assertNotIn("the way opens", out["prose"])
        self.assertNotIn("steps into", out["prose"])

    def test_turkish_stay_put_hit_stub_does_not_echo_or_open(self):
        notes = "Madolyonu alığ oymaların uğultusunu dinliyorum. Olduğum yerde kalıyorum"
        self.assertTrue(stay_put_deed(notes))
        self.assertEqual(table_deed(notes), "")
        locale, payload = storyteller_input(
            {
                "locale": "tr",
                "kind": "action",
                "actor_name": "Floc",
                "notes": notes,
                "total": 19,
                "success": True,
            }
        )
        out = fallback_storyteller(locale, payload, say=False)
        self.assertIn("yerinde kalır", out["prose"])
        self.assertNotIn("19", out["prose"])
        self.assertNotIn("yol açılır", out["prose"])
        self.assertNotIn("Madolyonu", out["prose"])
        self.assertNotIn("dinliyorum", out["prose"])

    def test_miss_stubs_change_the_room(self):
        notes = "Sesin kaynağına doğru yavaş adımlarla ilerliyorum"
        first = storyteller_input(
            {
                "locale": "tr",
                "kind": "action",
                "actor_name": "Floc",
                "notes": notes,
                "total": 4,
                "success": False,
            }
        )[1]
        a = fallback_storyteller("tr", first, say=False)["prose"]
        second = storyteller_input(
            {
                "locale": "tr",
                "kind": "action",
                "actor_name": "Floc",
                "notes": notes,
                "total": 8,
                "success": False,
                "prior": [a],
            }
        )[1]
        b = fallback_storyteller("tr", second, say=False)["prose"]
        self.assertNotEqual(a, b)
        self.assertNotIn("ilerliyorum", a)
        self.assertNotIn("Taş susar", a)
        self.assertNotIn("yol açılır", a)
        self.assertIn("kaçırır", a)
        self.assertIn("kaçırır", b)

    def test_hit_stub_does_not_open_the_way(self):
        locale, payload = storyteller_input(
            {
                "locale": "tr",
                "kind": "action",
                "actor_name": "Floc",
                "notes": "Yüksek bir sesle bağırıyorum KİM VAR ORDA",
                "total": 16,
                "success": True,
            }
        )
        out = fallback_storyteller(locale, payload, say=False)
        self.assertIn("Floc", out["prose"])
        self.assertNotIn("tamamlar", out["prose"])
        self.assertNotIn("16", out["prose"])
        self.assertNotIn("yol açılır", out["prose"])
        self.assertNotIn("bağırıyorum", out["prose"])

    def test_recover_truncated_json_prose(self):
        raw = (
            '{"prose":"Floc ayağa kalkar ve koridordaki metal sürtünmesine doğru bakar. '
            "Uğultu bir an kesilir. Madalyon avucunda soğur, kazınmış yazı yerinde kalır."
        )
        parsed = parse_storyteller_response(
            raw, "tr", success=True, notes="Ayağa kalkıp sesin kaynağını bulmaya çalışıyorum"
        )
        self.assertIsNotNone(parsed)
        assert parsed is not None
        self.assertIn("Floc ayağa kalkar", parsed["prose"])
        self.assertIn("Madalyon", parsed["prose"])

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
