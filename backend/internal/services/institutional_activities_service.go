package services

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"

	"github.com/jackc/pgx/v5/pgconn"
)

// ============================================================================
// ERRORES DE DOMINIO
// ============================================================================

var (
	ErrInstitutionalActivityInvalid = errors.New(
		"los datos de la actividad institucional no son válidos",
	)

	ErrInstitutionalScheduleInvalid = errors.New(
		"la programación de la actividad institucional no es válida",
	)

	ErrInstitutionalResourceNotFound = errors.New(
		"recurso no encontrado",
	)

	ErrInstitutionalResourceInactive = errors.New(
		"el recurso está inactivo",
	)

	ErrInstitutionalActivityBlocked = errors.New(
		"la actividad intersecta un bloqueo administrativo",
	)
)

// ============================================================================
// CREACIÓN
// ============================================================================

// CreateInstitutionalActivity crea una actividad institucional programada.
//
// Flujo MVP2:
//
//	validar usuario
//	     ↓
//	validar MANAGER/admin de la unidad
//	     ↓
//	validar actividad y recurso
//	     ↓
//	validar horarios estructurados
//	     ↓
//	verificar availability_blocks dentro de la transacción
//	     ↓
//	crear SCHEDULED
//
// IMPORTANTE:
//
// El solapamiento con reservas u otras actividades NO se rechaza aquí.
// Esos casos serán registrados por el detector de scheduling conflicts.
//
// availability_blocks, en cambio, representan indisponibilidad dura.
func CreateInstitutionalActivity(
	user models.LocalAuthUser,
	request models.CreateInstitutionalActivityRequest,
) (models.InstitutionalActivity, error) {
	if err := normalizeInstitutionalActivityRequest(&request); err != nil {
		return models.InstitutionalActivity{}, err
	}

	if err := validateInstitutionalActivityRequest(request); err != nil {
		return models.InstitutionalActivity{}, err
	}

	// La autorización se evalúa antes de consultar o modificar recursos.
	if err := EnsureInstitutionalUnitManager(
		user,
		request.UnitID,
	); err != nil {
		return models.InstitutionalActivity{}, err
	}

	resource, err := repositories.GetResourceByID(
		request.ResourceID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.InstitutionalActivity{},
				ErrInstitutionalResourceNotFound
		}

		return models.InstitutionalActivity{}, err
	}

	if !resource.IsActive {
		return models.InstitutionalActivity{},
			ErrInstitutionalResourceInactive
	}

	// Evitamos que dos reglas pertenecientes a la misma actividad generen
	// ocupaciones internas solapadas. Eso sería un request incoherente, no un
	// scheduling conflict entre entidades diferentes.
	if err := validateInstitutionalScheduleSelfOverlap(
		request.Schedules,
	); err != nil {
		return models.InstitutionalActivity{}, err
	}

	activity, err :=
		repositories.CreateInstitutionalActivityWithSchedules(
			request,
			user.ID,
			models.InstitutionalActivityStatusScheduled,
		)

	if err != nil {
		return models.InstitutionalActivity{},
			mapInstitutionalActivityRepositoryError(err)
	}

	return activity, nil
}

// ============================================================================
// CONSULTA
// ============================================================================

// GetInstitutionalActivitiesForUnit permite al administrador global o al
// MANAGER correspondiente consultar la programación de una unidad.
func GetInstitutionalActivitiesForUnit(
	user models.LocalAuthUser,
	unitID int,
) ([]models.InstitutionalActivity, error) {
	if err := EnsureInstitutionalUnitManager(
		user,
		unitID,
	); err != nil {
		return nil, err
	}

	return repositories.GetInstitutionalActivitiesByUnit(unitID)
}

// ============================================================================
// NORMALIZACIÓN
// ============================================================================

func normalizeInstitutionalActivityRequest(
	request *models.CreateInstitutionalActivityRequest,
) error {
	request.ActivityType =
		strings.ToUpper(strings.TrimSpace(request.ActivityType))

	request.Title = strings.TrimSpace(request.Title)
	request.Description = strings.TrimSpace(request.Description)

	for index := range request.Schedules {
		schedule := &request.Schedules[index]

		schedule.ScheduleType =
			strings.ToUpper(strings.TrimSpace(schedule.ScheduleType))

		schedule.StartTime =
			strings.TrimSpace(schedule.StartTime)

		schedule.EndTime =
			strings.TrimSpace(schedule.EndTime)

		if schedule.SpecificDate != nil {
			value := strings.TrimSpace(*schedule.SpecificDate)
			schedule.SpecificDate = &value
		}

		if schedule.ValidFrom != nil {
			value := strings.TrimSpace(*schedule.ValidFrom)
			schedule.ValidFrom = &value
		}

		if schedule.ValidTo != nil {
			value := strings.TrimSpace(*schedule.ValidTo)
			schedule.ValidTo = &value
		}
	}

	return nil
}

// ============================================================================
// VALIDACIÓN DE ACTIVIDAD
// ============================================================================

