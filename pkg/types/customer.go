package types

import (
	"time"

	"github.com/kofc7186/fundraiser-manager/pkg/square/types/models"
)

type Customer struct {
	EmailAddress      string          `json:"emailAddress"`
	Expiration        time.Time       `json:"expiration"`
	ID                string          `json:"id"`
	IdempotencyKeys   map[string]bool `json:"idempotencyKeys"`
	FirstName         string          `json:"firstName"`
	LastName          string          `json:"lastName"`
	PhoneNumber       string          `json:"phoneNumber"`
	KnightOfColumbus  bool            `json:"isKnight"`
	SquareUpdatedTime time.Time       `json:"squareUpdatedTime"`
}

func CreateInternalCustomerFromSquareCustomer(squareCustomer models.Customer) (*Customer, error) {
	r := &Customer{
		EmailAddress: squareCustomer.EmailAddress,
		ID:           squareCustomer.Id,
		FirstName:    squareCustomer.GivenName,
		LastName:     squareCustomer.FamilyName,
		PhoneNumber:  squareCustomer.PhoneNumber,
	}

	var err error
	// per https://developer.squareup.com/reference/square/payments-api/webhooks/refund.created
	// this will be an RFC3339 timestamp
	r.SquareUpdatedTime, err = time.Parse(time.RFC3339, squareCustomer.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return r, nil
}
