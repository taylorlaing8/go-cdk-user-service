package endtoendtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	cfm "cf-user/core/models"

	cfcu "cf-user/create-user"
	cfuu "cf-user/update-user"
)

func WhenWeGetUser(groupId string) (*cfm.User, error) {
	httpRequestUrl := fmt.Sprintf("%v/v1/users/%v", Fixture.ApiGatewayUrl, groupId)

	return sendRequestWithoutBody[cfm.User]("GET", httpRequestUrl)
}

func WhenWeCreateUser(request cfcu.CreateUserRequest) (*cfcu.CreateUserResponse, error) {
	httpRequestUrl := fmt.Sprintf("%v/v1/users", Fixture.ApiGatewayUrl)

	return sendRequestWithBody[cfcu.CreateUserRequest, cfcu.CreateUserResponse]("POST", httpRequestUrl, request)
}

func WhenWeDeleteUser(groupId string) (*bool, error) {
	httpRequestUrl := fmt.Sprintf("%v/v1/users/%v", Fixture.ApiGatewayUrl, groupId)

	return sendRequestWithoutBody[bool]("DELETE", httpRequestUrl)
}

func WhenWeUpdateUser(groupId string, request cfuu.UpdateUserRequest) (*bool, error) {
	httpRequestUrl := fmt.Sprintf("%v/v1/users/%v", Fixture.ApiGatewayUrl, groupId)

	return sendRequestWithBody[cfuu.UpdateUserRequest, bool]("PUT", httpRequestUrl, request)
}

// Utility Functions
func sendRequestWithBody[TRequest interface{}, TResponse interface{}](httpMethod string, url string, request TRequest) (*TResponse, error) {
	httpRequestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	httpRequest := createRequest(httpMethod, url, &httpRequestBody)

	httpResponse, err := handleRequest[TResponse](httpRequest)
	if err != nil {
		return nil, err
	}

	return httpResponse, nil
}

func sendRequestWithoutBody[TResponse interface{}](httpMethod string, url string) (*TResponse, error) {
	httpRequest := createRequest(httpMethod, url, nil)

	httpResponse, err := handleRequest[TResponse](httpRequest)
	if err != nil {
		return nil, err
	}

	return httpResponse, nil
}

func createRequest(httpMethod string, httpUrl string, requestBody *[]byte) *http.Request {
	var httpRequest *http.Request

	if requestBody == nil {
		httpRequest, _ = http.NewRequest(httpMethod, httpUrl, nil)
	} else {
		httpRequest, _ = http.NewRequest(httpMethod, httpUrl, bytes.NewBuffer(*requestBody))
	}

	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %v", GetCanaryAccessToken()))

	return httpRequest
}

func handleRequest[TResponse interface{}](httpRequest *http.Request) (*TResponse, error) {
	httpResponse, err := Fixture.HTTPClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	content, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	switch httpResponse.StatusCode {
	case http.StatusOK:
		var result TResponse
		json.Unmarshal(content, &result)

		return &result, nil
	case http.StatusNoContent:
		var result TResponse

		result = interface{}(true).(TResponse)
		return &result, nil
	default:
		return nil, fmt.Errorf("StatusCode: %v Response:\n%s", httpResponse.StatusCode, content)
	}
}
