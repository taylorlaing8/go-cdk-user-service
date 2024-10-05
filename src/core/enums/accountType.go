package enums

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator"
)

type AccountType int

const (
	PersonalAccount AccountType = iota
	BusinessAccount
)

func (priority AccountType) String() string {
	return [...]string{
		"Personal",
		"Business",
	}[priority]
}

func (priority AccountType) MarshalJSON() ([]byte, error) {
	return json.Marshal(priority.String())
}

func (accountType *AccountType) UnmarshalJSON(data []byte) error {
	var accountTypeString string = ""

	err := json.Unmarshal(data, &accountTypeString)
	accountType, err = GetAccountType(&accountTypeString)

	return err
}

func GetAccountType(accountType *string) (*AccountType, error) {
	var matchedAccountType AccountType

	if accountType == nil || len(*accountType) == 0 {
		matchedAccountType = PersonalAccount
	} else {
		switch *accountType {
		case PersonalAccount.String():
			matchedAccountType = PersonalAccount
		case BusinessAccount.String():
			matchedAccountType = BusinessAccount
		default:
			return nil, fmt.Errorf("no matching account type found for: %v", accountType)
		}
	}

	return &matchedAccountType, nil
}

func GetAccountTypeValidator() (string, func(fl validator.FieldLevel) bool) {
	return "is_account_type", func(fl validator.FieldLevel) bool {
		field := fl.Field().String()
		_, err := GetAccountType(&field)
		return err == nil
	}
}
