package endtoendtest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Update_Identity_Group_Should_Succeed(t *testing.T) {
	Fixture.DynamoDbStore.WipeTestData()

	args := GivenCreateUserArgs(nil)
	entityId, err := Fixture.DynamoDbStore.CreateUser(args)
	require.Nil(t, err)

	request := GivenUpdateUserRequest(nil)

	apiResponse, err := WhenWeUpdateUser(*entityId, *request)

	require.Nil(t, err)
	require.True(t, *apiResponse)
}
