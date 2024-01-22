package eventschemas

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kofc7186/fundraiser-manager/pkg/square/types/models"
)

const (
	SquareGetPaymentRequestedType  = "org.kofc7186.fundraiserManager.square.getPaymentRequested"
	SquareGetPaymentCompletedType  = "org.kofc7186.fundraiserManager.square.getPaymentCompleted"
	SquareGetOrderRequestedType    = "org.kofc7186.fundraiserManager.square.getOrderRequested"
	SquareGetOrderCompletedType    = "org.kofc7186.fundraiserManager.square.getOrderCompleted"
	SquareGetCustomerRequestedType = "org.kofc7186.fundraiserManager.square.getCustomerRequested"
	SquareGetCustomerCompletedType = "org.kofc7186.fundraiserManager.square.getCustomerCompleted"
)

func NewSquareGetPaymentRequested(id string) cloudevents.Event {
	event := newEvent(SquareGetPaymentRequestedType)
	event.SetSubject(id)

	return event
}

type SquareGetPaymentCompleted struct {
	Payment models.Payment `json:"payment"`
}

func NewSquareGetPaymentCompleted(squarePayment models.Payment) cloudevents.Event {
	event := newEvent(SquareGetPaymentCompletedType)
	event.SetSubject(squarePayment.Id)

	sgpc := &SquareGetPaymentCompleted{
		Payment: squarePayment,
	}

	_ = event.SetData(applicationJSON, sgpc)
	return event
}

func NewSquareGetOrderRequested(id string) cloudevents.Event {
	event := newEvent(SquareGetOrderRequestedType)
	event.SetSubject(id)

	return event
}

type SquareGetOrderCompleted struct {
	Order models.Order `json:"order"`
}

func NewSquareGetOrderCompleted(squareOrder models.Order) cloudevents.Event {
	event := newEvent(SquareGetOrderCompletedType)
	event.SetSubject(squareOrder.Id)

	sgoc := &SquareGetOrderCompleted{
		Order: squareOrder,
	}

	_ = event.SetData(applicationJSON, sgoc)
	return event
}

func NewSquareGetCustomerRequested(id string) cloudevents.Event {
	event := newEvent(SquareGetCustomerRequestedType)
	event.SetSubject(id)

	return event
}

type SquareGetCustomerCompleted struct {
	Customer models.Customer `json:"customer"`
}

func NewSquareGetCustomerCompleted(squareCustomer models.Customer) cloudevents.Event {
	event := newEvent(SquareGetCustomerCompletedType)
	event.SetSubject(squareCustomer.Id)

	sgcc := &SquareGetCustomerCompleted{
		Customer: squareCustomer,
	}

	_ = event.SetData(applicationJSON, sgcc)
	return event
}
