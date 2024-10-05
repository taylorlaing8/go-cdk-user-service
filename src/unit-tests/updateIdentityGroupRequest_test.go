package unittest

import (
	"strings"
	"testing"

	cfm "cf-user/core/models"
	cfuu "cf-user/update-user"

	"github.com/go-playground/validator"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func Test_UpdateUser_Should_Not_Have_Errors_For_Valid_Request(t *testing.T) {
	firstName := "John"
	lastName := "Doe"
	phoneNumber := "8011239088"

	addressOne := "123 Sunshine Street"
	addressTwo := "Unit 69"
	city := "Salt Lake City"
	state := "UT"
	postalCode := "84103"
	country := "USA"
	address := &cfm.Address{
		AddressOne: &addressOne,
		AddressTwo: &addressTwo,
		City:       &city,
		State:      &state,
		PostalCode: &postalCode,
		Country:    &country,
	}

	profileImageId := ulid.Make().String()
	biography := "Short bio about the incoming user account."

	request := &cfuu.UpdateUserRequest{
		FirstName:      &firstName,
		LastName:       &lastName,
		PhoneNumber:    &phoneNumber,
		PrimaryAddress: address,
		BillingAddress: address,
		ProfileImageId: &profileImageId,
		Biography:      &biography,
	}

	err := Validate.Struct(request)

	require.Nil(t, err)
}

func Test_UpdateUser_Should_Not_Have_Errors_For_Valid_Minimal_Request(t *testing.T) {
	request := &cfuu.UpdateUserRequest{}
	err := Validate.Struct(request)

	require.Nil(t, err)
}

func Test_UpdateUser_Should_Have_Errors_For_Base_Fields_With_Invalid_Max_Size(t *testing.T) {
	firstName := strings.Repeat("a", 301)
	lastName := strings.Repeat("a", 301)
	phoneNumber := strings.Repeat("a", 12)
	profileImageId := strings.Repeat("a", 27)
	biography := strings.Repeat("a", 4001)

	request := &cfuu.UpdateUserRequest{
		FirstName:      &firstName,
		LastName:       &lastName,
		PhoneNumber:    &phoneNumber,
		ProfileImageId: &profileImageId,
		Biography:      &biography,
	}

	err := Validate.Struct(request)
	require.NotNil(t, err)

	structErrors := err.(validator.ValidationErrors)
	require.Equal(t, 5, len(structErrors))

	AssertValidationError(t, err, "FirstName", "max")
	AssertValidationError(t, err, "LastName", "max")
	AssertValidationError(t, err, "PhoneNumber", "max")
	AssertValidationError(t, err, "ProfileImageId", "max")
	AssertValidationError(t, err, "Biography", "max")
}

func Test_UpdateUser_Should_Not_Have_Errors_For_Empty_Address(t *testing.T) {
	request := &cfuu.UpdateUserRequest{
		PrimaryAddress: &cfm.Address{},
	}

	err := Validate.Struct(request)
	require.Nil(t, err)
}

func Test_UpdateUser_Should_Have_Errors_For_Address_Fields_With_Invalid_Max_Size(t *testing.T) {
	addressOne := strings.Repeat("a", 301)
	addressTwo := strings.Repeat("a", 301)
	city := strings.Repeat("a", 201)
	state := strings.Repeat("a", 3)
	postalCode := strings.Repeat("a", 10)
	country := strings.Repeat("a", 4)

	request := &cfuu.UpdateUserRequest{
		PrimaryAddress: &cfm.Address{
			AddressOne: &addressOne,
			AddressTwo: &addressTwo,
			City:       &city,
			State:      &state,
			PostalCode: &postalCode,
			Country:    &country,
		},
	}

	err := Validate.Struct(request)
	require.NotNil(t, err)

	structErrors := err.(validator.ValidationErrors)
	require.Equal(t, 6, len(structErrors))

	AssertValidationError(t, err, "AddressOne", "max")
	AssertValidationError(t, err, "AddressTwo", "max")
	AssertValidationError(t, err, "City", "max")
	AssertValidationError(t, err, "State", "max")
	AssertValidationError(t, err, "PostalCode", "max")
	AssertValidationError(t, err, "Country", "max")
}

func Test_UpdateUser_Should_Have_Errors_For_Address_Fields_With_Invalid_Min_Size(t *testing.T) {
	addressOne := "123 Sunshine Street"
	addressTwo := "Unit 69"
	city := "Salt Lake City"

	state := strings.Repeat("a", 1)
	postalCode := strings.Repeat("a", 4)
	country := strings.Repeat("a", 2)

	request := &cfuu.UpdateUserRequest{
		PrimaryAddress: &cfm.Address{
			AddressOne: &addressOne,
			AddressTwo: &addressTwo,
			City:       &city,
			State:      &state,
			PostalCode: &postalCode,
			Country:    &country,
		},
	}

	err := Validate.Struct(request)
	require.NotNil(t, err)

	structErrors := err.(validator.ValidationErrors)
	require.Equal(t, 3, len(structErrors))

	AssertValidationError(t, err, "State", "min")
	AssertValidationError(t, err, "PostalCode", "min")
	AssertValidationError(t, err, "Country", "min")
}
