package deleteidentitygroup

import (
	"context"

	"github.com/aws/aws-lambda-go/events"

	cfc "cf-user/core"
	cfe "cf-user/core/enums"
)

type DeleteUserRequest interface{}

var LambdaConfig *cfc.LambdaConfig[DeleteUserRequest, bool]

func InitLambda(ddbStore *cfc.DynamoDbStore) {
	roleRequired := cfe.DeleteUser

	LambdaConfig = cfc.CreateLambaConfig[DeleteUserRequest, bool](roleRequired, ddbStore)
}

func Handler(ctx context.Context, apiRequest events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return LambdaConfig.FunctionHandler.HandleRequest(apiRequest, func(DeleteUserRequest) (*bool, *cfe.ResponseError) {
		userId, ok := apiRequest.PathParameters["userId"]
		if !ok {
			e := cfe.ErrorValidation("missing path parameter userId")
			return nil, &e
		}

		groupDeleted, err := LambdaConfig.DynamoDbStore.DeleteUser(userId)
		if err != nil {
			parseError := cfe.ErrorGetOrDefault(err)
			return nil, &parseError
		}

		return groupDeleted, nil
	}), nil
}
