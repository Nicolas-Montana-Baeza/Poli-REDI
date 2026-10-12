package handlers

import (
	"encoding/json"
	"errors"
	"poli-redi-api/internal/repositories"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type updateResourceImageRequest struct {
	ImageURL string `json:"imageUrl"`
}

type nullableResourceCapacity struct {
	Present bool
	Value   *int
}

func (field *nullableResourceCapacity) UnmarshalJSON(data []byte) error {
	field.Present = true

	if string(data) == "null" {
		field.Value = nil
		return nil
	}

	var value int

	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	field.Value = &value
	return nil
}

type updateResourceOperationalRequest struct {
	ReservationMode *string                  `json:"reservationMode"`
	Capacity        nullableResourceCapacity `json:"capacity"`
	IsActive        *bool                    `json:"isActive"`
}

func validResourceReservationMode(mode string) bool {
	switch mode {
	case "RESERVABLE",
		"OPEN_USE",
		"INFORMATIVE",
		"ADMIN_ONLY":
		return true
	default:
		return false
	}
}

func GetResources(c *fiber.Ctx) error {
	resources, err := repositories.GetAllResources()

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  "Error obteniendo recursos",
			"detail": err.Error(),
		})
	}

	return c.JSON(resources)
}

func UpdateResourceOperationalSettings(c *fiber.Ctx) error {
	resourceID, err := strconv.Atoi(c.Params("id"))

	if err != nil || resourceID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "recurso invalido",
		})
	}

	var request updateResourceOperationalRequest

	if err := decodeStrictJSON(c.Body(), &request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "datos invalidos o campos no permitidos",
		})
	}

	if request.ReservationMode == nil &&
		!request.Capacity.Present &&
		request.IsActive == nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "debes indicar al menos un campo para actualizar",
		})
	}

	if request.ReservationMode != nil {
		normalized := strings.ToUpper(
			strings.TrimSpace(*request.ReservationMode),
		)

		if !validResourceReservationMode(normalized) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "modo de reserva invalido",
			})
		}

		request.ReservationMode = &normalized
	}

	if request.Capacity.Present &&
		request.Capacity.Value != nil &&
		*request.Capacity.Value <= 0 {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "la capacidad debe ser mayor que cero o null",
		})
	}

	resource, err :=
		repositories.UpdateResourceOperationalSettings(
			resourceID,
			request.ReservationMode,
			request.Capacity.Present,
			request.Capacity.Value,
			request.IsActive,
		)

	if err != nil {
		switch {
		case errors.Is(
			err,
			repositories.ErrResourceNotFound,
		):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "recurso no encontrado",
			})

		case errors.Is(
			err,
			repositories.ErrGroupResourceMustRemainReservable,
		),
			errors.Is(
				err,
				repositories.ErrGroupResourceCapacityInvalid,
			):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})

		default:
			return c.Status(
				fiber.StatusInternalServerError,
			).JSON(fiber.Map{
				"error":  "no se pudo actualizar el recurso",
				"detail": err.Error(),
			})
		}
	}

	return c.JSON(resource)
}

func UpdateResourceImage(c *fiber.Ctx) error {
	resourceID, err := strconv.Atoi(c.Params("id"))

	if err != nil || resourceID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "recurso invalido",
		})
	}

	var request updateResourceImageRequest

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "datos invalidos",
		})
	}

	imageURL := strings.TrimSpace(request.ImageURL)

	if len(imageURL) > 500 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "la URL de imagen es demasiado larga",
		})
	}

	if imageURL != "" &&
		!strings.HasPrefix(imageURL, "http://") &&
		!strings.HasPrefix(imageURL, "https://") &&
		!strings.HasPrefix(imageURL, "/") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "usa una URL http, https o una ruta local que comience con /",
		})
	}

	resource, err := repositories.UpdateResourceImageURL(
		resourceID,
		imageURL,
	)

	if err != nil {
		if errors.Is(err, repositories.ErrResourceNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "recurso no encontrado",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "No se pudo actualizar la imagen",
			"detail": err.Error(),
		})
	}

	return c.JSON(resource)
}
