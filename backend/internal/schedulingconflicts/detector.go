package schedulingconflicts

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// ============================================================================
// TIPOS DE OCUPACIÓN
// ============================================================================
//
// El detector no necesita conocer detalles académicos ni de reservas.
// Solamente necesita saber:
//
//   - qué recurso está ocupado;
//   - qué entidad genera la ocupación;
//   - cuál es su intervalo concreto.
//
// Esto mantiene separado:
//
//   detección de conflictos
//            ↓
//   resolución administrativa
//
// y permite agregar nuevos tipos de ocupación en el futuro sin cambiar el
// algoritmo de agrupación.

type OccupancyKind string

const (
	OccupancyKindInstitutionalActivity OccupancyKind = "INSTITUTIONAL_ACTIVITY"
	OccupancyKindReservation           OccupancyKind = "RESERVATION"
)

// ============================================================================
// OCUPACIÓN CONCRETA
// ============================================================================

// Occupancy representa una ocurrencia concreta dentro del calendario.
//
// Para una actividad recurrente:
//
//	ActivityID + ScheduleID
//
// identifican la regla que originó la ocupación, mientras:
//
//	Start + End
//
// identifican la ocurrencia concreta.
//
// Key debe ser única para la ocurrencia.
//
// Ejemplos:
//
//	activity:12:schedule:4:2026-09-02T10:00:00-04:00
//	reservation:87
//
// La Key es una identidad operacional de Go; no reemplaza las FKs almacenadas
// posteriormente en scheduling_conflict_items.
type Occupancy struct {
	Key string

	ResourceID int
	Kind       OccupancyKind

	InstitutionalActivityID *int
	ScheduleID              *int
	ReservationID           *int

	Start time.Time
	End   time.Time
}

// ============================================================================
// COMPONENTE DE CONFLICTO
// ============================================================================

// Component representa un componente conectado del grafo de conflictos.
//
// Cada ocupación es conceptualmente un nodo.
//
// Existe una arista entre dos nodos cuando:
//
//   - pertenecen al mismo recurso;
//   - sus intervalos [start,end) se solapan.
//
// Por definición, un componente con menos de dos elementos no constituye
// conflicto.
type Component struct {
	ResourceID int
	Items      []Occupancy
}

// ============================================================================
// ERRORES
// ============================================================================

var (
	ErrInvalidOccupancy = errors.New(
		"ocupación de scheduling inválida",
	)

	ErrDuplicateOccupancy = errors.New(
		"ocupación de scheduling duplicada",
	)

	ErrInvalidComponent = errors.New(
		"componente de scheduling inválido",
	)
)

// ============================================================================
// DETECTOR
// ============================================================================

