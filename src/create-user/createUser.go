package createidentitygroup

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	cfc "cf-user/core"
	cfe "cf-user/core/enums"
	cfm "cf-user/core/models"
)

type CreateUserRequest struct {
	Username       *string      `json:"username" validate:"omitempty,max=300"`
	FirstName      *string      `json:"firstName" validate:"omitempty,max=300"`
	LastName       *string      `json:"lastName" validate:"omitempty,max=300"`
	PhoneNumber    *string      `json:"phoneNumber" validate:"omitempty,max=11"` // 10 digit number or 11 digit including country code
	EmailAddress   string       `json:"emailAddress" validate:"required,email"`
	PrimaryAddress *cfm.Address `json:"primaryAddress" validate:"omitempty"`
	BillingAddress *cfm.Address `json:"billingAddress" validate:"omitempty"`
	ProfileImageId *string      `json:"profileImageId" validate:"omitempty,max=26"` // 26 char ULID
	Biography      *string      `json:"biography" validate:"omitempty,max=4000"`
	AccountType    *string      `json:"accountType" validate:"omitempty,is_account_type"`
}

type CreateUserResponse struct {
	UserId string `json:"userId"`
}

var LambdaConfig *cfc.LambdaConfig[CreateUserRequest, CreateUserResponse]

func InitLambda(ddbStore *cfc.DynamoDbStore) {
	roleRequired := cfe.CreateUser

	LambdaConfig = cfc.CreateLambaConfig[CreateUserRequest, CreateUserResponse](roleRequired, ddbStore)

	LambdaConfig.FunctionHandler.Validate.RegisterValidation(cfe.GetAccountTypeValidator())
}

func Handler(ctx context.Context, apiRequest events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return LambdaConfig.FunctionHandler.HandleRequest(apiRequest, func(request CreateUserRequest) (*CreateUserResponse, *cfe.ResponseError) {
		user, _ := LambdaConfig.DynamoDbStore.GetUserByEmail(request.EmailAddress)
		if user != nil {
			e := cfe.ErrorValidation(fmt.Sprintf("User already exists with given email address: %v", request.EmailAddress))
			return nil, &e
		}

		LambdaConfig.FunctionHandler.Logger.Info("Creating User", "EmailAddress", request.EmailAddress)

		var username string
		if request.Username != nil {
			username = *request.Username
		} else {
			username = strings.Split(request.EmailAddress, "@")[0]
		}

		var accountType cfe.AccountType
		if request.AccountType != nil {
			aType, err := cfe.GetAccountType(request.AccountType)

			if err != nil {
				accountType = cfe.PersonalAccount
			} else {
				accountType = *aType
			}
		} else {
			accountType = cfe.PersonalAccount
		}

		createInput := &cfm.User{
			Username:       username,
			FirstName:      request.FirstName,
			LastName:       request.LastName,
			PhoneNumber:    request.PhoneNumber,
			EmailAddress:   request.EmailAddress,
			PrimaryAddress: request.PrimaryAddress,
			BillingAddress: request.BillingAddress,
			ProfileImageId: request.ProfileImageId,
			Biography:      request.Biography,
			AccountType:    accountType,
		}

		userId, err := LambdaConfig.DynamoDbStore.CreateUser(createInput)

		if err != nil {
			parseError := cfe.ErrorGetOrDefault(err)
			return nil, &parseError
		}

		return &CreateUserResponse{
			UserId: *userId,
		}, nil
	}), nil
}
