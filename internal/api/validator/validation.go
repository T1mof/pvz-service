package validator

import (
	"fmt"
	"strings"

	"pvz-service/internal/domain/models"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	_ = validate.RegisterValidation("itemtype", validateItemType)
	_ = validate.RegisterValidation("allowedcity", validateAllowedCity)
}

// ValidateStruct проверяет структуру на соответствие правилам валидации
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// FormatValidationErrors форматирует ошибки валидации в более понятный вид
func FormatValidationErrors(err error) string {
	if err == nil {
		return ""
	}

	var errMessages []string

	validationErrors := err.(validator.ValidationErrors)
	for _, e := range validationErrors {
		errMessages = append(errMessages, fmt.Sprintf(
			"Field '%s' failed validation: %s",
			e.Field(),
			e.Tag(),
		))
	}

	return strings.Join(errMessages, "; ")
}

// validateItemType проверяет, что тип товара допустимый
func validateItemType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return value == string(models.TypeElectronics) ||
		value == string(models.TypeClothes) ||
		value == string(models.TypeFootwear)
}

// validateAllowedCity проверяет, что город разрешен для создания ПВЗ
func validateAllowedCity(fl validator.FieldLevel) bool {
	city := fl.Field().String()
	return models.AllowedCities[city]
}
