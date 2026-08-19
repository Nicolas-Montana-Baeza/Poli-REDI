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
// manteniendo el extremo máximo alcanzado por el componente actual.
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

	seenKeys := make(map[string]struct{}, len(normalized))

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

	// El orden estable hace determinista tanto el resultado como los tests.
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

	components := []Component{}

	currentResourceID := normalized[0].ResourceID

	currentItems := []Occupancy{
		normalized[0],
	}

	// componentEnd representa el extremo máximo alcanzado por cualquier
	// ocupación del componente actual.
	//
	// Esto es lo que permite detectar cadenas A-B-C aunque A y C no se
	// intersecten directamente.
	componentEnd := normalized[0].End

	flushCurrent := func() {
		// Una sola ocupación no constituye conflicto.
		if len(currentItems) < 2 {
			return
		}

		items := make(
			[]Occupancy,
			len(currentItems),
		)

		copy(items, currentItems)

		components = append(
			components,
			Component{
				ResourceID: currentResourceID,
				Items:      items,
			},
		)
	}

	for _, occupancy := range normalized[1:] {

		// Cambiar de recurso siempre inicia otro componente.
		if occupancy.ResourceID != currentResourceID {
			flushCurrent()

			currentResourceID = occupancy.ResourceID
			currentItems = []Occupancy{
				occupancy,
			}
			componentEnd = occupancy.End

			continue
		}

		// Intervalos half-open:
		//
		//   [10:00,11:00)
		//   [11:00,12:00)
		//
		// NO se solapan.
		//
		// Por eso usamos Start.Before(componentEnd) y no <=.
		if occupancy.Start.Before(componentEnd) {
			currentItems = append(
				currentItems,
				occupancy,
			)

			if occupancy.End.After(componentEnd) {
				componentEnd = occupancy.End
			}

			continue
		}

		// No existe conexión temporal con el componente actual.
		flushCurrent()

		currentItems = []Occupancy{
			occupancy,
		}

		componentEnd = occupancy.End
	}

	flushCurrent()

	return components, nil
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
