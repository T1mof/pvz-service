package models

import (
	"time"

	"github.com/google/uuid"
)

// Допустимые города для создания ПВЗ
var AllowedCities = map[string]bool{
	"Москва":          true,
	"Санкт-Петербург": true,
	"Казань":          true,
}

type PVZ struct {
	ID               uuid.UUID `json:"id"`
	RegistrationDate time.Time `json:"registrationDate"`
	City             string    `json:"city" validate:"required"`
}

// PVZCreateRequest представляет запрос на создание ПВЗ
type PVZCreateRequest struct {
	City string `json:"city" validate:"required"`
}

// PVZListOptions представляет параметры для фильтрации списка ПВЗ
type PVZListOptions struct {
	Page      int       `json:"page" form:"page"`
	Limit     int       `json:"limit" form:"limit"`
	StartDate time.Time `json:"startDate" form:"startDate"`
	EndDate   time.Time `json:"endDate" form:"endDate"`
}

// PVZWithReceptionsResponse представляет ПВЗ со связанными приемками и товарами
type PVZWithReceptionsResponse struct {
	PVZ        *PVZ                     `json:"pvz"`
	Receptions []*ReceptionWithProducts `json:"receptions"`
}
