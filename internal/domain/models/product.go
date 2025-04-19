package models

import (
	"time"

	"github.com/google/uuid"
)

type ProductType string

const (
	TypeElectronics ProductType = "электроника"
	TypeClothes     ProductType = "одежда"
	TypeFootwear    ProductType = "обувь"
)

type Product struct {
	ID          uuid.UUID   `json:"id"`
	DateTime    time.Time   `json:"dateTime"`
	Type        ProductType `json:"type"`
	ReceptionID uuid.UUID   `json:"receptionId"`
	SequenceNum int         `json:"sequenceNum"`
}

// ProductCreateRequest представляет запрос на создание товара
type ProductCreateRequest struct {
	Type  ProductType `json:"type" validate:"required,oneof=электроника одежда обувь"`
	PVZID uuid.UUID   `json:"pvzId" validate:"required"`
}
