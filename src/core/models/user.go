package models

import (
	"time"

	cfe "cf-user/core/enums"
)

type User struct {
	PK             *string         `json:"omitempty"`
	SK             *string         `json:"omitempty"`
	UserId         string          `json:"userId"`
	Username       string          `json:"username"`
	FirstName      *string         `json:"firstName"`
	LastName       *string         `json:"lastName"`
	PhoneNumber    *string         `json:"phoneNumber"`
	EmailAddress   string          `json:"emailAddress"`
	PrimaryAddress *Address        `json:"primaryAddress"`
	BillingAddress *Address        `json:"billingAddress"`
	ProfileImageId *string         `json:"profileImageId"`
	Biography      *string         `json:"biography"`
	AccountType    cfe.AccountType `json:"accountType"`
	GSI1PK         *string         `json:"omitempty"`
	GSI1SK         *string         `json:"omitempty"`
	GSI2PK         *string         `json:"omitempty"`
	GSI2SK         *string         `json:"omitempty"`
	CreatedDate    time.Time       `json:"createdDate"`
	UpdatedDate    time.Time       `json:"updatedDate"`
}
