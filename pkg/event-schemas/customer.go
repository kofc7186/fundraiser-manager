package eventschemas

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kofc7186/fundraiser-manager/pkg/square/types/webhooks"
	"github.com/kofc7186/fundraiser-manager/pkg/types"
)

const (
	CustomerReceivedType = "org.kofc7186.fundraiserManager.customerReceived"
	CustomerUpdatedType  = "org.kofc7186.fundraiserManager.customerUpdated"
)

type BaseCustomer struct {
	Customer       *types.Customer `json:"customer"`
	IdempotencyKey string          `json:"idempotencyKey"`
}

type CustomerReceived struct {
	BaseCustomer
	Raw *webhooks.CustomerCreated `json:"raw"`
}

type CustomerUpdated struct {
	BaseCustomer
	Raw *webhooks.CustomerUpdated `json:"raw"`
}

func NewCustomerReceived(squareCustomerCreatedEvent *webhooks.CustomerCreated) (*cloudevents.Event, error) {
	event := newEvent(CustomerReceivedType)
	event.SetSubject(squareCustomerCreatedEvent.Data.Object.Customer.Id)

	p, err := types.CreateInternalCustomerFromSquareCustomer(squareCustomerCreatedEvent.Data.Object.Customer)
	if err != nil {
		return nil, err
	}

	pr := &CustomerReceived{
		BaseCustomer: BaseCustomer{
			Customer:       p,
			IdempotencyKey: squareCustomerCreatedEvent.EventID,
		},
		Raw: squareCustomerCreatedEvent,
	}

	if err = event.SetData(applicationJSON, pr); err != nil {
		return nil, err
	}
	return event, nil
}

func NewCustomerUpdated(squareCustomerUpdatedEvent *webhooks.CustomerUpdated) (*cloudevents.Event, error) {
	event := newEvent(CustomerUpdatedType)
	event.SetSubject(squareCustomerUpdatedEvent.Data.Object.Customer.Id)

	p, err := types.CreateInternalCustomerFromSquareCustomer(squareCustomerUpdatedEvent.Data.Object.Customer)
	if err != nil {
		return nil, err
	}

	pu := &CustomerUpdated{
		BaseCustomer: BaseCustomer{
			Customer:       p,
			IdempotencyKey: squareCustomerUpdatedEvent.EventID,
		},
		Raw: squareCustomerUpdatedEvent,
	}

	if err = event.SetData(applicationJSON, pu); err != nil {
		return nil, err
	}
	return event, nil
}
