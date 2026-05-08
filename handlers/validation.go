package handlers

import (
	"fmt"
	"net/http"
	"reflect"

	"ledger-system/repository"

	"github.com/go-playground/validator/v10"
)

// interface{} =
// {
//     type: repository.Account,
//     value: {...}
// }

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

// ValidateResponse validates outbound response payloads and writes an error if invalid.
func ValidateResponse(w http.ResponseWriter, resp interface{}) error {
	validate := validator.New()
	fmt.Println(reflect.TypeOf(resp))
	switch data := resp.(type) {
	case []repository.Account:
		for _, acc := range data {
			if err := validate.Struct(acc); err != nil {
				http.Error(w, "Invalid response payload", http.StatusInternalServerError)
				return err
			}
		}
	case repository.Account:
		if err := validate.Struct(data); err != nil {
			http.Error(w, "Invalid response payload", http.StatusInternalServerError)
			return err
		}
	case *repository.Account:
		if data != nil {
			if err := validate.Struct(data); err != nil {
				http.Error(w, "Invalid response payload", http.StatusInternalServerError)
				return err
			}
		}
	case []repository.Entry:
		for _, entry := range data {
			if err := validate.Struct(entry); err != nil {
				http.Error(w, "Invalid response payload", http.StatusInternalServerError)
				return err
			}
		}
	case repository.Entry:
		if err := validate.Struct(data); err != nil {
			http.Error(w, "Invalid response payload", http.StatusInternalServerError)
			return err
		}
	case *repository.Entry:
		if data != nil {
			if err := validate.Struct(data); err != nil {
				http.Error(w, "Invalid response payload", http.StatusInternalServerError)
				return err
			}
		}
	default:
		if err := validate.Struct(data); err != nil {
			http.Error(w, "Invalid response payload", http.StatusInternalServerError)
			return err
		}

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
