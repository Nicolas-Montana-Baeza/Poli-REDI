package appscope

import (
	"os"
	"strings"
)

const (
	MVP1 = "mvp1"
	Full = "full"
)

// Current keeps the local runtime on the deliberately small MVP1 surface.
// The legacy/full surface must be explicitly requested while it is being
// adapted to PostgreSQL.
func Current() string {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("MVP_SCOPE")), Full) {
		return Full
	}
	return MVP1
}

func IsMVP1() bool {
	return Current() == MVP1
}
