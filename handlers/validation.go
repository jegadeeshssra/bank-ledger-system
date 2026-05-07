package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// ValidateRequest validates a struct and writes appropriate error response
func ValidateRequest(w http.ResponseWriter, req interface{}) error {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errorMsg := fmt.Sprintf("Validation failed: %s", formatValidationErrors(validationErrors))
		http.Error(w, errorMsg, http.StatusBadRequest)
		return err
	}
	return nil
}

// formatValidationErrors converts validator errors to user-friendly messages
func formatValidationErrors(errs validator.ValidationErrors) string {
	var errMsg string
	for i, err := range errs {
		if i > 0 {
			errMsg += "; "
		}
		field := err.Field()
		tag := err.Tag()
		switch tag {
		case "required":
			errMsg += fmt.Sprintf("%s is required", field)
		case "min":
			errMsg += fmt.Sprintf("%s must have minimum length of %s", field, err.Param())
		case "max":
			errMsg += fmt.Sprintf("%s must have maximum length of %s", field, err.Param())
		case "len":
			errMsg += fmt.Sprintf("%s must have length of %s", field, err.Param())
		case "numeric":
			errMsg += fmt.Sprintf("%s must be a valid numeric value", field)
		case "gt":
			errMsg += fmt.Sprintf("%s must be greater than 0", field)
		case "nefield":
			errMsg += fmt.Sprintf("%s cannot be equal to %s", field, err.Param())
		default:
			errMsg += fmt.Sprintf("%s failed validation %s", field, tag)
		}
	}
	return errMsg
}
