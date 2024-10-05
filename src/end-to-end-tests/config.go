package endtoendtest

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	cfc "cf-user/core"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joho/godotenv"
)

type TestUtils struct {
	AwsConfig     *aws.Config
	DynamoDbStore *cfc.DynamoDbStore
	ApiGatewayUrl string
	TokenService  *TokenService
	HTTPClient    *http.Client
}

var Fixture TestUtils

func init() {
	Fixture = loadConfig()
}

var TestUserEntities []string = make([]string, 0)
var TestPermissionGroupEntities []string = make([]string, 0)
var TestAccessRoleEntities []string = make([]string, 0)

func loadConfig() TestUtils {
	utils := TestUtils{}

	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	service := os.Getenv("SERVICE")
	stage := os.Getenv("STAGE")
	region := os.Getenv("CDK_DEFAULT_REGION")

	os.Setenv("AUTHORIZER_CONFIG_PATH", "/authorizer/config")

	os.Setenv("USER_TABLE_NAME", fmt.Sprintf("%v-%v-app-user", service, stage))

	var cfg aws.Config

	cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Panicf("Failed to load config: %v", err.Error())
	}

	cred, err := cfg.Credentials.Retrieve(context.TODO())
	if err != nil || !cred.HasKeys() {
		// Running locally - fetch credentials from profile
		profile := "cf-dev"

		cfg, _ = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithSharedConfigProfile(profile),
		)
		cred, err = cfg.Credentials.Retrieve(context.TODO())
		if err != nil || !cred.HasKeys() {
			log.Panicf("Failed to retrieve credentials for profile %v", profile)
		}
	}

	utils.AwsConfig = &cfg

	ddbStore := dynamodb.NewFromConfig(*utils.AwsConfig)
	utils.DynamoDbStore = cfc.CreateDynamoDbStore(ddbStore)

	agClient := apigateway.NewFromConfig(*utils.AwsConfig)

	apiGatewayUrl, err := getApiGatewayUrl(agClient, service, stage, region)
	if err != nil {
		log.Fatalf(err.Error())
	}

	utils.ApiGatewayUrl = *apiGatewayUrl

	tokenService := loadTokenService(*utils.AwsConfig)
	utils.TokenService = &tokenService

	utils.HTTPClient = &http.Client{}

	return utils
}

func getApiGatewayUrl(agClient *apigateway.Client, service string, stage string, region string) (*string, error) {
	var limit int32 = 500
	getApisResponse, err := agClient.GetRestApis(context.TODO(), &apigateway.GetRestApisInput{
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}

	apiName := fmt.Sprintf("%v-%v-app", service, stage)
	var restApi *types.RestApi
	for _, api := range getApisResponse.Items {
		if strings.Contains(*api.Name, apiName) {
			restApi = &api
		}
	}
	if restApi == nil {
		return nil, fmt.Errorf("could not locate API Gateway %v", apiName)
	}

	apiGatewayUrl := fmt.Sprintf("https://%v.execute-api.%v.amazonaws.com/LIVE", *restApi.Id, region)
	return &apiGatewayUrl, nil
}
