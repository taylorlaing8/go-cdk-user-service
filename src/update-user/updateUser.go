package updateidentitygroup

import (
	"context"

	"github.com/aws/aws-lambda-go/events"

	cfc "cf-user/core"
	cfe "cf-user/core/enums"
	cfm "cf-user/core/models"
)

type UpdateUserRequest struct {
	FirstName      *string      `json:"firstName" validate:"omitempty,max=300"`
	LastName       *string      `json:"lastName" validate:"omitempty,max=300"`
	PhoneNumber    *string      `json:"phoneNumber" validate:"omitempty,max=11"`    // 10 digit number or 11 digit including country code
	PrimaryAddress *cfm.Address `json:"primaryAddress" validate:"omitempty"`        // should inherit validation?
	BillingAddress *cfm.Address `json:"billingAddress" validate:"omitempty"`        // should inherit validation?
	ProfileImageId *string      `json:"profileImageId" validate:"omitempty,max=26"` // 26 char ULID
	Biography      *string      `json:"biography" validate:"omitempty,max=4000"`
}

var LambdaConfig *cfc.LambdaConfig[UpdateUserRequest, bool]

func InitLambda(ddbStore *cfc.DynamoDbStore) {
	roleRequired := cfe.UpdateUser

	LambdaConfig = cfc.CreateLambaConfig[UpdateUserRequest, bool](roleRequired, ddbStore)
}

func Handler(ctx context.Context, apiRequest events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return LambdaConfig.FunctionHandler.HandleRequest(apiRequest, func(request UpdateUserRequest) (*bool, *cfe.ResponseError) {
		userId, ok := apiRequest.PathParameters["userId"]
		if !ok {
			e := cfe.ErrorValidation("missing path parameter userId")
			return nil, &e
		}

		LambdaConfig.FunctionHandler.Logger.Info("Updating User", "UserId", userId)

		user, _ := LambdaConfig.DynamoDbStore.GetUser(userId)
		if user == nil {
			e := cfe.ErrorNotFound()
			return nil, &e
		}

		updateInput := &cfm.User{
			FirstName:      request.FirstName,
			LastName:       request.LastName,
			PhoneNumber:    request.PhoneNumber,
			PrimaryAddress: request.PrimaryAddress,
			BillingAddress: request.BillingAddress,
			ProfileImageId: request.ProfileImageId,
			Biography:      request.Biography,
		}

		groupUpdated, err := LambdaConfig.DynamoDbStore.UpdateUser(userId, updateInput)

		if err != nil {
			parseError := cfe.ErrorGetOrDefault(err)
			return nil, &parseError
		}

		return groupUpdated, nil
	}), nil
}