// DetectConnectedComponents construye componentes conectados de conflictos.
//
// IMPORTANTE:
//
// No buscamos únicamente pares A-B.
//
// Ejemplo:
//
//	A 10:00 ───── 11:00
//	     B 10:30 ───── 11:30
//	                  C 11:15 ───── 12:00
//
// A no se solapa directamente con C.
//
// Sin embargo:
//
//	A ↔ B ↔ C
//
// forma un único componente conectado de tres elementos.
//
// Como los conflictos temporales forman un grafo de intervalos, podemos
// obtener sus componentes eficientemente ordenando por hora de inicio y
// manteniendo el extremo máximo alcanzado por el componente actual
func DetectConnectedComponents(
	occupancies []Occupancy,
) ([]Component, error) {
	if len(occupancies) == 0 {
		return []Component{}, nil
	}

	normalized := make(
		[]Occupancy,
		len(occupancies),
	)

	copy(normalized, occupancies)

	seenKeys := make(
		map[string]struct{},
		len(normalized),
	)

	for _, occupancy := range normalized {
		if err := validateOccupancy(occupancy); err != nil {
			return nil, err
		}

		if _, exists := seenKeys[occupancy.Key]; exists {
			return nil, fmt.Errorf(
				"%w: %s",
				ErrDuplicateOccupancy,
				occupancy.Key,
			)
		}

		seenKeys[occupancy.Key] = struct{}{}
	}

	// El orden estable mantiene resultados y tests deterministas.
	sort.SliceStable(
		normalized,
		func(i, j int) bool {
			left := normalized[i]
			right := normalized[j]

			if left.ResourceID != right.ResourceID {
				return left.ResourceID < right.ResourceID
			}

			if !left.Start.Equal(right.Start) {
				return left.Start.Before(right.Start)
			}

			if !left.End.Equal(right.End) {
				return left.End.Before(right.End)
			}

			return left.Key < right.Key
		},
	)

	// ========================================================================
	// GRAFO DE INCOMPATIBILIDADES
	// ========================================================================
	//
	// No basta con preguntar si dos intervalos se solapan.
	//
	// Una arista significa:
	//
	//     las dos ocupaciones son incompatibles
	//
	// En particular, reserva ↔ reserva no pertenece a este módulo.
	// La compatibilidad entre reservas RESERVABLE / OPEN_USE ya es protegida
	// por el subsistema de reservas.
	//
	// Para MVP2 scheduling_conflicts administra:
	//
	//     actividad ↔ actividad
	//     actividad ↔ reserva
	//
	// El volumen diario de ocupaciones es pequeño, por lo que O(n²) ofrece una
	// implementación mucho más explícita y segura que asumir que todo
	// solapamiento temporal constituye una arista.

	adjacency := make(
		[][]int,
		len(normalized),
	)

	for i := 0; i < len(normalized); i++ {
		left := normalized[i]

		for j := i + 1; j < len(normalized); j++ {
			right := normalized[j]

			// Como normalized está ordenado por recurso, no puede volver a
			// aparecer el recurso de left después de este punto.
			if right.ResourceID != left.ResourceID {
				break
			}

			// Como también está ordenado por inicio, una vez que el siguiente
			// intervalo comienza después del fin de left ya no puede existir
			// una arista directa con left.
			if !right.Start.Before(left.End) {
				break
			}

			if !schedulingOccupanciesConflict(
				left,
				right,
			) {
				continue
			}

			adjacency[i] = append(
				adjacency[i],
				j,
			)

			adjacency[j] = append(
				adjacency[j],
				i,
			)
		}
	}

	// ========================================================================
	// COMPONENTES CONECTADOS
	// ========================================================================

	visited := make(
		[]bool,
		len(normalized),
	)

	components := []Component{}

	for startIndex := range normalized {
		if visited[startIndex] {
			continue
		}

		visited[startIndex] = true

		stack := []int{
			startIndex,
		}

		componentIndexes := []int{}

		for len(stack) > 0 {
			lastIndex := len(stack) - 1

			current := stack[lastIndex]

			stack =
				stack[:lastIndex]

			componentIndexes = append(
				componentIndexes,
				current,
			)

			for _, neighbour := range adjacency[current] {

				if visited[neighbour] {
					continue
				}

				visited[neighbour] = true

				stack = append(
					stack,
					neighbour,
				)
			}
		}

		// Una ocupación aislada no constituye conflicto.
		if len(componentIndexes) < 2 {
			continue
		}

		items := make(
			[]Occupancy,
			0,
			len(componentIndexes),
		)

		for _, index := range componentIndexes {

			items = append(
				items,
				normalized[index],
			)
		}

		sort.SliceStable(
			items,
			func(i, j int) bool {
				if !items[i].Start.Equal(
					items[j].Start,
				) {
					return items[i].
						Start.
						Before(
							items[j].Start,
						)
				}

				if !items[i].End.Equal(
					items[j].End,
				) {
					return items[i].
						End.
						Before(
							items[j].End,
						)
				}

				return items[i].Key <
					items[j].Key
			},
		)

		components = append(
			components,
			Component{
				ResourceID: items[0].ResourceID,

				Items: items,
			},
		)
	}

	// Orden determinista entre distintos componentes.
	sort.SliceStable(
		components,
		func(i, j int) bool {
			if components[i].ResourceID !=
				components[j].ResourceID {

				return components[i].ResourceID <
					components[j].ResourceID
			}

			left :=
				components[i].Items[0]

			right :=
				components[j].Items[0]

			if !left.Start.Equal(right.Start) {
				return left.Start.Before(
					right.Start,
				)
			}

			return left.Key < right.Key
		},
	)

	return components, nil
}

