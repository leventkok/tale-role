"""Build synthetic/storyteller.jsonl with invariants the chat window cannot keep."""

from __future__ import annotations

import json
import random
import re
from itertools import product
from pathlib import Path

OUT = Path(__file__).resolve().parent / "synthetic" / "storyteller.jsonl"
EMAIL = re.compile(r"[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}", re.I)

ACTORS = [
    "Iri", "Cal", "Nera", "Voss", "Kade", "Sera", "Bram", "Lute", "Osa", "Fen",
    "Mira", "Tor", "Pell", "Rook", "Anwen", "Dax", "Hesh", "Ilka", "Jor", "Nyx",
    "Orin", "Quill", "Ryn", "Sael", "Tess", "Ulric", "Vela", "Wren", "Yara", "Zev",
    "Ash", "Bex", "Cora", "Dell", "Emry", "Faye",
]

ROOMS = [
    ("Ashwood", "Külkoru"),
    ("Saltgate", "Tuzkapı"),
    ("Mirekeep", "Bataklıkhisar"),
    ("Cinderhall", "Koru Salon"),
    ("Lowbridge", "Altköprü"),
    ("Rookspire", "Kalekarga"),
    ("Fenmarket", "Batakpazar"),
    ("Glasswell", "Camkuyu"),
    ("Redstair", "Kızılmerdiven"),
    ("Owlcleft", "Baykuş Yarık"),
    ("Ironwake", "Demiriz"),
    ("Dustcourt", "Tozavlu"),
    ("Moonwell", "Ay Kuyusu"),
    ("Harbark", "Limankabuk"),
    ("Thornfen", "Dikenbatak"),
    ("Silverdock", "Gümüşiskele"),
    ("Graveorchard", "Mezar Bahçe"),
    ("Windcut", "Rüzgaryar"),
    ("Amberpit", "Kehribar Çukur"),
    ("Nightrift", "Gece Yarık"),
]

EN_FORCE = [
    "force the oak door", "rip the nailed hatch", "kick the cellar lid",
    "shoulder the bar gate", "wrench the stuck lever", "heave the stone slab",
    "smash the crate wall", "pull the rusted portcullis", "ram the wagon tongue",
    "break the frozen latch", "haul the fallen beam", "pry the sealed coffer",
]
EN_CLIMB = [
    "climb the ivy wall", "scale the slick chimney", "shimmy up the rope",
    "mount the palisade", "ascend the cliff stair", "clamber over the rubble",
    "swarm the watch tower", "creep up the drainpipe", "top the orchard wall",
    "free-climb the wet rock",
]
EN_SNEAK = [
    "slip past the sentry", "ghost through the kitchen", "edge along the rafters",
    "crawl under the wagon", "melt into the market crowd", "pad across the gravel",
    "slide behind the tapestry", "duck the lantern arc", "skulk the alley mouth",
    "breathe quiet in the pantry",
]
EN_TALK = [
    "talk the guard into looking away", "haggle the horse trader down",
    "plead innocence to the magistrate", "charm the barmaid for gossip",
    "lie about the missing crate", "soothe the war hound", "bluff the checkpoint",
    "beg the priest to open the crypt", "flatter the proud duchess",
    "bargain for the last ferry",
]
EN_LOCK = [
    "pick the lock quietly", "rake the cheap padlock", "tickle the desk tumbler",
    "ease the window latch", "decode the cipher ring", "reset the broken rune",
    "solve the mirror riddle", "recall the tunnel turns", "map the star puzzle",
    "unseal the waxed letter",
]
EN_GUTS = [
    "hold through the poison feast", "shrug off the serpent venom",
    "march through the sandstorm", "endure the plague ward air",
    "stay seated through the sleeping draught", "keep moving on a bad ankle",
    "resist the siren hush", "stand after the blood loss", "breathe the smoke hall",
    "weather the mountain cold",
]
EN_NOTICE = [
    "notice the wine's off color", "catch the whisper down the hall",
    "sense the ambush ahead", "spot the missing ledger page",
    "hear the stair creak twice", "read the scuff on the lintel",
    "taste metal on the wind", "see the false brick seam",
    "feel eyes on the treeline", "clock the extra set of boots",
]
EN_ACTION_NOTES = EN_FORCE + EN_CLIMB + EN_SNEAK + EN_TALK + EN_LOCK + EN_GUTS + EN_NOTICE

