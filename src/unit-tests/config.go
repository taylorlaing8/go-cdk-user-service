package unittest

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator"
	"github.com/stretchr/testify/require"

	cfe "cf-user/core/enums"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New()

	// Register Customer Validation
	Validate.RegisterValidation(cfe.GetAccountTypeValidator())
}

func AssertValidationError(t *testing.T, validationError error, fieldName string, tagName string) {
	fieldLocated := false

	for _, err := range validationError.(validator.ValidationErrors) {
		if err.Field() == fieldName {
			fieldLocated = true

			require.Equal(t, fieldName, err.Field())
			require.Equal(t, tagName, err.Tag())
		}
	}

	require.Equal(t, true, fieldLocated, fmt.Sprintf("field did not fail validation: %v", fieldName))
}
