package endtoendtest

import (
	"testing"
	"time"

	cfe "cf-user/core/enums"

	"github.com/stretchr/testify/require"
)

func Test_Get_Identity_Group_Should_Succeed(t *testing.T) {
	Fixture.DynamoDbStore.WipeTestData()

	startTime := time.Now()
	args := GivenCreateUserArgs(nil)
	entityId, err := Fixture.DynamoDbStore.CreateUser(args)
	require.Nil(t, err)

	user, err := WhenWeGetUser(*entityId)
	require.Nil(t, err)

	accountType := cfe.PersonalAccount

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
	require.Equal(t, accountType, user.AccountType)
	require.WithinRange(t, user.CreatedDate, startTime, time.Now())
	require.Equal(t, user.CreatedDate, user.UpdatedDate)
}

func Test_Get_User_Should_Succeed_By_Email_Or_Username(t *testing.T) {
	Fixture.DynamoDbStore.WipeTestData()

	args := GivenCreateUserArgs(nil)
	_, err := Fixture.DynamoDbStore.CreateUser(args)
	require.Nil(t, err)

	// Get User by Email
	_, err = WhenWeGetUser(args.EmailAddress)
	require.Nil(t, err)

	// Get User by Username
	_, err = WhenWeGetUser(args.Username)
	require.Nil(t, err)
}