EN_TAILS = [
    "before dawn", "in the dark", "while the bell rings", "under wet canvas",
    "with a borrowed tool", "without waking the dog", "as the fog thickens",
    "after the last song", "during the shift change", "on a held breath",
    "against the wind", "with both hands", "through the smoke",
    "without a word", "after a count of three", "before the next bell",
    "while rain hits the roof", "as embers die", "on slick boots",
    "with a split glove", "after the laugh dies", "before anyone turns",
    "while the kettle sings", "as the choir starts", "on an empty stomach",
]

EN_PASS = [
    "skip this round", "let the turn slide by", "stand down in the brawl",
    "lower the blade and wait", "concede the move", "save my strength",
    "hold my seat and pass", "give the round away", "stay my hand",
    "let the others go first", "refuse the opening", "keep the powder dry",
    "yield the floor", "sit this clash out", "wave the action off",
]
EN_PASS_WHY = [
    "the wound is fresh", "the lantern is dying", "the child is still hidden",
    "the oath forbids it", "the rain blinds the shot", "the captain already moved",
    "the floor might give", "the spell is still cooling", "the hostage can hear us",
    "the rope is fraying", "the watch is too close", "the story needs a pause",
    "the barrel is empty", "the gate just opened", "the ally is mid-climb",
]

EN_WAIT_WHERE = [
    "the stair", "the well", "the ramparts", "the fog line", "the dock",
    "the chapel door", "the treeline", "the kitchen hatch", "the loft ladder",
    "the postern", "the bridge rope", "the market stall", "the crypt mouth",
    "the stable gate", "the window slit", "the cistern rim", "the bell tower",
    "the orchard wall", "the ferry slip", "the ash heap",
]
EN_WAIT_HOW = [
    "listen at", "watch", "hold at", "keep vigil on", "wait out the hour at",
    "ear to", "eyes on", "hold position by", "count breaths at", "stand quiet at",
]

TR_FORCE = [
    "meşe kapıyı omuzla", "çivilenmiş kapağı sök", "kiler kapağını tekmele",
    "demir parmaklığı zorla", "sıkışan kolu aşağı çek", "taş levhayı taşı",
    "sandık duvarını parçala", "paslı savak kapısını çek", "kağnı okunu vur",
    "donmuş mandalı kır", "düşen kirişi kaldır", "mühürlü sandığı kaldır",
]
TR_CLIMB = [
    "sarmaşıklı duvara tırman", "kaygan bacadan çık", "ipten yukarı süzül",
    "çit palankasına tırman", "kaya merdivenini çık", "molozun üstünden aş",
    "gözetleme kulesine çık", "iniş borusundan tırman", "bahçe duvarını aş",
    "ıslak kayaya tırman",
]
TR_SNEAK = [
    "nöbetçinin yanından sız", "mutfaktan hayalet gibi geç", "kiriş boyunca ilerle",
    "kağnının altından sürün", "pazar kalabalığına karış", "çakılın üstünde yürü",
    "halının arkasına kay", "fener yayından kaç", "çıkmaz sokak ağzında bekle",
    "kilerde nefesi tut",
]
TR_TALK = [
    "muhafızı başka tarafa bakmaya ikna et", "at tüccarını indir",
    "yargıca masum olduğunu söyle", "barmaidden dedikodu kopar",
    "eksik sandık için yalan söyle", "savaş köpeğini yatıştır", "kontrol noktasında blöf yap",
    "rahibi mahzeni açmaya ikna et", "kibirli düşesi pohpohla",
    "son vapur için pazarlık et",
]
TR_LOCK = [
    "kilidi sessizce maymuncukla", "ucuz asma kilidi tara", "masa mandalını gıdıkla",
    "pencere mandalını aç", "şifre yüzüğünü çöz", "kırık rünü kur",
    "ayna bilmecesini çöz", "tünel dönüşlerini hatırla", "yıldız haritasını çöz",
    "mühürlü mektubu aç",
]
TR_GUTS = [
    "zehirli sofrada yerinde kal", "yılan zehrini silk",
    "kum fırtınasında yürü", "veba koğuşunun havasına dayan",
    "uyku şerbetine diren", "buruk bilekle ilerle",
    "denizkızı sessizliğine diren", "kan kaybından sonra ayakta dur", "dumanlı salonda nefes al",
    "dağ soğuğuna dayan",
]
TR_NOTICE = [
    "şarabın renginin kaçtığını fark et", "koridordaki fısıltıyı yakala",
    "öndeki pusuyu sez", "defterin eksik sayfasını gör",
    "merdivenin iki kez gıcırdadığını duy", "sövedeki sıyrığı oku",
    "rüzgarda metal tadı al", "sahte tuğla derzini gör",
    "ağaç sınırında bakış hisset", "fazladan çizmeyi say",
]
TR_ACTION_NOTES = TR_FORCE + TR_CLIMB + TR_SNEAK + TR_TALK + TR_LOCK + TR_GUTS + TR_NOTICE