func validateInstitutionalActivityRequest(
	request models.CreateInstitutionalActivityRequest,
) error {
	if request.UnitID <= 0 ||
		request.ResourceID <= 0 ||
		request.Title == "" ||
		len(request.Schedules) == 0 {
		return ErrInstitutionalActivityInvalid
	}

	if !validInstitutionalActivityType(
		request.ActivityType,
	) {
		return ErrInstitutionalActivityInvalid
	}

	// Toda actividad con inscripción debe declarar capacidad.
	if request.RequiresEnrollment {
		if request.Capacity == nil ||
			*request.Capacity <= 0 {
			return ErrInstitutionalActivityInvalid
		}
	} else if request.Capacity != nil {
		return ErrInstitutionalActivityInvalid
	}

	// WORKSHOP utiliza obligatoriamente inscripción individual.
	if request.ActivityType ==
		models.InstitutionalActivityTypeWorkshop &&
		!request.RequiresEnrollment {
		return ErrInstitutionalActivityInvalid
	}

	for _, schedule := range request.Schedules {
		if err := validateInstitutionalSchedule(
			schedule,
		); err != nil {
			return err
		}
	}

	return nil
}

func validInstitutionalActivityType(
	activityType string,
) bool {
	switch activityType {

	case models.InstitutionalActivityTypeAcademicClass,
		models.InstitutionalActivityTypeWorkshop,
		models.InstitutionalActivityTypeTraining,
		models.InstitutionalActivityTypeEvent,
		models.InstitutionalActivityTypeChampionship,
		models.InstitutionalActivityTypeOther:

		return true

	default:
		return false
	}
}

// ============================================================================
// VALIDACIÓN DE SCHEDULE
// ============================================================================

func validateInstitutionalSchedule(
	schedule models.CreateInstitutionalScheduleRequest,
) error {
	start, err := time.Parse(
		"15:04",
		schedule.StartTime,
	)

	if err != nil {
		return ErrInstitutionalScheduleInvalid
	}

	end, err := time.Parse(
		"15:04",
		schedule.EndTime,
	)

	if err != nil || !end.After(start) {
		return ErrInstitutionalScheduleInvalid
	}

	switch schedule.ScheduleType {

	case models.InstitutionalScheduleTypeSingle:

		if schedule.SpecificDate == nil ||
			schedule.DayOfWeek != nil ||
			schedule.ValidFrom != nil ||
			schedule.ValidTo != nil {
			return ErrInstitutionalScheduleInvalid
		}

		if _, err := parseInstitutionalDate(
			*schedule.SpecificDate,
		); err != nil {
			return ErrInstitutionalScheduleInvalid
		}

		return nil

	case models.InstitutionalScheduleTypeWeekly:

		if schedule.SpecificDate != nil ||
			schedule.DayOfWeek == nil ||
			schedule.ValidFrom == nil ||
			schedule.ValidTo == nil {
			return ErrInstitutionalScheduleInvalid
		}

		if *schedule.DayOfWeek < 1 ||
			*schedule.DayOfWeek > 7 {
			return ErrInstitutionalScheduleInvalid
		}

		validFrom, err := parseInstitutionalDate(
			*schedule.ValidFrom,
		)

		if err != nil {
			return ErrInstitutionalScheduleInvalid
		}

		validTo, err := parseInstitutionalDate(
			*schedule.ValidTo,
		)

		if err != nil ||
			validTo.Before(validFrom) {
			return ErrInstitutionalScheduleInvalid
		}

		// Una recurrencia cuyo rango no contiene siquiera el día de semana
		// solicitado no generaría ninguna ocurrencia real.
		if !weeklyRangeContainsDay(
			validFrom,
			validTo,
			*schedule.DayOfWeek,
		) {
			return ErrInstitutionalScheduleInvalid
		}

		return nil

	default:
		return ErrInstitutionalScheduleInvalid
	}
}

// ============================================================================
// SOLAPAMIENTOS DENTRO DE LA MISMA ACTIVIDAD
// ============================================================================

type institutionalOccurrence struct {
	start time.Time
	end   time.Time
}

// validateInstitutionalScheduleSelfOverlap evita programaciones como:
//
//	lunes 10:00-11:30
//	lunes 11:00-12:00
//
// dentro de la MISMA actividad.
//
// Eso no constituye un conflicto entre ocupaciones distintas; representa una
// definición inconsistente de la propia actividad y debe rechazarse antes de
// persistirla.
func validateInstitutionalScheduleSelfOverlap(
	schedules []models.CreateInstitutionalScheduleRequest,
) error {
	occurrences := []institutionalOccurrence{}

	for _, schedule := range schedules {
		expanded, err :=
			expandInstitutionalSchedule(schedule)

		if err != nil {
			return ErrInstitutionalScheduleInvalid
		}

		occurrences = append(
			occurrences,
			expanded...,
		)
	}

	for i := 0; i < len(occurrences); i++ {
		for j := i + 1; j < len(occurrences); j++ {

			// Intervalos [start,end):
			//
			// 10:00-11:00 y 11:00-12:00 NO se solapan.
			if occurrences[i].start.Before(
				occurrences[j].end,
			) &&
				occurrences[j].start.Before(
					occurrences[i].end,
				) {
				return ErrInstitutionalScheduleInvalid
			}
		}
	}

	return nil
}

