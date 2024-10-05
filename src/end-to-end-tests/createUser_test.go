package endtoendtest

import (
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func Test_Create_Identity_Group_Should_Succeed(t *testing.T) {
	email := "user" + ulid.Make().String() + "@canary-classifind.com"

	Fixture.DynamoDbStore.WipeTestData()

	request := GivenCreateUserRequest(&email)

	response, err := WhenWeCreateUser(*request)

	require.Nil(t, err)
	require.NotNil(t, response)
}
