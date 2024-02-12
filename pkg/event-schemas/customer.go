package eventschemas

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/kofc7186/fundraiser-manager/pkg/square/types/webhooks"
	"github.com/kofc7186/fundraiser-manager/pkg/types"
)

const (
	CustomerCreatedFromSquareType = "org.kofc7186.fundraiserManager.square.customer.created"
	CustomerUpdatedFromSquareType = "org.kofc7186.fundraiserManager.square.customer.updated"
	CustomerCreatedType           = "org.kofc7186.fundraiserManager.customer.created"
	CustomerUpdatedType           = "org.kofc7186.fundraiserManager.customer.updated"
	CustomerDeletedType           = "org.kofc7186.fundraiserManager.customer.deleted"
)

type BaseCustomer struct {
	Customer       types.Customer `json:"customer"`
	IdempotencyKey string         `json:"idempotencyKey"`
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

	p, err := types.CreateInternalCustomerFromSquareCustomer(squareCustomerCreatedEvent.Data.Object.Customer)
	if err != nil {
		return nil, err
	}

	pr := &CustomerCreatedFromSquare{
		BaseCustomer: BaseCustomer{
			Customer:       *p,
			IdempotencyKey: squareCustomerCreatedEvent.EventID,
		},
		Raw: squareCustomerCreatedEvent,
	}

	if err = event.SetData(applicationJSON, pr); err != nil {
		return nil, err
	}
	return event, nil
}

func NewCustomerUpdatedFromSquare(squareCustomerUpdatedEvent *webhooks.CustomerUpdated) (*cloudevents.Event, error) {
	event := newEvent(CustomerUpdatedFromSquareType)
	event.SetSubject(squareCustomerUpdatedEvent.Data.Object.Customer.Id)

	p, err := types.CreateInternalCustomerFromSquareCustomer(squareCustomerUpdatedEvent.Data.Object.Customer)
	if err != nil {
		return nil, err
	}

	pu := &CustomerUpdatedFromSquare{
		BaseCustomer: BaseCustomer{
			Customer:       *p,
			IdempotencyKey: squareCustomerUpdatedEvent.EventID,
		},
		Raw: squareCustomerUpdatedEvent,
	}

	if err = event.SetData(applicationJSON, pu); err != nil {
		return nil, err
	}
	return event, nil
}

type CustomerCreated struct {
	BaseCustomer
}

func NewCustomerCreated(customer *types.Customer) (*cloudevents.Event, error) {
	event := newEvent(CustomerCreatedType)
	event.SetSubject(customer.ID)

	pc := &CustomerCreated{
		BaseCustomer: BaseCustomer{
			Customer:       *customer,
			IdempotencyKey: uuid.Must(uuid.NewV7()).String(),
		},
	}

	if err := event.SetData(applicationJSON, pc); err != nil {
		return nil, err
	}
	return event, nil
}

type CustomerUpdated struct {
	BaseCustomer
	OldCustomer   types.Customer `json:"oldCustomer"`
	UpdatedFields []string       `json:"updatedFields"`
}

func NewCustomerUpdated(oldCustomer, newCustomer *types.Customer, fieldMask []string) (*cloudevents.Event, error) {
	event := newEvent(CustomerUpdatedType)
	event.SetSubject(newCustomer.ID)

	pu := &CustomerUpdated{
		BaseCustomer: BaseCustomer{
			Customer:       *newCustomer,
			IdempotencyKey: uuid.Must(uuid.NewV7()).String(),
		},
		OldCustomer:   *oldCustomer,
		UpdatedFields: fieldMask,
	}

	if err := event.SetData(applicationJSON, pu); err != nil {
		return nil, err
	}
	return event, nil
}

type CustomerDeleted struct {
	BaseCustomer
}

func NewCustomerDeleted(customer *types.Customer) (*cloudevents.Event, error) {
	event := newEvent(CustomerDeletedType)
	event.SetSubject(customer.ID)

	pd := &CustomerDeleted{
		BaseCustomer: BaseCustomer{
			Customer:       *customer,
			IdempotencyKey: uuid.Must(uuid.NewV7()).String(),
		},
	}

	if err := event.SetData(applicationJSON, pd); err != nil {
		return nil, err
	}
	return event, nil
}