// ============================================================================
// EXPANSIÓN DE OCURRENCIAS
// ============================================================================

// expandInstitutionalSchedule convierte una regla estructurada en ocurrencias
// concretas.
//
// Esta misma idea será reutilizada después por el detector de conflictos.
// Los timestamps representan instantes reales en America/Santiago, incluyendo
// automáticamente los cambios de offset por horario de verano.
func expandInstitutionalSchedule(
	schedule models.CreateInstitutionalScheduleRequest,
) ([]institutionalOccurrence, error) {
	switch schedule.ScheduleType {

	case models.InstitutionalScheduleTypeSingle:

		date, err := parseInstitutionalDate(
			*schedule.SpecificDate,
		)

		if err != nil {
			return nil, err
		}

		occurrence, err :=
			buildInstitutionalOccurrence(
				date,
				schedule.StartTime,
				schedule.EndTime,
			)

		if err != nil {
			return nil, err
		}

		return []institutionalOccurrence{
			occurrence,
		}, nil

	case models.InstitutionalScheduleTypeWeekly:

		validFrom, err := parseInstitutionalDate(
			*schedule.ValidFrom,
		)

		if err != nil {
			return nil, err
		}

		validTo, err := parseInstitutionalDate(
			*schedule.ValidTo,
		)

		if err != nil {
			return nil, err
		}

		occurrences := []institutionalOccurrence{}

		for date := validFrom; !date.After(validTo); date =
			date.AddDate(0, 0, 1) {

			if isoWeekday(date) !=
				*schedule.DayOfWeek {
				continue
			}

			occurrence, err :=
				buildInstitutionalOccurrence(
					date,
					schedule.StartTime,
					schedule.EndTime,
				)

			if err != nil {
				return nil, err
			}

			occurrences = append(
				occurrences,
				occurrence,
			)
		}

		return occurrences, nil

	default:
		return nil, ErrInstitutionalScheduleInvalid
	}
}

func buildInstitutionalOccurrence(
	date time.Time,
	startValue string,
	endValue string,
) (institutionalOccurrence, error) {
	startClock, err := time.Parse(
		"15:04",
		startValue,
	)

	if err != nil {
		return institutionalOccurrence{}, err
	}

	endClock, err := time.Parse(
		"15:04",
		endValue,
	)

	if err != nil {
		return institutionalOccurrence{}, err
	}

	location := date.Location()

	start := time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		startClock.Hour(),
		startClock.Minute(),
		0,
		0,
		location,
	)

	end := time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		endClock.Hour(),
		endClock.Minute(),
		0,
		0,
		location,
	)

	if !end.After(start) {
		return institutionalOccurrence{},
			ErrInstitutionalScheduleInvalid
	}

	return institutionalOccurrence{
		start: start,
		end:   end,
	}, nil
}

// ============================================================================
// FECHAS / CALENDARIO
// ============================================================================

// Poli-REDI opera actualmente en America/Santiago.
//
// Más adelante podemos reemplazar este helper por businessclock.Location()
// directamente si decidimos hacer configurable también la recurrencia
// institucional. El módulo businessclock ya centraliza esa zona en el backend.
func parseInstitutionalDate(
	value string,
) (time.Time, error) {
	location, err :=
		time.LoadLocation("America/Santiago")

	if err != nil {
		return time.Time{}, err
	}

	return time.ParseInLocation(
		"2006-01-02",
		value,
		location,
	)
}

func isoWeekday(
	value time.Time,
) int {
	switch value.Weekday() {

	case time.Monday:
		return 1

	case time.Tuesday:
		return 2

	case time.Wednesday:
		return 3

	case time.Thursday:
		return 4

	case time.Friday:
		return 5

	case time.Saturday:
		return 6

	case time.Sunday:
		return 7

	default:
		return 0
	}
}

func weeklyRangeContainsDay(
	from time.Time,
	to time.Time,
	dayOfWeek int,
) bool {
	for date := from; !date.After(to); date =
		date.AddDate(0, 0, 1) {

		if isoWeekday(date) == dayOfWeek {
			return true
		}
	}

	return false
}

// ============================================================================
// ERRORES DE REPOSITORY
// ============================================================================

func mapInstitutionalActivityRepositoryError(
	err error,
) error {
	if errors.Is(
		err,
		repositories.ErrInstitutionalScheduleBlocked,
	) {
		return ErrInstitutionalActivityBlocked
	}

	var pgErr *pgconn.PgError

	if !errors.As(err, &pgErr) {
		return err
	}

	switch pgErr.Code {

	case "23503":
		return ErrInstitutionalActivityInvalid

	case "23514":
		return ErrInstitutionalActivityInvalid

	default:
		return err
	}
}
