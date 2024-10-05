package integrationtest

import (
	"testing"
	"time"

	cfe "cf-user/core/enums"
	cfm "cf-user/core/models"
	cfgu "cf-user/get-user"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func init() {
	cfgu.InitLambda(Fixture.DynamoDbStore)
}

func Test_Get_User_Should_Fail_When_Role_Missing(t *testing.T) {
	role := "cf:fake:role"
	userId := ulid.Make().String()

	apiResponse, err := WhenWeGetUser(userId, &role, nil)

	require.Nil(t, err)
	require.Equal(t, 403, apiResponse.StatusCode)
	require.Contains(t, apiResponse.Body, "You do not have the appropriate permissions to perform this action. Please check the appropriate documentation to ensure you have the correct permissions.")
}

func Test_Get_User_Should_Succeed(t *testing.T) {
	startTime := time.Now()
	role := cfe.ReadUser.String()

	Fixture.DynamoDbStore.WipeTestData()

	args := GivenCreateUserArgs(nil)
	entityId, err := Fixture.DynamoDbStore.CreateUser(args)
	require.Nil(t, err)

	apiResponse, err := WhenWeGetUser(*entityId, &role, nil)
	require.Nil(t, err)
	require.Equal(t, 200, apiResponse.StatusCode)

	var user = GetDataFromResponse[cfm.User](apiResponse)

	require.Equal(t, *entityId, user.UserId)
	require.Equal(t, args.Username, user.Username)
	require.Equal(t, args.FirstName, user.FirstName)
	require.Equal(t, args.LastName, user.LastName)
	require.Equal(t, args.PhoneNumber, user.PhoneNumber)
	require.Equal(t, args.EmailAddress, user.EmailAddress)
	require.Equal(t, args.PrimaryAddress, user.PrimaryAddress)
	require.Equal(t, args.BillingAddress, user.BillingAddress)
	require.Equal(t, args.ProfileImageId, user.ProfileImageId)
	require.Equal(t, args.Biography, user.Biography)
	require.Equal(t, args.AccountType, user.AccountType)
	require.WithinRange(t, user.CreatedDate, startTime, time.Now())
	require.Equal(t, user.CreatedDate, user.UpdatedDate)
}

func Test_Get_User_Should_Succeed_By_Email_Or_Username(t *testing.T) {
	role := cfe.ReadUser.String()

	Fixture.DynamoDbStore.WipeTestData()

	args := GivenCreateUserArgs(nil)
	entityId, err := Fixture.DynamoDbStore.CreateUser(args)
	require.Nil(t, err)

	// Get User by Email
	apiResponse, err := WhenWeGetUser(args.EmailAddress, &role, nil)
	require.Nil(t, err)
	require.Equal(t, 200, apiResponse.StatusCode)

	user := GetDataFromResponse[cfm.User](apiResponse)

	require.Equal(t, *entityId, user.UserId)

	// Get User by Username
	apiResponse, err = WhenWeGetUser(args.Username, &role, nil)
	require.Nil(t, err)
	require.Equal(t, 200, apiResponse.StatusCode)

	user = GetDataFromResponse[cfm.User](apiResponse)

	require.Equal(t, *entityId, user.UserId)
}
