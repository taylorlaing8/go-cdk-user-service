package models

type Address struct {
	AddressOne *string `json:"addressOne" validate:"omitempty,max=300"`
	AddressTwo *string `json:"addressTwo" validate:"omitempty,max=300"`
	City       *string `json:"city" validate:"omitempty,max=200"`
	State      *string `json:"state" validate:"omitempty,min=2,max=2"`      // state code (e.g. UT)
	PostalCode *string `json:"postalCode" validate:"omitempty,min=5,max=9"` // 5 digit zip, or zip with +4 code
	Country    *string `json:"country" validate:"omitempty,min=3,max=3"`    // country code (e.g. USA). Only USA supported currently
}
