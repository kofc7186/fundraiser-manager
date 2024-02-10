package eventschemas

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kofc7186/fundraiser-manager/pkg/square/types/webhooks"
	"github.com/kofc7186/fundraiser-manager/pkg/types"
)

const (
	PaymentReceivedType = "org.kofc7186.fundraiserManager.paymentReceived"
	PaymentUpdatedType  = "org.kofc7186.fundraiserManager.paymentUpdated"
)

type BasePayment struct {
	Payment        *types.Payment `json:"payment"`
	IdempotencyKey string         `json:"idempotencyKey"`
}

type PaymentReceived struct {
	BasePayment
	Raw *webhooks.PaymentCreated `json:"raw"`
}

type PaymentUpdated struct {
	BasePayment
	Raw *webhooks.PaymentUpdated `json:"raw"`
}

func NewPaymentReceived(squarePaymentCreatedEvent *webhooks.PaymentCreated) (*cloudevents.Event, error) {
	event := newEvent(PaymentReceivedType)
	event.SetSubject(squarePaymentCreatedEvent.Data.Object.Payment.Id)

	p, err := types.CreateInternalPaymentFromSquarePayment(squarePaymentCreatedEvent.Data.Object.Payment)
	if err != nil {
		return nil, err
	}

	pr := &PaymentReceived{
		BasePayment: BasePayment{
			Payment:        p,
			IdempotencyKey: squarePaymentCreatedEvent.EventID,
		},
		Raw: squarePaymentCreatedEvent,
	}

	if err = event.SetData(applicationJSON, pr); err != nil {
		return nil, err
	}
	return event, nil
}

func NewPaymentUpdated(squarePaymentUpdatedEvent *webhooks.PaymentUpdated) (*cloudevents.Event, error) {
	event := newEvent(PaymentUpdatedType)
	event.SetSubject(squarePaymentUpdatedEvent.Data.Object.Payment.Id)

	p, err := types.CreateInternalPaymentFromSquarePayment(squarePaymentUpdatedEvent.Data.Object.Payment)
	if err != nil {
		return nil, err
	}

	pu := &PaymentUpdated{
		BasePayment: BasePayment{
			Payment:        p,
			IdempotencyKey: squarePaymentUpdatedEvent.EventID,
		},
		Raw: squarePaymentUpdatedEvent,
	}

	if err = event.SetData(applicationJSON, pu); err != nil {
		return nil, err
	}
	return event, nil
}
