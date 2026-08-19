package models

import "time"

// ============================================================================
// UNIDADES INSTITUCIONALES
// ============================================================================

const (
	InstitutionalUnitTypeAcademicProgram     = "ACADEMIC_PROGRAM"
	InstitutionalUnitTypePostgraduateProgram = "POSTGRADUATE_PROGRAM"
	InstitutionalUnitTypeSportsUnit          = "SPORTS_UNIT"
	InstitutionalUnitTypeAdministrativeUnit  = "ADMINISTRATIVE_UNIT"
	InstitutionalUnitTypeOther               = "OTHER"
)

const (
	InstitutionalMembershipRoleMember  = "MEMBER"
	InstitutionalMembershipRoleManager = "MANAGER"
)

type InstitutionalUnit struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	UnitType  string    `json:"unitType"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type InstitutionalUnitMembership struct {
	ID        int       `json:"id"`
	UnitID    int       `json:"unitId"`
	UserID    int       `json:"userId"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Datos enriquecidos para administración.
	UnitName     string `json:"unitName,omitempty"`
	UserFullName string `json:"userFullName,omitempty"`
	UserEmail    string `json:"userEmail,omitempty"`
}

// ============================================================================
// ACTIVIDADES INSTITUCIONALES
// ============================================================================

const (
	InstitutionalActivityTypeAcademicClass = "ACADEMIC_CLASS"
	InstitutionalActivityTypeWorkshop      = "WORKSHOP"
	InstitutionalActivityTypeTraining      = "TRAINING"
	InstitutionalActivityTypeEvent         = "EVENT"
	InstitutionalActivityTypeChampionship  = "CHAMPIONSHIP"
	InstitutionalActivityTypeOther         = "OTHER"
)

const (
	InstitutionalActivityStatusDraft     = "DRAFT"
	InstitutionalActivityStatusScheduled = "SCHEDULED"
	InstitutionalActivityStatusCancelled = "CANCELLED"
	InstitutionalActivityStatusCompleted = "COMPLETED"
)

type InstitutionalActivity struct {
	ID         int `json:"id"`
	UnitID     int `json:"unitId"`
	ResourceID int `json:"resourceId"`

	ActivityType string `json:"activityType"`
	Title        string `json:"title"`
	Description  string `json:"description"`

	Status string `json:"status"`

	RequiresEnrollment bool `json:"requiresEnrollment"`
	Capacity           *int `json:"capacity,omitempty"`

	CreatedByUserID int `json:"createdByUserId"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Datos enriquecidos utilizados por frontend y administración.
	UnitName     string `json:"unitName,omitempty"`
	ResourceName string `json:"resourceName,omitempty"`
	CreatedBy    string `json:"createdBy,omitempty"`

	EnrollmentCount int `json:"enrollmentCount"`

	Schedules []InstitutionalActivitySchedule `json:"schedules,omitempty"`
}

// ============================================================================
// PROGRAMACIÓN
// ============================================================================

const (
	InstitutionalScheduleTypeSingle = "SINGLE"
	InstitutionalScheduleTypeWeekly = "WEEKLY"
)

type InstitutionalActivitySchedule struct {
	ID         int `json:"id"`
	ActivityID int `json:"activityId"`

	ScheduleType string `json:"scheduleType"`

	// SINGLE usa SpecificDate.
	SpecificDate *string `json:"specificDate,omitempty"`

	// WEEKLY usa DayOfWeek + ValidFrom + ValidTo.
	//
	// ISO 8601:
	//   1 = lunes
	//   ...
	//   7 = domingo
	DayOfWeek *int `json:"dayOfWeek,omitempty"`

	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`

	ValidFrom *string `json:"validFrom,omitempty"`
	ValidTo   *string `json:"validTo,omitempty"`

	IsActive bool `json:"isActive"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ============================================================================
// INSCRIPCIONES
// ============================================================================

const (
	InstitutionalEnrollmentStatusConfirmed = "CONFIRMED"
	InstitutionalEnrollmentStatusCancelled = "CANCELLED"
)

type InstitutionalActivityEnrollment struct {
	ID         int `json:"id"`
	ActivityID int `json:"activityId"`
	UserID     int `json:"userId"`

	Status string `json:"status"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Datos enriquecidos para talleres.
	UserFullName string `json:"userFullName,omitempty"`
	UserEmail    string `json:"userEmail,omitempty"`
	UserRUT      string `json:"userRut,omitempty"`
}

