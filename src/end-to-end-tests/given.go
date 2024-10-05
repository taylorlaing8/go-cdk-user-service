package endtoendtest

import (
	cfe "cf-user/core/enums"
	cfm "cf-user/core/models"
	"strings"

	cfcu "cf-user/create-user"
	cfuu "cf-user/update-user"

	"github.com/oklog/ulid/v2"
)

func GetCanaryAccessToken() string {
	return Fixture.TokenService.GetAccessToken()
}

func GivenCreateUserRequest(emailAddress *string) *cfcu.CreateUserRequest {
	var email string
	if emailAddress != nil {
		email = *emailAddress
	} else {
		email = "user" + ulid.Make().String() + "@canary-classifind.com"
	}

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

	return &cfcu.CreateUserRequest{
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
}

func GivenCreateUserArgs(emailAddress *string) *cfm.User {
	request := GivenCreateUserRequest(emailAddress)

	accountType, _ := cfe.GetAccountType(request.AccountType)

	return &cfm.User{
		Username:       *request.Username,
		FirstName:      request.FirstName,
		LastName:       request.LastName,
		PhoneNumber:    request.PhoneNumber,
		EmailAddress:   request.EmailAddress,
		PrimaryAddress: request.PrimaryAddress,
		BillingAddress: request.BillingAddress,
		ProfileImageId: request.ProfileImageId,
		Biography:      request.Biography,
		AccountType:    *accountType,
	}
}

func GivenCreateMinimalUserRequest(emailAddress *string) *cfcu.CreateUserRequest {
	var email string
	if emailAddress != nil {
		email = *emailAddress
	} else {
		email = "user" + ulid.Make().String() + "@canary-classifind.com"
	}

	return &cfcu.CreateUserRequest{
		EmailAddress: email,
	}
}

func GivenUpdateUserRequest(priority *string) *cfuu.UpdateUserRequest {
	firstName := "UPDATE:John"
	lastName := "UPDATE:Doe"
	phoneNumber := "8014561233"

	addressOne := "UPDATE:123 Sunshine Street"
	addressTwo := "UPDATE:Unit 69"
	city := "Salt Lake City"
	state := "UT"
	postalCode := "84101"
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
	biography := "UPDATE:Short bio about the incoming user account."

	return &cfuu.UpdateUserRequest{
		FirstName:      &firstName,
		LastName:       &lastName,
		PhoneNumber:    &phoneNumber,
		PrimaryAddress: address,
		BillingAddress: address,
		ProfileImageId: &profileImageId,
		Biography:      &biography,
	}
}
