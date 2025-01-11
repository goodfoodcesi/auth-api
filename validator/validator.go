package validator

import (
	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

type Validator struct {
	validator *validator.Validate
}

func NewValidator() *Validator {
	return &Validator{
		validator: validator.New(),
	}
}

func (v *Validator) Validate(i interface{}) []ValidationError {
	var errors []ValidationError

	err := v.validator.Struct(i)
	if err == nil {
		return nil
	}

	validationErrors := err.(validator.ValidationErrors)
	for _, err := range validationErrors {
		errors = append(errors, ValidationError{
			Field: err.Field(),
			Tag:   err.Tag(),
			Value: err.Param(),
		})
	}

	return errors
}
