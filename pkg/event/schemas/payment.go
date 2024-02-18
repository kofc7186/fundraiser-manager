package schemas

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/kofc7186/fundraiser-manager/pkg/square/types/webhooks"
	"github.com/kofc7186/fundraiser-manager/pkg/types/payment"
)

const (
	PaymentCreatedFromSquareType = "org.kofc7186.fundraiserManager.square.payment.created"
	PaymentUpdatedFromSquareType = "org.kofc7186.fundraiserManager.square.payment.updated"
	PaymentCreatedType           = "org.kofc7186.fundraiserManager.payment.created"
	PaymentUpdatedType           = "org.kofc7186.fundraiserManager.payment.updated"
	PaymentDeletedType           = "org.kofc7186.fundraiserManager.payment.deleted"
)

type BasePayment struct {
	Payment        *payment.Payment `json:"payment"`
	IdempotencyKey string           `json:"idempotencyKey"`
}

type PaymentReceivedFromSquare struct {
	BasePayment
	Raw *webhooks.PaymentCreated `json:"raw"`
}

type PaymentUpdatedFromSquare struct {
	BasePayment
	Raw *webhooks.PaymentUpdated `json:"raw"`
}

func NewPaymentCreatedFromSquare(squarePaymentCreatedEvent *webhooks.PaymentCreated) (*cloudevents.Event, error) {
	event := newEvent(PaymentCreatedFromSquareType)
	event.SetSubject(squarePaymentCreatedEvent.Data.Object.Payment.Id)

	p, err := payment.CreateInternalPaymentFromSquarePayment(squarePaymentCreatedEvent.Data.Object.Payment)
	if err != nil {
		return nil, err
	}

	pr := &PaymentReceivedFromSquare{
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

func NewPaymentUpdatedFromSquare(squarePaymentUpdatedEvent *webhooks.PaymentUpdated) (*cloudevents.Event, error) {
	event := newEvent(PaymentUpdatedFromSquareType)
	event.SetSubject(squarePaymentUpdatedEvent.Data.Object.Payment.Id)

	p, err := payment.CreateInternalPaymentFromSquarePayment(squarePaymentUpdatedEvent.Data.Object.Payment)
	if err != nil {
		return nil, err
	}

	pu := &PaymentUpdatedFromSquare{
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

type PaymentCreated struct {
	BasePayment
}

func NewPaymentCreated(payment *payment.Payment) (*cloudevents.Event, error) {
	event := newEvent(PaymentCreatedType)
	event.SetSubject(payment.ID)

	pc := &PaymentCreated{
		BasePayment: BasePayment{
			Payment:        payment,
			IdempotencyKey: uuid.Must(uuid.NewV7()).String(),
		},
	}

	if err := event.SetData(applicationJSON, pc); err != nil {
		return nil, err
	}
	return event, nil
}

type PaymentUpdated struct {
	BasePayment
	OldPayment    *payment.Payment `json:"oldPayment"`
	UpdatedFields []string         `json:"updatedFields"`
}

func NewPaymentUpdated(oldPayment, newPayment *payment.Payment, fieldMask []string) (*cloudevents.Event, error) {
	event := newEvent(PaymentUpdatedType)
	event.SetSubject(newPayment.ID)

	pu := &PaymentUpdated{
		BasePayment: BasePayment{
			Payment:        newPayment,
			IdempotencyKey: uuid.Must(uuid.NewV7()).String(),
		},
		OldPayment:    oldPayment,
		UpdatedFields: fieldMask,
	}

	if err := event.SetData(applicationJSON, pu); err != nil {
		return nil, err
	}
	return event, nil
}

type PaymentDeleted struct {
	BasePayment
}

func NewPaymentDeleted(payment *payment.Payment) (*cloudevents.Event, error) {
	event := newEvent(PaymentDeletedType)
	event.SetSubject(payment.ID)

	pd := &PaymentDeleted{
		BasePayment: BasePayment{
			Payment:        payment,
			IdempotencyKey: uuid.Must(uuid.NewV7()).String(),
		},
	}

	if err := event.SetData(applicationJSON, pd); err != nil {
		return nil, err
	}
	return event, nil
}