// ============================================================================
// CONFLICTOS DE PROGRAMACIÓN
// ============================================================================

const (
	SchedulingConflictStatusPending  = "PENDING"
	SchedulingConflictStatusResolved = "RESOLVED"
)

const (
	SchedulingResolutionSourceManual = "MANUAL"
	SchedulingResolutionSourcePolicy = "POLICY"
	SchedulingResolutionSourceMixed  = "MIXED"
)

const (
	SchedulingItemResolutionPending    = "PENDING"
	SchedulingItemResolutionKeep       = "KEEP"
	SchedulingItemResolutionAllow      = "ALLOW"
	SchedulingItemResolutionReschedule = "RESCHEDULE"
	SchedulingItemResolutionCancel     = "CANCEL"
)

const (
	SchedulingConflictItemTypeInstitutionalActivity = "INSTITUTIONAL_ACTIVITY"
	SchedulingConflictItemTypeReservation           = "RESERVATION"
)

type SchedulingConflict struct {
	ID         int    `json:"id"`
	ResourceID int    `json:"resourceId"`
	Status     string `json:"status"`

	ResolutionSource  *string `json:"resolutionSource,omitempty"`
	ResolutionSummary string  `json:"resolutionSummary,omitempty"`

	ResolvedByUserID *int       `json:"resolvedByUserId,omitempty"`
	ResolvedAt       *time.Time `json:"resolvedAt,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Datos enriquecidos para la pantalla administrativa.
	ResourceName string                   `json:"resourceName,omitempty"`
	Items        []SchedulingConflictItem `json:"items,omitempty"`
}

type SchedulingConflictItem struct {
	ID         int `json:"id"`
	ConflictID int `json:"conflictId"`

	InstitutionalActivityID *int `json:"institutionalActivityId,omitempty"`
	ScheduleID              *int `json:"scheduleId,omitempty"`
	ReservationID           *int `json:"reservationId,omitempty"`

	// Snapshot de la ocurrencia exacta que participó en el conflicto.
	// Estos valores no deben modificarse si posteriormente cambia el horario.
	OccurrenceStart time.Time `json:"occurrenceStart"`
	OccurrenceEnd   time.Time `json:"occurrenceEnd"`

	Resolution       string  `json:"resolution"`
	ResolutionSource *string `json:"resolutionSource,omitempty"`

	ResolvedByUserID *int       `json:"resolvedByUserId,omitempty"`
	ResolvedAt       *time.Time `json:"resolvedAt,omitempty"`
	ResolutionNote   string     `json:"resolutionNote,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Tipo derivado para que el frontend no tenga que inferirlo comprobando FKs.
	EntityType string `json:"entityType,omitempty"`

	// Datos descriptivos enriquecidos.
	Title    string `json:"title,omitempty"`
	UnitName string `json:"unitName,omitempty"`
}

// ============================================================================
// REQUESTS
// ============================================================================

type CreateInstitutionalUnitRequest struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	UnitType string `json:"unitType"`
}

type CreateInstitutionalActivityRequest struct {
	UnitID       int    `json:"unitId"`
	ResourceID   int    `json:"resourceId"`
	ActivityType string `json:"activityType"`

	Title       string `json:"title"`
	Description string `json:"description"`

	RequiresEnrollment bool `json:"requiresEnrollment"`
	Capacity           *int `json:"capacity,omitempty"`

	Schedules []CreateInstitutionalScheduleRequest `json:"schedules"`
}

type CreateInstitutionalScheduleRequest struct {
	ScheduleType string `json:"scheduleType"`

	SpecificDate *string `json:"specificDate,omitempty"`
	DayOfWeek    *int    `json:"dayOfWeek,omitempty"`

	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`

	ValidFrom *string `json:"validFrom,omitempty"`
	ValidTo   *string `json:"validTo,omitempty"`
}

type AddInstitutionalUnitMembershipRequest struct {
	UserID int    `json:"userId"`
	Role   string `json:"role"`
}

type ResolveSchedulingConflictItemRequest struct {
	Resolution string `json:"resolution"`

	// Obligatoria a nivel de servicio para decisiones administrativas.
	ResolutionNote string `json:"resolutionNote"`

	// RESCHEDULE actúa únicamente sobre la ocurrencia concreta involucrada
	// en el conflicto. No modifica toda una regla WEEKLY.
	NewDate      *string `json:"newDate,omitempty"`
	NewStartTime *string `json:"newStartTime,omitempty"`
	NewEndTime   *string `json:"newEndTime,omitempty"`
}