TR_TAILS = [
    "şafaktan önce", "karanlıkta", "çan çalarken", "ıslak branda altında",
    "ödünç aletle", "köpeği uyandırmadan", "sis koyulaşırken",
    "son şarkıdan sonra", "nöbet değişiminde", "nefesi tutarak",
    "rüzgara karşı", "iki elle", "dumanın içinden",
    "sözsüz", "üç sayınca", "bir sonraki çandan önce",
    "yağmur çatıyı vururken", "korlar sönerken", "kaygan çizmeyle",
    "yarık eldivenle", "kahkaha bitince", "kimse dönmeden",
    "çaydanlık ötünce", "koro başlarken", "aç karnına",
]

TR_PASS = [
    "bu turu pas geç", "eli bırak", "kavgada dur",
    "kılıcı indir ve bekle", "hamleyi teslim et", "gücünü sakla",
    "otur ve geç", "turu başkasına bırak", "elini tut",
    "önce diğerleri gitsin", "açılışı reddet", "barutu kuru tut",
    "sözü bırak", "bu çatışmayı otur", "eylemi sav",
]
TR_PASS_WHY = [
    "yara taze", "fener sönmek üzere", "çocuk hâlâ gizli",
    "yemin yasaklıyor", "yağmur nişanı bozuyor", "yüzbaşı zaten oynadı",
    "taban çökebilir", "büyü hâlâ soğuyor", "rehine bizi duyabilir",
    "ip çözülüyor", "nöbet çok yakın", "hikaye bir durak istiyor",
    "fıçı boş", "kapı yeni açıldı", "yoldaş tırmanışta",
]

TR_WAIT_WHERE = [
    "merdiven", "kuyu", "sur", "sis çizgisi", "iskele",
    "şapel kapısı", "ağaç sınırı", "mutfak kapağı", "asma merdiven",
    "arka kapı", "köprü halatı", "pazar tezgahı", "mahzene ağız",
    "ahır kapısı", "mazgal", "sarnıç kenarı", "çan kulesi",
    "bahçe duvarı", "vapur iskelesi", "kül yığını",
]
TR_WAIT_HOW = [
    "dinle", "gözetle", "bekle", "nöbet tut", "saati geçir",
    "kulağını ver", "gözünü dik", "konumunu koru", "nefes say", "sessiz dur",
]

