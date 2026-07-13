package handlers

import (
	"poli-redi-api/internal/repositories"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type updateResourceImageRequest struct {
	ImageURL string `json:"imageUrl"`
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "No se pudo actualizar la imagen",
			"detail": err.Error(),
		})
	}

	return c.JSON(resource)
}
