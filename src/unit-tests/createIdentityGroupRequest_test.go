package unittest

import (
	"strings"
	"testing"

	cfe "cf-user/core/enums"
	cfm "cf-user/core/models"
	cfcu "cf-user/create-user"

	"github.com/go-playground/validator"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func Test_CreateUser_Should_Not_Have_Errors_For_Valid_Request(t *testing.T) {
	email := "user" + ulid.Make().String() + "@canary-classifind.com"
	username := strings.Split(email, "@")[0]
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
	accountType := cfe.PersonalAccount.String()

	request := &cfcu.CreateUserRequest{
		Username:       &username,
		FirstName:      &firstName,
		LastName:       &lastName,
		PhoneNumber:    &phoneNumber,
		EmailAddress:   email,
		PrimaryAddress: address,
		BillingAddress: address,
		ProfileImageId: &profileImageId,
		Biography:      &biography,
		AccountType:    &accountType,
	}

	err := Validate.Struct(request)

	require.Nil(t, err)
}

func Test_CreateUser_Should_Not_Have_Errors_For_Valid_Minimal_Request(t *testing.T) {
	email := "user" + ulid.Make().String() + "@canary-classifind.com"

	request := &cfcu.CreateUserRequest{
		EmailAddress: email,
	}
	err := Validate.Struct(request)

	require.Nil(t, err)
}

func Test_CreateUser_Should_Have_Errors_For_Invalid_Request_Missing_Email(t *testing.T) {
	request := &cfcu.CreateUserRequest{}

	err := Validate.Struct(request)
	require.NotNil(t, err)

	structErrors := err.(validator.ValidationErrors)
	require.Equal(t, 1, len(structErrors))

	AssertValidationError(t, err, "EmailAddress", "required")
}

func Test_CreateUser_Should_Have_Errors_For_Invalid_Request_Using_Invalid_Email(t *testing.T) {
	email := ulid.Make().String() + "classifindcom"

	request := &cfcu.CreateUserRequest{
		EmailAddress: email,
	}
	err := Validate.Struct(request)

	structErrors := err.(validator.ValidationErrors)
	require.Equal(t, 1, len(structErrors))

	AssertValidationError(t, err, "EmailAddress", "email")
}

func Test_CreateUser_Should_Not_Have_Errors_For_Valid_Request_AccountType(t *testing.T) {
	email := "user" + ulid.Make().String() + "@canary-classifind.com"

	accountTypes := []cfe.AccountType{cfe.PersonalAccount, cfe.BusinessAccount}

	for _, accountType := range accountTypes {
		accountTypeString := accountType.String()

		request := &cfcu.CreateUserRequest{
			EmailAddress: email,
			AccountType:  &accountTypeString,
		}

		err := Validate.Struct(request)

		require.Nil(t, err)
	}
}

func Test_CreateUser_Should_Have_Errors_For_Invalid_Request_AccountType(t *testing.T) {
	email := "user" + ulid.Make().String() + "@canary-classifind.com"

	accountTypes := []string{"NoAccountType", "Invalid", "Fake"}

	for _, accountType := range accountTypes {
		request := &cfcu.CreateUserRequest{
			EmailAddress: email,
			AccountType:  &accountType,
		}

		err := Validate.Struct(request)
		require.NotNil(t, err)

		structErrors := err.(validator.ValidationErrors)
		require.Equal(t, 1, len(structErrors))

		AssertValidationError(t, err, "AccountType", "is_account_type")
	}
}

func Test_CreateUser_Should_Have_Errors_For_Base_Fields_With_Invalid_Max_Size(t *testing.T) {
	email := "user" + ulid.Make().String() + "@canary-classifind.com"

	username := strings.Repeat("a", 301)
	firstName := strings.Repeat("a", 301)
	lastName := strings.Repeat("a", 301)
	phoneNumber := strings.Repeat("a", 12)
	profileImageId := strings.Repeat("a", 27)
	biography := strings.Repeat("a", 4001)

	request := &cfcu.CreateUserRequest{
		Username:       &username,
		FirstName:      &firstName,
		LastName:       &lastName,
		PhoneNumber:    &phoneNumber,
		EmailAddress:   email,
		ProfileImageId: &profileImageId,
		Biography:      &biography,
	}

	err := Validate.Struct(request)
	require.NotNil(t, err)

	structErrors := err.(validator.ValidationErrors)
	require.Equal(t, 6, len(structErrors))

	AssertValidationError(t, err, "Username", "max")
	AssertValidationError(t, err, "FirstName", "max")
	AssertValidationError(t, err, "LastName", "max")
	AssertValidationError(t, err, "PhoneNumber", "max")
	AssertValidationError(t, err, "ProfileImageId", "max")
	AssertValidationError(t, err, "Biography", "max")
}

func Test_CreateUser_Should_Not_Have_Errors_For_Empty_Address(t *testing.T) {
	email := "user" + ulid.Make().String() + "@canary-classifind.com"

	request := &cfcu.CreateUserRequest{
		EmailAddress:   email,
		PrimaryAddress: &cfm.Address{},
	}

	err := Validate.Struct(request)
	require.Nil(t, err)
}

func Test_CreateUser_Should_Have_Errors_For_Address_Fields_With_Invalid_Max_Size(t *testing.T) {
	email := "user" + ulid.Make().String() + "@canary-classifind.com"

	addressOne := strings.Repeat("a", 301)
	addressTwo := strings.Repeat("a", 301)
	city := strings.Repeat("a", 201)
	state := strings.Repeat("a", 3)
	postalCode := strings.Repeat("a", 10)
	country := strings.Repeat("a", 4)

	request := &cfcu.CreateUserRequest{
		EmailAddress: email,
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

func Test_CreateUser_Should_Have_Errors_For_Address_Fields_With_Invalid_Min_Size(t *testing.T) {
	email := "user" + ulid.Make().String() + "@canary-classifind.com"

	addressOne := "123 Sunshine Street"
	addressTwo := "Unit 69"
	city := "Salt Lake City"

	state := strings.Repeat("a", 1)
	postalCode := strings.Repeat("a", 4)
	country := strings.Repeat("a", 2)

	request := &cfcu.CreateUserRequest{
		EmailAddress: email,
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