EN_OK = [
    "The bar splinters and cold air spills through.",
    "The latch yields; a gap opens.",
    "The engine's number holds and the way clears.",
    "Stone dust rains down as it gives.",
    "No shout follows. It worked.",
    "The hinge screams once, then obeys.",
    "A narrow path appears where none was.",
    "The watch does not turn. Good.",
    "Hands come away bloody but the job is done.",
    "The room answers. The thing moves.",
]
EN_FAIL = [
    "Nothing yields. The watch stirs.",
    "The lock holds. A pin snaps.",
    "Boots scrape closer. The attempt dies.",
    "The wood does not care.",
    "Someone coughs on the far side.",
    "The tool slips. Time is gone.",
    "A dog answers from the yard.",
    "The rune stays dark.",
    "Rain washes the chance away.",
    "The body will not take more of this.",
]
TR_OK = [
    "Menteşe inler, içeri soğuk dolar.",
    "Mandal teslim olur; bir aralık açılır.",
    "Motorun sayısı tutar, yol açılır.",
    "Taş tozu yağar, şey yerinden oynar.",
    "Çığlık gelmez. İş bitti.",
    "Menteşe bir kez bağırır, sonra uyar.",
    "Olmayan bir yarık belirir.",
    "Nöbet dönmez. İyi.",
    "Eller kanlıdır ama iş olur.",
    "Oda cevap verir. Şey hareket eder.",
]
TR_FAIL = [
    "Hiçbir şey teslim olmaz. Nöbet kımıldar.",
    "Kilit durur. Bir pim kopar.",
    "Çizmeler yaklaşır. Deneme ölür.",
    "Ahşap aldırmaz.",
    "Ötede biri öksürür.",
    "Alet kayar. Zaman biter.",
    "Avludan bir köpek cevaplar.",
    "Rün karanlık kalır.",
    "Yağmur şansı siler.",
    "Beden daha fazlasını almaz.",
]

NPC = [
    ("npc-guard-1", "Stay where you are.", "Olduğun yerde kal."),
    ("npc-guard-2", "I heard the hinge.", "Menteşeyi duydum."),
    ("npc-innkeep-1", "We saw nothing.", "Bir şey görmedik."),
    ("npc-priest-1", "The crypt stays shut.", "Mahzen kapalı kalır."),
    ("npc-trader-1", "That price is final.", "O fiyat kesin."),
    ("npc-hound-1", "(a low growl)", "(alçak bir hırıltı)"),
    ("npc-dock-1", "Tide's turning.", "Gelgit dönüyor."),
    ("npc-child-1", "Don't look down.", "Aşağı bakma."),
    ("npc-captain-1", "Hold the line.", "Hattı tutun."),
    ("npc-scribe-1", "The page is missing.", "Sayfa yok."),
]


def combine(stems: list[str], tails: list[str], joiner: str = " ") -> list[str]:
    return [f"{s}{joiner}{t}" for s, t in product(stems, tails)]


def unique_seq(rng: random.Random, pool: list[str], n: int, used: set[str]) -> list[str]:
    leftover = [p for p in pool if p not in used]
    rng.shuffle(leftover)
    if len(leftover) < n:
        raise RuntimeError(f"pool too small: need {n}, have {len(leftover)}")
    picked = leftover[:n]
    used.update(picked)
    return picked


def d20(rng: random.Random) -> tuple[list[int], int, bool]:
    roll = rng.randint(1, 20)
    mod = rng.randint(-2, 5)
    total = max(1, roll + mod)
    return [roll], total, total >= 12


def d6x2(rng: random.Random) -> tuple[list[int], int, bool]:
    a, b = rng.randint(1, 6), rng.randint(1, 6)
    mod = rng.randint(-2, 5)
    total = max(1, a + b + mod)
    return [a, b], total, total >= 8


def companions(rng: random.Random, actor: str) -> list[str]:
    others = [a for a in ACTORS if a != actor]
    n = rng.randint(0, 2)
    extra = rng.sample(others, n) if n else []
    return [actor, *extra]


