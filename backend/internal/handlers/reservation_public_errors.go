package handlers

import (
	"database/sql"
	"errors"
	"regexp"

	"poli-redi-api/internal/repositories"
	"poli-redi-api/internal/services"

	"github.com/gofiber/fiber/v2"
)

// Solo las validaciones conocidas pueden entregar su mensaje al cliente.
// Los errores inesperados nunca se devuelven literalmente.
var reservationValidationMessages = map[string]bool{
	"selecciona una instalación":                                                   true,
	"selecciona una fecha y hora de inicio":                                        true,
	"no puedes crear reservas en el pasado":                                        true,
	"selecciona una duracion de 30 a 180 minutos en intervalos de 30":              true,
	"el usuario se encuentra bloqueado o no tiene RUT habilitante":                 true,
	"el recurso no esta disponible para reservas":                                  true,
	"el horario o la duracion no estan permitidos por la politica vigente":         true,
	"la fecha esta fuera de la ventana reservable":                                 true,
	"el recurso esta bloqueado en ese horario":                                     true,
	"el recurso tiene programación institucional en ese horario":                   true,
	"no existe una política de reservas vigente":                                   true,
	"el usuario ya tiene una reserva en ese horario":                               true,
	"el recurso ya esta reservado en ese horario":                                  true,
	"usuario, recurso o actividad no existe, o los datos no cumplen restricciones": true,
	"ya existe un registro con esos datos":                                         true,
	"la reserva compitio con otra operacion; intenta nuevamente":                   true,
}

var reservationValidationPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^la fecha debe estar entre [0-9]{4}-[0-9]{2}-[0-9]{2} y [0-9]{4}-[0-9]{2}-[0-9]{2}$`),
	regexp.MustCompile(`^la hora de inicio debe usar intervalos de [0-9]+ minutos$`),
	regexp.MustCompile(`^la jornada de reservas comienza a las [0-9]{2}:[0-9]{2}$`),
	regexp.MustCompile(`^la hora de inicio debe ser anterior a las [0-9]{2}:[0-9]{2}$`),
	regexp.MustCompile(`^la reserva debe finalizar a mas tardar a las [0-9]{2}:[0-9]{2}$`),
}

func createReservationPublicError(err error) (int, string) {
	if errors.Is(err, sql.ErrNoRows) {
		return fiber.StatusBadRequest, "la instalación seleccionada no existe"
	}

	if errors.Is(err, repositories.ErrResourceNotAllowedByPolicy) {
		return fiber.StatusBadRequest,
			"el recurso no esta permitido por la politica vigente"
	}

	var frequency repositories.RequestFrequencyError
	if errors.As(err, &frequency) {
		return fiber.StatusBadRequest, frequency.Error()
	}

	message := err.Error()

	if reservationValidationMessages[message] {
		return fiber.StatusBadRequest, message
	}

	for _, pattern := range reservationValidationPatterns {
		if pattern.MatchString(message) {
			return fiber.StatusBadRequest, message
		}
	}

	return fiber.StatusInternalServerError, "no se pudo crear la reserva"
}

func cancelReservationPublicError(err error) (int, string) {
	if errors.Is(err, services.ErrReservationNotFound) {
		return fiber.StatusNotFound, "reserva no encontrada"
	}

	if errors.Is(err, repositories.ErrReservationForbidden) ||
		err.Error() == "no tienes permisos para cancelar esta reserva" {
		return fiber.StatusForbidden, "no tienes permisos para cancelar esta reserva"
	}

	if errors.Is(err, repositories.ErrReservationFinalized) {
		return fiber.StatusBadRequest, repositories.ErrReservationFinalized.Error()
	}

	if errors.Is(err, repositories.ErrReservationNotCancellable) {
		return fiber.StatusBadRequest, repositories.ErrReservationNotCancellable.Error()
	}

	if err.Error() == "no se pudo identificar la reserva" {
		return fiber.StatusBadRequest, "no se pudo identificar la reserva"
	}

	return fiber.StatusInternalServerError, "no se pudo cancelar la reserva"
}
