package eventschemas

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kofc7186/fundraiser-manager/pkg/square/types/webhooks"
)

const (
	PaymentReceivedType = "org.kofc7186.fundraiserManager.paymentReceived"
	PaymentUpdatedType  = "org.kofc7186.fundraiserManager.paymentUpdated"
)

type PaymentReceived struct {
	PaymentID string                   `json:"paymentID"`
	OrderID   string                   `json:"orderID"`
	Raw       *webhooks.PaymentCreated `json:"raw"`
}

func NewPaymentReceived(squarePaymentCreatedEvent *webhooks.PaymentCreated) cloudevents.Event {
	event := newEvent(PaymentReceivedType)

	pr := &PaymentReceived{
		PaymentID: squarePaymentCreatedEvent.Data.Object.Payment.Id,
		OrderID:   squarePaymentCreatedEvent.Data.Object.Payment.OrderId,
		Raw:       squarePaymentCreatedEvent,
	}
	_ = event.SetData(applicationJSON, pr)
	return event
}

type PaymentUpdated struct {
	PaymentID string                   `json:"paymentID"`
	OrderID   string                   `json:"orderID"`
	Raw       *webhooks.PaymentUpdated `json:"raw"`
}

func NewPaymentUpdated(squarePaymentUpdatedEvent *webhooks.PaymentUpdated) cloudevents.Event {
	event := newEvent(PaymentUpdatedType)

	pr := &PaymentUpdated{
		PaymentID: squarePaymentUpdatedEvent.Data.Object.Payment.Id,
		OrderID:   squarePaymentUpdatedEvent.Data.Object.Payment.OrderId,
		Raw:       squarePaymentUpdatedEvent,
	}
	_ = event.SetData(applicationJSON, pr)
	return event
}
