package integrationtest

import (
	"context"
	"encoding/json"

	cfcu "cf-user/create-user"
	cfdu "cf-user/delete-user"
	cfgu "cf-user/get-user"
	cfuu "cf-user/update-user"

	"github.com/aws/aws-lambda-go/events"
)

// Identity Groups
func WhenWeGetUser(userId string, permissions *string, requesterId *string) (events.APIGatewayProxyResponse, error) {
	apiRequest := CreateGetRequest(permissions, requesterId)
	apiRequest.PathParameters["userId"] = userId

	return cfgu.Handler(context.TODO(), *apiRequest)
}

func WhenWeCreateUser(request *cfcu.CreateUserRequest, permissions *string, requesterId *string) (events.APIGatewayProxyResponse, error) {
	apiRequest := createPostRequest(request, permissions, requesterId)

	return cfcu.Handler(context.TODO(), *apiRequest)
}

func WhenWeUpdateUser(userId string, request *cfuu.UpdateUserRequest, permissions *string, requesterId *string) (events.APIGatewayProxyResponse, error) {
	apiRequest := createPutRequest(request, permissions, requesterId)
	apiRequest.PathParameters["userId"] = userId

	return cfuu.Handler(context.TODO(), *apiRequest)
}

func WhenWeDeleteUser(userId string, permissions *string, requesterId *string) (events.APIGatewayProxyResponse, error) {
	apiRequest := createDeleteRequest(permissions, requesterId)
	apiRequest.PathParameters["userId"] = userId

	return cfdu.Handler(context.TODO(), *apiRequest)
}

// Helper Functions
func createPostRequest[T interface{}](body T, permissions *string, requesterId *string) *events.APIGatewayProxyRequest {
	apiRequest := createRequest("POST", permissions, requesterId)

	val, err := json.Marshal(&body)
	if err != nil {
		panic("Failed to serialize request body")
	}

	apiRequest.Body = string(val)
	return apiRequest
}

func createPutRequest[T interface{}](body T, permissions *string, requesterId *string) *events.APIGatewayProxyRequest {
	apiRequest := createRequest("PUT", permissions, requesterId)

	val, err := json.Marshal(&body)
	if err != nil {
		panic("Failed to serialize request body")
	}

	apiRequest.Body = string(val)
	return apiRequest
}

func createDeleteRequest(permissions *string, requesterId *string) *events.APIGatewayProxyRequest {
	apiRequest := createRequest("DELETE", permissions, requesterId)
	return apiRequest
}

func CreateGetRequest(permissions *string, requesterId *string) *events.APIGatewayProxyRequest {
	apiRequest := createRequest("GET", permissions, requesterId)
	return apiRequest
}

func createRequest(httpMethod string, permissions *string, requesterId *string) *events.APIGatewayProxyRequest {
	apiRequest := &events.APIGatewayProxyRequest{
		HTTPMethod: httpMethod,
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: make(map[string]interface{}),
		},
	}

	if permissions != nil {
		apiRequest.RequestContext.Authorizer["permissions"] = *permissions
	}
	if requesterId != nil {
		apiRequest.RequestContext.Authorizer["requesterOid"] = *requesterId
	}

	apiRequest.PathParameters = make(map[string]string)
	apiRequest.QueryStringParameters = make(map[string]string)

	return apiRequest
}

func GetDataFromResponse[TResponse interface{}](apiResponse events.APIGatewayProxyResponse) *TResponse {
	var response TResponse
	json.Unmarshal([]byte(apiResponse.Body), &response)

	return &response
}
