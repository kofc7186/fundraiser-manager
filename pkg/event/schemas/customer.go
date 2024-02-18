package schemas

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/kofc7186/fundraiser-manager/pkg/square/types/webhooks"
	"github.com/kofc7186/fundraiser-manager/pkg/types/customer"
)

const (
	CustomerCreatedFromSquareType = "org.kofc7186.fundraiserManager.square.customer.created"
	CustomerUpdatedFromSquareType = "org.kofc7186.fundraiserManager.square.customer.updated"
	CustomerCreatedType           = "org.kofc7186.fundraiserManager.customer.created"
	CustomerUpdatedType           = "org.kofc7186.fundraiserManager.customer.updated"
	CustomerDeletedType           = "org.kofc7186.fundraiserManager.customer.deleted"
)

type BaseCustomer struct {
	Customer       *customer.Customer `json:"customer"`
	IdempotencyKey string             `json:"idempotencyKey"`
}

type CustomerCreatedFromSquare struct {
	BaseCustomer
	Raw *webhooks.CustomerCreated `json:"raw"`
}

type CustomerUpdatedFromSquare struct {
	BaseCustomer
	Raw *webhooks.CustomerUpdated `json:"raw"`
}

func NewCustomerCreatedFromSquare(squareCustomerCreatedEvent *webhooks.CustomerCreated) (*cloudevents.Event, error) {
	event := newEvent(CustomerCreatedFromSquareType)
	event.SetSubject(squareCustomerCreatedEvent.Data.Object.Customer.Id)

	c, err := customer.CreateInternalCustomerFromSquareCustomer(squareCustomerCreatedEvent.Data.Object.Customer)
	if err != nil {
		return nil, err
	}

	cc := &CustomerCreatedFromSquare{
		BaseCustomer: BaseCustomer{
			Customer:       c,
			IdempotencyKey: squareCustomerCreatedEvent.EventID,
		},
		Raw: squareCustomerCreatedEvent,
	}

	if err = event.SetData(applicationJSON, cc); err != nil {
		return nil, err
	}
	return event, nil
}

func NewCustomerUpdatedFromSquare(squareCustomerUpdatedEvent *webhooks.CustomerUpdated) (*cloudevents.Event, error) {
	event := newEvent(CustomerUpdatedFromSquareType)
	event.SetSubject(squareCustomerUpdatedEvent.Data.Object.Customer.Id)

	c, err := customer.CreateInternalCustomerFromSquareCustomer(squareCustomerUpdatedEvent.Data.Object.Customer)
	if err != nil {
		return nil, err
	}

	cu := &CustomerUpdatedFromSquare{
		BaseCustomer: BaseCustomer{
			Customer:       c,
			IdempotencyKey: squareCustomerUpdatedEvent.EventID,
		},
		Raw: squareCustomerUpdatedEvent,
	}

	if err = event.SetData(applicationJSON, cu); err != nil {
		return nil, err
	}
	return event, nil
}

type CustomerCreated struct {
	BaseCustomer
}

func NewCustomerCreated(customer *customer.Customer) (*cloudevents.Event, error) {
	event := newEvent(CustomerCreatedType)
	event.SetSubject(customer.ID)

	cc := &CustomerCreated{
		BaseCustomer: BaseCustomer{
			Customer:       customer,
			IdempotencyKey: uuid.Must(uuid.NewV7()).String(),
		},
	}

	if err := event.SetData(applicationJSON, cc); err != nil {
		return nil, err
	}
	return event, nil
}

type CustomerUpdated struct {
	BaseCustomer
	OldCustomer   *customer.Customer `json:"oldCustomer"`
	UpdatedFields []string           `json:"updatedFields"`
}

func NewCustomerUpdated(oldCustomer, newCustomer *customer.Customer, fieldMask []string) (*cloudevents.Event, error) {
	event := newEvent(CustomerUpdatedType)
	event.SetSubject(newCustomer.ID)

	cu := &CustomerUpdated{
		BaseCustomer: BaseCustomer{
			Customer:       newCustomer,
			IdempotencyKey: uuid.Must(uuid.NewV7()).String(),
		},
		OldCustomer:   oldCustomer,
		UpdatedFields: fieldMask,
	}

	if err := event.SetData(applicationJSON, cu); err != nil {
		return nil, err
	}
	return event, nil
}

type CustomerDeleted struct {
	BaseCustomer
}

func NewCustomerDeleted(customer *customer.Customer) (*cloudevents.Event, error) {
	event := newEvent(CustomerDeletedType)
	event.SetSubject(customer.ID)

	cd := &CustomerDeleted{
		BaseCustomer: BaseCustomer{
			Customer:       customer,
			IdempotencyKey: uuid.Must(uuid.NewV7()).String(),
		},
	}

	if err := event.SetData(applicationJSON, cd); err != nil {
		return nil, err
	}
	return event, nil
}
