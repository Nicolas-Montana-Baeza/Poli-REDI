package validators

import (
	"strings"
)

func NormalizeRUT(value string) string {
	value = strings.TrimSpace(strings.ToUpper(value))
	value = strings.ReplaceAll(value, ".", "")
	value = strings.ReplaceAll(value, " ", "")

	if value == "" {
		return ""
	}

	parts := strings.Split(value, "-")

	if len(parts) == 2 {
		return parts[0] + "-" + parts[1]
	}

	if len(value) < 2 {
		return value
	}

	return value[:len(value)-1] + "-" + value[len(value)-1:]
}

func IsValidRUT(value string) bool {
	value = NormalizeRUT(value)

	parts := strings.Split(value, "-")

	if len(parts) != 2 {
		return false
	}

	number := parts[0]
	verifier := parts[1]

	if len(number) < 7 || len(number) > 8 || len(verifier) != 1 {
		return false
	}

	for _, char := range number {
		if char < '0' || char > '9' {
			return false
		}
	}

	if verifier != "K" {
		for _, char := range verifier {
			if char < '0' || char > '9' {
				return false
			}
		}
	}

	sum := 0
	multiplier := 2

	for index := len(number) - 1; index >= 0; index-- {
		sum += int(number[index]-'0') * multiplier
		multiplier++

		if multiplier > 7 {
			multiplier = 2
		}
	}

	expectedValue := 11 - (sum % 11)
	expected := ""

	switch expectedValue {
	case 11:
		expected = "0"
	case 10:
		expected = "K"
	default:
		expected = string(rune('0' + expectedValue))
	}

	return verifier == expected
}
