package integrationtest

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joho/godotenv"

	cfc "cf-user/core"
)

type TestUtils struct {
	DynamoDbStore *cfc.DynamoDbStore
}

var Fixture TestUtils

func init() {
	Fixture = loadConfig()
}

func loadConfig() TestUtils {
	utils := TestUtils{}

	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	service := os.Getenv("SERVICE")
	stage := os.Getenv("STAGE")
	region := os.Getenv("CDK_DEFAULT_REGION")

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

	ddbStore := dynamodb.NewFromConfig(cfg)
	utils.DynamoDbStore = cfc.CreateDynamoDbStore(ddbStore)

	return utils
}