def prose_action(locale: str, actor: str, room: str, notes: str, total: int, ok: bool, tag: str) -> str:
    if locale == "en":
        if ok:
            return f"{actor} tries to {notes} in {room}. The engine's {total} lands true. {tag}"
        return f"{actor} tries to {notes} in {room}. The die reads {total}. {tag}"
    if ok:
        return f"{actor}, {room} içinde, {notes}. Motorun {total}'i tutar. {tag}"
    return f"{actor}, {room} içinde, {notes}. Zar {total} der. {tag}"


def prose_pass(locale: str, actor: str, room: str, notes: str, tag: str) -> str:
    if locale == "en":
        return f"{actor} in {room} chooses to {notes}. No dice this time. {tag}"
    return f"{actor}, {room} içinde, {notes}. Bu kez zar yok. {tag}"


def prose_wait(locale: str, actor: str, room: str, notes: str, tag: str) -> str:
    if locale == "en":
        return f"{actor} waits in {room}, {notes}, and does not roll. {tag}"
    return f"{actor} {room} içinde bekler, {notes}, zar atılmaz. {tag}"


def slots() -> list[dict]:
    rows: list[dict] = []
    for locale in ("en", "tr"):
        for _ in range(600):
            rows.append({"locale": locale, "kind": "action", "dice": "d20"})
        for _ in range(100):
            rows.append({"locale": locale, "kind": "action", "dice": "2d6"})
        for _ in range(150):
            rows.append({"locale": locale, "kind": "pass", "dice": "d20"})
        for _ in range(150):
            rows.append({"locale": locale, "kind": "wait", "dice": "d20"})
    return rows


def build(seed: int = 42) -> list[dict]:
    rng = random.Random(seed)
    plan = slots()
    rng.shuffle(plan)

    used_notes: set[str] = set()
    used_prose: set[str] = set()
    en_action = unique_seq(rng, combine(EN_ACTION_NOTES, EN_TAILS), 700, used_notes)
    tr_action = unique_seq(rng, combine(TR_ACTION_NOTES, TR_TAILS), 700, used_notes)
    en_pass = unique_seq(rng, combine(EN_PASS, EN_PASS_WHY, " because "), 150, used_notes)
    tr_pass = unique_seq(rng, combine(TR_PASS, TR_PASS_WHY, " çünkü "), 150, used_notes)
    en_wait = unique_seq(
        rng,
        combine(
            [f"{h} {w}" for h, w in product(EN_WAIT_HOW, EN_WAIT_WHERE)],
            ["quietly", "a while", "until the lamp dies", "in the wet", "without a word"],
        ),
        150,
        used_notes,
    )
    tr_wait = unique_seq(
        rng,
        combine(
            [f"{w} {h}" for h, w in product(TR_WAIT_HOW, TR_WAIT_WHERE)],
            ["sessizce", "bir süre", "lamba sönene dek", "ıslaklıkta", "sözsüz"],
        ),
        150,
        used_notes,
    )

    ai = {"en": 0, "tr": 0}
    pi = {"en": 0, "tr": 0}
    wi = {"en": 0, "tr": 0}

    npc_idx = list(range(2000))
    rng.shuffle(npc_idx)
    npc_set = set(npc_idx[:250])

    rows: list[dict] = []
    for i, spec in enumerate(plan):
        locale = spec["locale"]
        kind = spec["kind"]
        dice = spec["dice"]
        actor = ACTORS[i % len(ACTORS)]
        room_en, room_tr = ROOMS[i % len(ROOMS)]
        room = room_en if locale == "en" else room_tr
        hold = "Masa durur." if locale == "tr" else "The table waits."

        if kind == "action":
            notes = (en_action if locale == "en" else tr_action)[ai[locale]]
            ai[locale] += 1
            rolls, total, success = d20(rng) if dice == "d20" else d6x2(rng)
            pool = (EN_OK if locale == "en" else TR_OK) if success else (EN_FAIL if locale == "en" else TR_FAIL)
            text = prose_action(locale, actor, room, notes, total, success, pool[i % len(pool)])
        elif kind == "pass":
            notes = (en_pass if locale == "en" else tr_pass)[pi[locale]]
            pi[locale] += 1
            rolls, total, success = [], 0, None
            dice = "d20"
            text = prose_pass(locale, actor, room, notes, hold)
        else:
            notes = (en_wait if locale == "en" else tr_wait)[wi[locale]]
            wi[locale] += 1
            rolls, total, success = [], 0, None
            dice = "d20"
            text = prose_wait(locale, actor, room, notes, hold)

        if text.lower() in used_prose:
            text = f"{text} ({actor} {i})"
        used_prose.add(text.lower())

        npc_lines = []
        if i in npc_set:
            nid, en_t, tr_t = NPC[i % len(NPC)]
            npc_lines = [{"npc_id": nid, "text": en_t if locale == "en" else tr_t}]

        rows.append(
            {
                "locale": locale,
                "input": {
                    "actor": actor,
                    "room": room,
                    "kind": kind,
                    "dice": dice,
                    "rolls": rolls,
                    "total": total,
                    "success": success,
                    "notes": notes,
                    "presence": companions(rng, actor),
                },
                "output": {"prose": text, "npc_lines": npc_lines},
            }
        )
    return rows


