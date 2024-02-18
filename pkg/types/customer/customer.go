package customer

import (
	"time"

	"github.com/kofc7186/fundraiser-manager/pkg/square/types/models"
)

type Customer struct {
	EmailAddress      string          `json:"emailAddress" firestore:"emailAddress"`
	Expiration        time.Time       `json:"expiration" firestore:"expiration"`
	ID                string          `json:"id" firestore:"id"`
	IdempotencyKeys   map[string]bool `json:"idempotencyKeys" firestore:"idempotencyKeys"`
	FirstName         string          `json:"firstName" firestore:"firstName"`
	LastName          string          `json:"lastName" firestore:"lastName"`
	PhoneNumber       string          `json:"phoneNumber" firestore:"phoneNumber"`
	KnightOfColumbus  bool            `json:"isKnight" firestore:"isKnight"`
	SquareUpdatedTime time.Time       `json:"squareUpdatedTime" firestore:"squareUpdatedTime"`
	Version           int64           `json:"version" firestore:"version"`
}

func CreateInternalCustomerFromSquareCustomer(squareCustomer models.Customer) (*Customer, error) {
	r := &Customer{
		EmailAddress: squareCustomer.EmailAddress,
		ID:           squareCustomer.Id,
		FirstName:    squareCustomer.GivenName,
		LastName:     squareCustomer.FamilyName,
		PhoneNumber:  squareCustomer.PhoneNumber,
		Version:      squareCustomer.Version,
	}

	var err error
	// per https://developer.squareup.com/reference/square/payments-api/webhooks/customer.created
	// this will be an RFC3339 timestamp
	r.SquareUpdatedTime, err = time.Parse(time.RFC3339, squareCustomer.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return r, nil
}
