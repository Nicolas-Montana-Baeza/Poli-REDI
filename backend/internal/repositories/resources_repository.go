package repositories

import "poli-redi-api/internal/models"

func GetAllResources() []models.Resource {
	return []models.Resource{
		{
			ID:     1,
			Name:   "Cancha 1",
			Type:   "Exterior",
			Status: "available",
		},
		{
			ID:     2,
			Name:   "Cancha 2",
			Type:   "Exterior",
			Status: "busy",
		},
		{
			ID:     3,
			Name:   "Cancha 3",
			Type:   "Exterior",
			Status: "available",
		},
		{
			ID:     4,
			Name:   "Gimnasio",
			Type:   "Interior",
			Status: "maintenance",
		},
		{
			ID:     5,
			Name:   "Piscina",
			Type:   "Interior",
			Status: "available",
		},
	}
}
