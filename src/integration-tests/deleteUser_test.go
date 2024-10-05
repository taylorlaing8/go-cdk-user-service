package integrationtest

import (
	"testing"

	cfe "cf-user/core/enums"
	cfdu "cf-user/delete-user"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func init() {
	cfdu.InitLambda(Fixture.DynamoDbStore)
}

func Test_Delete_Identity_Group_Should_Fail_When_Role_Missing(t *testing.T) {
	role := "cf:fake:role"
	userId := ulid.Make().String()

	apiResponse, err := WhenWeDeleteUser(userId, &role, nil)

	require.Nil(t, err)
	require.Equal(t, 403, apiResponse.StatusCode)
	require.Contains(t, apiResponse.Body, "You do not have the appropriate permissions to perform this action. Please check the appropriate documentation to ensure you have the correct permissions.")
}

func Test_Delete_Identity_Group_Should_Succeed(t *testing.T) {
	role := cfe.DeleteUser.String()

	Fixture.DynamoDbStore.WipeTestData()

	args := GivenCreateUserArgs(nil)
	entityId, err := Fixture.DynamoDbStore.CreateUser(args)
	require.Nil(t, err)

	apiResponse, err := WhenWeDeleteUser(*entityId, &role, nil)
	require.Nil(t, err)
	require.Equal(t, 204, apiResponse.StatusCode)
}
