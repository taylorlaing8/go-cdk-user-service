package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	cfe "cf-user/core/enums"
	cfm "cf-user/core/models"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/oklog/ulid/v2"
)

type DynamoDbStore struct {
	dynamoDb  *dynamodb.Client
	tableName string
}

var gsi1IndexName = "GSI1"
var gsi2IndexName = "GSI2"

func CreateDynamoDbStore(dynamoDb *dynamodb.Client) *DynamoDbStore {
	tableName := os.Getenv("USER_TABLE_NAME")

	return &DynamoDbStore{
		dynamoDb:  dynamoDb,
		tableName: tableName,
	}
}

type DynamoDb interface {
	WipeTestData() error

	GetUser(userId string) (*cfm.User, error)
	GetUserByUsername(username string) (*cfm.User, error)
	GetUserByEmail(emailAddress string) (*cfm.User, error)
	CreateUser(group *cfm.User) (*string, error)
	UpdateUser(userId string, group *cfm.User) (*bool, error)
	DeleteUser(userId string) (*bool, error)
}

func (DynamoDbStore *DynamoDbStore) WipeTestData() error {
	stage := os.Getenv("STAGE")
	if stage == "prod" || stage == "stage" {
		return errors.New("cannot delete data in prod or staging")
	}

	canaryDomain, _ := attributevalue.Marshal("@canary-classifind.com")

	filterExpression := "contains(EmailAddress, :email_address_domain)"

	scanInput := &dynamodb.ScanInput{
		TableName:        &DynamoDbStore.tableName,
		FilterExpression: &filterExpression,
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email_address_domain": canaryDomain,
		},
	}

	page, err := DynamoDbStore.dynamoDb.Scan(context.TODO(), scanInput)
	if err != nil {
		return fmt.Errorf("unable to fetch canary users: %v", err.Error())
	}

	users := []cfm.User{}
	err = attributevalue.UnmarshalListOfMaps(page.Items, &users)
	if err != nil {
		return fmt.Errorf("unable to parse users from page: %v", err)
	}

	for _, user := range users {
		DynamoDbStore.DeleteUser(user.UserId)
	}

	return nil
}

func (DynamoDbStore *DynamoDbStore) GetUser(userId string) (*cfm.User, error) {
	pkAttribute, _ := attributevalue.Marshal(fmt.Sprintf("USER#%v", userId))
	skAttribute, _ := attributevalue.Marshal(fmt.Sprintf("USER#%v", userId))

	queryInput := &dynamodb.GetItemInput{
		TableName: &DynamoDbStore.tableName,
		Key: map[string]types.AttributeValue{
			"PK": pkAttribute,
			"SK": skAttribute,
		},
	}

	response, err := DynamoDbStore.dynamoDb.GetItem(context.TODO(), queryInput)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch user: %v", err.Error())
	}
	if response.Item == nil {
		return nil, cfe.ErrorNotFound()
	}

	var user cfm.User
	err = attributevalue.UnmarshalMap(response.Item, &user)
	if err != nil {
		return nil, fmt.Errorf("unable to parse user from returned item: %v", err)
	}

	return &user, nil
}

func (DynamoDbStore *DynamoDbStore) GetUserByUsername(username string) (*cfm.User, error) {
	gsi1pkAttribute, _ := attributevalue.Marshal(fmt.Sprintf("USERNAME#%v", username))
	gsi1skAttribute, _ := attributevalue.Marshal("USER#")

	keyCondition := "GSI1PK = :gsi1pk and begins_with(GSI1SK, :gsi1sk)"

	queryInput := &dynamodb.QueryInput{
		TableName:              &DynamoDbStore.tableName,
		IndexName:              &gsi1IndexName,
		KeyConditionExpression: &keyCondition,
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi1pk": gsi1pkAttribute,
			":gsi1sk": gsi1skAttribute,
		},
	}

	page, err := DynamoDbStore.dynamoDb.Query(context.TODO(), queryInput)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch user(s) by username: %v", err.Error())
	}

	users := []cfm.User{}
	err = attributevalue.UnmarshalListOfMaps(page.Items, &users)
	if err != nil {
		return nil, fmt.Errorf("unable to parse users from page: %v", err)
	}

	if len(users) < 1 {
		return nil, fmt.Errorf("no user found with given username: %v", username)
	}

	return &users[0], nil
}

func (DynamoDbStore *DynamoDbStore) GetUserByEmail(emailAddress string) (*cfm.User, error) {
	gsi2pkAttribute, _ := attributevalue.Marshal(fmt.Sprintf("EMAIL_ADDRESS#%v", emailAddress))
	gsi2skAttribute, _ := attributevalue.Marshal("USER#")

	keyCondition := "GSI2PK = :gsi2pk and begins_with(GSI2SK, :gsi2sk)"

	queryInput := &dynamodb.QueryInput{
		TableName:              &DynamoDbStore.tableName,
		IndexName:              &gsi2IndexName,
		KeyConditionExpression: &keyCondition,
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi2pk": gsi2pkAttribute,
			":gsi2sk": gsi2skAttribute,
		},
	}

	page, err := DynamoDbStore.dynamoDb.Query(context.TODO(), queryInput)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch user(s) by email address: %v", err.Error())
	}

	users := []cfm.User{}
	err = attributevalue.UnmarshalListOfMaps(page.Items, &users)
	if err != nil {
		return nil, fmt.Errorf("unable to parse users from page: %v", err)
	}

	if len(users) < 1 {
		return nil, fmt.Errorf("no user found with given email address: %v", emailAddress)
	}

	return &users[0], nil
}

