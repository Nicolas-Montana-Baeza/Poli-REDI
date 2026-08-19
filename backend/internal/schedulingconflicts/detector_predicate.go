package schedulingconflicts

import (
	"fmt"
	"sort"
)

// ============================================================================
// PREDICADO DE CONFLICTO
// ============================================================================
//
// DetectConnectedComponents mantiene el comportamiento estándar.
//
// Esta variante permite que la capa de persistencia suprima aristas que ya
// fueron autorizadas administrativamente mediante ALLOW + ALLOW.
//
// Es importante distinguir:
//
//	solapamiento temporal
//
// de:
//
//	incompatibilidad
//
// Dos ocupaciones pueden solaparse y aun así ser compatibles.
type ConflictPredicate func(
	left Occupancy,
	right Occupancy,
) bool

// DefaultConflictPredicate expone la misma regla utilizada por el detector
// estándar:
//
//	actividad ↔ actividad = conflicto
//	actividad ↔ reserva   = conflicto
//	reserva ↔ reserva     = no pertenece a scheduling_conflicts
func DefaultConflictPredicate(
	left Occupancy,
	right Occupancy,
) bool {
	return schedulingOccupanciesConflict(
		left,
		right,
	)
}

// DetectConnectedComponentsWithPredicate construye los componentes del grafo
// utilizando una regla de aristas suministrada por el llamador.
//
// Esto permite agregar conocimiento administrativo sin introducir acceso a BD
// dentro del paquete puro schedulingconflicts.
func DetectConnectedComponentsWithPredicate(
	occupancies []Occupancy,
	predicate ConflictPredicate,
) ([]Component, error) {
	if predicate == nil {
		predicate =
			DefaultConflictPredicate
	}

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
		if err := validateOccupancy(
			occupancy,
		); err != nil {
			return nil, err
		}

		if _, exists :=
			seenKeys[occupancy.Key]; exists {

			return nil, fmt.Errorf(
				"%w: %s",
				ErrDuplicateOccupancy,
				occupancy.Key,
			)
		}

		seenKeys[occupancy.Key] =
			struct{}{}
	}

	sort.SliceStable(
		normalized,
		func(i, j int) bool {
			left := normalized[i]
			right := normalized[j]

			if left.ResourceID !=
				right.ResourceID {

				return left.ResourceID <
					right.ResourceID
			}

			if !left.Start.Equal(
				right.Start,
			) {
				return left.Start.Before(
					right.Start,
				)
			}

			if !left.End.Equal(
				right.End,
			) {
				return left.End.Before(
					right.End,
				)
			}

			return left.Key < right.Key
		},
	)

	adjacency := make(
		[][]int,
		len(normalized),
	)

	for i := 0; i < len(normalized); i++ {
		left := normalized[i]

		for j := i + 1; j < len(normalized); j++ {

			right := normalized[j]

			if right.ResourceID !=
				left.ResourceID {
				break
			}

			// El orden temporal permite cortar cuando ya no puede existir
			// solapamiento directo con left.
			if !right.Start.Before(
				left.End,
			) {
				break
			}

			if !predicate(
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

	visited := make(
		[]bool,
		len(normalized),
	)

	components := []Component{}

	for root := range normalized {
		if visited[root] {
			continue
		}

		visited[root] = true

		stack := []int{
			root,
		}

		indexes := []int{}

		for len(stack) > 0 {
			last :=
				len(stack) - 1

			current :=
				stack[last]

			stack =
				stack[:last]

			indexes = append(
				indexes,
				current,
			)

			for _, neighbour := range adjacency[current] {

				if visited[neighbour] {
					continue
				}

				visited[neighbour] =
					true

				stack = append(
					stack,
					neighbour,
				)
			}
		}

		if len(indexes) < 2 {
			continue
		}

		items := make(
			[]Occupancy,
			0,
			len(indexes),
		)

		for _, index := range indexes {

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

			if !left.Start.Equal(
				right.Start,
			) {
				return left.Start.Before(
					right.Start,
				)
			}

			return left.Key < right.Key
		},
	)

	return components, nil
}
