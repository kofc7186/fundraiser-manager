package eventschemas

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kofc7186/fundraiser-manager/pkg/square/types/webhooks"
	"github.com/kofc7186/fundraiser-manager/pkg/types"
)

const (
	RefundReceivedType = "org.kofc7186.fundraiserManager.refundReceived"
	RefundUpdatedType  = "org.kofc7186.fundraiserManager.refundUpdated"
)

type BaseRefund struct {
	Refund         *types.Refund `json:"refund"`
	IdempotencyKey string        `json:"idempotencyKey"`
}

type RefundReceived struct {
	BaseRefund
	Raw *webhooks.RefundCreated `json:"raw"`
}

type RefundUpdated struct {
	BaseRefund
	Raw *webhooks.RefundUpdated `json:"raw"`
}

func NewRefundReceived(squareRefundCreatedEvent *webhooks.RefundCreated) (*cloudevents.Event, error) {
	event := newEvent(RefundReceivedType)
	event.SetSubject(squareRefundCreatedEvent.Data.Object.Refund.Id)

	p, err := types.CreateInternalRefundFromSquareRefund(squareRefundCreatedEvent.Data.Object.Refund)
	if err != nil {
		return nil, err
	}

	pr := &RefundReceived{
		BaseRefund: BaseRefund{
			Refund:         p,
			IdempotencyKey: squareRefundCreatedEvent.EventID,
		},
		Raw: squareRefundCreatedEvent,
	}

	if err = event.SetData(applicationJSON, pr); err != nil {
		return nil, err
	}
	return event, nil
}

func NewRefundUpdated(squareRefundUpdatedEvent *webhooks.RefundUpdated) (*cloudevents.Event, error) {
	event := newEvent(RefundUpdatedType)
	event.SetSubject(squareRefundUpdatedEvent.Data.Object.Refund.Id)

	p, err := types.CreateInternalRefundFromSquareRefund(squareRefundUpdatedEvent.Data.Object.Refund)
	if err != nil {
		return nil, err
	}

	pu := &RefundUpdated{
		BaseRefund: BaseRefund{
			Refund:         p,
			IdempotencyKey: squareRefundUpdatedEvent.EventID,
		},
		Raw: squareRefundUpdatedEvent,
	}

	if err = event.SetData(applicationJSON, pu); err != nil {
		return nil, err
	}
	return event, nil
}
