package integrationtest

import (
	"fmt"
	"testing"
	"time"

	cfe "cf-user/core/enums"
	cfuu "cf-user/update-user"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func init() {
	cfuu.InitLambda(Fixture.DynamoDbStore)
}

func Test_Update_User_Should_Fail_When_Role_Missing(t *testing.T) {
	role := "cf:fake:role"
	userId := ulid.Make().String()
	email := "user" + userId + "@canary-classifind.com"

	request := GivenUpdateUserRequest(&email)

	apiResponse, err := WhenWeUpdateUser(userId, request, &role, nil)

	require.Nil(t, err)
	require.Equal(t, 403, apiResponse.StatusCode)
	require.Contains(t, apiResponse.Body, "You do not have the appropriate permissions to perform this action. Please check the appropriate documentation to ensure you have the correct permissions.")
}

func Test_Update_User_Should_Succeed(t *testing.T) {
	startTime := time.Now()

	username := "user" + ulid.Make().String()
	email := username + "@canary-classifind.com"
	role := cfe.UpdateUser.String()

	Fixture.DynamoDbStore.WipeTestData()

	args := GivenCreateUserArgs(&email)
	entityId, err := Fixture.DynamoDbStore.CreateUser(args)
	require.Nil(t, err)

	request := GivenUpdateUserRequest(nil)

	apiResponse, err := WhenWeUpdateUser(*entityId, request, &role, nil)
	require.Nil(t, err)
	require.Equal(t, 204, apiResponse.StatusCode)

	user, err := Fixture.DynamoDbStore.GetUser(*entityId)
	require.Nil(t, err)

	accountType := user.AccountType.String()

	require.Equal(t, *user.PK, fmt.Sprintf("USER#%v", *entityId))
	require.Equal(t, *user.SK, fmt.Sprintf("USER#%v", *entityId))

	require.Equal(t, *entityId, user.UserId)
	require.Equal(t, username, user.Username)
	require.Equal(t, request.FirstName, user.FirstName)
	require.Equal(t, request.LastName, user.LastName)
	require.Equal(t, request.PhoneNumber, user.PhoneNumber)
	require.Equal(t, email, user.EmailAddress)
	require.Equal(t, request.PrimaryAddress, user.PrimaryAddress)
	require.Equal(t, request.BillingAddress, user.BillingAddress)
	require.Equal(t, request.ProfileImageId, user.ProfileImageId)
	require.Equal(t, request.Biography, user.Biography)
	require.Equal(t, args.AccountType.String(), accountType)
	require.Equal(t, fmt.Sprintf("USERNAME#%v", username), *user.GSI1PK)
	require.Equal(t, fmt.Sprintf("USER#%v", *entityId), *user.GSI1SK)
	require.Equal(t, fmt.Sprintf("EMAIL_ADDRESS#%v", email), *user.GSI2PK)
	require.Equal(t, fmt.Sprintf("USER#%v", *entityId), *user.GSI2SK)
	require.WithinRange(t, user.CreatedDate, startTime, time.Now())
	require.True(t, user.UpdatedDate.After(user.CreatedDate))
}
