package integrationtest

import (
	"fmt"
	"strings"
	"testing"
	"time"

	cfe "cf-user/core/enums"
	cfcu "cf-user/create-user"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func init() {
	cfcu.InitLambda(Fixture.DynamoDbStore)
}

func Test_Create_User_Should_Fail_When_Role_Missing(t *testing.T) {
	role := "cf:fake:role"
	email := "user" + ulid.Make().String() + "@canary-classifind.com"

	request := GivenCreateUserRequest(&email)

	apiResponse, err := WhenWeCreateUser(request, &role, nil)

	require.Nil(t, err)
	require.Equal(t, 403, apiResponse.StatusCode)
	require.Contains(t, apiResponse.Body, "You do not have the appropriate permissions to perform this action. Please check the appropriate documentation to ensure you have the correct permissions.")
}

func Test_Create_User_Should_Succeed(t *testing.T) {
	startTime := time.Now()
	role := cfe.CreateUser.String()

	Fixture.DynamoDbStore.WipeTestData()

	request := GivenCreateUserRequest(nil)

	apiResponse, err := WhenWeCreateUser(request, &role, nil)
	require.Nil(t, err)
	require.Equal(t, 200, apiResponse.StatusCode)

	var response = GetDataFromResponse[cfcu.CreateUserResponse](apiResponse)

	user, err := Fixture.DynamoDbStore.GetUser(response.UserId)
	require.Nil(t, err)

	accountType := user.AccountType.String()

	require.Equal(t, *user.PK, fmt.Sprintf("USER#%v", response.UserId))
	require.Equal(t, *user.SK, fmt.Sprintf("USER#%v", response.UserId))
	require.Equal(t, response.UserId, user.UserId)
	require.Equal(t, *request.Username, user.Username)
	require.Equal(t, request.FirstName, user.FirstName)
	require.Equal(t, request.LastName, user.LastName)
	require.Equal(t, request.PhoneNumber, user.PhoneNumber)
	require.Equal(t, request.EmailAddress, user.EmailAddress)
	require.Equal(t, request.PrimaryAddress, user.PrimaryAddress)
	require.Equal(t, request.BillingAddress, user.BillingAddress)
	require.Equal(t, request.ProfileImageId, user.ProfileImageId)
	require.Equal(t, request.Biography, user.Biography)
	require.Equal(t, *request.AccountType, accountType)
	require.Equal(t, fmt.Sprintf("USERNAME#%v", *request.Username), *user.GSI1PK)
	require.Equal(t, fmt.Sprintf("USER#%v", response.UserId), *user.GSI1SK)
	require.Equal(t, fmt.Sprintf("EMAIL_ADDRESS#%v", request.EmailAddress), *user.GSI2PK)
	require.Equal(t, fmt.Sprintf("USER#%v", response.UserId), *user.GSI2SK)
	require.WithinRange(t, user.CreatedDate, startTime, time.Now())
	require.Equal(t, user.CreatedDate, user.UpdatedDate)
}

func Test_Create_User_Should_Succeed_With_Minimal_Request(t *testing.T) {
	startTime := time.Now()
	role := cfe.CreateUser.String()

	Fixture.DynamoDbStore.WipeTestData()

	request := GivenCreateMinimalUserRequest(nil)

	apiResponse, err := WhenWeCreateUser(request, &role, nil)
	require.Nil(t, err)
	require.Equal(t, 200, apiResponse.StatusCode)

	var response = GetDataFromResponse[cfcu.CreateUserResponse](apiResponse)

	user, err := Fixture.DynamoDbStore.GetUser(response.UserId)
	require.Nil(t, err)

	accountType := user.AccountType.String()

	require.Equal(t, fmt.Sprintf("USER#%v", response.UserId), *user.PK)
	require.Equal(t, fmt.Sprintf("USER#%v", response.UserId), *user.SK)
	require.Equal(t, response.UserId, user.UserId)
	require.Equal(t, strings.Split(request.EmailAddress, "@")[0], user.Username)
	require.Nil(t, user.FirstName)
	require.Nil(t, user.LastName)
	require.Nil(t, user.PhoneNumber)
	require.Equal(t, request.EmailAddress, user.EmailAddress)
	require.Nil(t, user.PrimaryAddress)
	require.Nil(t, user.BillingAddress)
	require.Nil(t, user.ProfileImageId)
	require.Nil(t, user.Biography)
	require.Equal(t, cfe.PersonalAccount.String(), accountType)
	require.Equal(t, fmt.Sprintf("USERNAME#%v", strings.Split(request.EmailAddress, "@")[0]), *user.GSI1PK)
	require.Equal(t, fmt.Sprintf("USER#%v", response.UserId), *user.GSI1SK)
	require.Equal(t, fmt.Sprintf("EMAIL_ADDRESS#%v", request.EmailAddress), *user.GSI2PK)
	require.Equal(t, fmt.Sprintf("USER#%v", response.UserId), *user.GSI2SK)
	require.WithinRange(t, user.CreatedDate, startTime, time.Now())
	require.Equal(t, user.CreatedDate, user.UpdatedDate)
}
