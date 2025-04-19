package models

import (
	"time"

	"github.com/google/uuid"
)

type ReceptionStatus string

const (
	StatusInProgress ReceptionStatus = "in_progress"
	StatusClosed     ReceptionStatus = "close"
)

type Reception struct {
	ID       uuid.UUID       `json:"id"`
	DateTime time.Time       `json:"dateTime"`
	PVZID    uuid.UUID       `json:"pvzId"`
	Status   ReceptionStatus `json:"status"`
	Products []*Product      `json:"products,omitempty"`
}

// ReceptionCreateRequest представляет запрос на создание приемки
type ReceptionCreateRequest struct {
	PVZID uuid.UUID `json:"pvzId" validate:"required"`
}

// ReceptionWithProducts представляет приемку вместе со списком товаров
type ReceptionWithProducts struct {
	Reception *Reception `json:"reception"`
	Products  []*Product `json:"products"`
}
