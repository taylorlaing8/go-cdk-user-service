package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	cfe "cf-user/core/enums"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/go-playground/validator"
)

type LambdaConfig[TRequest interface{}, TResponse interface{}] struct {
	DynamoDbStore   *DynamoDbStore
	FunctionHandler *FunctionHandler[TRequest, TResponse]
}

type FunctionHandler[TRequest interface{}, TResponse interface{}] struct {
	service      string
	stage        string
	roleRequired cfe.LambdaRole
	coldstart    bool
	Validate     *validator.Validate
	Logger       *slog.Logger
}

func CreateLambaConfig[TRequest interface{}, TResponse interface{}](roleRequired cfe.LambdaRole, ddbStore *DynamoDbStore) *LambdaConfig[TRequest, TResponse] {
	lambdaConfig := LambdaConfig[TRequest, TResponse]{}

	service := os.Getenv("SERVICE")
	stage := os.Getenv("STAGE")

	if ddbStore != nil {
		lambdaConfig.DynamoDbStore = ddbStore
	} else {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			log.Panicf("Unable to load SDK config, %v", err.Error())
		}

		ddbStore := dynamodb.NewFromConfig(cfg)
		lambdaConfig.DynamoDbStore = CreateDynamoDbStore(ddbStore)
	}

	lambdaConfig.FunctionHandler = &FunctionHandler[TRequest, TResponse]{
		service:      service,
		stage:        stage,
		roleRequired: roleRequired,
		coldstart:    true,
		Validate:     validator.New(),
		Logger:       nil,
	}

	return &lambdaConfig
}

func (handler *FunctionHandler[TRequest, TResponse]) HandleRequest(apiRequest events.APIGatewayProxyRequest, callback func(req TRequest) (*TResponse, *cfe.ResponseError)) (handlerResponse events.APIGatewayProxyResponse) {
	defer func() {
		handler.coldstart = false
	}()

	defer func() {
		if r := recover(); r != nil {
			var respError cfe.ResponseError

			switch rVal := r.(type) {
			case string:
				respError = cfe.ErrorUnhandled(rVal)
			case error:
				respError = cfe.ErrorUnhandled(rVal.Error())
			default:
				respError = cfe.ErrorUnhandled(fmt.Sprint(rVal))
			}

			handlerResponse = respError.ApiResponse()
		}
	}()

	traceId := os.Getenv("_X_AMZN_TRACE_ID")

	logAttr := make([]slog.Attr, 0)
	logAttr = append(logAttr, slog.String("FunctionName", lambdacontext.FunctionName))
	logAttr = append(logAttr, slog.String("FunctionRequestId", apiRequest.RequestContext.RequestID))
	logAttr = append(logAttr, slog.String("FunctionVersion", lambdacontext.FunctionVersion))
	logAttr = append(logAttr, slog.Int("FunctionMemoryLimitInMB", lambdacontext.MemoryLimitInMB))
	logAttr = append(logAttr, slog.Bool("ColdStart", handler.coldstart))
	logAttr = append(logAttr, slog.String("XRayTraceId", traceId))
	logAttr = append(logAttr, slog.String("Service", handler.service))
	logAttr = append(logAttr, slog.String("Stage", handler.stage))

	handler.Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil).WithAttrs(logAttr))
	slog.SetDefault(handler.Logger)

	reqContext := make(map[string]interface{})
	reqContext["AccountId"] = apiRequest.RequestContext.AccountID
	reqContext["Stage"] = apiRequest.RequestContext.Stage
	reqContext["DomainName"] = apiRequest.RequestContext.DomainName
	reqContext["RequestID"] = apiRequest.RequestContext.RequestID
	reqContext["Identity"] = apiRequest.RequestContext.Identity
	reqContext["ResourcePath"] = apiRequest.RequestContext.ResourcePath
	reqContext["Path"] = apiRequest.RequestContext.Path
	reqContext["HTTPMethod"] = apiRequest.RequestContext.HTTPMethod
	reqContext["RequestTime"] = apiRequest.RequestContext.RequestTime
	reqContext["Body"] = apiRequest.Body

	handler.Logger.Info("Properties", "Request", reqContext)

	validRole := handler.roleRequired.ExistsInAuthContext(apiRequest.RequestContext.Authorizer)
	if !validRole {
		return cfe.ErrorAuthorization("You do not have the appropriate permissions to perform this action. Please check the appropriate documentation to ensure you have the correct permissions.").ApiResponse()
	}

	requestMethod := apiRequest.HTTPMethod

	var requestValue TRequest

	if !(requestMethod == "GET" || requestMethod == "DELETE") {
		requestBody := strings.TrimSpace(apiRequest.Body)
		if len(requestBody) <= 0 {
			return cfe.ErrorValidation("Body was null or empty.").ApiResponse()
		}

		err := json.Unmarshal([]byte(requestBody), &requestValue)
		if err != nil {
			return cfe.ErrorValidation("Body contains invalid payload.").ApiResponse()
		}

		err = handler.Validate.Struct(requestValue)
		if err != nil {
			e := cfe.ErrorValidation("Request failed validation.")
			for _, err := range err.(validator.ValidationErrors) {
				e.AddData(fmt.Sprintf("Field: %v (%v %v)", err.Field(), err.Tag(), err.Param()))
			}

			return e.ApiResponse()
		}
	}

	response, respError := callback(requestValue)
	if respError != nil {
		return respError.ApiResponse()
	}

	if responseSuccess, ok := interface{}(*response).(bool); ok {
		if responseSuccess {
			return events.APIGatewayProxyResponse{
				StatusCode: 204,
			}
		} else {
			return cfe.ErrorUnhandled("Request failed.").ApiResponse()
		}
	}

	val, err := json.Marshal(&response)
	if err != nil {
		return cfe.ErrorUnhandled("Unable to serialize response payload.").ApiResponse()
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(val),
	}
}
