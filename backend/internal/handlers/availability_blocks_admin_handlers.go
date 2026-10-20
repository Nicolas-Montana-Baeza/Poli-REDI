package handlers

import (
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"

	"github.com/gofiber/fiber/v2"
)

type createAdminBlockRequest struct {
	ResourceID int    `json:"resourceId"`
	BlockType  string `json:"blockType"`
	Reason     string `json:"reason"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type deactivateAdminBlockRequest struct {
	IsActive *bool `json:"isActive"`
}

func validAdminBlockType(value string) bool {
	switch value {
	case "MAINTENANCE", "ADMINISTRATIVE", "EVENT", "CLOSED", "OTHER":
		return true
	default:
		return false
	}
}

// Los instantes deben incluir zona/offset para evitar ambiguedad en cambios DST.
func parseAdminBlockTimes(startValue, endValue string) (time.Time, time.Time, error) {
	start, startErr := time.Parse(time.RFC3339Nano, startValue)
	end, endErr := time.Parse(time.RFC3339Nano, endValue)
	if startErr != nil || endErr != nil || !end.After(start) || end.Sub(start) > 31*24*time.Hour {
		return time.Time{}, time.Time{}, errors.New("rango invalido; usa RFC3339 con zona y maximo 31 dias")
	}
	return start, end, nil
}

func GetAdminAvailabilityBlocks(c *fiber.Ctx) error {
	from, to, err := availabilityRange(c.Query("from"), c.Query("to"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	blocks, err := repositories.ListAdminAvailabilityBlocks(from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no se pudieron consultar los bloqueos"})
	}
	return c.JSON(blocks)
}

func CreateAdminAvailabilityBlock(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "usuario no autenticado"})
	}
	var request createAdminBlockRequest
	if err := decodeStrictJSON(c.Body(), &request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "datos invalidos o campos no permitidos"})
	}
	request.BlockType = strings.TrimSpace(request.BlockType)
	request.Reason = strings.TrimSpace(request.Reason)
	if request.ResourceID <= 0 || !validAdminBlockType(request.BlockType) || utf8.RuneCountInString(request.Reason) > 500 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "recurso, tipo o motivo invalido"})
	}
	start, end, err := parseAdminBlockTimes(request.StartTime, request.EndTime)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	block, err := repositories.CreateAdminAvailabilityBlock(models.AdminAvailabilityBlock{
		ResourceID: request.ResourceID, CreatedByUserID: user.ID,
		BlockType: request.BlockType, Reason: request.Reason, StartTime: start, EndTime: end,
	})
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrBlockOverlap), errors.Is(err, repositories.ErrBlockResourceUnavailable):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no se pudo crear el bloqueo"})
		}
	}
	return c.Status(fiber.StatusCreated).JSON(block)
}

func DeactivateAdminAvailabilityBlock(c *fiber.Ctx) error {
	blockID, err := strconv.Atoi(c.Params("id"))
	if err != nil || blockID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bloqueo invalido"})
	}
	var request deactivateAdminBlockRequest
	if err := decodeStrictJSON(c.Body(), &request); err != nil || request.IsActive == nil || *request.IsActive {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "solo se admite isActive: false"})
	}
	block, err := repositories.DeactivateAdminAvailabilityBlock(blockID)
	if errors.Is(err, repositories.ErrBlockNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no se pudo desactivar el bloqueo"})
	}
	return c.JSON(block)
}
