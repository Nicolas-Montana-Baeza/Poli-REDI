package validators

import (
	"strings"
	"unicode"
)

var blockedActivityTerms = []string{
	"puta",
	"puto",
	"putas",
	"putos",
	"mierda",
	"weon",
	"weona",
	"weones",
	"weonas",
	"wea",
	"weas",
	"ctm",
	"maricon",
	"marica",
	"pico",
	"pene",
	"sexo",
	"fuck",
	"shit",
	"bitch",
}

func NormalizeActivityName(value string) string {
	words := strings.Fields(value)

	if len(words) == 0 {
		return ""
	}

	return strings.Join(words, " ")
}

func IsValidActivityName(value string) bool {
	name := NormalizeActivityName(value)
	runes := []rune(name)

	if len(runes) < 3 || len(runes) > 120 {
		return false
	}

	hasLetter := false

	for _, char := range runes {
		if unicode.IsLetter(char) {
			hasLetter = true
			continue
		}

		if unicode.IsSpace(char) {
			continue
		}

		return false
	}

	if !hasLetter {
		return false
	}

	return !HasBlockedActivityTerms(name)
}

func HasBlockedActivityTerms(value string) bool {
	terms := map[string]bool{}

	for _, term := range blockedActivityTerms {
		terms[term] = true
	}

	for _, word := range strings.Fields(normalizeModerationText(value)) {
		if terms[word] {
			return true
		}
	}

	return false
}

func normalizeModerationText(value string) string {
	replacer := strings.NewReplacer(
		"á", "a",
		"é", "e",
		"í", "i",
		"ó", "o",
		"ú", "u",
		"ü", "u",
		"ñ", "n",
	)

	return replacer.Replace(strings.ToLower(NormalizeActivityName(value)))
}
