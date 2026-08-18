package appscope

import (
	"os"
	"strings"
)

const (
	MVP1 = "mvp1"
	MVP2 = "mvp2"
	Full = "full"
)

// Current determina qué superficie funcional puede exponer la aplicación.
//
// El valor por defecto sigue siendo MVP1 para mantener un comportamiento
// conservador cuando MVP_SCOPE no está configurado.
//
// MVP2 habilita únicamente las funcionalidades que ya fueron migradas y
// validadas sobre PostgreSQL.
//
// Full queda reservado para la superficie legacy completa mientras termina
// su adaptación progresiva.
func Current() string {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("MVP_SCOPE"))) {
	case MVP2:
		return MVP2

	case Full:
		return Full

	default:
		return MVP1
	}
}

// IsMVP1 permite mantener la superficie mínima y estable del MVP1.
func IsMVP1() bool {
	return Current() == MVP1
}

// IsMVP2 indica que la aplicación puede utilizar funcionalidades propias
// del MVP2 ya migradas a PostgreSQL.
func IsMVP2() bool {
	return Current() == MVP2
}

// IsFull identifica explícitamente la superficie legacy completa.
//
// No debe utilizarse como sinónimo de MVP2 porque puede incluir módulos que
// todavía no han sido reconstruidos o validados sobre PostgreSQL.
func IsFull() bool {
	return Current() == Full
}

// HasMVP2 habilita funcionalidades que pertenecen al MVP2 tanto en el scope
// incremental "mvp2" como en el scope completo "full".
func HasMVP2() bool {
	scope := Current()

	return scope == MVP2 ||
		scope == Full
}
