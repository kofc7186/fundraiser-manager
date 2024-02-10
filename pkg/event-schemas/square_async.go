package eventschemas

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kofc7186/fundraiser-manager/pkg/square/types/models"
	"github.com/kofc7186/fundraiser-manager/pkg/types"
)

const (
	SquareGetPaymentRequestedType  = "org.kofc7186.fundraiserManager.square.getPaymentRequested"
	SquareGetPaymentCompletedType  = "org.kofc7186.fundraiserManager.square.getPaymentCompleted"
	SquareGetOrderRequestedType    = "org.kofc7186.fundraiserManager.square.getOrderRequested"
	SquareGetOrderCompletedType    = "org.kofc7186.fundraiserManager.square.getOrderCompleted"
	SquareGetCustomerRequestedType = "org.kofc7186.fundraiserManager.square.getCustomerRequested"
	SquareGetCustomerCompletedType = "org.kofc7186.fundraiserManager.square.getCustomerCompleted"
)

func NewSquareGetPaymentRequested(id string) *cloudevents.Event {
	event := newEvent(SquareGetPaymentRequestedType)
	event.SetSubject(id)

	return event
}

type SquareGetPaymentCompleted struct {
	BasePayment
}

func NewSquareGetPaymentCompleted(squarePayment models.Payment) (*cloudevents.Event, error) {
	event := newEvent(SquareGetPaymentCompletedType)
	event.SetSubject(squarePayment.Id)

	payment, err := types.CreateInternalPaymentFromSquarePayment(squarePayment)
	if err != nil {
		return nil, err
	}

	sgpc := &SquareGetPaymentCompleted{
		BasePayment: BasePayment{
			Payment:        payment,
			IdempotencyKey: "",
		},
	}

	_ = event.SetData(applicationJSON, sgpc)
	return event, nil
}

func NewSquareGetOrderRequested(id string) *cloudevents.Event {
	event := newEvent(SquareGetOrderRequestedType)
	event.SetSubject(id)

	return event
}

type SquareGetOrderCompleted struct {
	BaseOrder
}

func NewSquareGetOrderCompleted(squareOrder models.Order) (*cloudevents.Event, error) {
	event := newEvent(SquareGetOrderCompletedType)
	event.SetSubject(squareOrder.Id)

	order, err := types.CreateInternalOrderFromSquareOrder(squareOrder)
	if err != nil {
		return nil, err
	}

	sgoc := &SquareGetOrderCompleted{
		BaseOrder: BaseOrder{
			Order:          order,
			IdempotencyKey: "",
		},
	}

	_ = event.SetData(applicationJSON, sgoc)
	return event, nil
}

func NewSquareGetCustomerRequested(id string) *cloudevents.Event {
	event := newEvent(SquareGetCustomerRequestedType)
	event.SetSubject(id)

	return event
}

type SquareGetCustomerCompleted struct {
	BaseCustomer
}

func NewSquareGetCustomerCompleted(squareCustomer models.Customer) (*cloudevents.Event, error) {
	event := newEvent(SquareGetCustomerCompletedType)
	event.SetSubject(squareCustomer.Id)

	customer, err := types.CreateInternalCustomerFromSquareCustomer(squareCustomer)
	if err != nil {
		return nil, err
	}

	sgcc := &SquareGetCustomerCompleted{
		BaseCustomer: BaseCustomer{
			Customer:       customer,
			IdempotencyKey: "",
		},
	}

	_ = event.SetData(applicationJSON, sgcc)
	return event, nil
}
