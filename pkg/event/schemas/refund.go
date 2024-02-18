package schemas

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/kofc7186/fundraiser-manager/pkg/square/types/webhooks"
	"github.com/kofc7186/fundraiser-manager/pkg/types/refund"
)

const (
	RefundCreatedFromSquareType = "org.kofc7186.fundraiserManager.square.refund.created"
	RefundUpdatedFromSquareType = "org.kofc7186.fundraiserManager.square.refund.updated"
	RefundCreatedType           = "org.kofc7186.fundraiserManager.refund.created"
	RefundUpdatedType           = "org.kofc7186.fundraiserManager.refund.updated"
	RefundDeletedType           = "org.kofc7186.fundraiserManager.refund.deleted"
)

type BaseRefund struct {
	Refund         *refund.Refund `json:"refund"`
	IdempotencyKey string         `json:"idempotencyKey"`
}

type RefundCreatedFromSquare struct {
	BaseRefund
	Raw *webhooks.RefundCreated `json:"raw"`
}

type RefundUpdatedFromSquare struct {
	BaseRefund
	Raw *webhooks.RefundUpdated `json:"raw"`
}

func NewRefundCreatedFromSquare(squareRefundCreatedEvent *webhooks.RefundCreated) (*cloudevents.Event, error) {
	event := newEvent(RefundCreatedFromSquareType)
	event.SetSubject(squareRefundCreatedEvent.Data.Object.Refund.Id)

	r, err := refund.CreateInternalRefundFromSquareRefund(squareRefundCreatedEvent.Data.Object.Refund)
	if err != nil {
		return nil, err
	}

	rc := &RefundCreatedFromSquare{
		BaseRefund: BaseRefund{
			Refund:         r,
			IdempotencyKey: squareRefundCreatedEvent.EventID,
		},
		Raw: squareRefundCreatedEvent,
	}

	if err = event.SetData(applicationJSON, rc); err != nil {
		return nil, err
	}
	return event, nil
}

func NewRefundUpdatedFromSquare(squareRefundUpdatedEvent *webhooks.RefundUpdated) (*cloudevents.Event, error) {
	event := newEvent(RefundUpdatedFromSquareType)
	event.SetSubject(squareRefundUpdatedEvent.Data.Object.Refund.Id)

	r, err := refund.CreateInternalRefundFromSquareRefund(squareRefundUpdatedEvent.Data.Object.Refund)
	if err != nil {
		return nil, err
	}

	ru := &RefundUpdatedFromSquare{
		BaseRefund: BaseRefund{
			Refund:         r,
			IdempotencyKey: squareRefundUpdatedEvent.EventID,
		},
		Raw: squareRefundUpdatedEvent,
	}

	if err = event.SetData(applicationJSON, ru); err != nil {
		return nil, err
	}
	return event, nil
}

type RefundCreated struct {
	BaseRefund
}

func NewRefundCreated(refund *refund.Refund) (*cloudevents.Event, error) {
	event := newEvent(RefundCreatedType)
	event.SetSubject(refund.ID)

	rc := &RefundCreated{
		BaseRefund: BaseRefund{
			Refund:         refund,
			IdempotencyKey: uuid.Must(uuid.NewV7()).String(),
		},
	}

	if err := event.SetData(applicationJSON, rc); err != nil {
		return nil, err
	}
	return event, nil
}

type RefundUpdated struct {
	BaseRefund
	OldRefund     *refund.Refund `json:"oldRefund"`
	UpdatedFields []string       `json:"updatedFields"`
}

func NewRefundUpdated(oldRefund, newRefund *refund.Refund, fieldMask []string) (*cloudevents.Event, error) {
	event := newEvent(RefundUpdatedType)
	event.SetSubject(newRefund.ID)

	ru := &RefundUpdated{
		BaseRefund: BaseRefund{
			Refund:         newRefund,
			IdempotencyKey: uuid.Must(uuid.NewV7()).String(),
		},
		OldRefund:     oldRefund,
		UpdatedFields: fieldMask,
	}

	if err := event.SetData(applicationJSON, ru); err != nil {
		return nil, err
	}
	return event, nil
}

type RefundDeleted struct {
	BaseRefund
}

func NewRefundDeleted(refund *refund.Refund) (*cloudevents.Event, error) {
	event := newEvent(RefundDeletedType)
	event.SetSubject(refund.ID)

	rd := &RefundDeleted{
		BaseRefund: BaseRefund{
			Refund:         refund,
			IdempotencyKey: uuid.Must(uuid.NewV7()).String(),
		},
	}

	if err := event.SetData(applicationJSON, rd); err != nil {
		return nil, err
	}
	return event, nil
}
