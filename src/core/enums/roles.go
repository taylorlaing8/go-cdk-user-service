package enums

import (
	"strings"
)

type LambdaRole int

const (
	CreateUser LambdaRole = iota
	UpdateUser
	DeleteUser
	ReadUser
)

func (role LambdaRole) String() string {
	return [...]string{
		"cf:create:user",
		"cf:update:user",
		"cf:delete:user",
		"cf:read:user",
	}[role]
}

func (role LambdaRole) ExistsInAuthContext(authContext map[string]interface{}) bool {
	permissions, ok := authContext["permissions"]
	if !ok {
		return false
	}

	permissionsArray := strings.Split(permissions.(string), ",")
	for _, r := range permissionsArray {
		if r == role.String() {
			return true
		}
	}

	return false
}
