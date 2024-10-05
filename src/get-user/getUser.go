package getidentitygroup

import (
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/oklog/ulid/v2"

	cfc "cf-user/core"
	cfe "cf-user/core/enums"
	cfm "cf-user/core/models"
)

type GetUserRequest interface{}

var LambdaConfig *cfc.LambdaConfig[GetUserRequest, cfm.User]

func InitLambda(ddbStore *cfc.DynamoDbStore) {
	roleRequired := cfe.ReadUser

	LambdaConfig = cfc.CreateLambaConfig[GetUserRequest, cfm.User](roleRequired, ddbStore)
}

func Handler(ctx context.Context, apiRequest events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return LambdaConfig.FunctionHandler.HandleRequest(apiRequest, func(GetUserRequest) (*cfm.User, *cfe.ResponseError) {
		userId, ok := apiRequest.PathParameters["userId"]
		if !ok {
			e := cfe.ErrorValidation("missing path parameter userId")
			return nil, &e
		}

		var user *cfm.User

		_, err := ulid.ParseStrict(userId)
		if err == nil {
			user, err = LambdaConfig.DynamoDbStore.GetUser(userId)
		} else if strings.Contains(userId, "@") {
			user, err = LambdaConfig.DynamoDbStore.GetUserByEmail(userId)
		} else {
			user, err = LambdaConfig.DynamoDbStore.GetUserByUsername(userId)
		}

		if err != nil {
			parseError := cfe.ErrorGetOrDefault(err)
			return nil, &parseError
		}

		return user, nil
	}), nil
}