func (DynamoDbStore *DynamoDbStore) CreateUser(user *cfm.User) (*string, error) {
	now := time.Now().UTC()
	userId := ulid.Make().String()

	pk := fmt.Sprintf("USER#%v", userId)
	sk := fmt.Sprintf("USER#%v", userId)
	gsi1pk := fmt.Sprintf("USERNAME#%v", user.Username)
	gsi1sk := fmt.Sprintf("USER#%v", userId)
	gsi2pk := fmt.Sprintf("EMAIL_ADDRESS#%v", user.EmailAddress)
	gsi2sk := fmt.Sprintf("USER#%v", userId)

	pkAttribute, _ := attributevalue.Marshal(pk)

	user.PK = &pk
	user.SK = &sk
	user.UserId = userId
	user.GSI1PK = &gsi1pk
	user.GSI1SK = &gsi1sk
	user.GSI2PK = &gsi2pk
	user.GSI2SK = &gsi2sk
	user.CreatedDate = now
	user.UpdatedDate = now

	conditionExpression := "PK <> :pk"

	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return nil, fmt.Errorf("unable to convert User to Attribute Value map: %v", err.Error())
	}

	putInput := &dynamodb.PutItemInput{
		TableName:           &DynamoDbStore.tableName,
		Item:                item,
		ConditionExpression: &conditionExpression,
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": pkAttribute,
		},
	}

	_, err = DynamoDbStore.dynamoDb.PutItem(context.TODO(), putInput)
	if err != nil {
		return nil, fmt.Errorf("unable to create user: %v", err.Error())
	}

	return &user.UserId, nil
}

func (DynamoDbStore *DynamoDbStore) UpdateUser(userId string, group *cfm.User) (*bool, error) {
	now := time.Now().UTC()

	pk := fmt.Sprintf("USER#%v", userId)
	sk := fmt.Sprintf("USER#%v", userId)

	pkAttribute, _ := attributevalue.Marshal(pk)
	skAttribute, _ := attributevalue.Marshal(sk)

	group.PK = &pk
	group.SK = &sk
	group.UpdatedDate = now

	conditionExpression := "attribute_exists(PK) and attribute_exists(SK)"

	updateFields := []string{
		"FirstName",
		"LastName",
		"PhoneNumber",
		"PrimaryAddress",
		"BillingAddress",
		"ProfileImageId",
		"Biography",
		"UpdatedDate",
	}
	updateExpression, attributeNames, attributeValues, err := extractAttributeUpdateValues(group, updateFields...)
	if err != nil {
		return nil, fmt.Errorf("unable to convert User to Attribute Value map: %v", err.Error())
	}

	updateInput := &dynamodb.UpdateItemInput{
		TableName: &DynamoDbStore.tableName,
		Key: map[string]types.AttributeValue{
			"PK": pkAttribute,
			"SK": skAttribute,
		},
		ConditionExpression:       &conditionExpression,
		UpdateExpression:          updateExpression,
		ExpressionAttributeNames:  *attributeNames,
		ExpressionAttributeValues: *attributeValues,
	}

	_, err = DynamoDbStore.dynamoDb.UpdateItem(context.TODO(), updateInput)
	if err != nil {
		return nil, fmt.Errorf("unable to update user: %v", err.Error())
	}

	success := true
	return &success, nil
}

func (DynamoDbStore *DynamoDbStore) DeleteUser(userId string) (*bool, error) {
	pkAttribute, _ := attributevalue.Marshal(fmt.Sprintf("USER#%v", userId))
	skAttribute, _ := attributevalue.Marshal(fmt.Sprintf("USER#%v", userId))

	deleteInput := &dynamodb.DeleteItemInput{
		TableName: &DynamoDbStore.tableName,
		Key: map[string]types.AttributeValue{
			"PK": pkAttribute,
			"SK": skAttribute,
		},
	}

	_, err := DynamoDbStore.dynamoDb.DeleteItem(context.TODO(), deleteInput)
	if err != nil {
		return nil, fmt.Errorf("unable to delete user: %v", err.Error())
	}

	success := true
	return &success, nil
}

// DynamoDb Helper Functions
func extractAttributeUpdateValues[T interface{}](entity T, fieldNames ...string) (expression *string, attributeNames *map[string]string, attributeValues *map[string]types.AttributeValue, err error) {
	updateExpression := make([]string, 0)
	expressionAttributeNames := make(map[string]string)
	expressionAttributeValues := make(map[string]types.AttributeValue)

	entityVal := reflect.ValueOf(entity).Elem()

	for _, field := range fieldNames {
		attributeName := fmt.Sprintf("#%v", field)
		attributeValue := fmt.Sprintf(":%v", field)

		fieldValueAttribute, _ := attributevalue.Marshal(entityVal.FieldByName(field).Interface())

		updateExpression = append(updateExpression, fmt.Sprintf("%v = %v", attributeName, attributeValue))

		expressionAttributeNames[attributeName] = field
		expressionAttributeValues[attributeValue] = fieldValueAttribute
	}

	updateExpressionVal := "SET " + strings.Join(updateExpression, ", ")

	return &updateExpressionVal, &expressionAttributeNames, &expressionAttributeValues, err
}
