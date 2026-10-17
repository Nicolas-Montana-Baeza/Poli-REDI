package handlers

import (
	"encoding/json"
	"errors"
	"poli-redi-api/internal/repositories"
	"strconv"
	"strings"
	"unicode/utf8"

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

type resourceRequiredText struct {
	Present bool
	Value   *string
}

func (field *resourceRequiredText) UnmarshalJSON(data []byte) error {
	field.Present = true
	return json.Unmarshal(data, &field.Value)
}

func normalizeRequiredResourceText(
	field resourceRequiredText,
	maxLength int,
) (*string, bool) {
	if !field.Present {
		return nil, true
	}
	if field.Value == nil {
		return nil, false
	}

	value := strings.TrimSpace(*field.Value)
	if value == "" || utf8.RuneCountInString(value) > maxLength {
		return nil, false
	}
	return &value, true
}

type updateResourceOperationalRequest struct {
	Name            resourceRequiredText     `json:"name"`
	Type            resourceRequiredText     `json:"type"`
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
	query := c.Queries()
	filters := repositories.ResourceFilters{}

	if raw, exists := query["venueId"]; exists {
		venueID, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil || venueID <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "venueId debe ser un entero positivo",
			})
		}
		filters.VenueID = &venueID
	}

	if raw, exists := query["type"]; exists {
		resourceType := strings.TrimSpace(raw)
		if resourceType == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "type no puede estar vacio",
			})
		}
		filters.Type = &resourceType
	}

	if raw, exists := query["isActive"]; exists {
		var isActive bool
		switch strings.TrimSpace(raw) {
		case "true":
			isActive = true
		case "false":
			isActive = false
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "isActive debe ser true o false",
			})
		}
		filters.IsActive = &isActive
	}

	resources, err := repositories.GetResourcesFiltered(filters)

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

	if !request.Name.Present &&
		!request.Type.Present &&
		request.ReservationMode == nil &&
		!request.Capacity.Present &&
		request.IsActive == nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "debes indicar al menos un campo para actualizar",
		})
	}

	name, validName := normalizeRequiredResourceText(
		request.Name, 120,
	)
	resourceType, validType := normalizeRequiredResourceText(
		request.Type, 80,
	)

	if !validName || !validType {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "nombre o tipo invalido",
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
		repositories.UpdateResourceConfiguration(
			resourceID,
			repositories.ResourceConfigurationPatch{
				Name:            name,
				Type:            resourceType,
				ReservationMode: request.ReservationMode,
				CapacityPresent: request.Capacity.Present,
				Capacity:        request.Capacity.Value,
				IsActive:        request.IsActive,
			},
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

		case errors.Is(err, repositories.ErrResourceDuplicate),
			errors.Is(err, repositories.ErrResourceConcurrentUpdate),
			errors.Is(err, repositories.ErrResourceVenueInactive),
			errors.Is(err, repositories.ErrGroupResourceMustRemainReservable),
			errors.Is(err, repositories.ErrGroupResourceCapacityInvalid):
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