// schedulingOccupanciesConflict define cuándo dos ocupaciones generan una
// arista dentro del grafo administrativo.
//
// Un simple solapamiento temporal no es suficiente.
//
// Las reservas entre sí son gestionadas por el subsistema de reservas.
// Esto es especialmente importante para OPEN_USE, donde distintos usuarios
// pueden utilizar simultáneamente el mismo recurso.
//
// Una actividad institucional sí es incompatible con otra actividad o con una
// reserva existente sobre el mismo recurso.
func schedulingOccupanciesConflict(
	left Occupancy,
	right Occupancy,
) bool {
	if left.ResourceID != right.ResourceID {
		return false
	}

	// Intervalos half-open [start,end).
	if !left.Start.Before(right.End) ||
		!right.Start.Before(left.End) {

		return false
	}

	if left.Kind ==
		OccupancyKindReservation &&
		right.Kind ==
			OccupancyKindReservation {

		return false
	}

	return true
}

// ============================================================================
// INTERVALO PROTEGIDO
// ============================================================================

// ProtectedInterval devuelve la unión temporal efectiva protegida frente a
// terceros por un componente.
//
// Ejemplo:
//
//	A = 10:00-11:30
//	B = 10:30-12:00
//	C = 11:00-13:00
//
// Las intersecciones internas son diferentes, pero mientras las tres
// ocupaciones permanezcan activas:
//
//	protected interval = 10:00-13:00
//
// La disponibilidad no debe confundirse con el intervalo exacto de
// intersección entre dos elementos.
func ProtectedInterval(
	component Component,
) (time.Time, time.Time, error) {
	if component.ResourceID <= 0 ||
		len(component.Items) < 2 {
		return time.Time{},
			time.Time{},
			ErrInvalidComponent
	}

	start := component.Items[0].Start
	end := component.Items[0].End

	for _, occupancy := range component.Items {
		if err := validateOccupancy(
			occupancy,
		); err != nil {
			return time.Time{},
				time.Time{},
				err
		}

		if occupancy.ResourceID !=
			component.ResourceID {
			return time.Time{},
				time.Time{},
				ErrInvalidComponent
		}

		if occupancy.Start.Before(start) {
			start = occupancy.Start
		}

		if occupancy.End.After(end) {
			end = occupancy.End
		}
	}

	return start, end, nil
}

// ============================================================================
// VALIDACIÓN
// ============================================================================

func validateOccupancy(
	occupancy Occupancy,
) error {
	if strings.TrimSpace(occupancy.Key) == "" ||
		occupancy.ResourceID <= 0 ||
		occupancy.Start.IsZero() ||
		occupancy.End.IsZero() ||
		!occupancy.End.After(occupancy.Start) {
		return ErrInvalidOccupancy
	}

	switch occupancy.Kind {

	case OccupancyKindInstitutionalActivity:

		if occupancy.InstitutionalActivityID == nil ||
			occupancy.ScheduleID == nil ||
			occupancy.ReservationID != nil {
			return ErrInvalidOccupancy
		}

		if *occupancy.InstitutionalActivityID <= 0 ||
			*occupancy.ScheduleID <= 0 {
			return ErrInvalidOccupancy
		}

	case OccupancyKindReservation:

		if occupancy.InstitutionalActivityID != nil ||
			occupancy.ScheduleID != nil ||
			occupancy.ReservationID == nil {
			return ErrInvalidOccupancy
		}

		if *occupancy.ReservationID <= 0 {
			return ErrInvalidOccupancy
		}

	default:
		return ErrInvalidOccupancy
	}

	return nil
}
