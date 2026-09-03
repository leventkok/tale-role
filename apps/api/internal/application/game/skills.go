package game

import "strings"

// Tale Core skills. Ability is the parent trait. Keep in sync with packages/game-schema.
var skillAbility = map[string]string{
	"athletics":     "str",
	"acrobatics":    "dex",
	"stealth":       "dex",
	"arcana":        "int",
	"investigation": "int",
	"history":       "int",
	"perception":    "wis",
	"insight":       "wis",
	"survival":      "wis",
	"persuasion":    "cha",
	"deception":     "cha",
	"intimidation":  "cha",
}

func KnownSkill(id string) bool {
	_, ok := skillAbility[strings.ToLower(id)]
	if ok {
		return true
	}
	_, err := (Stats{}).Skill(id)
	return err == nil
}

func ValidSkills(ids []string) bool {
	if len(ids) == 0 {
		return true
	}
	if len(ids) > 3 {
		return false
	}
	seen := map[string]struct{}{}
	for _, id := range ids {
		id = strings.ToLower(strings.TrimSpace(id))
		if _, ok := skillAbility[id]; !ok {
			return false
		}
		if _, dup := seen[id]; dup {
			return false
		}
		seen[id] = struct{}{}
	}
	return true
}

func Proficiency(level int) int {
	if level < 1 {
		level = 1
	}
	return 2 + (level-1)/4
}

func (ch *Character) HasSkill(id string) bool {
	want := strings.ToLower(id)
	for _, s := range ch.Skills {
		if strings.ToLower(s) == want {
			return true
		}
	}
	return false
}

func (ch *Character) CheckBonus(skill string) (int, error) {
	id := strings.ToLower(strings.TrimSpace(skill))
	if id == "" {
		id = "str"
	}
	ability := id
	if parent, ok := skillAbility[id]; ok {
		ability = parent
	}
	bonus, err := ch.Stats.Skill(ability)
	if err != nil {
		return 0, err
	}
	if _, isSkill := skillAbility[id]; isSkill && ch.HasSkill(id) {
		bonus += Proficiency(ch.Level)
	}
	return bonus, nil
}

func MaxHP(stats Stats, level int) int {
	if level < 1 {
		level = 1
	}
	return 8 + stats.CON + level
}

func (ch *Character) GrantXP(n int) {
	if n <= 0 {
		return
	}
	if ch.Level < 1 {
		ch.Level = 1
	}
	ch.XP += n
	for ch.XP >= 100*ch.Level {
		ch.XP -= 100 * ch.Level
		ch.Level++
		ch.HP = MaxHP(ch.Stats, ch.Level)
	}
}
