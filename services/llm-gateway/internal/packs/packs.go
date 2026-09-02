package packs

const (
	V1      = "v1"
	V1Terse = "v1-terse"
	Stub    = "stub"
	Hub     = "hub"
	Local   = "local" // accepted alias of Hub
)

func Known(id string) bool {
	return id == V1 || id == V1Terse
}

func Default() string {
	return V1
}

func NormalizeAdapter(id string) string {
	if id == Local {
		return Hub
	}
	return id
}

// Voice is a public, PII-free prompt fragment. Weights stay out of git.
func Voice(pack, locale string) string {
	if locale != "tr" {
		locale = "en"
	}
	if pack == V1Terse {
		if locale == "tr" {
			return "v1-terse: tek cümle. Zar ve HP uydurma. Görünmez izleyici yok."
		}
		return "v1-terse: one sentence. Never invent dice or HP. Omit invisible spectators."
	}
	if locale == "tr" {
		return "v1: kısa edebi tasvir. Zar ve HP uydurma. Görünmez izleyici asla bağlamda yok."
	}
	return "v1: short literary narration. Never invent dice or HP. Omit invisible spectators from context."
}