def validate(rows: list[dict]) -> None:
    if len(rows) != 2000:
        raise SystemExit(f"count {len(rows)}")
    loc = {"en": 0, "tr": 0}
    kinds = {"action": 0, "pass": 0, "wait": 0}
    two = {"en": 0, "tr": 0}
    notes, prose = set(), set()
    npc_n = 0
    for row in rows:
        loc[row["locale"]] += 1
        inp, out = row["input"], row["output"]
        kinds[inp["kind"]] += 1
        if inp["dice"] == "2d6":
            two[row["locale"]] += 1
        blob = json.dumps(row, ensure_ascii=False)
        if EMAIL.search(blob) or "system_admin" in blob:
            raise SystemExit("leak")
        nkey, pkey = inp["notes"].casefold(), out["prose"].casefold()
        if nkey in notes or pkey in prose:
            raise SystemExit("duplicate notes/prose")
        notes.add(nkey)
        prose.add(pkey)
        if inp["kind"] == "action":
            if inp["dice"] == "d20":
                if len(inp["rolls"]) != 1 or not (1 <= inp["rolls"][0] <= 20):
                    raise SystemExit("bad d20")
                if inp["success"] != (inp["total"] >= 12):
                    raise SystemExit("d20 success mismatch")
            else:
                if inp["dice"] != "2d6" or len(inp["rolls"]) != 2:
                    raise SystemExit("bad 2d6")
                if inp["success"] != (inp["total"] >= 8):
                    raise SystemExit("2d6 success mismatch")
            if inp["success"] is True and str(inp["total"]) not in out["prose"]:
                raise SystemExit("success missing total digits")
        else:
            if inp["rolls"] != [] or inp["total"] != 0 or inp["success"] is not None:
                raise SystemExit("pass/wait dice")
            if re.search(r"\b(d20|2d6)\b", out["prose"]):
                raise SystemExit("pass/wait named dice")
        if out["npc_lines"]:
            npc_n += 1
            for line in out["npc_lines"]:
                if line["npc_id"] == "system_admin" or not line["text"]:
                    raise SystemExit("bad npc")
    if loc != {"en": 1000, "tr": 1000}:
        raise SystemExit(loc)
    if kinds != {"action": 1400, "pass": 300, "wait": 300}:
        raise SystemExit(kinds)
    if two != {"en": 100, "tr": 100}:
        raise SystemExit(two)
    if not (240 <= npc_n <= 260):
        raise SystemExit(f"npc {npc_n}")


def main() -> None:
    rows = build()
    validate(rows)
    OUT.write_text(
        "\n".join(json.dumps(r, ensure_ascii=False, separators=(",", ":")) for r in rows) + "\n",
        encoding="utf-8",
    )
    print(f"wrote {len(rows)} lines -> {OUT}")


if __name__ == "__main__":
    main()
